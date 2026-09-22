// Package main 是 GoForge 秒杀链路的一致性校验工具。
//
// 它不测"性能"，只回答三个问题：
//  1. 高并发下会不会超卖？
//  2. 同一个用户会不会重复下单？
//  3. Redis 的秒杀库存扣减和 MySQL 的真源库存、订单，能不能对上账？
//
// 用法（在 server/ 目录下执行）：
//
//	go run ./cmd/seckill-check
//	go run ./cmd/seckill-check -users=300 -stock=100
//	go run ./cmd/seckill-check -skuA=1 -stockA=100 -usersA=200
//
// 前置条件：docker compose 起的网关、MySQL、Redis 都已在运行。
// 密码从 deploy/compose/.env 读取，不在命令行和输出中回显。
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/zhian9/GoForge/server/internal/pkg/utils"
)

// ---------- 命令行参数 ----------

var (
	baseURL     = flag.String("base", "http://localhost:8080", "API 网关地址")
	mysqlHost   = flag.String("mysqlHost", "127.0.0.1", "宿主机 MySQL 地址")
	mysqlPort   = flag.Int("mysqlPort", 3307, "宿主机 MySQL 端口（compose 把 3306 映射到 3307）")
	mysqlUser   = flag.String("mysqlUser", "goforge", "MySQL 用户")
	mysqlDB     = flag.String("mysqlDB", "go_forge", "MySQL 库名")
	redisAddr   = flag.String("redisAddr", "127.0.0.1:6379", "宿主机 Redis 地址")
	redisDB     = flag.Int("redisDB", 0, "Redis 库号")
	envFile     = flag.String("env", "", ".env 路径（默认自动查找 deploy/compose/.env）")
	redisPassF  = flag.String("redisPass", "", "Redis 密码（留空则从 .env 读取）")
	mysqlPassF  = flag.String("mysqlPass", "", "MySQL 密码（留空则从 .env 读取）")
	skuA        = flag.Int64("skuA", 1, "场景 A 使用的 SKU")
	stockA      = flag.Int("stockA", 100, "场景 A 的秒杀库存")
	usersA      = flag.Int("usersA", 200, "场景 A 的并发用户数（应大于库存）")
	skuB        = flag.Int64("skuB", 2, "场景 B 使用的 SKU")
	stockB      = flag.Int("stockB", 5, "场景 B 的秒杀库存")
	dupTimes    = flag.Int("dupTimes", 10, "场景 B 同一用户并发提交次数")
	skuC        = flag.Int64("skuC", 3, "场景 C 使用的 SKU")
	stockC      = flag.Int("stockC", 1, "场景 C 的秒杀库存")
	asyncWait   = flag.Duration("asyncWait", 25*time.Second, "等待 Kafka 异步落单的最长时间")
	httpTimeout = flag.Duration("httpTimeout", 20*time.Second, "单个请求超时")
	cleanup     = flag.Bool("cleanup", true, "结束后删除本次创建的秒杀活动（不回滚已消耗的真实库存）")
)

// runSeed 每次运行唯一，用来生成测试用户 ID。
// 必须每轮唯一：秒杀消费端的幂等校验是"该用户是否已为该 SKU 建过订单"，
// 如果复用用户 ID，上一轮成功的用户这一轮会被正确地跳过处理，
// 表现出来就像"丢单"，从而污染结论。
var runSeed int64

// jwtSecret 用于给测试用户签发 token。
// 抢购接口现在要求带 token（身份只从 token 取），所以压测也要以「已登录用户」的身份发起。
var jwtSecret string

func uniqueUserID(offset int64) int64 {
	return runSeed*1000 + offset
}

// ---------- 响应结构 ----------

// 秒杀接口响应（网关返回 camelCase）
type seckillResp struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Success bool   `json:"success"`
		OrderNo string `json:"orderNo"`
		Message string `json:"message"`
	} `json:"data"`
}

// 创建活动响应
// 注意：网关用 protobuf 的 json 编码，int64 会被序列化成字符串（如 "12" 而不是 12），
// 所以这里用 string 接收再转换，否则会报 cannot unmarshal string into int64。
type createActResp struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID    string `json:"id"`
		SkuID string `json:"skuId"`
		Stock int32  `json:"stock"`
	} `json:"data"`
}

