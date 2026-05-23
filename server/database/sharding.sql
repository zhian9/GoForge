-- ============================================
-- 分库分表脚本示例
-- ============================================
-- 说明：此脚本展示如何创建分表，实际使用时需要根据业务需求调整
-- ============================================

-- ============================================
-- 一、用户表分表（按 user_id % 16 分表）
-- ============================================

-- 创建用户分表 user_0 到 user_15
DELIMITER $$

DROP PROCEDURE IF EXISTS create_user_sharding_tables$$

CREATE PROCEDURE create_user_sharding_tables()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 16 DO
        SET @sql = CONCAT('
            CREATE TABLE `user_', i, '` LIKE `user`;
            ALTER TABLE `user_', i, '` COMMENT = ''用户表分表_', i, ''';
        ');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET i = i + 1;
END WHILE;
END$$

DELIMITER ;

-- 执行创建分表
CALL create_user_sharding_tables();

-- 删除存储过程
DROP PROCEDURE IF EXISTS create_user_sharding_tables;

-- ============================================
-- 二、订单表分表（按月分表）
-- ============================================

-- 创建订单分表（示例：2024年1月到12月）
DELIMITER $$

DROP PROCEDURE IF EXISTS create_order_sharding_tables$$

CREATE PROCEDURE create_order_sharding_tables()
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE table_suffix VARCHAR(6);
    WHILE i <= 12 DO
        SET table_suffix = CONCAT('2024', LPAD(i, 2, '0'));
        SET @sql = CONCAT('
            CREATE TABLE `orders_', table_suffix, '` LIKE `orders`;
            ALTER TABLE `order_', table_suffix, '` COMMENT = ''订单表分表_', table_suffix, ''';
        ');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET i = i + 1;
END WHILE;
END$$

DELIMITER ;

-- 执行创建分表
CALL create_order_sharding_tables();

-- 删除存储过程
DROP PROCEDURE IF EXISTS create_order_sharding_tables;

-- ============================================
-- 三、订单商品项表分表（与订单表对应）
-- ============================================

DELIMITER $$

DROP PROCEDURE IF EXISTS create_order_item_sharding_tables$$

CREATE PROCEDURE create_order_item_sharding_tables()
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE table_suffix VARCHAR(6);
    WHILE i <= 12 DO
        SET table_suffix = CONCAT('2024', LPAD(i, 2, '0'));
        SET @sql = CONCAT('
            CREATE TABLE `order_item_', table_suffix, '` LIKE `order_item`;
            ALTER TABLE `order_item_', table_suffix, '` COMMENT = ''订单商品项表分表_', table_suffix, ''';
        ');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET i = i + 1;
END WHILE;
END$$

DELIMITER ;

-- 执行创建分表
CALL create_order_item_sharding_tables();

-- 删除存储过程
DROP PROCEDURE IF EXISTS create_order_item_sharding_tables;

-- ============================================
-- 四、订单日志表分表（与订单表对应）
-- ============================================

DELIMITER $$

DROP PROCEDURE IF EXISTS create_order_log_sharding_tables$$

CREATE PROCEDURE create_order_log_sharding_tables()
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE table_suffix VARCHAR(6);
    WHILE i <= 12 DO
        SET table_suffix = CONCAT('2024', LPAD(i, 2, '0'));
        SET @sql = CONCAT('
            CREATE TABLE `order_log_', table_suffix, '` LIKE `order_log`;
            ALTER TABLE `order_log_', table_suffix, '` COMMENT = ''订单日志表分表_', table_suffix, ''';
        ');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
SET i = i + 1;
END WHILE;
END$$

DELIMITER ;

-- 执行创建分表
CALL create_order_log_sharding_tables();

-- 删除存储过程
DROP PROCEDURE IF EXISTS create_order_log_sharding_tables;

-- ============================================
-- 使用说明
-- ============================================
-- 1. 用户表分表：根据 user_id % 16 路由到对应分表
--    示例：user_id = 12345，则路由到 user_9 (12345 % 16 = 9)
--
-- 2. 订单表分表：根据订单创建时间路由到对应月份分表
--    示例：订单创建时间为 2024-03-15，则路由到 order_202403
--
-- 3. 应用层需要实现分表路由逻辑
-- ============================================
