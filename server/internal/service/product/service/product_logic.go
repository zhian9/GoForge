package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/zhian9/GoForge/server/internal/pkg/cache"
	apperrors "github.com/zhian9/GoForge/server/internal/pkg/errors"
	"github.com/zhian9/GoForge/server/internal/pkg/mq"
	"github.com/zhian9/GoForge/server/internal/pkg/outbox"
	"github.com/zhian9/GoForge/server/internal/service/product/model"
	"github.com/zhian9/GoForge/server/internal/service/product/repository"

	"gorm.io/gorm"
)

func isDuplicateSkuCodeErr(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	// MySQL: Error 1062 (23000): Duplicate entry 'xxx' for key 'sku.uk_sku_code'
	return strings.Contains(s, "Duplicate entry") && strings.Contains(s, "uk_sku_code")
}

// ProductLogic 商品业务逻辑
type ProductLogic struct {
	db           *gorm.DB
	outboxRepo   *outbox.Repo
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
	skuRepo      repository.SkuRepository
	bannerRepo   repository.BannerRepository
	cache        *cache.CacheOperations
	mqProducer   *mq.Producer
}

// NewProductLogic 创建商品业务逻辑
func NewProductLogic(
	db *gorm.DB,
	outboxRepo *outbox.Repo,
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	skuRepo repository.SkuRepository,
	bannerRepo repository.BannerRepository,
	cache *cache.CacheOperations,
	mqProducer *mq.Producer,
) *ProductLogic {
	return &ProductLogic{
		db:           db,
		outboxRepo:   outboxRepo,
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		skuRepo:      skuRepo,
		bannerRepo:   bannerRepo,
		cache:        cache,
		mqProducer:   mqProducer,
	}
}

// GetProductRequest 获取商品详情请求
type GetProductRequest struct {
	ID      uint64
	SpuCode string
}

// GetProductResponse 获取商品详情响应
type GetProductResponse struct {
	Product *model.Product
	Skus    []*model.Sku
}

// GetProduct 获取商品详情（带缓存）
func (l *ProductLogic) GetProduct(ctx context.Context, req *GetProductRequest) (*GetProductResponse, error) {
	// 检查 repository 是否初始化
	if l.productRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化，请检查数据库配置")
	}

	var productID uint64
	var err error

	// 确定商品ID
	if req.ID > 0 {
		productID = req.ID
	} else if req.SpuCode != "" {
		product, err := l.productRepo.GetBySpuCode(ctx, req.SpuCode)
		if err != nil {
			return nil, apperrors.NewInternalError("查询商品失败: " + err.Error())
		}
		if product == nil {
			return nil, apperrors.NewError(apperrors.CodeNotFound, "商品不存在")
		}
		productID = product.ID
	} else {
		return nil, apperrors.NewInvalidParamError("商品ID或SPU编码不能为空")
	}

	// 尝试从缓存获取
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixProductDetail, productID)
		var cachedResp GetProductResponse
		if err := l.cache.GetJSON(ctx, cacheKey, &cachedResp); err == nil {
			return &cachedResp, nil
		}
	}

	// 从数据库查询
	product, err := l.productRepo.GetByID(ctx, productID)
	if err != nil {
		if err == gorm.ErrRecordNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
		}
		return nil, apperrors.NewInternalError("查询商品失败: " + err.Error())
	}
	if product == nil {
		return nil, apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
	}

	// 获取SKU列表
	skus, err := l.skuRepo.GetByProductID(ctx, product.ID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询SKU列表失败: " + err.Error())
	}

	// 口径统一：详情页返回的 product.price / product.stock 用“上架 SKU 聚合值”兜底
	// - price: 上架 SKU 最低价
	// - stock: 上架 SKU 总库存
	// 注意：这里只改返回值，不落库
	if len(skus) > 0 {
		minPrice := math.MaxFloat64
		totalStock := 0
		hasActive := false
		for _, s := range skus {
			if s == nil || s.Status != 1 {
				continue
			}
			hasActive = true
			if s.Price > 0 && s.Price < minPrice {
				minPrice = s.Price
			}
			if s.Stock > 0 {
				totalStock += s.Stock
			}
		}
		if hasActive {
			if minPrice != math.MaxFloat64 {
				product.Price = minPrice
			}
			product.Stock = totalStock
		}
	}

	resp := &GetProductResponse{
		Product: product,
		Skus:    skus,
	}

	// 写入缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixProductDetail, productID)
		_ = l.cache.Set(ctx, cacheKey, resp, 1*time.Hour)
	}

	return resp, nil
}

// ListProductsRequest 获取商品列表请求
type ListProductsRequest struct {
	CategoryID uint64
	BrandID    uint64
	Keyword    string
	Status     int8
	IsHot      int8 // -1-全部, 0-否, 1-是
	Page       int
	PageSize   int
	Sort       string
}

// ListProductsResponse 获取商品列表响应
type ListProductsResponse struct {
	Products   []*model.Product
	Page       int
	PageSize   int
	Total      int64
	TotalPages int
}

// ListProducts 获取商品列表（带缓存）
func (l *ProductLogic) ListProducts(ctx context.Context, req *ListProductsRequest) (*ListProductsResponse, error) {
	// 检查 repository 是否初始化
	if l.productRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化，请检查数据库配置")
	}

	// 参数验证
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.Sort == "" {
		req.Sort = "default"
	}

	// 兼容：点击“主分类”时，需要过滤出其所有子分类（以及更深层级）下的商品
	// 前端仍然只传 category_id=主分类ID，后端自动展开为 category_id IN (...)
	var expandedCategoryIDs []uint64
	if req.CategoryID > 0 && l.categoryRepo != nil {
		ids, err := l.collectDescendantCategoryIDs(ctx, req.CategoryID)
		if err == nil && len(ids) > 1 {
			expandedCategoryIDs = ids
		}
	}

	// 构建缓存键
	if l.cache != nil {
		// v2：避免老缓存（精确匹配 category_id）导致改完逻辑后短时间仍返回旧结果
		cacheKey := fmt.Sprintf("%sv2:%d:%d:%d:%d:%s:%d:%d:%s",
			cache.KeyPrefixProductList,
			req.CategoryID,
			req.BrandID,
			req.Status,
			req.IsHot,
			req.Keyword,
			req.Page,
			req.PageSize,
			req.Sort,
		)
		var cachedResp ListProductsResponse
		if err := l.cache.GetJSON(ctx, cacheKey, &cachedResp); err == nil {
			return &cachedResp, nil
		}
	}

	// 构建查询请求
	repoReq := &repository.ListProductsRequest{
		CategoryID:  req.CategoryID,
		CategoryIDs: expandedCategoryIDs,
		BrandID:     req.BrandID,
		Keyword:     req.Keyword,
		Status:      req.Status,
		IsHot:       req.IsHot,
		Page:        req.Page,
		PageSize:    req.PageSize,
		Sort:        req.Sort,
	}

	// 查询商品列表
	products, total, err := l.productRepo.List(ctx, repoReq)
	if err != nil {
		return nil, apperrors.NewInternalError("查询商品列表失败: " + err.Error())
	}

	// 口径统一：列表页返回的 product.price / product.stock 用“上架 SKU 聚合值”兜底（避免前端只拿到 SPU 原值）
	// 使用批量聚合，避免 N+1
	if l.skuRepo != nil && len(products) > 0 {
		ids := make([]uint64, 0, len(products))
		for _, p := range products {
			if p != nil && p.ID > 0 {
				ids = append(ids, p.ID)
			}
		}
		if len(ids) > 0 {
			agg, err := l.skuRepo.GetAggByProductIDs(ctx, ids, 1)
			if err == nil && len(agg) > 0 {
				for _, p := range products {
					if p == nil {
						continue
					}
					if a, ok := agg[p.ID]; ok {
						if a.MinPrice > 0 {
							p.Price = a.MinPrice
						}
						// 总库存可能很大，这里做个保底转换
						if a.TotalStock >= 0 {
							if a.TotalStock > int64(^uint(0)>>1) {
								p.Stock = int(^uint(0) >> 1)
							} else {
								p.Stock = int(a.TotalStock)
							}
						}
						// 如果你未来要在列表展示“原价”，可以用 a.MinOriginalPrice 去兜底 p.OriginalPrice
					}
				}
			}
		}
	}

	// 计算总页数
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	resp := &ListProductsResponse{
		Products:   products,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	// 写入缓存（10分钟）
	if l.cache != nil {
		cacheKey := fmt.Sprintf("%sv2:%d:%d:%d:%d:%s:%d:%d:%s",
			cache.KeyPrefixProductList,
			req.CategoryID,
			req.BrandID,
			req.Status,
			req.IsHot,
			req.Keyword,
			req.Page,
			req.PageSize,
			req.Sort,
		)
		_ = l.cache.Set(ctx, cacheKey, resp, 10*time.Minute)
	}

	return resp, nil
}