// 单个请求的结果
type reqResult struct {
	success  bool // 业务成功：code==0 && data.success
	code     int32
	message  string
	httpCode int
	err      error
	latency  time.Duration
}

// 一条断言
type check struct {
	name   string
	ok     bool
	detail string
}

var allChecks []check

func addCheck(name string, ok bool, detail string) {
	allChecks = append(allChecks, check{name, ok, detail})
	mark := "[PASS]"
	if !ok {
		mark = "[FAIL]"
	}
	fmt.Printf("  %s %s\n", mark, name)
	if detail != "" {
		fmt.Printf("         %s\n", detail)
	}
}

// ---------- 主流程 ----------

func main() {
	flag.Parse()
	ctx := context.Background()

	fmt.Println(strings.Repeat("=", 66))
	fmt.Println("GoForge 秒杀一致性校验（不测性能，只校对账）")
	fmt.Println(strings.Repeat("=", 66))

	// 1. 读取 .env（不回显密码）
	env := loadEnv(*envFile)
	fmt.Printf("[环境] 配置文件 %s（读到 %d 个键）\n", describeEnvSource(*envFile), len(env))
	fmt.Printf("[环境] 读到的键名：%s\n", strings.Join(envKeys(env), ", "))

	// 2. 连接 MySQL
	dbPass := firstNonEmpty(*mysqlPassF, env["DB_PASSWORD"], env["MYSQL_PASSWORD"], "123456")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		*mysqlUser, dbPass, *mysqlHost, *mysqlPort, *mysqlDB)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		fatalf("连接 MySQL 失败：%v\n提示：确认 docker compose 已启动，且 -mysqlPort 指向宿主机映射端口（默认 3307）", err)
	}
	if err := db.Raw("SELECT 1").Scan(new(int)).Error; err != nil {
		fatalf("MySQL 连通性检查失败：%v", err)
	}

	// 3. 连接 Redis
	redisPass := firstNonEmpty(*redisPassF, env["REDIS_PASSWORD"], "508065")
	fmt.Printf("[环境] Redis 密码来源：%s（长度 %d，sha256 前 12 位 %s）\n",
		passSource(*redisPassF, env["REDIS_PASSWORD"]), len(redisPass), shortHash(redisPass))
	rdb := redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: redisPass,
		DB:       *redisDB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		fatalf("连接 Redis 失败：%v\n提示：当前使用的密码长度为 %d。若 .env 读到的键数为 0，说明没找到配置文件，"+
			"可用 -redisPass=xxx 显式指定，或 -env=<绝对路径> 指定 .env", err, len(redisPass))
	}

	// 4. HTTP 客户端：池子开大，避免 200 并发时连接不够导致假失败
	client := &http.Client{
		Timeout: *httpTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        512,
			MaxIdleConnsPerHost: 512,
			MaxConnsPerHost:     512,
			IdleConnTimeout:     60 * time.Second,
		},
	}

	fmt.Printf("[环境] 网关   %s\n", *baseURL)
	fmt.Printf("[环境] MySQL  %s:%d/%s (user=%s)\n", *mysqlHost, *mysqlPort, *mysqlDB, *mysqlUser)
	fmt.Printf("[环境] Redis  %s db=%d\n", *redisAddr, *redisDB)
	fmt.Printf("[环境] 密码来源 %s（值不回显）\n", describeEnvSource(*envFile))
	runSeed = (time.Now().UnixNano() / 1000) % 100_000_000
	fmt.Printf("[环境] 本轮 user_id 种子 %d（每轮唯一，避免被幂等逻辑跳过）\n", runSeed)
	jwtSecret = firstNonEmpty(env["JWT_SECRET"], "goforge-jwt-secret")
	fmt.Printf("[环境] 抢购将携带 JWT（密钥来源 %s，长度 %d）\n",
		map[bool]string{true: ".env 的 JWT_SECRET", false: "内置默认值"}[env["JWT_SECRET"] != ""], len(jwtSecret))
	fmt.Println()

	// 5. 三个场景
	scenarioA(ctx, client, db, rdb)
	scenarioB(ctx, client, db, rdb)
	scenarioC(ctx, client, db, rdb)

	// 6. 汇总
	printSummary()
}

// ---------- 场景 A：并发抢购，校验不超卖、不丢单 ----------

