package cache

import "fmt"

// 缓存键前缀定义
const (
	// 用户相关
	KeyPrefixUserInfo    = "user:info:"    // user:info:{user_id}
	KeyPrefixUserSession = "user:session:" // user:session:{token}
	KeyPrefixUserAddress = "user:address:" // user:address:{user_id}
	KeyPrefixVerifyCode  = "verify:code:"  // verify:code:{phone/email}:{type}
	KeyPrefixLoginFail   = "login:fail:"   // login:fail:{username}

	// 商品相关
	KeyPrefixProductDetail = "product:detail:" // product:detail:{product_id}
	KeyPrefixSkuInfo       = "sku:info:"       // sku:info:{sku_id}
	KeyPrefixProductList   = "product:list:"   // product:list:{category_id}:{page}:{page_size}:{sort}
	KeyPrefixCategoryTree  = "category:tree"   // category:tree
	KeyPrefixProductHot    = "product:hot:"    // product:hot:{category_id}

	// 库存相关
	KeyPrefixInventoryStock = "inventory:stock:" // inventory:stock:{sku_id}
	KeyPrefixInventoryLock  = "inventory:lock:"  // inventory:lock:{order_id}:{sku_id}
	KeyPrefixInventoryAlert = "inventory:alert:" // inventory:alert:{sku_id}

	// 订单相关
	KeyPrefixOrderDetail = "order:detail:" // order:detail:{order_id}
	KeyPrefixOrderList   = "order:list:"   // order:list:{user_id}:{status}:{page}
	KeyPrefixOrderSeq    = "order:seq:"    // order:seq:{date}

	// 购物车相关
	KeyPrefixCart = "cart:" // cart:{user_id}

	// 支付相关
	KeyPrefixPaymentInfo = "payment:info:" // payment:info:{payment_no}
	KeyPrefixPaymentLock = "payment:lock:" // payment:lock:{order_id}

	// 营销相关
	KeyPrefixCouponInfo   = "coupon:info:"   // coupon:info:{coupon_id}
	KeyPrefixUserCoupon   = "user:coupon:"   // user:coupon:{user_id}:{status}
	KeyPrefixSeckillStock = "seckill:stock:" // seckill:stock:{product_id}
	KeyPrefixSeckillInfo  = "seckill:info:"  // seckill:info:{activity_id}

	// 搜索相关
	KeyPrefixSearchHotwords = "search:hotwords" // search:hotwords
	KeyPrefixSearchHistory  = "search:history:" // search:history:{user_id}

	// 推荐相关
	KeyPrefixRecommendUser    = "recommend:user:"    // recommend:user:{user_id}
	KeyPrefixRecommendSimilar = "recommend:similar:" // recommend:similar:{product_id}

	// 分布式锁
	KeyPrefixLock = "lock:" // lock:{resource}:{id}
)

// BuildKey 构建缓存键
func BuildKey(prefix string, parts ...interface{}) string {
	key := prefix
	for _, part := range parts {
		key += fmt.Sprintf("%v", part)
	}
	return key
}