// collectDescendantCategoryIDs 收集某个分类的所有子孙分类 ID（包含自身）
// - 仅依赖 parent_id 关系，支持任意层级
// - 用 visited 防止异常数据导致死循环
func (l *ProductLogic) collectDescendantCategoryIDs(ctx context.Context, rootID uint64) ([]uint64, error) {
	if rootID == 0 || l.categoryRepo == nil {
		return nil, nil
	}
	visited := map[uint64]bool{}
	queue := []uint64{rootID}
	out := make([]uint64, 0, 16)

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if id == 0 || visited[id] {
			continue
		}
		visited[id] = true
		out = append(out, id)

		children, err := l.categoryRepo.GetByParentID(ctx, id)
		if err != nil {
			return nil, err
		}
		for _, c := range children {
			if c != nil && c.ID > 0 && !visited[c.ID] {
				queue = append(queue, c.ID)
			}
		}
	}
	return out, nil
}

// GetSkuRequest 获取SKU详情请求
type GetSkuRequest struct {
	ID      uint64
	SkuCode string
}

// GetSkuResponse 获取SKU详情响应
type GetSkuResponse struct {
	Sku *model.Sku
}

// GetSku 获取SKU详情（带缓存）
func (l *ProductLogic) GetSku(ctx context.Context, req *GetSkuRequest) (*GetSkuResponse, error) {
	// 检查 repository 是否初始化
	if l.skuRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化，请检查数据库配置")
	}

	var skuID uint64
	var err error

	// 确定SKU ID
	if req.ID > 0 {
		skuID = req.ID
	} else if req.SkuCode != "" {
		sku, err := l.skuRepo.GetBySkuCode(ctx, req.SkuCode)
		if err != nil {
			return nil, apperrors.NewInternalError("查询SKU失败: " + err.Error())
		}
		if sku == nil {
			return nil, apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
		}
		skuID = sku.ID
	} else {
		return nil, apperrors.NewInvalidParamError("SKU ID或SKU编码不能为空")
	}

	// 尝试从缓存获取
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixSkuInfo, skuID)
		var cachedResp GetSkuResponse
		if err := l.cache.GetJSON(ctx, cacheKey, &cachedResp); err == nil {
			return &cachedResp, nil
		}
	}

	// 从数据库查询
	sku, err := l.skuRepo.GetByID(ctx, skuID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询SKU失败: " + err.Error())
	}
	if sku == nil {
		return nil, apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
	}

	resp := &GetSkuResponse{
		Sku: sku,
	}

	// 写入缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixSkuInfo, skuID)
		_ = l.cache.Set(ctx, cacheKey, resp, 1*time.Hour)
	}

	return resp, nil
}

// ListSkusRequest 获取SKU列表请求
type ListSkusRequest struct {
	ProductID uint64
	Status    int8 // -1-全部, 0-下架, 1-上架
	Page      int
	PageSize  int
}

// ListSkusResponse 获取SKU列表响应
type ListSkusResponse struct {
	Skus       []*model.Sku
	Page       int
	PageSize   int
	Total      int64
	TotalPages int
}