func scenarioA(ctx context.Context, client *http.Client, db *gorm.DB, rdb *redis.Client) {
	fmt.Printf("--- 场景 A：%d 个不同用户并发抢 %d 件库存（SKU %d）---\n", *usersA, *stockA, *skuA)

	baseline := skuStock(db, *skuA)
	orderBaseline := maxOrderID(db)
	fmt.Printf("  [准备] 抢购前 sku.stock = %d，订单表最大 id = %d\n", baseline, orderBaseline)

	n, err := cleanUserKeys(ctx, rdb, *skuA)
	if err != nil {
		fatalf("清理 Redis 防重 key 失败：%v", err)
	}
	fmt.Printf("  [准备] 清理历史防重 key %d 个\n", n)

	actID, err := createActivity(client, *skuA, *stockA, "9.90")
	if err != nil {
		fatalf("创建秒杀活动失败：%v", err)
	}
	fmt.Printf("  [准备] 创建秒杀活动 id=%d，stock=%d\n", actID, *stockA)

	// 活动创建时后端会把库存预热进 Redis，先确认预热成功
	preheated, _ := rdb.Get(ctx, fmt.Sprintf("seckill:stock:%d", *skuA)).Int64()
	addCheck("Redis 秒杀库存已预热", preheated == int64(*stockA),
		fmt.Sprintf("seckill:stock:%d = %d（期望 %d）", *skuA, preheated, *stockA))

	// 并发发压
	userIDs := make([]int64, *usersA)
	for i := range userIDs {
		userIDs[i] = uniqueUserID(int64(i) + 1)
	}
	fmt.Printf("  [执行] 并发发起 %d 个抢购请求 ...\n", *usersA)
	results := fire(client, *skuA, userIDs, 1)

	success, soldOut, dup, other, netErr := classify(results)
	fmt.Printf("  [执行] 成功 %d / 已抢光 %d / 重复 %d / 其它业务失败 %d / 网络错误 %d\n",
		success, soldOut, dup, other, netErr)
	printMessageBreakdown(results)
	printLatency(results)

	// 等待 Kafka 异步落单
	fmt.Printf("  [对账] 等待异步落单（最长 %s）...\n", *asyncWait)
	orderCnt := waitForStableNewOrderCount(db, *skuA, orderBaseline, *asyncWait)
	itemQty := newOrderItemQuantitySum(db, *skuA, orderBaseline)
	afterStock := skuStock(db, *skuA)
	redisStock, _ := rdb.Get(ctx, fmt.Sprintf("seckill:stock:%d", *skuA)).Int64()
	redisDeducted := preheated - redisStock
	dbDeducted := baseline - afterStock

	addCheck("成功响应数 == 秒杀库存", success == *stockA,
		fmt.Sprintf("成功 %d，期望 %d", success, *stockA))
	addCheck("Redis 扣减数 == 成功响应数", redisDeducted == int64(success),
		fmt.Sprintf("Redis 扣了 %d 件，但只有 %d 个请求返回成功 —— 差值就是「扣了库存却没成功返回」的请求数",
			redisDeducted, success))
	addCheck("新增秒杀订单数 == 成功响应数", orderCnt == int64(success),
		fmt.Sprintf("新增订单 %d 笔，成功响应 %d 个", orderCnt, success))
	addCheck("新增订单商品数量合计 == 成功响应数", itemQty == int64(success),
		fmt.Sprintf("SUM(order_item.quantity) = %d，成功响应 %d 个", itemQty, success))
	addCheck("MySQL 真源库存减少量 == 新增订单数", dbDeducted == orderCnt,
		fmt.Sprintf("sku.stock 减少 %d，新增订单 %d 笔 —— 两者不一致说明真源与订单脱节", dbDeducted, orderCnt))
	addCheck("Redis 秒杀库存已归零", redisStock == 0,
		fmt.Sprintf("seckill:stock:%d = %d，期望 0", *skuA, redisStock))
	addCheck("MySQL 真源库存未出现负数", afterStock >= 0,
		fmt.Sprintf("sku.stock = %d", afterStock))
	if *cleanup {
		_ = deleteActivity(client, actID)
		fmt.Printf("  [清理] 已删除压测活动 id=%d（已消耗的 sku.stock 不会回滚）\n", actID)
	}
	fmt.Println()
}

// ---------- 场景 B：同一用户并发重复抢，校验幂等 ----------

