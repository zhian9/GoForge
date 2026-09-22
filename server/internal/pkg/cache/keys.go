package cache

import "fmt"

// 缓存键前缀定义
const (
	// 用户相关
	KeyPrefixUserInfo    = "user:info:"    // user:info:{user_id}
	KeyPrefixUserSession = "user:session:" // user:session:{token}
	KeyPrefixUserAddress = "user:address:" // user:address:{user_id}

	// 商品相关
	KeyPrefixProductDetail = "product:detail:" // product:detail:{product_id}
	KeyPrefixSkuInfo       = "sku:info:"       // sku:info:{sku_id}
	KeyPrefixProductList   = "product:list:"   // product:list:{category_id}:{page}:{page_size}:{sort}
	KeyPrefixCategoryTree  = "category:tree"   // category:tree

	// 库存相关
	KeyPrefixInventoryStock = "inventory:stock:" // inventory:stock:{sku_id}

	// 订单相关
	KeyPrefixOrderDetail = "order:detail:" // order:detail:{order_id}
	KeyPrefixOrderList   = "order:list:"   // order:list:{user_id}:{status}:{page}
	KeyPrefixOrderSeq    = "order:seq:"    // order:seq:{date}

	// 支付相关
	KeyPrefixPaymentSeq = "payment:seq:" // payment:seq:{date}
)

// BuildKey 构建缓存键
func BuildKey(prefix string, parts ...interface{}) string {
	key := prefix
	for _, part := range parts {
		key += fmt.Sprintf("%v", part)
	}
	return key
}