// ListSkus 获取SKU列表
func (l *ProductLogic) ListSkus(ctx context.Context, req *ListSkusRequest) (*ListSkusResponse, error) {
	if l.skuRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	// 参数验证和默认值处理
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	listReq := &repository.ListSkusRequest{
		ProductID: req.ProductID,
		Status:    req.Status,
		Page:      page,
		PageSize:  pageSize,
	}

	skus, total, err := l.skuRepo.List(ctx, listReq)
	if err != nil {
		return nil, apperrors.NewInternalError("查询SKU列表失败: " + err.Error())
	}

	// 兜底：用户侧经常按 product_id + status=1 查询 SKU。
	// 如果该商品一个 SKU 都没有，会导致前端无法选择规格、无法加入购物车。
	// 这里自动创建一个“默认 SKU”（仅在该商品完全没有 SKU 时才创建）。
	if req.ProductID > 0 && req.Status == 1 && total == 0 {
		existing, err := l.skuRepo.GetByProductID(ctx, req.ProductID)
		if err == nil && len(existing) == 0 && l.productRepo != nil {
			product, err := l.productRepo.GetByID(ctx, req.ProductID)
			if err == nil && product != nil {
				defaultSku := &model.Sku{
					ProductID: req.ProductID,
					SkuCode:   fmt.Sprintf("SKU%d-DEFAULT", req.ProductID),
					Name:      product.Name,
					Specs:     "{}",
					Price:     product.Price,
					Stock:     product.Stock,
					Status:    1,
				}
				if product.OriginalPrice != nil {
					defaultSku.OriginalPrice = product.OriginalPrice
				}
				// 优先用本地主图
				if product.LocalMainImage != "" {
					defaultSku.Image = product.LocalMainImage
				} else {
					defaultSku.Image = product.MainImage
				}

				_ = l.skuRepo.Create(ctx, defaultSku) // 忽略重复创建错误（并发情况下）
				l.ensureInventory(ctx, defaultSku.ID, defaultSku.Stock)

				// 清理商品详情缓存，避免 SKU 列表和详情不一致
				if l.cache != nil {
					_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, req.ProductID))
				}

				// 重新查询一次（确保拿到 ID / 最新数据）
				skus, total, _ = l.skuRepo.List(ctx, listReq)
			}
		}
	}

	// 调试日志
	fmt.Printf("[ProductLogic.ListSkus] ProductID=%d, Status=%d, Page=%d, PageSize=%d, Total=%d, Found=%d\\n",
		req.ProductID, req.Status, page, pageSize, total, len(skus))

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 && total > 0 {
		totalPages = 1
	}

	return &ListSkusResponse{
		Skus:       skus,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// CreateSkuRequest 创建SKU请求
type CreateSkuRequest struct {
	ProductID     uint64
	SkuCode       string
	Name          string
	Specs         map[string]string // 规格属性
	Price         float64
	OriginalPrice *float64
	Stock         int
	Image         string
	Weight        *float64
	Volume        *float64
	Status        int8
}

// CreateSkuResponse 创建SKU响应
type CreateSkuResponse struct {
	Sku *model.Sku
}

// CreateSku 创建SKU
func (l *ProductLogic) CreateSku(ctx context.Context, req *CreateSkuRequest) (*CreateSkuResponse, error) {
	if l.skuRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}
	// Outbox 要求：SKU 变更需要触发商品索引更新
	if l.db != nil && l.outboxRepo != nil {
		var outSku *model.Sku
		if err := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			productRepoTx := repository.NewProductRepository(tx)
			skuRepoTx := repository.NewSkuRepository(tx)

			// 验证必填字段
			if req.ProductID == 0 {
				return apperrors.NewInvalidParamError("商品ID不能为空")
			}
			if req.SkuCode == "" {
				return apperrors.NewInvalidParamError("SKU编码不能为空")
			}
			if req.Name == "" {
				return apperrors.NewInvalidParamError("SKU名称不能为空")
			}
			if len(req.Specs) == 0 {
				return apperrors.NewInvalidParamError("规格属性不能为空")
			}
			if req.Price <= 0 {
				return apperrors.NewInvalidParamError("价格必须大于0")
			}

			// 检查商品是否存在
			product, err := productRepoTx.GetByID(ctx, req.ProductID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
				}
				return apperrors.NewInternalError("查询商品失败: " + err.Error())
			}
			if product == nil {
				return apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
			}

			// sku_code 唯一：如果存在软删除记录，则“恢复并覆盖字段”
			existingAny, err := skuRepoTx.GetBySkuCodeUnscoped(ctx, req.SkuCode)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.NewInternalError("查询SKU失败: " + err.Error())
			}
			if existingAny != nil {
				// 未删除：直接提示已存在
				if !existingAny.DeletedAt.Valid {
					return apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在")
				}
				// 已删除：恢复并更新（等价于重新创建）
				specsJSON, err := json.Marshal(req.Specs)
				if err != nil {
					return apperrors.NewInternalError("规格属性格式错误: " + err.Error())
				}
				now := time.Now()
				updates := map[string]any{
					"product_id":     req.ProductID,
					"sku_code":       req.SkuCode,
					"name":           req.Name,
					"specs":          string(specsJSON),
					"price":          req.Price,
					"original_price": req.OriginalPrice,
					"stock":          req.Stock,
					"image":          req.Image,
					"weight":         req.Weight,
					"volume":         req.Volume,
					"status":         req.Status,
					"updated_at":     now,
				}
				if err := skuRepoTx.RestoreAndUpdateByID(ctx, existingAny.ID, updates); err != nil {
					if isDuplicateSkuCodeErr(err) {
						return apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在")
					}
					return apperrors.NewInternalError("恢复SKU失败: " + err.Error())
				}
				restored, _ := skuRepoTx.GetByID(ctx, existingAny.ID)
				if restored == nil {
					restored = existingAny
				}
				outSku = restored

				// 同事务写 outbox
				payloadBytes, _ := json.Marshal(map[string]any{"product_id": req.ProductID})
				payload := string(payloadBytes)
				evt := &outbox.Event{
					AggregateType: "product",
					AggregateID:   fmt.Sprintf("%d", req.ProductID),
					EventType:     outbox.EventProductUpserted,
					Payload:       &payload,
					Status:        outbox.StatusPending,
				}
				return l.outboxRepo.CreateInTx(ctx, tx, evt)
			}

			// 将规格属性转换为JSON
			specsJSON, err := json.Marshal(req.Specs)
			if err != nil {
				return apperrors.NewInternalError("规格属性格式错误: " + err.Error())
			}

			now := time.Now()
			sku := &model.Sku{
				ProductID:     req.ProductID,
				SkuCode:       req.SkuCode,
				Name:          req.Name,
				Specs:         string(specsJSON),
				Price:         req.Price,
				OriginalPrice: req.OriginalPrice,
				Stock:         req.Stock,
				Image:         req.Image,
				Weight:        req.Weight,
				Volume:        req.Volume,
				Status:        req.Status,
				CreatedAt:     now,
				UpdatedAt:     now,
			}

			if err := skuRepoTx.Create(ctx, sku); err != nil {
				if isDuplicateSkuCodeErr(err) {
					return apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在（可能是之前删除过的记录），请更换SKU编码")
				}
				return apperrors.NewInternalError("创建SKU失败: " + err.Error())
			}
			outSku = sku

			// 同事务写 outbox
			payloadBytes, _ := json.Marshal(map[string]any{"product_id": sku.ProductID})
			payload := string(payloadBytes)
			evt := &outbox.Event{
				AggregateType: "product",
				AggregateID:   fmt.Sprintf("%d", sku.ProductID),
				EventType:     outbox.EventProductUpserted,
				Payload:       &payload,
				Status:        outbox.StatusPending,
			}
			return l.outboxRepo.CreateInTx(ctx, tx, evt)
		}); err != nil {
			// 业务错误直接透传
			return nil, err
		}

		// 清除缓存（事务提交后）
		if outSku != nil && l.cache != nil {
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixSkuInfo, outSku.ID))
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, outSku.ProductID))
			_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
		}
		// 同步初始化库存记录
		if outSku != nil {
			l.ensureInventory(ctx, outSku.ID, outSku.Stock)
		}

		return &CreateSkuResponse{Sku: outSku}, nil
	}

	// 验证必填字段
	if req.ProductID == 0 {
		return nil, apperrors.NewInvalidParamError("商品ID不能为空")
	}
	if req.SkuCode == "" {
		return nil, apperrors.NewInvalidParamError("SKU编码不能为空")
	}
	if req.Name == "" {
		return nil, apperrors.NewInvalidParamError("SKU名称不能为空")
	}
	if len(req.Specs) == 0 {
		return nil, apperrors.NewInvalidParamError("规格属性不能为空")
	}
	if req.Price <= 0 {
		return nil, apperrors.NewInvalidParamError("价格必须大于0")
	}

	// 检查商品是否存在
	product, err := l.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
		}
		return nil, apperrors.NewInternalError("查询商品失败: " + err.Error())
	}
	if product == nil {
		return nil, apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
	}

	// sku_code 唯一：如果存在软删除记录，则“恢复并覆盖字段”（用户期望：删了能用同一个 sku_code 重新添加）
	// 如果存在未删除记录，则报已存在。
	existingAny, err := l.skuRepo.GetBySkuCodeUnscoped(ctx, req.SkuCode)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.NewInternalError("查询SKU失败: " + err.Error())
	}
	if existingAny != nil {
		// 未删除：直接提示已存在
		if !existingAny.DeletedAt.Valid {
			return nil, apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在")
		}
		// 已删除：恢复并更新（等价于重新创建）
		specsJSON, err := json.Marshal(req.Specs)
		if err != nil {
			return nil, apperrors.NewInternalError("规格属性格式错误: " + err.Error())
		}
		now := time.Now()
		updates := map[string]any{
			"product_id":     req.ProductID,
			"sku_code":       req.SkuCode,
			"name":           req.Name,
			"specs":          string(specsJSON),
			"price":          req.Price,
			"original_price": req.OriginalPrice,
			"stock":          req.Stock,
			"image":          req.Image,
			"weight":         req.Weight,
			"volume":         req.Volume,
			"status":         req.Status,
			"updated_at":     now,
		}
		if err := l.skuRepo.RestoreAndUpdateByID(ctx, existingAny.ID, updates); err != nil {
			if isDuplicateSkuCodeErr(err) {
				return nil, apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在")
			}
			return nil, apperrors.NewInternalError("恢复SKU失败: " + err.Error())
		}
		restored, _ := l.skuRepo.GetByID(ctx, existingAny.ID)
		if restored == nil {
			restored = existingAny
		}
		// 清除缓存
		if l.cache != nil {
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixSkuInfo, existingAny.ID))
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, req.ProductID))
			_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
		}
		return &CreateSkuResponse{Sku: restored}, nil
	}

	// 将规格属性转换为JSON
	specsJSON, err := json.Marshal(req.Specs)
	if err != nil {
		return nil, apperrors.NewInternalError("规格属性格式错误: " + err.Error())
	}

	now := time.Now()
	sku := &model.Sku{
		ProductID:     req.ProductID,
		SkuCode:       req.SkuCode,
		Name:          req.Name,
		Specs:         string(specsJSON),
		Price:         req.Price,
		OriginalPrice: req.OriginalPrice,
		Stock:         req.Stock,
		Image:         req.Image,
		Weight:        req.Weight,
		Volume:        req.Volume,
		Status:        req.Status,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := l.skuRepo.Create(ctx, sku); err != nil {
		if isDuplicateSkuCodeErr(err) {
			return nil, apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在（可能是之前删除过的记录），请更换SKU编码")
		}
		return nil, apperrors.NewInternalError("创建SKU失败: " + err.Error())
	}
	// 同步初始化库存记录
	l.ensureInventory(ctx, sku.ID, sku.Stock)

	// 清除缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixSkuInfo, sku.ID)
		_ = l.cache.Delete(ctx, cacheKey)
		// SKU 变更会影响商品详情（规格/价格）与商品列表（最低价/总库存）
		_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, sku.ProductID))
		_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
	}

	return &CreateSkuResponse{
		Sku: sku,
	}, nil
}

// ensureInventory 确保 SKU 有对应的库存记录并与 sku.stock 同步（创建或更新 SKU 时调用）
func (l *ProductLogic) ensureInventory(ctx context.Context, skuID uint64, stock int) {
	if l.db == nil || skuID == 0 {
		return
	}
	var count int64
	if err := l.db.WithContext(ctx).Table("inventory").Where("sku_id = ?", skuID).Count(&count).Error; err != nil {
		return
	}
	now := time.Now()
	if count == 0 {
		// 创建库存记录
		_ = l.db.WithContext(ctx).Table("inventory").Create(map[string]interface{}{
			"sku_id":             skuID,
			"total_stock":        stock,
			"available_stock":    stock,
			"locked_stock":       0,
			"sold_stock":         0,
			"low_stock_threshold": 10,
			"created_at":         now,
			"updated_at":         now,
		}).Error
	} else {
		// 同步库存：total = sku.stock，available = total - sold - locked
		_ = l.db.WithContext(ctx).Exec(
			"UPDATE inventory SET total_stock = ?, available_stock = ? - sold_stock - locked_stock, updated_at = ? WHERE sku_id = ?",
			stock, stock, now, skuID,
		).Error
	}
}