func scenarioB(ctx context.Context, client *http.Client, db *gorm.DB, rdb *redis.Client) {
	fmt.Printf("--- 场景 B：同一用户并发提交 %d 次（SKU %d，库存 %d）---\n", *dupTimes, *skuB, *stockB)

	baseline := skuStock(db, *skuB)
	orderBaseline := maxOrderID(db)
	if _, err := cleanUserKeys(ctx, rdb, *skuB); err != nil {
		fatalf("清理 Redis 防重 key 失败：%v", err)
	}
	actID, err := createActivity(client, *skuB, *stockB, "19.90")
	if err != nil {
		fatalf("创建秒杀活动失败：%v", err)
	}
	fmt.Printf("  [准备] 活动 id=%d，抢购前 sku.stock = %d\n", actID, baseline)

	uid := uniqueUserID(500000)
	sameUser := make([]int64, *dupTimes)
	for i := range sameUser {
		sameUser[i] = uid
	}
	fmt.Printf("  [执行] 同一 user_id=%d 并发提交 %d 次 ...\n", uid, *dupTimes)
	results := fire(client, *skuB, sameUser, 1)

	success, soldOut, dup, other, netErr := classify(results)
	fmt.Printf("  [执行] 成功 %d / 已抢光 %d / 重复 %d / 其它业务失败 %d / 网络错误 %d\n",
		success, soldOut, dup, other, netErr)

	fmt.Printf("  [对账] 等待异步落单 ...\n")
	waitForStableNewOrderCount(db, *skuB, orderBaseline, *asyncWait)
	userOrders := newUserOrderCount(db, *skuB, uid, orderBaseline)
	redisStock, _ := rdb.Get(ctx, fmt.Sprintf("seckill:stock:%d", *skuB)).Int64()

	addCheck("同一用户只成功 1 次", success == 1,
		fmt.Sprintf("成功 %d 次，期望 1 次", success))
	addCheck("该用户只产生 1 笔订单", userOrders == 1,
		fmt.Sprintf("该用户该 SKU 的秒杀订单数 = %d，期望 1", userOrders))
	addCheck("Redis 库存只扣了 1 件", redisStock == int64(*stockB-1),
		fmt.Sprintf("seckill:stock:%d = %d，期望 %d", *skuB, redisStock, *stockB-1))
	if *cleanup {
		_ = deleteActivity(client, actID)
		fmt.Printf("  [清理] 已删除压测活动 id=%d\n", actID)
	}
	fmt.Println()
}

// ---------- 场景 C：quantity 边界，校验扣减与落库一致 ----------

func scenarioC(ctx context.Context, client *http.Client, db *gorm.DB, rdb *redis.Client) {
	fmt.Printf("--- 场景 C：quantity=0 边界（SKU %d，库存 %d）---\n", *skuC, *stockC)

	if _, err := cleanUserKeys(ctx, rdb, *skuC); err != nil {
		fatalf("清理 Redis 防重 key 失败：%v", err)
	}
	actID, err := createActivity(client, *skuC, *stockC, "29.90")
	if err != nil {
		fatalf("创建秒杀活动失败：%v", err)
	}
	fmt.Printf("  [准备] 活动 id=%d\n", actID)

	uid := uniqueUserID(900000)
	orderBaseline := maxOrderID(db)
	fmt.Printf("  [执行] user_id=%d 提交 quantity=0 ...\n", uid)
	r := doSeckill(client, *skuC, uid, 0)
	if r.err != nil {
		fatalf("请求失败：%v", r.err)
	}
	fmt.Printf("  [执行] code=%d message=%q\n", r.code, r.message)

	fmt.Printf("  [对账] 等待异步落单 ...\n")
	waitForStableNewOrderCount(db, *skuC, orderBaseline, *asyncWait)
	redisStock, _ := rdb.Get(ctx, fmt.Sprintf("seckill:stock:%d", *skuC)).Int64()
	dropped := int64(*stockC) - redisStock
	orderExists := newUserOrderCount(db, *skuC, uid, orderBaseline)
	qty := lastOrderItemQuantity(db, *skuC, uid, orderBaseline)

	fmt.Printf("  [对账] Redis 扣减 = %d 件；本次新增该用户订单 = %d 笔；订单上的数量 = %d 件\n",
		dropped, orderExists, qty)
	if dropped == qty {
		addCheck("Redis 扣减数量 == 落库订单数量", true,
			fmt.Sprintf("Redis 扣 %d 件，订单记 %d 件，账目一致", dropped, qty))
	} else {
		reason := "订单未落库（丢单）"
		if orderExists > 0 {
			reason = "订单已落库但数量记为 0（字段不一致）"
		}
		addCheck("Redis 扣减数量 == 落库订单数量", false,
			fmt.Sprintf("Redis 扣 %d 件，订单记 %d 件（%s）—— 根因：seckill_service.go 里"+
				"Lua 脚本用的是兜底后的 quantity，发 Kafka 消息却用了原始 req.Quantity",
				dropped, qty, reason))
	}
	if *cleanup {
		_ = deleteActivity(client, actID)
		fmt.Printf("  [清理] 已删除压测活动 id=%d\n", actID)
	}
	fmt.Println()
}

