@echo off
REM ============================================
REM 数据库初始化脚本 (Windows)
REM ============================================

setlocal enabledelayedexpansion

REM 配置变量
set DB_HOST=%DB_HOST%
if "%DB_HOST%"=="" set DB_HOST=localhost

set DB_PORT=%DB_PORT%
if "%DB_PORT%"=="" set DB_PORT=3306

set DB_USER=%DB_USER%
if "%DB_USER%"=="" set DB_USER=root

set DB_PASS=%DB_PASS%
if "%DB_PASS%"=="" set DB_PASS=123456
set DB_NAME=%DB_NAME%
if "%DB_NAME%"=="" set DB_NAME=go_forge

echo ============================================
echo Go Microservice E-commerce Project - Database Initialization
echo ============================================
echo Database Host: %DB_HOST%
echo Database Port: %DB_PORT%
echo Database User: %DB_USER%
echo Database Name: %DB_NAME%
echo ============================================

REM 检查 MySQL 客户端是否安装
where mysql >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: mysql client not found, please install MySQL client first
    exit /b 1
)

REM 构建 MySQL 连接命令
set MYSQL_CMD=mysql -h%DB_HOST% -P%DB_PORT% -u%DB_USER%
if not "%DB_PASS%"=="" (
    set MYSQL_CMD=%MYSQL_CMD% -p%DB_PASS%
)

REM 创建数据库
echo Creating database...
%MYSQL_CMD% -e "CREATE DATABASE IF NOT EXISTS %DB_NAME% DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>nul
if %errorlevel% neq 0 (
    echo Error: Failed to create database
    exit /b 1
)
echo [OK] Database created successfully

REM 导入表结构
echo Importing table structure...
if not exist "%~dp0..serverdatabaseschema.sql" (
    echo Error: ..serverdatabaseschema.sql file not found
    exit /b 1
)
%MYSQL_CMD% %DB_NAME% < "%~dp0..serverdatabaseschema.sql" 2>nul
if %errorlevel% neq 0 (
    echo Error: Failed to import table structure
    exit /b 1
)
echo [OK] Table structure imported successfully

REM 询问是否创建分表
set /p CREATE_SHARDING=Create sharding tables? (y/n):
if /i "%CREATE_SHARDING%"=="y" (
    if not exist "%~dp0..serverdatabasesharding.sql" (
        echo Warning: ..serverdatabasesharding.sql file not found, skipping...
    ) else (
        echo Creating sharding tables...
        %MYSQL_CMD% %DB_NAME% < "%~dp0..serverdatabasesharding.sql" 2>nul
        if %errorlevel% neq 0 (
            echo Error: Failed to create sharding tables
            exit /b 1
        )
        echo [OK] Sharding tables created successfully
    )
)

echo ============================================
echo Database initialization completed!
echo ============================================

endlocal