// UpdateSkuRequest 更新SKU请求
type UpdateSkuRequest struct {
	ID            uint64
	SkuCode       string
	Name          string
	Specs         map[string]string
	Price         float64
	OriginalPrice *float64
	Stock         int
	Image         string
	Weight        *float64
	Volume        *float64
	Status        int8
}

// UpdateSkuResponse 更新SKU响应
type UpdateSkuResponse struct {
	Sku *model.Sku
}

// UpdateSku 更新SKU
func (l *ProductLogic) UpdateSku(ctx context.Context, req *UpdateSkuRequest) (*UpdateSkuResponse, error) {
	if l.skuRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}
	// Outbox 要求：SKU 变更需要触发商品索引更新
	if l.db != nil && l.outboxRepo != nil {
		var outSku *model.Sku
		if err := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			skuRepoTx := repository.NewSkuRepository(tx)
			if req.ID == 0 {
				return apperrors.NewInvalidParamError("SKU ID不能为空")
			}
			sku, err := skuRepoTx.GetByID(ctx, req.ID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
				}
				return apperrors.NewInternalError("查询SKU失败: " + err.Error())
			}
			if sku == nil {
				return apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
			}

			// 更新字段（复用原逻辑）
			if req.SkuCode != "" {
				if req.SkuCode != sku.SkuCode {
					existingSku, err := skuRepoTx.GetBySkuCode(ctx, req.SkuCode)
					if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
						return apperrors.NewInternalError("查询SKU失败: " + err.Error())
					}
					if existingSku != nil && existingSku.ID != req.ID {
						return apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在")
					}
					deletedSku, err := skuRepoTx.GetBySkuCodeUnscoped(ctx, req.SkuCode)
					if err == nil && deletedSku != nil && deletedSku.ID != req.ID {
						return apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在（可能是之前删除过的记录），请更换SKU编码")
					}
				}
				sku.SkuCode = req.SkuCode
			}
			if req.Name != "" {
				sku.Name = req.Name
			}
			if len(req.Specs) > 0 {
				specsJSON, err := json.Marshal(req.Specs)
				if err != nil {
					return apperrors.NewInternalError("规格属性格式错误: " + err.Error())
				}
				sku.Specs = string(specsJSON)
			}
			if req.Price > 0 {
				sku.Price = req.Price
			}
			if req.OriginalPrice != nil {
				sku.OriginalPrice = req.OriginalPrice
			}
			if req.Stock >= 0 {
				sku.Stock = req.Stock
			}
			if req.Image != "" {
				sku.Image = req.Image
			}
			if req.Weight != nil {
				sku.Weight = req.Weight
			}
			if req.Volume != nil {
				sku.Volume = req.Volume
			}
			if req.Status >= 0 {
				sku.Status = req.Status
			}
			sku.UpdatedAt = time.Now()

			if err := skuRepoTx.Update(ctx, sku); err != nil {
				if isDuplicateSkuCodeErr(err) {
					return apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在（可能是之前删除过的记录），请更换SKU编码")
				}
				return apperrors.NewInternalError("更新SKU失败: " + err.Error())
			}
			outSku = sku

			// 同事务写 outbox
			payloadBytes, _ := json.Marshal(map[string]any{"product_id": sku.ProductID})
			payload := string(payloadBytes)
			evt := &outbox.Event{
				AggregateType: "product",
				AggregateID:   fmt.Sprintf("%d", sku.ProductID),
				EventType:     outbox.EventProductUpserted,
				Payload:       &payload,
				Status:        outbox.StatusPending,
			}
			return l.outboxRepo.CreateInTx(ctx, tx, evt)
		}); err != nil {
			return nil, err
		}

		// 清除缓存
		if outSku != nil && l.cache != nil {
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixSkuInfo, outSku.ID))
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, outSku.ProductID))
			_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
		}
		// 同步库存记录
		if outSku != nil {
			l.ensureInventory(ctx, outSku.ID, outSku.Stock)
		}

		return &UpdateSkuResponse{Sku: outSku}, nil
	}

	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("SKU ID不能为空")
	}

	// 获取现有SKU
	sku, err := l.skuRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
		}
		return nil, apperrors.NewInternalError("查询SKU失败: " + err.Error())
	}
	if sku == nil {
		return nil, apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
	}

	// 更新字段
	if req.SkuCode != "" {
		// 检查SKU编码是否与其他SKU重复
		if req.SkuCode != "" && req.SkuCode != sku.SkuCode {
			existingSku, err := l.skuRepo.GetBySkuCode(ctx, req.SkuCode)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.NewInternalError("查询SKU失败: " + err.Error())
			}
			if existingSku != nil && existingSku.ID != req.ID {
				return nil, apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在")
			}
			// 软删除记录也会占用唯一索引
			deletedSku, err := l.skuRepo.GetBySkuCodeUnscoped(ctx, req.SkuCode)
			if err == nil && deletedSku != nil && deletedSku.ID != req.ID {
				return nil, apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在（可能是之前删除过的记录），请更换SKU编码")
			}
		}
		sku.SkuCode = req.SkuCode
	}
	if req.Name != "" {
		sku.Name = req.Name
	}
	if len(req.Specs) > 0 {
		specsJSON, err := json.Marshal(req.Specs)
		if err != nil {
			return nil, apperrors.NewInternalError("规格属性格式错误: " + err.Error())
		}
		sku.Specs = string(specsJSON)
	}
	if req.Price > 0 {
		sku.Price = req.Price
	}
	if req.OriginalPrice != nil {
		sku.OriginalPrice = req.OriginalPrice
	}
	if req.Stock >= 0 {
		sku.Stock = req.Stock
	}
	if req.Image != "" {
		sku.Image = req.Image
	}
	if req.Weight != nil {
		sku.Weight = req.Weight
	}
	if req.Volume != nil {
		sku.Volume = req.Volume
	}
	if req.Status >= 0 {
		sku.Status = req.Status
	}

	sku.UpdatedAt = time.Now()

	if err := l.skuRepo.Update(ctx, sku); err != nil {
		if isDuplicateSkuCodeErr(err) {
			return nil, apperrors.NewError(apperrors.CodeAlreadyExists, "SKU编码已存在（可能是之前删除过的记录），请更换SKU编码")
		}
		return nil, apperrors.NewInternalError("更新SKU失败: " + err.Error())
	}
	// 同步库存记录
	l.ensureInventory(ctx, sku.ID, sku.Stock)

	// 清除缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixSkuInfo, sku.ID)
		_ = l.cache.Delete(ctx, cacheKey)
		_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, sku.ProductID))
		_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
	}

	return &UpdateSkuResponse{
		Sku: sku,
	}, nil
}

// DeleteSkuRequest 删除SKU请求
type DeleteSkuRequest struct {
	ID uint64
}

// DeleteSkuResponse 删除SKU响应
type DeleteSkuResponse struct{}

// DeleteSku 删除SKU（软删除）
func (l *ProductLogic) DeleteSku(ctx context.Context, req *DeleteSkuRequest) (*DeleteSkuResponse, error) {
	if l.skuRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}
	// Outbox 要求：SKU 变更需要触发商品索引更新
	if l.db != nil && l.outboxRepo != nil {
		var productID uint64
		if err := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			skuRepoTx := repository.NewSkuRepository(tx)
			if req.ID == 0 {
				return apperrors.NewInvalidParamError("SKU ID不能为空")
			}
			sku, err := skuRepoTx.GetByID(ctx, req.ID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
				}
				return apperrors.NewInternalError("查询SKU失败: " + err.Error())
			}
			if sku == nil {
				return apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
			}
			productID = sku.ProductID

			if err := skuRepoTx.Delete(ctx, req.ID); err != nil {
				return apperrors.NewInternalError("删除SKU失败: " + err.Error())
			}

			payloadBytes, _ := json.Marshal(map[string]any{"product_id": productID})
			payload := string(payloadBytes)
			evt := &outbox.Event{
				AggregateType: "product",
				AggregateID:   fmt.Sprintf("%d", productID),
				EventType:     outbox.EventProductUpserted,
				Payload:       &payload,
				Status:        outbox.StatusPending,
			}
			return l.outboxRepo.CreateInTx(ctx, tx, evt)
		}); err != nil {
			return nil, err
		}

		// 清除缓存
		if l.cache != nil {
			_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixSkuInfo, req.ID))
			if productID > 0 {
				_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, productID))
			}
			_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
		}

		return &DeleteSkuResponse{}, nil
	}

	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("SKU ID不能为空")
	}

	// 检查SKU是否存在
	sku, err := l.skuRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
		}
		return nil, apperrors.NewInternalError("查询SKU失败: " + err.Error())
	}
	if sku == nil {
		return nil, apperrors.NewError(apperrors.CodeNotFound, "SKU不存在")
	}

	// 执行软删除
	if err := l.skuRepo.Delete(ctx, req.ID); err != nil {
		return nil, apperrors.NewInternalError("删除SKU失败: " + err.Error())
	}

	// 清除缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixSkuInfo, req.ID)
		_ = l.cache.Delete(ctx, cacheKey)
		_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixProductDetail, sku.ProductID))
		_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
	}

	return &DeleteSkuResponse{}, nil
}

