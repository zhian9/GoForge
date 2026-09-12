SET NAMES utf8mb4;

-- ===== 管理员账号 =====
-- 账号: admin / 密码: admin123
INSERT INTO user (id, username, nickname, status, member_level, points, is_admin, created_at, updated_at) VALUES
(1, 'admin', '管理员', 1, 0, 0, 1, NOW(), NOW());

INSERT INTO credential (user_id, credential_type, credential_key, credential_value, extra, created_at, updated_at) VALUES
(1, 1, 'admin', '$2a$10$Tf2GYbDstDsrKX4WNv8BP.OpCqxeM6VVRvY7xLUkXDWhBVcnzeppO', '{}', NOW(), NOW());

-- ===== 分类数据 =====
INSERT INTO category (id, parent_id, name, level, sort, icon, image, status, created_at, updated_at) VALUES
(1, 0, '手机', 1, 1, '', 'https://picsum.photos/seed/phone/100/100', 1, NOW(), NOW()),
(2, 0, '电脑', 1, 2, '', 'https://picsum.photos/seed/laptop/100/100', 1, NOW(), NOW()),
(3, 0, '家电', 1, 3, '', 'https://picsum.photos/seed/homeapp/100/100', 1, NOW(), NOW()),
(4, 0, '服饰', 1, 4, '', 'https://picsum.photos/seed/cloth/100/100', 1, NOW(), NOW()),
(5, 0, '食品', 1, 5, '', 'https://picsum.photos/seed/food/100/100', 1, NOW(), NOW()),
(6, 0, '家居', 1, 6, '', 'https://picsum.photos/seed/home/100/100', 1, NOW(), NOW()),
(11, 1, '智能手机', 2, 1, '', 'https://picsum.photos/seed/smartphone/100/100', 1, NOW(), NOW()),
(12, 1, '功能手机', 2, 2, '', 'https://picsum.photos/seed/feature/100/100', 1, NOW(), NOW()),
(13, 1, '手机配件', 2, 3, '', 'https://picsum.photos/seed/acc/100/100', 1, NOW(), NOW()),
(21, 2, '笔记本', 2, 1, '', 'https://picsum.photos/seed/notebook/100/100', 1, NOW(), NOW()),
(22, 2, '台式机', 2, 2, '', 'https://picsum.photos/seed/desktop/100/100', 1, NOW(), NOW()),
(23, 2, '平板电脑', 2, 3, '', 'https://picsum.photos/seed/tablet/100/100', 1, NOW(), NOW());

-- ===== Banner 数据 =====
INSERT INTO banner (id, title, description, image, image_local, link, link_type, sort, status, start_time, end_time, created_at, updated_at) VALUES
(1, '618年中大促', '全场五折起', 'https://picsum.photos/seed/banner1/800/400', '', '/products', 2, 10, 1, '2026-01-01 00:00:00', '2026-12-31 23:59:59', NOW(), NOW()),
(2, '新品首发', '最新旗舰手机', 'https://picsum.photos/seed/banner2/800/400', '', '/products', 2, 9, 1, '2026-01-01 00:00:00', '2026-12-31 23:59:59', NOW(), NOW()),
(3, '限时秒杀', '每日10点开抢', 'https://picsum.photos/seed/banner3/800/400', '', '/seckill', 2, 8, 1, '2026-01-01 00:00:00', '2026-12-31 23:59:59', NOW(), NOW());

