package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Lua脚本定义

// LuaScriptInventoryDeduct 库存扣减脚本（原子操作，防超卖）
// KEYS[1]: 库存key (inventory:stock:{sku_id})
// ARGV[1]: 扣减数量
// 返回: >=0 成功(扣减后库存), -1 库存不足, -2 库存key不存在
const LuaScriptInventoryDeduct = `
	local stock = redis.call('GET', KEYS[1])
	if not stock then
		return -2
	end
	local deduct = tonumber(ARGV[1])
	if not deduct or deduct <= 0 then
		return -1
	end
	stock = tonumber(stock)
	if stock < deduct then
		return -1
	end
	return redis.call('DECRBY', KEYS[1], deduct)
`

// LuaScriptInventoryRollback 库存回退脚本（原子操作）
// KEYS[1]: 库存key (inventory:stock:{sku_id})
// ARGV[1]: 回退数量
// 返回: 回退后的库存数量
const LuaScriptInventoryRollback = `
	local rollback = tonumber(ARGV[1])
	if not rollback or rollback <= 0 then
		rollback = 1
	end
	return redis.call('INCRBY', KEYS[1], rollback)
`

// LuaScriptCartAdd 购物车添加商品脚本
// KEYS[1]: 购物车key (cart:{user_id})
// ARGV[1]: sku_id
// ARGV[2]: 数量
// 返回: 操作后的数量
const LuaScriptCartAdd = `
	local sku_id = ARGV[1]
	local quantity = tonumber(ARGV[2])
	local current = redis.call('HGET', KEYS[1], sku_id)
	if current then
		local data = cjson.decode(current)
		data.quantity = data.quantity + quantity
		redis.call('HSET', KEYS[1], sku_id, cjson.encode(data))
		return data.quantity
	else
		local data = {quantity = quantity, selected = 1}
		redis.call('HSET', KEYS[1], sku_id, cjson.encode(data))
		return quantity
	end
`

// LuaScriptCartUpdate 购物车更新商品数量脚本
// KEYS[1]: 购物车key (cart:{user_id})
// ARGV[1]: sku_id
// ARGV[2]: 新数量
// 返回: 1成功，0失败
const LuaScriptCartUpdate = `
	local sku_id = ARGV[1]
	local quantity = tonumber(ARGV[2])
	local current = redis.call('HGET', KEYS[1], sku_id)
	if current then
		local data = cjson.decode(current)
		data.quantity = quantity
		redis.call('HSET', KEYS[1], sku_id, cjson.encode(data))
		return 1
	else
		return 0
	end
`

// LuaScriptCartRemove 购物车删除商品脚本
// KEYS[1]: 购物车key (cart:{user_id})
// ARGV[1]: sku_id
// 返回: 1成功，0失败
const LuaScriptCartRemove = `
	local sku_id = ARGV[1]
	return redis.call('HDEL', KEYS[1], sku_id)
`

// LuaScriptLockRelease 安全释放锁脚本
// KEYS[1]: 锁key
// ARGV[1]: 锁的值
// 返回: 1成功释放，0失败（锁值不匹配或不存在）
const LuaScriptLockRelease = `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
`

// LuaScriptSeckill 秒杀脚本（防超卖 + 防重复 + 支持购买数量）
// KEYS[1]: 库存key (seckill:stock:{skuId})
// KEYS[2]: 用户key (seckill:user:{skuId}:{uid})
// ARGV[1]: 购买数量 (默认1)
// 返回: >=0 成功(剩余库存), -1 库存不足, -2 重复抢购
const LuaScriptSeckill = `
	if redis.call("exists", KEYS[2]) == 1 then
		return -2
	end

	local quantity = tonumber(ARGV[1])
	if not quantity or quantity <= 0 then
		quantity = 1
	end

	local stock = tonumber(redis.call("get", KEYS[1]) or "-1")
	if stock < quantity then
		return -1
	end

	local new_stock = redis.call("decrby", KEYS[1], quantity)
	redis.call("set", KEYS[2], 1)
	redis.call("expire", KEYS[2], 86400)

	return new_stock
`

// LuaScriptSeckillRollback 秒杀预扣回滚脚本（原子）
//
// 使用场景：Redis 已经扣了库存、也标记了用户，但后续步骤失败（例如 Kafka 发送失败）。
// 此时如果只是返回错误而不回滚，会造成两个后果：
//  1. 库存被扣掉但没有产生订单 —— 库存凭空蒸发；
//  2. 防重 key 仍然存在 —— 用户在 24 小时内无法重试，既没抢到又被锁死。
//
// 所以「退库存」和「删防重标记」必须在同一个脚本里原子完成，否则会退了一半。
//
// KEYS[1]: 库存key (seckill:stock:{skuId})
// KEYS[2]: 用户key (seckill:user:{skuId}:{uid})
// ARGV[1]: 回滚数量
// 返回: 回滚后的库存
const LuaScriptSeckillRollback = `
	local quantity = tonumber(ARGV[1])
	if not quantity or quantity <= 0 then
		quantity = 1
	end

	local new_stock = redis.call("incrby", KEYS[1], quantity)
	redis.call("del", KEYS[2])

	return new_stock
`

// ExecuteLuaScript 执行Lua脚本
func ExecuteLuaScript(ctx context.Context, client *redis.Client, script string, keys []string, args ...interface{}) (interface{}, error) {
	// 使用 Eval 执行原始 Lua 脚本
	// 参数顺序：ctx, script, keys, args...
	return client.Eval(ctx, script, keys, args...).Result()
}