// GetCategoryListRequest 获取类目列表请求
type GetCategoryListRequest struct {
	ParentID uint64
	Level    int8
	Status   int8
}

// GetCategoryListResponse 获取类目列表响应
type GetCategoryListResponse struct {
	Categories []*model.Category
}

// GetCategoryList 获取类目列表
func (l *ProductLogic) GetCategoryList(ctx context.Context, req *GetCategoryListRequest) (*GetCategoryListResponse, error) {
	// 检查 repository 是否初始化
	if l.categoryRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化，请检查数据库配置")
	}

	var categories []*model.Category
	var err error

	if req.ParentID > 0 {
		// 根据父ID查询子类目
		categories, err = l.categoryRepo.GetByParentID(ctx, req.ParentID)
	} else {
		// 查询所有类目
		categories, err = l.categoryRepo.GetAll(ctx, req.Status)
	}

	if err != nil {
		return nil, apperrors.NewInternalError("查询类目列表失败: " + err.Error())
	}

	return &GetCategoryListResponse{
		Categories: categories,
	}, nil
}

// GetCategoryTreeRequest 获取类目树请求
type GetCategoryTreeRequest struct {
	Status int8
}

// GetCategoryTreeResponse 获取类目树响应
type GetCategoryTreeResponse struct {
	Categories []*CategoryTreeNode
}

// CategoryTreeNode 类目树节点
type CategoryTreeNode struct {
	*model.Category
	Children []*CategoryTreeNode `json:"children"`
}

// GetCategoryTree 获取类目树（带缓存）
func (l *ProductLogic) GetCategoryTree(ctx context.Context, req *GetCategoryTreeRequest) (*GetCategoryTreeResponse, error) {
	// 检查 repository 是否初始化
	if l.categoryRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化，请检查数据库配置")
	}

	// 构建缓存 key（包含 status 参数，避免不同 status 的缓存冲突）
	cacheKey := fmt.Sprintf("%s:status:%d", cache.KeyPrefixCategoryTree, req.Status)

	// 尝试从缓存获取
	if l.cache != nil {
		var cachedResp GetCategoryTreeResponse
		if err := l.cache.GetJSON(ctx, cacheKey, &cachedResp); err == nil {
			return &cachedResp, nil
		}
	}

	// 获取所有类目
	allCategories, err := l.categoryRepo.GetAll(ctx, req.Status)
	if err != nil {
		return nil, apperrors.NewInternalError("查询类目失败: " + err.Error())
	}

	// 构建类目树
	tree := buildCategoryTree(allCategories)

	resp := &GetCategoryTreeResponse{
		Categories: tree,
	}

	// 写入缓存（24小时）
	if l.cache != nil {
		_ = l.cache.Set(ctx, cacheKey, resp, 24*time.Hour)
	}

	return resp, nil
}