-- ===== 商品数据 =====
INSERT INTO product (id, spu_code, name, subtitle, category_id, main_image, local_main_image, images, detail, price, original_price, stock, sales, status, is_hot, sort, created_at, updated_at) VALUES
(1, 'SPU001', '小米 15 Pro', '骁龙8至尊版 徕卡光学', 11, 'https://picsum.photos/seed/product1/400/400', '', '["https://picsum.photos/seed/prod1a/400/400","https://picsum.photos/seed/prod1b/400/400"]', '<p>旗舰手机</p>', 4999.00, 5299.00, 1000, 500, 1, 1, 100, NOW(), NOW()),
(2, 'SPU002', '华为 Mate 70 Pro', '麒麟芯片 卫星通信', 11, 'https://picsum.photos/seed/product2/400/400', '', '["https://picsum.photos/seed/prod2a/400/400"]', '<p>华为旗舰</p>', 6999.00, 7499.00, 800, 1200, 1, 1, 99, NOW(), NOW()),
(3, 'SPU003', 'iPhone 16 Pro Max', 'A18 Pro芯片 钛金属', 11, 'https://picsum.photos/seed/product3/400/400', '', '["https://picsum.photos/seed/prod3a/400/400"]', '<p>苹果旗舰</p>', 8999.00, NULL, 500, 2300, 1, 1, 98, NOW(), NOW()),
(4, 'SPU004', 'MacBook Pro 16', 'M4 Max芯片 32GB', 21, 'https://picsum.photos/seed/product4/400/400', '', '["https://picsum.photos/seed/prod4a/400/400"]', '<p>专业笔记本</p>', 19999.00, NULL, 300, 450, 1, 1, 97, NOW(), NOW()),
(5, 'SPU005', 'ThinkPad X1 Carbon', 'Ultra 9 14英寸', 21, 'https://picsum.photos/seed/product5/400/400', '', '["https://picsum.photos/seed/prod5a/400/400"]', '<p>商务轻薄本</p>', 10999.00, 12999.00, 200, 180, 1, 1, 96, NOW(), NOW()),
(6, 'SPU006', 'iPad Pro M4', '11英寸 Liquid Retina', 23, 'https://picsum.photos/seed/product6/400/400', '', '["https://picsum.photos/seed/prod6a/400/400"]', '<p>最强平板</p>', 6799.00, NULL, 400, 800, 1, 1, 95, NOW(), NOW()),
(7, 'SPU007', '海尔冰箱 456L', '风冷无霜 一级能效', 3, 'https://picsum.photos/seed/product7/400/400', '', '["https://picsum.photos/seed/prod7a/400/400"]', '<p>大容量冰箱</p>', 3299.00, 3999.00, 600, 320, 1, 0, 50, NOW(), NOW()),
(8, 'SPU008', '格力空调 1.5匹', '新一级能效 变频冷暖', 3, 'https://picsum.photos/seed/product8/400/400', '', '["https://picsum.photos/seed/prod8a/400/400"]', '<p>节能空调</p>', 2699.00, 2999.00, 500, 1500, 1, 1, 49, NOW(), NOW()),
(9, 'SPU009', '耐克 Air Max 270', '气垫运动鞋', 4, 'https://picsum.photos/seed/product9/400/400', '', '["https://picsum.photos/seed/prod9a/400/400"]', '<p>经典运动鞋</p>', 899.00, 1199.00, 2000, 5000, 1, 1, 48, NOW(), NOW()),
(10, 'SPU010', '三只松鼠坚果礼盒', '每日坚果 混合装', 5, 'https://picsum.photos/seed/product10/400/400', '', '["https://picsum.photos/seed/prod10a/400/400"]', '<p>健康零食</p>', 99.00, 139.00, 3000, 8000, 1, 1, 47, NOW(), NOW());

