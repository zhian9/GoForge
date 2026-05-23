@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ============================================
echo 启动所有微服务
echo ============================================
echo.

REM ============================================
REM 服务配置
REM 格式:
REM 服务名|配置文件|端口
REM ============================================

set services[0]=user-service|configs/dev/user-config.yaml|8080
set services[1]=product-service|configs/dev/product-config.yaml|8081
set services[2]=inventory-service|configs/dev/inventory-config.yaml|8084
set services[3]=cart-service|configs/dev/cart-config.yaml|8085
set services[4]=order-service|configs/dev/order-config.yaml|8082
set services[5]=payment-service|configs/dev/payment-config.yaml|8083

REM 可选扩展服务
set services[6]=promotion-service|configs/dev/promotion-config.yaml|8086
set services[7]=review-service|configs/dev/review-config.yaml|8087
set services[8]=logistics-service|configs/dev/logistics-config.yaml|8088
set services[9]=message-service|configs/dev/message-config.yaml|8089
set services[10]=search-service|configs/dev/search-config.yaml|8090
set services[11]=recommend-service|configs/dev/recommend-config.yaml|8091
set services[12]=file-service|configs/dev/file-config.yaml|8092
set services[13]=job-service|configs/dev/job-config.yaml|8093

REM ============================================
REM 创建日志目录
REM ============================================

if not exist logs (
    mkdir logs
)

echo 检查配置文件...
echo.

REM ============================================
REM 检查配置文件
REM ============================================

for /L %%i in (0,1,13) do (
    for /f "tokens=1,2,3 delims=|" %%a in ("!services[%%i]!") do (
        if not exist "%%b" (
            echo [错误] 配置文件不存在: %%b
            pause
            exit /b 1
        )
    )
)

echo 配置文件检查通过
echo.

echo 开始启动服务...
echo.

REM ============================================
REM 启动服务
REM ============================================

for /L %%i in (0,1,13) do (

    for /f "tokens=1,2,3 delims=|" %%a in ("!services[%%i]!") do (

        echo 启动 %%a ^(端口: %%c^)

        REM 如果存在 exe 则优先运行 exe
        if exist "bin\%%a.exe" (

            start "%%a" cmd /k ^
            "bin\%%a.exe -f %%b"

        ) else (

            start "%%a" cmd /k ^
            "go run cmd/%%a/main.go -f %%b"

        )

        timeout /t 2 >nul
    )
)

echo.
echo ============================================
echo 所有服务已启动
echo ============================================
echo.

echo 服务列表:
echo.

for /L %%i in (0,1,13) do (
    for /f "tokens=1,2,3 delims=|" %%a in ("!services[%%i]!") do (
        echo %%a: http://localhost:%%c
    )
)

echo.
echo 关闭窗口不会自动停止服务
echo 请手动关闭各个 CMD 窗口
echo.

pause