// buildCategoryTree 构建类目树
func buildCategoryTree(categories []*model.Category) []*CategoryTreeNode {
	// 创建映射表
	categoryMap := make(map[uint64]*CategoryTreeNode)
	for _, cat := range categories {
		categoryMap[cat.ID] = &CategoryTreeNode{
			Category: cat,
			Children: []*CategoryTreeNode{},
		}
	}

	// 构建树结构
	var rootNodes []*CategoryTreeNode
	for _, cat := range categories {
		node := categoryMap[cat.ID]
		if cat.ParentID == 0 {
			// 根节点
			rootNodes = append(rootNodes, node)
		} else {
			// 子节点
			if parent, ok := categoryMap[cat.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return rootNodes
}

// clearCategoryTreeCache 清除所有 status 的分类树缓存
func (l *ProductLogic) clearCategoryTreeCache(ctx context.Context) {
	if l.cache == nil {
		return
	}
	// 清除常见的 status 值的缓存（-1: 所有, 0: 禁用, 1: 启用, 2: 其他状态）
	statuses := []int8{-1, 0, 1, 2}
	for _, status := range statuses {
		cacheKey := fmt.Sprintf("%s:status:%d", cache.KeyPrefixCategoryTree, status)
		_ = l.cache.Delete(ctx, cacheKey)
	}
}

// parseJSONArray 解析JSON数组字符串
func parseJSONArray(jsonStr string) ([]string, error) {
	if jsonStr == "" {
		return []string{}, nil
	}
	var result []string
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}

// parseJSONMap 解析JSON对象字符串
func parseJSONMap(jsonStr string) (map[string]string, error) {
	if jsonStr == "" {
		return map[string]string{}, nil
	}
	var result map[string]string
	err := json.Unmarshal([]byte(jsonStr), &result)
	return result, err
}

// CreateProductRequest 创建商品请求
type CreateProductRequest struct {
	Name           string
	Subtitle       string
	CategoryID     uint64
	BrandID        *uint64
	MainImage      string
	LocalMainImage string
	Images         []string
	LocalImages    []string
	Detail         string
	Price          float64
	OriginalPrice  float64
	Stock          int
	Status         int8
	IsHot          int8
}

// CreateProductResponse 创建商品响应
type CreateProductResponse struct {
	Product *model.Product
}

// CreateProduct 创建商品（管理后台）
func (l *ProductLogic) CreateProduct(ctx context.Context, req *CreateProductRequest) (*CreateProductResponse, error) {
	if l.productRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}
	// Outbox 要求：商品写库与 outbox 同事务
	if l.db != nil && l.outboxRepo != nil {
		// 参数验证（复用原逻辑）
		if req.Name == "" {
			return nil, apperrors.NewInvalidParamError("商品名称不能为空")
		}
		if req.CategoryID == 0 {
			return nil, apperrors.NewInvalidParamError("分类ID不能为空")
		}
		if req.Price <= 0 {
			return nil, apperrors.NewInvalidParamError("价格必须大于0")
		}

		// 转换图片列表为JSON
		imagesJSON := "[]"
		if len(req.Images) > 0 {
			imagesBytes, err := json.Marshal(req.Images)
			if err != nil {
				return nil, apperrors.NewInternalError("图片列表格式错误: " + err.Error())
			}
			imagesJSON = string(imagesBytes)
		}

		// 转换本地图片列表为JSON
		localImagesJSON := "[]"
		if len(req.LocalImages) > 0 {
			localImagesBytes, err := json.Marshal(req.LocalImages)
			if err != nil {
				return nil, apperrors.NewInternalError("本地图片列表格式错误: " + err.Error())
			}
			localImagesJSON = string(localImagesBytes)
		}

		spuCode := fmt.Sprintf("SPU%08d", time.Now().Unix()%100000000)
		now := time.Now()
		product := &model.Product{
			SpuCode:        spuCode,
			Name:           req.Name,
			Subtitle:       req.Subtitle,
			CategoryID:     req.CategoryID,
			BrandID:        req.BrandID,
			MainImage:      req.MainImage,
			LocalMainImage: req.LocalMainImage,
			Images:         imagesJSON,
			LocalImages:    localImagesJSON,
			Detail:         req.Detail,
			Price:          req.Price,
			Stock:          req.Stock,
			Status:         req.Status,
			IsHot:          req.IsHot,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if req.OriginalPrice > 0 {
			product.OriginalPrice = &req.OriginalPrice
		}

		if err := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			productRepoTx := repository.NewProductRepository(tx)
			if err := productRepoTx.Create(ctx, product); err != nil {
				return apperrors.NewInternalError("创建商品失败: " + err.Error())
			}

			payloadBytes, _ := json.Marshal(map[string]any{"product_id": product.ID})
			payload := string(payloadBytes)
			evt := &outbox.Event{
				AggregateType: "product",
				AggregateID:   fmt.Sprintf("%d", product.ID),
				EventType:     outbox.EventProductUpserted,
				Payload:       &payload,
				Status:        outbox.StatusPending,
			}
			return l.outboxRepo.CreateInTx(ctx, tx, evt)
		}); err != nil {
			return nil, err
		}

		// 清除相关缓存
		if l.cache != nil {
			_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
			cacheKey := cache.BuildKey(cache.KeyPrefixProductDetail, product.ID)
			_ = l.cache.Delete(ctx, cacheKey)
		}

		return &CreateProductResponse{Product: product}, nil
	}

	// 参数验证
	if req.Name == "" {
		return nil, apperrors.NewInvalidParamError("商品名称不能为空")
	}
	if req.CategoryID == 0 {
		return nil, apperrors.NewInvalidParamError("分类ID不能为空")
	}
	if req.Price <= 0 {
		return nil, apperrors.NewInvalidParamError("价格必须大于0")
	}

	// 转换图片列表为JSON
	imagesJSON := "[]"
	if len(req.Images) > 0 {
		imagesBytes, err := json.Marshal(req.Images)
		if err != nil {
			return nil, apperrors.NewInternalError("图片列表格式错误: " + err.Error())
		}
		imagesJSON = string(imagesBytes)
	}

	// 转换本地图片列表为JSON
	localImagesJSON := "[]"
	if len(req.LocalImages) > 0 {
		localImagesBytes, err := json.Marshal(req.LocalImages)
		if err != nil {
			return nil, apperrors.NewInternalError("本地图片列表格式错误: " + err.Error())
		}
		localImagesJSON = string(localImagesBytes)
	}

	// 生成SPU编码（简化处理，实际应该使用更复杂的生成规则）
	spuCode := fmt.Sprintf("SPU%08d", time.Now().Unix()%100000000)

	// 创建商品
	now := time.Now()
	product := &model.Product{
		SpuCode:        spuCode,
		Name:           req.Name,
		Subtitle:       req.Subtitle,
		CategoryID:     req.CategoryID,
		BrandID:        req.BrandID,
		MainImage:      req.MainImage,
		LocalMainImage: req.LocalMainImage,
		Images:         imagesJSON,
		LocalImages:    localImagesJSON,
		Detail:         req.Detail,
		Price:          req.Price,
		Stock:          req.Stock,
		Status:         req.Status,
		IsHot:          req.IsHot,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if req.OriginalPrice > 0 {
		product.OriginalPrice = &req.OriginalPrice
	}

	if err := l.productRepo.Create(ctx, product); err != nil {
		return nil, apperrors.NewInternalError("创建商品失败: " + err.Error())
	}

	// 清除相关缓存
	if l.cache != nil {
		// 清除所有商品列表缓存（使用通配符匹配）
		_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
		// 清除商品详情缓存（如果存在）
		cacheKey := cache.BuildKey(cache.KeyPrefixProductDetail, product.ID)
		_ = l.cache.Delete(ctx, cacheKey)
	}

	return &CreateProductResponse{
		Product: product,
	}, nil
}

// UpdateProductRequest 更新商品请求
type UpdateProductRequest struct {
	ID             uint64
	Name           string
	Subtitle       string
	CategoryID     uint64
	BrandID        *uint64
	MainImage      string
	LocalMainImage string
	Images         []string
	LocalImages    []string
	Detail         string
	Price          float64
	OriginalPrice  float64
	Stock          int
	Status         int8
	IsHot          int8
}

// UpdateProductResponse 更新商品响应
type UpdateProductResponse struct {
	Product *model.Product
}

// UpdateProduct 更新商品（管理后台）
func (l *ProductLogic) UpdateProduct(ctx context.Context, req *UpdateProductRequest) (*UpdateProductResponse, error) {
	if l.productRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}
	// Outbox 要求：商品写库与 outbox 同事务
	if l.db != nil && l.outboxRepo != nil {
		if req.ID == 0 {
			return nil, apperrors.NewInvalidParamError("商品ID不能为空")
		}
		var updated *model.Product
		if err := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			productRepoTx := repository.NewProductRepository(tx)
			product, err := productRepoTx.GetByID(ctx, req.ID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
				}
				return apperrors.NewInternalError("查询商品失败: " + err.Error())
			}
			if product == nil {
				return apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
			}

			if req.Name != "" {
				product.Name = req.Name
			}
			if req.Subtitle != "" {
				product.Subtitle = req.Subtitle
			}
			if req.CategoryID > 0 {
				product.CategoryID = req.CategoryID
			}
			if req.BrandID != nil {
				product.BrandID = req.BrandID
			}
			if req.MainImage != "" {
				product.MainImage = req.MainImage
			}
			if req.LocalMainImage != "" {
				product.LocalMainImage = req.LocalMainImage
			}
			if req.Images != nil {
				imagesBytes, err := json.Marshal(req.Images)
				if err != nil {
					return apperrors.NewInternalError("图片列表格式错误: " + err.Error())
				}
				product.Images = string(imagesBytes)
			}
			if req.LocalImages != nil {
				localImagesBytes, err := json.Marshal(req.LocalImages)
				if err != nil {
					return apperrors.NewInternalError("本地图片列表格式错误: " + err.Error())
				}
				product.LocalImages = string(localImagesBytes)
			}
			if product.LocalImages == "" {
				product.LocalImages = "[]"
			}
			if product.Images == "" {
				product.Images = "[]"
			}
			if req.Detail != "" {
				product.Detail = req.Detail
			}
			if req.Price > 0 {
				product.Price = req.Price
			}
			if req.OriginalPrice > 0 {
				product.OriginalPrice = &req.OriginalPrice
			}
			if req.Status >= 0 && req.Status != -1 {
				product.Status = req.Status
			}
			if req.IsHot >= 0 && req.IsHot != -1 {
				product.IsHot = req.IsHot
			}
			product.Stock = req.Stock
			product.UpdatedAt = time.Now()

			if err := productRepoTx.Update(ctx, product); err != nil {
				return apperrors.NewInternalError("更新商品失败: " + err.Error())
			}
			updated = product

			payloadBytes, _ := json.Marshal(map[string]any{"product_id": product.ID})
			payload := string(payloadBytes)
			evt := &outbox.Event{
				AggregateType: "product",
				AggregateID:   fmt.Sprintf("%d", product.ID),
				EventType:     outbox.EventProductUpserted,
				Payload:       &payload,
				Status:        outbox.StatusPending,
			}
			return l.outboxRepo.CreateInTx(ctx, tx, evt)
		}); err != nil {
			return nil, err
		}

		// 清除相关缓存
		if l.cache != nil {
			cacheKey := cache.BuildKey(cache.KeyPrefixProductDetail, req.ID)
			_ = l.cache.Delete(ctx, cacheKey)
			_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
		}

		return &UpdateProductResponse{Product: updated}, nil
	}

	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("商品ID不能为空")
	}

	// 获取现有商品
	product, err := l.productRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
		}
		return nil, apperrors.NewInternalError("查询商品失败: " + err.Error())
	}
	if product == nil {
		return nil, apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
	}

	// 更新字段：只更新请求中明确提供的字段，未提供的字段保持原值
	// 这是标准的 PATCH 更新逻辑：根据 ID 查询记录，只更新提供的字段

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Subtitle != "" {
		product.Subtitle = req.Subtitle
	}
	// 分类ID：如果提供了且 > 0 则更新
	if req.CategoryID > 0 {
		product.CategoryID = req.CategoryID
	}
	if req.BrandID != nil {
		product.BrandID = req.BrandID
	}

	// 图片字段：只有当明确提供时才更新（非空字符串或非 nil）
	// 如果请求中没有提供图片字段（空字符串），保持原值不变
	if req.MainImage != "" {
		product.MainImage = req.MainImage
	}
	if req.LocalMainImage != "" {
		product.LocalMainImage = req.LocalMainImage
	}
	if req.Images != nil {
		imagesBytes, err := json.Marshal(req.Images)
		if err != nil {
			return nil, apperrors.NewInternalError("图片列表格式错误: " + err.Error())
		}
		product.Images = string(imagesBytes)
	}
	if req.LocalImages != nil {
		localImagesBytes, err := json.Marshal(req.LocalImages)
		if err != nil {
			return nil, apperrors.NewInternalError("本地图片列表格式错误: " + err.Error())
		}
		product.LocalImages = string(localImagesBytes)
	}

	// 确保LocalImages和Images字段始终是有效的JSON（不能是空字符串）
	if product.LocalImages == "" {
		product.LocalImages = "[]"
	}
	if product.Images == "" {
		product.Images = "[]"
	}

	if req.Detail != "" {
		product.Detail = req.Detail
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.OriginalPrice > 0 {
		product.OriginalPrice = &req.OriginalPrice
	}
	// 直接赋值：因为我们已经根据 ID 查询到了原记录
	// 如果前端没有传递某个字段，proto 默认值是 0（对于 int32）或空字符串（对于 string）
	// 但前端在部分更新时会明确传递需要更新的字段，未传递的字段保持原值
	// 对于 Status 和 IsHot：使用 -1 表示"不更新"（前端明确传递 -1），其他值（>= 0）表示要更新
	if req.Status >= 0 && req.Status != -1 {
		product.Status = req.Status
	}
	if req.IsHot >= 0 && req.IsHot != -1 {
		product.IsHot = req.IsHot
	}
	// Stock：直接赋值（前端只传递需要更新的字段，不传递的字段保持原值）
	// 如果前端没有传递 stock，proto 默认值是 0，但前端在部分更新时不会传递不需要更新的字段
	// 所以如果 stock 是 0，说明前端明确要设置为 0
	product.Stock = req.Stock

	product.UpdatedAt = time.Now()

	if err := l.productRepo.Update(ctx, product); err != nil {
		return nil, apperrors.NewInternalError("更新商品失败: " + err.Error())
	}

	// 清除相关缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixProductDetail, req.ID)
		_ = l.cache.Delete(ctx, cacheKey)
		// 清除所有商品列表缓存
		_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
	}

	return &UpdateProductResponse{
		Product: product,
	}, nil
}