-- ===== SKU 数据 =====
INSERT INTO sku (id, product_id, sku_code, name, specs, price, original_price, stock, image, weight, volume, status, created_at, updated_at) VALUES
(1, 1, 'SKU001-01', '小米15 Pro 12+256 白色', '{"颜色":"白色","内存":"12+256"}', 4999.00, 5299.00, 300, 'https://picsum.photos/seed/sku1/300/300', NULL, NULL, 1, NOW(), NOW()),
(2, 1, 'SKU001-02', '小米15 Pro 16+512 黑色', '{"颜色":"黑色","内存":"16+512"}', 5499.00, 5799.00, 200, 'https://picsum.photos/seed/sku2/300/300', NULL, NULL, 1, NOW(), NOW()),
(3, 2, 'SKU002-01', '华为Mate70 Pro 12+512 银', '{"颜色":"银色","内存":"12+512"}', 6999.00, 7499.00, 400, 'https://picsum.photos/seed/sku3/300/300', NULL, NULL, 1, NOW(), NOW()),
(4, 3, 'SKU003-01', 'iPhone16 Pro Max 256GB 钛色', '{"颜色":"钛色","存储":"256GB"}', 8999.00, NULL, 200, 'https://picsum.photos/seed/sku4/300/300', NULL, NULL, 1, NOW(), NOW()),
(5, 4, 'SKU004-01', 'MacBook Pro 16 M4 Max 32GB', '{"芯片":"M4 Max","内存":"32GB"}', 19999.00, NULL, 100, 'https://picsum.photos/seed/sku5/300/300', NULL, NULL, 1, NOW(), NOW()),
(6, 7, 'SKU007-01', '海尔456L 白色', '{"颜色":"白色"}', 3299.00, 3999.00, 300, 'https://picsum.photos/seed/sku7/300/300', NULL, NULL, 1, NOW(), NOW()),
(7, 8, 'SKU008-01', '格力1.5匹 变频', '{"型号":"1.5匹"}', 2699.00, 2999.00, 200, 'https://picsum.photos/seed/sku8/300/300', NULL, NULL, 1, NOW(), NOW()),
(8, 9, 'SKU009-01', 'AirMax270 42码', '{"尺码":"42"}', 899.00, 1199.00, 500, 'https://picsum.photos/seed/sku9/300/300', NULL, NULL, 1, NOW(), NOW()),
(9, 10, 'SKU010-01', '三只松鼠礼盒 750g', '{"规格":"750g"}', 99.00, 139.00, 1000, 'https://picsum.photos/seed/sku10/300/300', NULL, NULL, 1, NOW(), NOW()),
(10, 5, 'SKU005-01', 'ThinkPad X1 16GB+512GB', '{"内存":"16GB","存储":"512GB"}', 10999.00, 12999.00, 80, 'https://picsum.photos/seed/sku6/300/300', NULL, NULL, 1, NOW(), NOW()),
(11, 6, 'SKU006-01', 'iPad Pro M4 256GB WiFi', '{"存储":"256GB","网络":"WiFi"}', 6799.00, NULL, 150, 'https://picsum.photos/seed/sku11/300/300', NULL, NULL, 1, NOW(), NOW());

-- ===== 库存数据 =====
INSERT INTO inventory (sku_id, total_stock, available_stock, locked_stock, sold_stock, low_stock_threshold, created_at, updated_at) VALUES
(1, 300, 300, 0, 0, 10, NOW(), NOW()),
(2, 200, 200, 0, 0, 10, NOW(), NOW()),
(3, 400, 400, 0, 0, 10, NOW(), NOW()),
(4, 200, 200, 0, 0, 10, NOW(), NOW()),
(5, 100, 100, 0, 0, 10, NOW(), NOW()),
(6, 300, 300, 0, 0, 10, NOW(), NOW()),
(7, 200, 200, 0, 0, 10, NOW(), NOW()),
(8, 500, 500, 0, 0, 10, NOW(), NOW()),
(9, 1000, 1000, 0, 0, 10, NOW(), NOW()),
(10, 80, 80, 0, 0, 5, NOW(), NOW()),
(11, 150, 150, 0, 0, 5, NOW(), NOW());

-- ===== 秒杀活动数据 =====
-- status: 0-禁用, 1-启用；运行状态由 start_time/end_time 与当前时间动态计算
INSERT INTO seckill_activity (name, sku_id, seckill_price, stock, start_time, end_time, status, created_at, updated_at) VALUES
('限时秒杀：小米15 Pro 12+256', 1, 3999.00, 100, UNIX_TIMESTAMP() - 3600, UNIX_TIMESTAMP() + 86400, 1, NOW(), NOW()),
('限时秒杀：华为 Mate 70 Pro', 2, 5999.00, 80, UNIX_TIMESTAMP() - 3600, UNIX_TIMESTAMP() + 86400, 1, NOW(), NOW()),
('限时秒杀：海尔冰箱 456L', 7, 2699.00, 50, UNIX_TIMESTAMP() - 1800, UNIX_TIMESTAMP() + 43200, 1, NOW(), NOW()),
('即将开抢：三只松鼠坚果礼盒', 9, 79.00, 200, UNIX_TIMESTAMP() + 3600, UNIX_TIMESTAMP() + 172800, 1, NOW(), NOW());
