-- ============================================================
-- 演示数据刷新脚本
--
-- 为什么需要它：演示数据里有两类「会过期」的东西 ——
--   1) 秒杀场次的时间窗口（进行中/未开始/已结束是按 start_time / end_time 实时判断的）
--   2) 待支付订单（job 服务每 30 分钟会把超时未支付订单自动取消）
-- 演示前跑一次，保证「限时秒杀」有正在进行的场次、订单列表里有「待支付」状态。
--
-- 用法（在项目根目录执行）：
--   docker cp deploy/scripts/refresh-demo-data.sql goforge-mysql:/tmp/refresh.sql
--   docker exec goforge-mysql sh -c "mysql -uroot -p123456 --default-character-set=utf8mb4 go_forge < /tmp/refresh.sql"
--
-- 同时刷新 Redis 秒杀库存（让「已抢 %」有内容）：
--   docker exec goforge-redis redis-cli -a 123456 SETEX seckill:stock:1 86400 26
--   docker exec goforge-redis redis-cli -a 123456 SETEX seckill:stock:3 86400 41
--   docker exec goforge-redis redis-cli -a 123456 SETEX seckill:stock:4 86400 7
--   docker exec goforge-redis redis-cli -a 123456 SETEX seckill:stock:6 86400 78
-- ============================================================

SET NAMES utf8mb4;

-- 1) 进行中的秒杀场次：窗口滚动到「现在」附近
UPDATE seckill_activity SET start_time = UNIX_TIMESTAMP() - 7200,  end_time = UNIX_TIMESTAMP() + 21600, status = 1 WHERE id = 54;
UPDATE seckill_activity SET start_time = UNIX_TIMESTAMP() - 5400,  end_time = UNIX_TIMESTAMP() + 25200, status = 1 WHERE id = 55;
UPDATE seckill_activity SET start_time = UNIX_TIMESTAMP() - 3600,  end_time = UNIX_TIMESTAMP() + 18000, status = 1 WHERE id = 56;
UPDATE seckill_activity SET start_time = UNIX_TIMESTAMP() - 1800,  end_time = UNIX_TIMESTAMP() + 28800, status = 1 WHERE id = 57;

-- 2) 即将开始的场次
UPDATE seckill_activity SET start_time = UNIX_TIMESTAMP() + 3600,  end_time = UNIX_TIMESTAMP() + 86400,  status = 1 WHERE id = 58;
UPDATE seckill_activity SET start_time = UNIX_TIMESTAMP() + 7200,  end_time = UNIX_TIMESTAMP() + 90000,  status = 1 WHERE id = 59;
UPDATE seckill_activity SET start_time = UNIX_TIMESTAMP() + 10800, end_time = UNIX_TIMESTAMP() + 100000, status = 1 WHERE id = 60;

-- 3) 已结束的场次
UPDATE seckill_activity SET start_time = UNIX_TIMESTAMP() - 172800, end_time = UNIX_TIMESTAMP() - 86400,  status = 1 WHERE id = 61;
UPDATE seckill_activity SET start_time = UNIX_TIMESTAMP() - 216000, end_time = UNIX_TIMESTAMP() - 108000, status = 1 WHERE id = 62;
UPDATE seckill_activity SET start_time = UNIX_TIMESTAMP() - 259200, end_time = UNIX_TIMESTAMP() - 144000, status = 1 WHERE id = 63;

-- 4) 待支付订单：重置为「刚刚下单」，避免被 job 服务按 30 分钟超时取消
UPDATE orders
SET status = 1, cancel_time = NULL, cancel_reason = NULL,
    created_at = NOW() - INTERVAL 2 MINUTE, updated_at = NOW() - INTERVAL 2 MINUTE
WHERE id = 697;

-- 5) 购物车示例商品（若为空则补两件，方便演示结算流程）
--    购物车存在 Redis，这里只保证数据库侧订单/商品一致；加购请走接口或页面操作。

SELECT '秒杀场次' AS item, SUM(status = 1) AS total FROM seckill_activity
UNION ALL SELECT '进行中', COUNT(*) FROM seckill_activity WHERE status = 1 AND start_time <= UNIX_TIMESTAMP() AND end_time >= UNIX_TIMESTAMP()
UNION ALL SELECT '未开始', COUNT(*) FROM seckill_activity WHERE status = 1 AND start_time > UNIX_TIMESTAMP()
UNION ALL SELECT '已结束', COUNT(*) FROM seckill_activity WHERE status = 1 AND end_time < UNIX_TIMESTAMP()
UNION ALL SELECT '待支付订单', COUNT(*) FROM orders WHERE user_id = 4 AND status = 1;