// DeleteProductRequest 删除商品请求
type DeleteProductRequest struct {
	ID uint64
}

// DeleteProductResponse 删除商品响应
type DeleteProductResponse struct {
}

// DeleteProduct 删除商品（管理后台）
func (l *ProductLogic) DeleteProduct(ctx context.Context, req *DeleteProductRequest) (*DeleteProductResponse, error) {
	if l.productRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}
	// Outbox 要求：删除与 outbox 同事务（下游删除 ES 文档）
	if l.db != nil && l.outboxRepo != nil {
		if req.ID == 0 {
			return nil, apperrors.NewInvalidParamError("商品ID不能为空")
		}
		if err := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			productRepoTx := repository.NewProductRepository(tx)
			// 确认存在
			product, err := productRepoTx.GetByID(ctx, req.ID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
				}
				return apperrors.NewInternalError("查询商品失败: " + err.Error())
			}
			if product == nil {
				return apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
			}

			if err := productRepoTx.Delete(ctx, req.ID); err != nil {
				return apperrors.NewInternalError("删除商品失败: " + err.Error())
			}

			payloadBytes, _ := json.Marshal(map[string]any{"product_id": req.ID})
			payload := string(payloadBytes)
			evt := &outbox.Event{
				AggregateType: "product",
				AggregateID:   fmt.Sprintf("%d", req.ID),
				EventType:     outbox.EventProductDeleted,
				Payload:       &payload,
				Status:        outbox.StatusPending,
			}
			return l.outboxRepo.CreateInTx(ctx, tx, evt)
		}); err != nil {
			return nil, err
		}

		// 清除相关缓存
		if l.cache != nil {
			cacheKey := cache.BuildKey(cache.KeyPrefixProductDetail, req.ID)
			_ = l.cache.Delete(ctx, cacheKey)
			_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
		}

		return &DeleteProductResponse{}, nil
	}

	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("商品ID不能为空")
	}

	// 检查商品是否存在
	product, err := l.productRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
		}
		return nil, apperrors.NewInternalError("查询商品失败: " + err.Error())
	}
	if product == nil {
		return nil, apperrors.NewError(apperrors.CodeProductNotFound, "商品不存在")
	}

	// 删除商品（软删除）
	if err := l.productRepo.Delete(ctx, req.ID); err != nil {
		return nil, apperrors.NewInternalError("删除商品失败: " + err.Error())
	}

	// 清除相关缓存
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixProductDetail, req.ID)
		_ = l.cache.Delete(ctx, cacheKey)
		// 清除所有商品列表缓存
		_ = l.cache.DeletePattern(ctx, cache.KeyPrefixProductList+"*")
	}

	return &DeleteProductResponse{}, nil
}

// ============================================
// 分类CRUD操作
// ============================================

// GetCategoryRequest 获取类目详情请求
type GetCategoryRequest struct {
	ID uint64
}

// GetCategoryResponse 获取类目详情响应
type GetCategoryResponse struct {
	Category *model.Category
}

// GetCategory 获取类目详情
func (l *ProductLogic) GetCategory(ctx context.Context, req *GetCategoryRequest) (*GetCategoryResponse, error) {
	if l.categoryRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("类目ID不能为空")
	}

	category, err := l.categoryRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeCategoryNotFound, "类目不存在")
		}
		return nil, apperrors.NewInternalError("查询类目失败: " + err.Error())
	}

	return &GetCategoryResponse{
		Category: category,
	}, nil
}

// CreateCategoryRequest 创建类目请求
type CreateCategoryRequest struct {
	ParentID    uint64
	Name        string
	Level       int8
	Sort        int
	Icon        string
	IconLocal   string
	Image       string
	ImageLocal  string
	Description string
	Status      int8
}

// CreateCategoryResponse 创建类目响应
type CreateCategoryResponse struct {
	Category *model.Category
}

// CreateCategory 创建类目（管理后台）
func (l *ProductLogic) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*CreateCategoryResponse, error) {
	if l.categoryRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	// 参数验证
	if req.Name == "" {
		return nil, apperrors.NewInvalidParamError("类目名称不能为空")
	}
	if req.Level <= 0 {
		req.Level = 1 // 默认一级分类
	}

	// 如果指定了父类目，验证父类目是否存在
	if req.ParentID > 0 {
		parent, err := l.categoryRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.NewError(apperrors.CodeCategoryNotFound, "父类目不存在")
			}
			return nil, apperrors.NewInternalError("查询父类目失败: " + err.Error())
		}
		// 自动设置层级为父类目层级+1
		if req.Level <= parent.Level {
			req.Level = parent.Level + 1
		}
	} else {
		req.ParentID = 0 // 确保顶级分类的parent_id为0
		req.Level = 1
	}

	// 创建类目
	now := time.Now()
	category := &model.Category{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Level:       req.Level,
		Sort:        req.Sort,
		Icon:        req.Icon,
		IconLocal:   req.IconLocal,
		Image:       req.Image,
		ImageLocal:  req.ImageLocal,
		Description: req.Description,
		Status:      req.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := l.categoryRepo.Create(ctx, category); err != nil {
		return nil, apperrors.NewInternalError("创建类目失败: " + err.Error())
	}

	// 清除类目树缓存（清除所有 status 的缓存）
	l.clearCategoryTreeCache(ctx)

	return &CreateCategoryResponse{
		Category: category,
	}, nil
}

// UpdateCategoryRequest 更新类目请求
type UpdateCategoryRequest struct {
	ID          uint64
	ParentID    uint64
	Name        string
	Level       int8
	Sort        int
	Icon        string
	IconLocal   string
	Image       string
	ImageLocal  string
	Description string
	Status      int8
}

// UpdateCategoryResponse 更新类目响应
type UpdateCategoryResponse struct {
	Category *model.Category
}

