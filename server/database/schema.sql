SET NAMES utf8mb4;

-- ============================================
-- Go 微服务电商项目数据库模型
-- 数据库: MySQL 8.0
-- 字符集: utf8mb4
-- 排序规则: utf8mb4_unicode_ci
-- ============================================

-- ============================================
-- 一、用户域服务 (user-service)
-- ============================================

-- 用户表
CREATE TABLE `user` (
                        `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
                        `username` VARCHAR(50) NOT NULL COMMENT '用户名',
                        `nickname` VARCHAR(50) DEFAULT NULL COMMENT '昵称',
                        `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号',
                        `email` VARCHAR(100) DEFAULT NULL COMMENT '邮箱',
                        `avatar` VARCHAR(255) DEFAULT NULL COMMENT '头像URL',
                        `gender` TINYINT DEFAULT 0 COMMENT '性别: 0-未知, 1-男, 2-女',
                        `birthday` DATE DEFAULT NULL COMMENT '生日',
                        `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-正常',
                        `member_level` TINYINT DEFAULT 0 COMMENT '会员等级: 0-普通, 1-VIP1, 2-VIP2, 3-VIP3',
                        `points` INT DEFAULT 0 COMMENT '积分',
                        `balance` DECIMAL(10, 2) DEFAULT 0 COMMENT '余额（系统内货币，单位元）',
                        `is_admin` TINYINT DEFAULT 0 COMMENT '0-普通用户 1-管理员',
                        `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                        `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                        `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `uk_username` (`username`),
                        UNIQUE KEY `uk_phone` (`phone`),
                        UNIQUE KEY `uk_email` (`email`),
                        KEY `idx_status` (`status`),
                        KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 用户地址表
CREATE TABLE `address` (
                           `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '地址ID',
                           `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                           `receiver_name` VARCHAR(50) NOT NULL COMMENT '收货人姓名',
                           `receiver_phone` VARCHAR(20) NOT NULL COMMENT '收货人电话',
                           `province` VARCHAR(50) NOT NULL COMMENT '省份',
                           `city` VARCHAR(50) NOT NULL COMMENT '城市',
                           `district` VARCHAR(50) NOT NULL COMMENT '区县',
                           `detail` VARCHAR(200) NOT NULL COMMENT '详细地址',
                           `postal_code` VARCHAR(10) DEFAULT NULL COMMENT '邮编',
                           `is_default` TINYINT DEFAULT 0 COMMENT '是否默认: 0-否, 1-是',
                           `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                           `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                           `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
                           PRIMARY KEY (`id`),
                           KEY `idx_user_id` (`user_id`),
                           KEY `idx_is_default` (`is_default`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户地址表';

-- 用户凭证表（密码、第三方登录等）
CREATE TABLE `credential` (
                              `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '凭证ID',
                              `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                              `credential_type` TINYINT NOT NULL COMMENT '凭证类型: 1-密码, 2-微信, 3-支付宝, 4-QQ',
                              `credential_key` VARCHAR(100) NOT NULL COMMENT '凭证标识（手机号/邮箱/第三方openid）',
                              `credential_value` VARCHAR(255) DEFAULT NULL COMMENT '凭证值（加密后的密码）',
                              `extra` JSON DEFAULT NULL COMMENT '扩展信息（第三方用户信息等）',
                              `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                              `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                              PRIMARY KEY (`id`),
                              UNIQUE KEY `uk_user_type_key` (`user_id`, `credential_type`, `credential_key`),
                              KEY `idx_credential_key` (`credential_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户凭证表';

-- ============================================
-- 二、商品域服务 (product-service)
-- ============================================

-- 商品类目表
CREATE TABLE `category` (
                            `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '类目ID',
                            `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '父类目ID，0表示顶级类目',
                            `name` VARCHAR(100) NOT NULL COMMENT '类目名称',
                            `level` TINYINT NOT NULL COMMENT '类目层级: 1-一级, 2-二级, 3-三级',
                            `sort` INT DEFAULT 0 COMMENT '排序值，越大越靠前',
                            `icon` VARCHAR(255) DEFAULT NULL COMMENT '类目图标',
                            `icon_local` VARCHAR(255) DEFAULT NULL COMMENT '类目本地图标路径',
                            `image` VARCHAR(255) DEFAULT NULL COMMENT '类目图片',
                            `image_local` VARCHAR(255) DEFAULT NULL COMMENT '类目本地图片路径',
                            `description` TEXT COMMENT '类目描述',
                            `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
                            `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                            `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                            PRIMARY KEY (`id`),
                            KEY `idx_parent_id` (`parent_id`),
                            KEY `idx_level` (`level`),
                            KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品类目表';

-- 商品表（SPU）
CREATE TABLE `product` (
                           `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '商品ID',
                           `spu_code` VARCHAR(50) NOT NULL COMMENT 'SPU编码',
                           `name` VARCHAR(200) NOT NULL COMMENT '商品名称',
                           `subtitle` VARCHAR(200) DEFAULT NULL COMMENT '副标题',
                           `category_id` BIGINT UNSIGNED NOT NULL COMMENT '类目ID',
                           `brand_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '品牌ID',
                           `main_image` VARCHAR(255) NOT NULL COMMENT '主图',
                           `local_main_image` VARCHAR(255) DEFAULT NULL COMMENT '本地主图路径',
                           `images` JSON DEFAULT NULL COMMENT '商品图片列表',
                           `local_images` JSON DEFAULT NULL COMMENT '本地图片路径列表',
                           `detail` TEXT COMMENT '商品详情',
                           `price` DECIMAL(10, 2) NOT NULL COMMENT '商品价格（最低SKU价格）',
                           `original_price` DECIMAL(10, 2) DEFAULT NULL COMMENT '原价',
                           `stock` INT DEFAULT 0 COMMENT '总库存（所有SKU库存之和）',
                           `sales` INT DEFAULT 0 COMMENT '销量',
                           `status` TINYINT DEFAULT 1 COMMENT '状态: 0-下架, 1-上架, 2-待审核',
                           `is_hot` TINYINT DEFAULT 0 COMMENT '是否热门: 0-否, 1-是',
                           `sort` INT DEFAULT 0 COMMENT '排序值',
                           `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                           `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                           `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
                           PRIMARY KEY (`id`),
                           UNIQUE KEY `uk_spu_code` (`spu_code`),
                           KEY `idx_category_id` (`category_id`),
                           KEY `idx_brand_id` (`brand_id`),
                           KEY `idx_status` (`status`),
                           KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品表(SPU)';

-- Banner表（横幅广告）
CREATE TABLE `banner` (
                           `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Banner ID',
                           `title` VARCHAR(200) DEFAULT NULL COMMENT '标题',
                           `description` VARCHAR(500) DEFAULT NULL COMMENT '描述',
                           `image` VARCHAR(255) NOT NULL COMMENT '封面图片URL',
                           `image_local` VARCHAR(255) DEFAULT NULL COMMENT '本地图片路径',
                           `link` VARCHAR(500) DEFAULT NULL COMMENT '跳转链接',
                           `link_type` TINYINT DEFAULT 1 COMMENT '1-商品详情, 2-分类页面, 3-外部链接, 4-无链接',
                           `sort` INT DEFAULT 0 COMMENT '排序值',
                           `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
                           `start_time` DATETIME DEFAULT NULL COMMENT '开始时间',
                           `end_time` DATETIME DEFAULT NULL COMMENT '结束时间',
                           `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                           `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                           PRIMARY KEY (`id`),
                           KEY `idx_sort` (`sort`),
                           KEY `idx_status` (`status`),
                           KEY `idx_start_time` (`start_time`),
                           KEY `idx_end_time` (`end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Banner表';

-- SKU表
CREATE TABLE `sku` (
                       `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'SKU ID',
                       `product_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
                       `sku_code` VARCHAR(50) NOT NULL COMMENT 'SKU编码',
                       `name` VARCHAR(200) NOT NULL COMMENT 'SKU名称',
                       `specs` JSON NOT NULL COMMENT '规格属性（如：{"颜色":"红色","尺寸":"L"}）',
                       `price` DECIMAL(10, 2) NOT NULL COMMENT '价格',
                       `original_price` DECIMAL(10, 2) DEFAULT NULL COMMENT '原价',
                       `stock` INT DEFAULT 0 COMMENT '库存（基础库存，实时库存由库存服务管理）',
                       `image` VARCHAR(255) DEFAULT NULL COMMENT 'SKU图片',
                       `weight` DECIMAL(8, 2) DEFAULT NULL COMMENT '重量(kg)',
                       `volume` DECIMAL(8, 2) DEFAULT NULL COMMENT '体积(立方米)',
                       `status` TINYINT DEFAULT 1 COMMENT '状态: 0-下架, 1-上架',
                       `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                       `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                       `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
                       PRIMARY KEY (`id`),
                       UNIQUE KEY `uk_sku_code` (`sku_code`),
                       KEY `idx_product_id` (`product_id`),
                       KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='SKU表';

-- 商品属性表
CREATE TABLE `attr` (
                        `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '属性ID',
                        `category_id` BIGINT UNSIGNED NOT NULL COMMENT '类目ID',
                        `name` VARCHAR(50) NOT NULL COMMENT '属性名称',
                        `type` TINYINT NOT NULL COMMENT '属性类型: 1-规格属性, 2-销售属性, 3-基础属性',
                        `input_type` TINYINT NOT NULL COMMENT '输入类型: 1-单选, 2-多选, 3-文本输入',
                        `values` JSON DEFAULT NULL COMMENT '属性可选值列表',
                        `sort` INT DEFAULT 0 COMMENT '排序值',
                        `is_required` TINYINT DEFAULT 0 COMMENT '是否必填: 0-否, 1-是',
                        `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                        `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                        PRIMARY KEY (`id`),
                        KEY `idx_category_id` (`category_id`),
                        KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品属性表';

-- ============================================
-- 三、库存服务 (inventory-service)
-- ============================================

-- 库存表
CREATE TABLE `inventory` (
                             `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '库存ID',
                             `sku_id` BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
                             `total_stock` INT NOT NULL DEFAULT 0 COMMENT '总库存',
                             `available_stock` INT NOT NULL DEFAULT 0 COMMENT '可用库存',
                             `locked_stock` INT NOT NULL DEFAULT 0 COMMENT '锁定库存（预占）',
                             `sold_stock` INT NOT NULL DEFAULT 0 COMMENT '已售库存',
                             `low_stock_threshold` INT DEFAULT 10 COMMENT '低库存预警阈值',
                             `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                             `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `uk_sku_id` (`sku_id`),
                             KEY `idx_available_stock` (`available_stock`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='库存表';

-- 库存流水表
CREATE TABLE `inventory_log` (
                                 `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流水ID',
                                 `sku_id` BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
                                 `order_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '订单ID',
                                 `type` TINYINT NOT NULL COMMENT '操作类型: 1-入库, 2-出库, 3-锁定, 4-解锁, 5-扣减, 6-回退',
                                 `quantity` INT NOT NULL COMMENT '数量（正数表示增加，负数表示减少）',
                                 `before_stock` INT NOT NULL COMMENT '操作前库存',
                                 `after_stock` INT NOT NULL COMMENT '操作后库存',
                                 `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
                                 `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                 PRIMARY KEY (`id`),
                                 KEY `idx_sku_id` (`sku_id`),
                                 KEY `idx_order_id` (`order_id`),
                                 KEY `idx_type` (`type`),
                                 KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='库存流水表';

-- ============================================
-- 四、订单服务 (order-service)
-- ============================================

-- 订单表
CREATE TABLE `orders` (
                          `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '订单ID',
                          `order_no` VARCHAR(32) NOT NULL COMMENT '订单号',
                          `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                          `order_type` TINYINT DEFAULT 1 COMMENT '订单类型: 1-普通订单, 2-秒杀订单, 3-拼团订单',
                          `status` TINYINT NOT NULL COMMENT '订单状态: 0-已取消, 1-待支付, 2-待发货, 3-待收货, 4-已完成, 5-已退款',
                          `total_amount` DECIMAL(10, 2) NOT NULL COMMENT '订单总金额',
                          `pay_amount` DECIMAL(10, 2) NOT NULL COMMENT '实付金额',
                          `discount_amount` DECIMAL(10, 2) DEFAULT 0 COMMENT '优惠金额',
                          `freight_amount` DECIMAL(10, 2) DEFAULT 0 COMMENT '运费',
                          `receiver_name` VARCHAR(50) NOT NULL COMMENT '收货人姓名',
                          `receiver_phone` VARCHAR(20) NOT NULL COMMENT '收货人电话',
                          `receiver_address` VARCHAR(500) NOT NULL COMMENT '收货地址',
                          `payment_method` TINYINT DEFAULT NULL COMMENT '支付方式: 1-微信, 2-支付宝, 3-银联',
                          `payment_time` DATETIME DEFAULT NULL COMMENT '支付时间',
                          `delivery_time` DATETIME DEFAULT NULL COMMENT '发货时间',
                          `receive_time` DATETIME DEFAULT NULL COMMENT '收货时间',
                          `cancel_time` DATETIME DEFAULT NULL COMMENT '取消时间',
                          `cancel_reason` VARCHAR(255) DEFAULT NULL COMMENT '取消原因',
                          `remark` VARCHAR(500) DEFAULT NULL COMMENT '订单备注',
                          `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                          `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                          PRIMARY KEY (`id`),
                          UNIQUE KEY `uk_order_no` (`order_no`),
                          KEY `idx_user_id` (`user_id`),
                          KEY `idx_status` (`status`),
                          KEY `idx_created_at` (`created_at`),
                          KEY `idx_payment_time` (`payment_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- 订单商品项表
CREATE TABLE `order_item` (
                              `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '订单项ID',
                              `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
                              `order_no` VARCHAR(32) NOT NULL COMMENT '订单号',
                              `product_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
                              `product_name` VARCHAR(200) NOT NULL COMMENT '商品名称',
                              `sku_id` BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
                              `sku_code` VARCHAR(50) NOT NULL COMMENT 'SKU编码',
                              `sku_name` VARCHAR(200) NOT NULL COMMENT 'SKU名称',
                              `sku_image` VARCHAR(255) DEFAULT NULL COMMENT 'SKU图片',
                              `sku_specs` JSON DEFAULT NULL COMMENT 'SKU规格',
                              `price` DECIMAL(10, 2) NOT NULL COMMENT '单价',
                              `quantity` INT NOT NULL COMMENT '数量',
                              `total_amount` DECIMAL(10, 2) NOT NULL COMMENT '小计金额',
                              `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                              PRIMARY KEY (`id`),
                              KEY `idx_order_id` (`order_id`),
                              KEY `idx_order_no` (`order_no`),
                              KEY `idx_product_id` (`product_id`),
                              KEY `idx_sku_id` (`sku_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单商品项表';

-- 订单操作日志表
CREATE TABLE `order_log` (
                             `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '日志ID',
                             `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
                             `order_no` VARCHAR(32) NOT NULL COMMENT '订单号',
                             `operator_type` TINYINT NOT NULL COMMENT '操作人类型: 1-用户, 2-系统, 3-管理员',
                             `operator_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '操作人ID',
                             `action` VARCHAR(50) NOT NULL COMMENT '操作动作',
                             `before_status` TINYINT DEFAULT NULL COMMENT '操作前状态',
                             `after_status` TINYINT DEFAULT NULL COMMENT '操作后状态',
                             `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
                             `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                             PRIMARY KEY (`id`),
                             KEY `idx_order_id` (`order_id`),
                             KEY `idx_order_no` (`order_no`),
                             KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单操作日志表';

-- ============================================
-- 五、支付服务 (payment-service)
-- ============================================

-- 支付单表
CREATE TABLE `payment` (
                           `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '支付单ID',
                           `payment_no` VARCHAR(32) NOT NULL COMMENT '支付单号',
                           `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
                           `order_no` VARCHAR(32) NOT NULL COMMENT '订单号',
                           `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                           `amount` DECIMAL(10, 2) NOT NULL COMMENT '支付金额',
                           `payment_method` TINYINT NOT NULL COMMENT '支付方式: 1-微信, 2-支付宝, 3-银联',
                           `status` TINYINT NOT NULL COMMENT '支付状态: 0-待支付, 1-支付成功, 2-支付失败, 3-已退款',
                           `third_party_no` VARCHAR(100) DEFAULT NULL COMMENT '第三方支付单号',
                           `third_party_response` JSON DEFAULT NULL COMMENT '第三方支付响应信息',
                           `paid_at` DATETIME DEFAULT NULL COMMENT '支付时间',
                           `expire_at` DATETIME DEFAULT NULL COMMENT '支付过期时间',
                           `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                           `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                           PRIMARY KEY (`id`),
                           UNIQUE KEY `uk_payment_no` (`payment_no`),
                           KEY `idx_order_id` (`order_id`),
                           KEY `idx_order_no` (`order_no`),
                           KEY `idx_user_id` (`user_id`),
                           KEY `idx_status` (`status`),
                           KEY `idx_third_party_no` (`third_party_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付单表';

-- 支付流水表
CREATE TABLE `payment_log` (
                               `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流水ID',
                               `payment_id` BIGINT UNSIGNED NOT NULL COMMENT '支付单ID',
                               `payment_no` VARCHAR(32) NOT NULL COMMENT '支付单号',
                               `action` VARCHAR(50) NOT NULL COMMENT '操作动作: create, pay, refund, cancel',
                               `amount` DECIMAL(10, 2) NOT NULL COMMENT '金额',
                               `before_status` TINYINT DEFAULT NULL COMMENT '操作前状态',
                               `after_status` TINYINT DEFAULT NULL COMMENT '操作后状态',
                               `request_data` JSON DEFAULT NULL COMMENT '请求数据',
                               `response_data` JSON DEFAULT NULL COMMENT '响应数据',
                               `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注',
                               `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                               PRIMARY KEY (`id`),
                               KEY `idx_payment_id` (`payment_id`),
                               KEY `idx_payment_no` (`payment_no`),
                               KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付流水表';

-- ============================================
-- 六、营销服务 (promotion-service)
-- ============================================

-- 优惠券表
CREATE TABLE `coupon` (
                          `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '优惠券ID',
                          `name` VARCHAR(100) NOT NULL COMMENT '优惠券名称',
                          `type` TINYINT NOT NULL COMMENT '优惠券类型: 1-满减券, 2-折扣券, 3-免运费券',
                          `discount_type` TINYINT NOT NULL COMMENT '优惠类型: 1-固定金额, 2-百分比折扣',
                          `discount_value` DECIMAL(10, 2) NOT NULL COMMENT '优惠值',
                          `min_amount` DECIMAL(10, 2) DEFAULT 0 COMMENT '最低使用金额',
                          `max_discount` DECIMAL(10, 2) DEFAULT NULL COMMENT '最大优惠金额（折扣券使用）',
                          `total_count` INT DEFAULT -1 COMMENT '发放总数，-1表示不限',
                          `used_count` INT DEFAULT 0 COMMENT '已使用数量',
                          `per_user_limit` INT DEFAULT 1 COMMENT '每人限领数量',
                          `valid_start_time` DATETIME NOT NULL COMMENT '有效期开始时间',
                          `valid_end_time` DATETIME NOT NULL COMMENT '有效期结束时间',
                          `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
                          `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                          `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                          PRIMARY KEY (`id`),
                          KEY `idx_status` (`status`),
                          KEY `idx_valid_time` (`valid_start_time`, `valid_end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='优惠券表';

-- 用户优惠券表
CREATE TABLE `user_coupon` (
                               `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
                               `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                               `coupon_id` BIGINT UNSIGNED NOT NULL COMMENT '优惠券ID',
                               `status` TINYINT DEFAULT 0 COMMENT '状态: 0-未使用, 1-已使用, 2-已过期',
                               `order_id` BIGINT UNSIGNED DEFAULT NULL COMMENT '使用订单ID',
                               `used_at` DATETIME DEFAULT NULL COMMENT '使用时间',
                               `expire_at` DATETIME NOT NULL COMMENT '过期时间',
                               `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间',
                               PRIMARY KEY (`id`),
                               KEY `idx_user_id` (`user_id`),
                               KEY `idx_coupon_id` (`coupon_id`),
                               KEY `idx_status` (`status`),
                               KEY `idx_expire_at` (`expire_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户优惠券表';

-- 促销活动表
CREATE TABLE `promotion` (
                             `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '活动ID',
                             `name` VARCHAR(100) NOT NULL COMMENT '活动名称',
                             `type` TINYINT NOT NULL COMMENT '活动类型: 1-满减, 2-折扣, 3-秒杀, 4-拼团',
                             `rule` JSON NOT NULL COMMENT '活动规则（JSON格式）',
                             `product_ids` JSON DEFAULT NULL COMMENT '参与商品ID列表',
                             `category_ids` JSON DEFAULT NULL COMMENT '参与类目ID列表',
                             `start_time` DATETIME NOT NULL COMMENT '开始时间',
                             `end_time` DATETIME NOT NULL COMMENT '结束时间',
                             `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
                             `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                             `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                             PRIMARY KEY (`id`),
                             KEY `idx_type` (`type`),
                             KEY `idx_status` (`status`),
                             KEY `idx_time` (`start_time`, `end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='促销活动表';

-- 积分表
CREATE TABLE `points` (
                          `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '积分账户ID',
                          `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                          `total` BIGINT NOT NULL DEFAULT 0 COMMENT '总积分',
                          `used` BIGINT NOT NULL DEFAULT 0 COMMENT '已用积分',
                          `available` BIGINT NOT NULL DEFAULT 0 COMMENT '可用积分',
                          `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                          `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                          PRIMARY KEY (`id`),
                          UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分账户表';

-- 余额流水表
CREATE TABLE `balance_log` (
                              `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '流水ID',
                              `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                              `order_no` VARCHAR(32) DEFAULT NULL COMMENT '关联订单号',
                              `type` TINYINT NOT NULL COMMENT '类型: 1-充值, 2-支付, 3-退款, 4-赠送',
                              `amount` DECIMAL(10, 2) NOT NULL COMMENT '变动金额（正增负减）',
                              `before_balance` DECIMAL(10, 2) NOT NULL COMMENT '变动前余额',
                              `after_balance` DECIMAL(10, 2) NOT NULL COMMENT '变动后余额',
                              `remark` VARCHAR(255) DEFAULT NULL COMMENT '备注',
                              `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                              PRIMARY KEY (`id`),
                              KEY `idx_user_id` (`user_id`),
                              KEY `idx_order_no` (`order_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='余额流水表';

-- ============================================
-- 七、评价服务 (review-service)
-- ============================================

-- 评价表
CREATE TABLE `review` (
                          `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评价ID',
                          `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                          `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
                          `order_item_id` BIGINT UNSIGNED NOT NULL COMMENT '订单项ID',
                          `product_id` BIGINT UNSIGNED NOT NULL COMMENT '商品ID',
                          `sku_id` BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
                          `rating` TINYINT NOT NULL COMMENT '评分: 1-5星',
                          `content` TEXT COMMENT '评价内容',
                          `images` JSON DEFAULT NULL COMMENT '评价图片列表',
                          `videos` JSON DEFAULT NULL COMMENT '评价视频列表',
                          `status` TINYINT DEFAULT 1 COMMENT '状态: 0-隐藏, 1-显示',
                          `reply_content` TEXT DEFAULT NULL COMMENT '商家回复内容',
                          `reply_time` DATETIME DEFAULT NULL COMMENT '商家回复时间',
                          `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                          `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                          PRIMARY KEY (`id`),
                          KEY `idx_user_id` (`user_id`),
                          KEY `idx_order_id` (`order_id`),
                          KEY `idx_product_id` (`product_id`),
                          KEY `idx_sku_id` (`sku_id`),
                          KEY `idx_rating` (`rating`),
                          KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评价表';

-- 评价回复表（用户对评价的回复）
CREATE TABLE `review_reply` (
                                `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '回复ID',
                                `review_id` BIGINT UNSIGNED NOT NULL COMMENT '评价ID',
                                `user_id` BIGINT UNSIGNED NOT NULL COMMENT '回复用户ID',
                                `content` TEXT NOT NULL COMMENT '回复内容',
                                `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '父回复ID，0表示直接回复评价',
                                `status` TINYINT DEFAULT 1 COMMENT '状态: 0-隐藏, 1-显示',
                                `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                PRIMARY KEY (`id`),
                                KEY `idx_review_id` (`review_id`),
                                KEY `idx_user_id` (`user_id`),
                                KEY `idx_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评价回复表';

-- ============================================
-- 八、物流服务 (logistics-service)
-- ============================================

-- 物流信息表
CREATE TABLE `logistics` (
                             `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '物流ID',
                             `order_id` BIGINT UNSIGNED NOT NULL COMMENT '订单ID',
                             `order_no` VARCHAR(32) NOT NULL COMMENT '订单号',
                             `logistics_company` VARCHAR(50) NOT NULL COMMENT '物流公司',
                             `logistics_no` VARCHAR(50) NOT NULL COMMENT '物流单号',
                             `receiver_name` VARCHAR(50) NOT NULL COMMENT '收货人姓名',
                             `receiver_phone` VARCHAR(20) NOT NULL COMMENT '收货人电话',
                             `receiver_address` VARCHAR(500) NOT NULL COMMENT '收货地址',
                             `sender_name` VARCHAR(50) DEFAULT NULL COMMENT '发货人姓名',
                             `sender_phone` VARCHAR(20) DEFAULT NULL COMMENT '发货人电话',
                             `sender_address` VARCHAR(500) DEFAULT NULL COMMENT '发货地址',
                             `status` TINYINT DEFAULT 0 COMMENT '物流状态: 0-待发货, 1-已发货, 2-运输中, 3-已送达, 4-异常',
                             `current_location` VARCHAR(200) DEFAULT NULL COMMENT '当前位置',
                             `tracking_info` JSON DEFAULT NULL COMMENT '物流跟踪信息',
                             `shipped_at` DATETIME DEFAULT NULL COMMENT '发货时间',
                             `delivered_at` DATETIME DEFAULT NULL COMMENT '送达时间',
                             `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                             `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                             PRIMARY KEY (`id`),
                             KEY `idx_order_id` (`order_id`),
                             KEY `idx_order_no` (`order_no`),
                             KEY `idx_logistics_no` (`logistics_no`),
                             KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='物流信息表';

-- ============================================
-- 九、Outbox（事务消息表，用于异步同步到 ES 等）
-- ============================================

CREATE TABLE `outbox_event` (
                                `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '事件ID',
                                `aggregate_type` VARCHAR(32) NOT NULL COMMENT '聚合类型，如 product',
                                `aggregate_id` VARCHAR(64) NOT NULL COMMENT '聚合ID，如 product_id',
                                `event_type` VARCHAR(64) NOT NULL COMMENT '事件类型，如 product.upserted',
                                `payload` JSON DEFAULT NULL COMMENT '事件负载（可选）',
                                `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0-待投递, 1-已投递, 2-投递失败',
                                `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
                                `last_error` VARCHAR(255) DEFAULT NULL COMMENT '最后一次错误',
                                `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                `sent_at` DATETIME DEFAULT NULL COMMENT '投递时间',
                                PRIMARY KEY (`id`),
                                KEY `idx_status_id` (`status`, `id`),
                                KEY `idx_aggregate` (`aggregate_type`, `aggregate_id`),
                                KEY `idx_event_type` (`event_type`),
                                KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='事务消息Outbox表';

-- ============================================
-- 九、消息服务 (message-service)
-- ============================================

-- 消息表
CREATE TABLE `message` (
                           `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息ID',
                           `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                           `type` TINYINT NOT NULL COMMENT '消息类型: 1-系统通知, 2-订单消息, 3-营销消息, 4-物流消息',
                           `title` VARCHAR(200) NOT NULL COMMENT '消息标题',
                           `content` TEXT NOT NULL COMMENT '消息内容',
                           `link` VARCHAR(500) DEFAULT NULL COMMENT '跳转链接',
                           `is_read` TINYINT DEFAULT 0 COMMENT '是否已读: 0-未读, 1-已读',
                           `read_at` DATETIME DEFAULT NULL COMMENT '阅读时间',
                           `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                           PRIMARY KEY (`id`),
                           KEY `idx_user_id` (`user_id`),
                           KEY `idx_type` (`type`),
                           KEY `idx_is_read` (`is_read`),
                           KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表';

-- ============================================
-- 十、购物车服务 (cart-service)
-- 注意：购物车主要存储在Redis中，此表用于持久化备份
-- ============================================

-- 购物车表
CREATE TABLE `cart` (
                        `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '购物车ID',
                        `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
                        `sku_id` BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
                        `product_name` VARCHAR(200) DEFAULT NULL COMMENT '商品名称（冗余）',
                        `price` DECIMAL(10, 2) DEFAULT NULL COMMENT '加入时单价',
                        `product_image` VARCHAR(500) DEFAULT NULL COMMENT '商品图片（冗余）',
                        `quantity` INT NOT NULL DEFAULT 1 COMMENT '数量',
                        `is_selected` TINYINT DEFAULT 1 COMMENT '是否选中: 0-未选中, 1-选中',
                        `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                        `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                        `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `uk_user_sku` (`user_id`, `sku_id`),
                        KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='购物车表';

-- ============================================
-- 十一、秒杀服务 (seckill-service)
-- ============================================

-- 秒杀活动表
CREATE TABLE `seckill_activity` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Activity ID',
    `name` VARCHAR(200) NOT NULL COMMENT 'Activity name',
    `sku_id` BIGINT UNSIGNED NOT NULL COMMENT 'SKU ID',
    `seckill_price` DECIMAL(10,2) NOT NULL COMMENT 'Seckill price',
    `stock` INT NOT NULL DEFAULT 0 COMMENT 'Initial stock',
    `start_time` BIGINT NOT NULL COMMENT 'Start unix timestamp',
    `end_time` BIGINT NOT NULL COMMENT 'End unix timestamp',
    `status` TINYINT DEFAULT 1 COMMENT '0-disabled, 1-enabled',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_sku_id` (`sku_id`),
    KEY `idx_time` (`start_time`, `end_time`),
    KEY `idx_status` (`status`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Seckill activity';

-- ============================================
-- 十二、其他辅助表
-- ============================================

-- 品牌表
CREATE TABLE `brand` (
                         `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '品牌ID',
                         `name` VARCHAR(100) NOT NULL COMMENT '品牌名称',
                         `logo` VARCHAR(255) DEFAULT NULL COMMENT '品牌Logo',
                         `description` TEXT COMMENT '品牌描述',
                         `sort` INT DEFAULT 0 COMMENT '排序值',
                         `status` TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
                         `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                         `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                         PRIMARY KEY (`id`),
                         UNIQUE KEY `uk_name` (`name`),
                         KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='品牌表';

-- 系统配置表
CREATE TABLE `system_config` (
                                 `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '配置ID',
                                 `config_key` VARCHAR(100) NOT NULL COMMENT '配置键',
                                 `config_value` TEXT COMMENT '配置值',
                                 `config_type` VARCHAR(50) DEFAULT 'string' COMMENT '配置类型: string, number, json, boolean',
                                 `description` VARCHAR(255) DEFAULT NULL COMMENT '配置描述',
                                 `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                 `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                 PRIMARY KEY (`id`),
                                 UNIQUE KEY `uk_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- ============================================
-- 索引优化说明
-- ============================================
-- 1. 所有表都包含 created_at 和 updated_at 字段用于审计
-- 2. 软删除表使用 deleted_at 字段
-- 3. 外键关系通过应用层维护，不使用数据库外键约束（提高性能）
-- 4. 订单表、用户表等大表考虑分库分表策略
-- 5. 根据实际查询场景调整索引
-- ============================================