// ---------- 发压与统计 ----------

// fire 用 at-least-once 的方式并发发起请求：先让所有 goroutine 就绪，再统一放行，
// 尽量让请求同时到达，避免"一个一个发"导致并发度不够。
func fire(client *http.Client, sku int64, userIDs []int64, quantity int32) []reqResult {
	results := make([]reqResult, len(userIDs))
	startGate := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(len(userIDs))
	for i := range userIDs {
		go func(idx int) {
			defer wg.Done()
			<-startGate
			results[idx] = doSeckill(client, sku, userIDs[idx], quantity)
		}(i)
	}
	close(startGate)
	wg.Wait()
	return results
}

func doSeckill(client *http.Client, sku, userID int64, quantity int32) reqResult {
	payload, _ := json.Marshal(map[string]any{
		"user_id":  userID,
		"sku_id":   sku,
		"quantity": quantity,
	})

	t0 := time.Now()
	req, err := http.NewRequest(http.MethodPost, *baseURL+"/api/v1/seckill", bytes.NewReader(payload))
	if err != nil {
		return reqResult{err: err, latency: time.Since(t0)}
	}
	req.Header.Set("Content-Type", "application/json")

	// 每个用户用自己的 token：服务端只认 token 里的 user_id
	if jwtSecret != "" {
		if token, err := utils.GenerateToken(uint64(userID),
			fmt.Sprintf("loadtest-%d", userID), jwtSecret, 3600); err == nil {
			req.Header.Set("Authorization", "Bearer "+token)
		} else {
			return reqResult{err: fmt.Errorf("签发测试 token 失败: %w", err), latency: time.Since(t0)}
		}
	}

	resp, err := client.Do(req)
	latency := time.Since(t0)
	if err != nil {
		return reqResult{err: err, latency: latency}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var r seckillResp
	_ = json.Unmarshal(body, &r)

	return reqResult{
		success:  r.Code == 0 && r.Data.Success,
		code:     r.Code,
		message:  firstNonEmpty(r.Data.Message, r.Message),
		httpCode: resp.StatusCode,
		latency:  latency,
	}
}

// classify 按业务语义分类结果
// 注意：秒杀服务里「重复抢购」的顶层 message 是"不可重复抢购"，但 data.message 是
// "您已参加过本次秒杀活动"，所以这里两个关键词都要匹配，否则会被误判成其它失败。
func classify(results []reqResult) (success, soldOut, dup, other, netErr int) {
	for _, r := range results {
		msg := r.message
		switch {
		case r.err != nil:
			netErr++
		case r.success:
			success++
		case containsAny(msg, "抢光", "库存不足"):
			soldOut++
		case containsAny(msg, "不可重复", "重复抢购", "已参加过"):
			dup++
		default:
			other++
		}
	}
	return
}

// printMessageBreakdown 打印失败响应的具体文案分布。
// 这一项很关键：只知道"失败了多少"没用，必须知道"失败在什么原因上"。
func printMessageBreakdown(results []reqResult) {
	counts := map[string]int{}
	for _, r := range results {
		if r.err != nil {
			counts["[网络错误] "+r.err.Error()]++
			continue
		}
		if r.success {
			continue
		}
		key := fmt.Sprintf("code=%d %s", r.code, r.message)
		if r.httpCode != 0 && r.httpCode != 200 {
			key += fmt.Sprintf(" (HTTP %d)", r.httpCode)
		}
		counts[key]++
	}
	if len(counts) == 0 {
		return
	}
	type kv struct {
		k string
		n int
	}
	list := make([]kv, 0, len(counts))
	for k, n := range counts {
		list = append(list, kv{k, n})
	}
	sort.Slice(list, func(i, j int) bool { return list[i].n > list[j].n })
	fmt.Println("  [归因] 失败原因分布：")
	for _, item := range list {
		fmt.Printf("         %4d 次  %s\n", item.n, item.k)
	}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// printLatency 打印本机观测到的延迟分位，并明确它不是容量指标
func printLatency(results []reqResult) {
	var ds []time.Duration
	for _, r := range results {
		if r.err == nil {
			ds = append(ds, r.latency)
		}
	}
	if len(ds) == 0 {
		return
	}
	sort.Slice(ds, func(i, j int) bool { return ds[i] < ds[j] })
	fmt.Printf("  [观测] 单机延迟 P50=%v P95=%v P99=%v max=%v（单机自测，不能当容量指标）\n",
		percentile(ds, 50).Round(time.Millisecond),
		percentile(ds, 95).Round(time.Millisecond),
		percentile(ds, 99).Round(time.Millisecond),
		ds[len(ds)-1].Round(time.Millisecond))
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p/100*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// ---------- 数据库 / Redis 辅助 ----------

func skuStock(db *gorm.DB, sku int64) int64 {
	var stock int64
	if err := db.Raw("SELECT stock FROM sku WHERE id = ?", sku).Scan(&stock).Error; err != nil {
		fatalf("查询 sku.stock 失败：%v", err)
	}
	return stock
}

const orderCountSQL = `SELECT COUNT(*) FROM orders o
	JOIN order_item oi ON oi.order_id = o.id
	WHERE oi.sku_id = ? AND o.order_type = 2`

// maxOrderID 记录基线，后续只统计"本次运行新增"的订单，
// 避免历史数据（或上一次中断的运行）污染断言。
func maxOrderID(db *gorm.DB) int64 {
	var id int64
	_ = db.Raw("SELECT COALESCE(MAX(id), 0) FROM orders").Scan(&id).Error
	return id
}

func newOrderCount(db *gorm.DB, sku, afterOrderID int64) int64 {
	var c int64
	_ = db.Raw(orderCountSQL+` AND o.id > ?`, sku, afterOrderID).Scan(&c).Error
	return c
}

func newOrderItemQuantitySum(db *gorm.DB, sku, afterOrderID int64) int64 {
	var q int64
	_ = db.Raw(`SELECT COALESCE(SUM(oi.quantity), 0) FROM orders o
		JOIN order_item oi ON oi.order_id = o.id
		WHERE oi.sku_id = ? AND o.order_type = 2 AND o.id > ?`, sku, afterOrderID).Scan(&q).Error
	return q
}

func newUserOrderCount(db *gorm.DB, sku, userID, afterOrderID int64) int64 {
	var c int64
	_ = db.Raw(orderCountSQL+` AND o.user_id = ? AND o.id > ?`, sku, userID, afterOrderID).Scan(&c).Error
	return c
}

func lastOrderItemQuantity(db *gorm.DB, sku, userID, afterOrderID int64) int64 {
	var q int64
	_ = db.Raw(`SELECT oi.quantity FROM orders o
		JOIN order_item oi ON oi.order_id = o.id
		WHERE oi.sku_id = ? AND o.user_id = ? AND o.id > ?
		ORDER BY o.id DESC LIMIT 1`, sku, userID, afterOrderID).Scan(&q).Error
	return q
}

// waitForStableOrderCount 轮询订单数，连续 3 次不变即认为异步落单结束。
// 秒杀接口返回的是"处理中"，订单由 Kafka 消费者异步创建，所以对账前必须等待。
// waitForStableNewOrderCount 轮询"本次新增"订单数，连续 3 次不变即认为异步落单结束。
func waitForStableNewOrderCount(db *gorm.DB, sku, afterOrderID int64, maxWait time.Duration) int64 {
	deadline := time.Now().Add(maxWait)
	last, stable := int64(-1), 0
	for time.Now().Before(deadline) {
		c := newOrderCount(db, sku, afterOrderID)
		if c == last {
			stable++
			if stable >= 3 {
				return c
			}
		} else {
			stable, last = 0, c
		}
		time.Sleep(500 * time.Millisecond)
	}
	return last
}

// cleanUserKeys 清理该 SKU 的历史防重 key，否则重复运行场景 B 会一直"不可重复抢购"
func cleanUserKeys(ctx context.Context, rdb *redis.Client, sku int64) (int, error) {
	pattern := fmt.Sprintf("seckill:user:%d:*", sku)
	var cursor uint64
	deleted := 0
	for {
		keys, next, err := rdb.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			return deleted, err
		}
		if len(keys) > 0 {
			if err := rdb.Del(ctx, keys...).Err(); err != nil {
				return deleted, err
			}
			deleted += len(keys)
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return deleted, nil
}

// ---------- 调用创建活动接口 ----------

func createActivity(client *http.Client, sku int64, stock int, price string) (int64, error) {
	now := time.Now().Unix()
	payload, _ := json.Marshal(map[string]any{
		"name":          fmt.Sprintf("压测活动-sku%d-%d", sku, now),
		"sku_id":        sku,
		"seckill_price": price,
		"stock":         stock,
		"start_time":    now - 5,    // 立刻开始
		"end_time":      now + 3600, // 一小时后结束
		"enable_status": 1,
	})

	resp, err := client.Post(*baseURL+"/api/v1/seckill/activities", "application/json", bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	var r createActResp
	if err := json.Unmarshal(body, &r); err != nil {
		return 0, fmt.Errorf("解析响应失败：%v，原始响应：%s", err, string(body))
	}
	if r.Code != 0 {
		return 0, fmt.Errorf("业务失败 code=%d message=%s", r.Code, r.Message)
	}
	id, err := strconv.ParseInt(r.Data.ID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("解析活动 id 失败：%v（原始值 %q）", err, r.Data.ID)
	}
	return id, nil
}

// deleteActivity 尽力删除本次压测创建的活动，避免污染后台列表。
// 注意：删除活动不会回滚已被消耗的 sku.stock —— 那是真实业务数据的变化。
func deleteActivity(client *http.Client, id int64) error {
	req, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/api/v1/seckill/activities/%d", *baseURL, id), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// ---------- .env 读取（只取需要的键，不回显值） ----------

func loadEnv(explicit string) map[string]string {
	candidates := []string{}
	if explicit != "" {
		candidates = append(candidates, explicit)
	}
	candidates = append(candidates,
		filepath.Join("deploy", "compose", ".env"),
		filepath.Join("..", "deploy", "compose", ".env"),
		".env",
		filepath.Join("..", ".env"),
	)

	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		env := map[string]string{}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			k, v, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			env[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
		}
		usedEnvPath = p
		return env
	}
	return map[string]string{}
}

var usedEnvPath string

func describeEnvSource(explicit string) string {
	if usedEnvPath != "" {
		return usedEnvPath
	}
	return "未找到 .env（回退到 compose 默认密码）"
}

// ---------- 汇总 ----------

func printSummary() {
	passed, failed := 0, 0
	for _, c := range allChecks {
		if c.ok {
			passed++
		} else {
			failed++
		}
	}

	fmt.Println(strings.Repeat("=", 66))
	fmt.Printf("汇总：%d 项通过，%d 项失败\n", passed, failed)
	if failed > 0 {
		fmt.Println("失败项：")
		for _, c := range allChecks {
			if !c.ok {
				fmt.Printf("  - %s：%s\n", c.name, c.detail)
			}
		}
		fmt.Println()
		fmt.Println("提示：FAIL 不等于你写错了系统，它说明这条链路存在真实问题——")
		fmt.Println("      重点不是指标全绿，而是把「怎么发现、根因是什么、怎么修」记录清楚。")
	} else {
		fmt.Println("全部通过：秒杀链路在本次并发下未出现超卖、重复下单或账目不一致。")
	}
	fmt.Println(strings.Repeat("=", 66))
}

// ---------- 小工具 ----------

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func envKeys(env map[string]string) []string {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// passSource 只描述密码来源，不输出密码本身
func passSource(flagVal, envVal string) string {
	switch {
	case flagVal != "":
		return "命令行参数"
	case envVal != "":
		return ".env 中的 REDIS_PASSWORD"
	default:
		return "内置默认值（未从 .env 读到）"
	}
}

// shortHash 输出密钥摘要前 12 位，便于比对配置来源，且不泄露明文
func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:12]
}

func fatalf(format string, args ...any) {
	fmt.Printf("\n[ERROR] "+format+"\n", args...)
	os.Exit(1)
}