// UpdateCategory 更新类目（管理后台）
func (l *ProductLogic) UpdateCategory(ctx context.Context, req *UpdateCategoryRequest) (*UpdateCategoryResponse, error) {
	if l.categoryRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("类目ID不能为空")
	}

	// 获取现有类目
	category, err := l.categoryRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeCategoryNotFound, "类目不存在")
		}
		return nil, apperrors.NewInternalError("查询类目失败: " + err.Error())
	}

	// 更新字段
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.ParentID > 0 {
		// 验证父类目是否存在
		parent, err := l.categoryRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperrors.NewError(apperrors.CodeCategoryNotFound, "父类目不存在")
			}
			return nil, apperrors.NewInternalError("查询父类目失败: " + err.Error())
		}
		// 防止将类目设置为自己的子类目（简单检查，实际应该递归检查）
		if parent.ParentID == req.ID {
			return nil, apperrors.NewInvalidParamError("不能将类目设置为自己的子类目")
		}
		category.ParentID = req.ParentID
		if req.Level > 0 {
			category.Level = req.Level
		} else {
			category.Level = parent.Level + 1
		}
	} else if req.Level > 0 {
		category.Level = req.Level
	}
	if req.Sort > 0 {
		category.Sort = req.Sort
	}
	if req.Icon != "" {
		category.Icon = req.Icon
	}
	if req.IconLocal != "" {
		category.IconLocal = req.IconLocal
	}
	if req.Image != "" {
		category.Image = req.Image
	}
	if req.ImageLocal != "" {
		category.ImageLocal = req.ImageLocal
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if req.Status >= 0 {
		category.Status = req.Status
	}

	category.UpdatedAt = time.Now()

	if err := l.categoryRepo.Update(ctx, category); err != nil {
		return nil, apperrors.NewInternalError("更新类目失败: " + err.Error())
	}

	// 清除类目树缓存（清除所有 status 的缓存）
	l.clearCategoryTreeCache(ctx)

	return &UpdateCategoryResponse{
		Category: category,
	}, nil
}

// DeleteCategoryRequest 删除类目请求
type DeleteCategoryRequest struct {
	ID uint64
}

// DeleteCategoryResponse 删除类目响应
type DeleteCategoryResponse struct {
}

// DeleteCategory 删除类目（管理后台）
func (l *ProductLogic) DeleteCategory(ctx context.Context, req *DeleteCategoryRequest) (*DeleteCategoryResponse, error) {
	if l.categoryRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("类目ID不能为空")
	}

	// 检查类目是否存在
	_, err := l.categoryRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeCategoryNotFound, "类目不存在")
		}
		return nil, apperrors.NewInternalError("查询类目失败: " + err.Error())
	}

	// 检查是否有子类目
	children, err := l.categoryRepo.GetByParentID(ctx, req.ID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询子类目失败: " + err.Error())
	}
	if len(children) > 0 {
		return nil, apperrors.NewInvalidParamError("该类目下存在子类目，无法删除")
	}

	// 检查是否有商品使用该类目
	if l.productRepo != nil {
		// 这里可以添加检查商品是否使用该类目的逻辑
		// 暂时允许删除，实际应该检查
	}

	// 删除类目
	if err := l.categoryRepo.Delete(ctx, req.ID); err != nil {
		return nil, apperrors.NewInternalError("删除类目失败: " + err.Error())
	}

	// 清除类目树缓存（清除所有 status 的缓存）
	l.clearCategoryTreeCache(ctx)

	return &DeleteCategoryResponse{}, nil
}

// ==================== Banner 相关方法 ====================

// ListBannersRequest 获取Banner列表请求
type ListBannersRequest struct {
	Status int8
	Limit  int
}

// ListBannersResponse 获取Banner列表响应
type ListBannersResponse struct {
	Banners []*model.Banner
}

// ListBanners 获取Banner列表
func (l *ProductLogic) ListBanners(ctx context.Context, req *ListBannersRequest) (*ListBannersResponse, error) {
	if l.bannerRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	banners, err := l.bannerRepo.GetAll(ctx, req.Status, req.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError("查询Banner列表失败: " + err.Error())
	}

	return &ListBannersResponse{
		Banners: banners,
	}, nil
}

// GetBannerRequest 获取Banner详情请求
type GetBannerRequest struct {
	ID uint64
}

// GetBannerResponse 获取Banner详情响应
type GetBannerResponse struct {
	Banner *model.Banner
}

// GetBanner 获取Banner详情
func (l *ProductLogic) GetBanner(ctx context.Context, req *GetBannerRequest) (*GetBannerResponse, error) {
	if l.bannerRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("Banner ID不能为空")
	}

	banner, err := l.bannerRepo.GetByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeNotFound, "Banner不存在")
		}
		return nil, apperrors.NewInternalError("查询Banner失败: " + err.Error())
	}

	return &GetBannerResponse{
		Banner: banner,
	}, nil
}

// CreateBannerRequest 创建Banner请求
type CreateBannerRequest struct {
	Title       string
	Description string
	Image       string
	ImageLocal  string
	Link        string
	LinkType    int8
	Sort        int
	Status      int8
	StartTime   *time.Time
	EndTime     *time.Time
}

// CreateBannerResponse 创建Banner响应
type CreateBannerResponse struct {
	Banner *model.Banner
}

// CreateBanner 创建Banner
func (l *ProductLogic) CreateBanner(ctx context.Context, req *CreateBannerRequest) (*CreateBannerResponse, error) {
	if l.bannerRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	if req.Image == "" && req.ImageLocal == "" {
		return nil, apperrors.NewInvalidParamError("封面图片不能为空")
	}

	banner := &model.Banner{
		Title:       req.Title,
		Description: req.Description,
		Image:       req.Image,
		ImageLocal:  req.ImageLocal,
		Link:        req.Link,
		LinkType:    req.LinkType,
		Sort:        req.Sort,
		Status:      req.Status,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}

	if err := l.bannerRepo.Create(ctx, banner); err != nil {
		return nil, apperrors.NewInternalError("创建Banner失败: " + err.Error())
	}

	return &CreateBannerResponse{
		Banner: banner,
	}, nil
}

// UpdateBannerRequest 更新Banner请求
type UpdateBannerRequest struct {
	ID          uint64
	Title       string
	Description string
	Image       string
	ImageLocal  string
	Link        string
	LinkType    int8
	Sort        int
	Status      int8
	StartTime   *time.Time
	EndTime     *time.Time
}

// UpdateBannerResponse 更新Banner响应
type UpdateBannerResponse struct {
	Banner *model.Banner
}

// UpdateBanner 更新Banner
func (l *ProductLogic) UpdateBanner(ctx context.Context, req *UpdateBannerRequest) (*UpdateBannerResponse, error) {
	if l.bannerRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("Banner ID不能为空")
	}

	banner, err := l.bannerRepo.GetByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeNotFound, "Banner不存在")
		}
		return nil, apperrors.NewInternalError("查询Banner失败: " + err.Error())
	}

	// 更新字段
	if req.Title != "" {
		banner.Title = req.Title
	}
	if req.Description != "" {
		banner.Description = req.Description
	}
	if req.Image != "" {
		banner.Image = req.Image
	}
	if req.ImageLocal != "" {
		banner.ImageLocal = req.ImageLocal
	}
	if req.Link != "" {
		banner.Link = req.Link
	}
	if req.LinkType >= 0 {
		banner.LinkType = req.LinkType
	}
	if req.Sort > 0 {
		banner.Sort = req.Sort
	}
	if req.Status >= 0 {
		banner.Status = req.Status
	}
	if req.StartTime != nil {
		banner.StartTime = req.StartTime
	}
	if req.EndTime != nil {
		banner.EndTime = req.EndTime
	}

	banner.UpdatedAt = time.Now()

	if err := l.bannerRepo.Update(ctx, banner); err != nil {
		return nil, apperrors.NewInternalError("更新Banner失败: " + err.Error())
	}

	return &UpdateBannerResponse{
		Banner: banner,
	}, nil
}

// DeleteBannerRequest 删除Banner请求
type DeleteBannerRequest struct {
	ID uint64
}

// DeleteBannerResponse 删除Banner响应
type DeleteBannerResponse struct {
}

// DeleteBanner 删除Banner
func (l *ProductLogic) DeleteBanner(ctx context.Context, req *DeleteBannerRequest) (*DeleteBannerResponse, error) {
	if l.bannerRepo == nil {
		return nil, apperrors.NewInternalError("数据库连接未初始化")
	}

	if req.ID == 0 {
		return nil, apperrors.NewInvalidParamError("Banner ID不能为空")
	}

	// 检查Banner是否存在
	_, err := l.bannerRepo.GetByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodeNotFound, "Banner不存在")
		}
		return nil, apperrors.NewInternalError("查询Banner失败: " + err.Error())
	}

	// 删除Banner
	if err := l.bannerRepo.Delete(ctx, req.ID); err != nil {
		return nil, apperrors.NewInternalError("删除Banner失败: " + err.Error())
	}

	return &DeleteBannerResponse{}, nil
}
