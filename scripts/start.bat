@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

:: GoForge E-Commerce One-Click Start
:: Usage: start.bat           full start
::        start.bat nobuild   skip build
::        start.bat noinfra   skip docker

set PROJECT_ROOT=%~dp0..\server
set BIN_DIR=%PROJECT_ROOT%\bin
set LOG_DIR=%PROJECT_ROOT%\logs
set CONFIG_DIR=%~dp0..\configs\dev
set SKIP_BUILD=false
set SKIP_INFRA=false

:: Load .env
if exist "%~dp0..\.env" for /f "usebackq tokens=1,2 delims==" %%a in ("%~dp0..\.env") do (
  if not "%%a"=="" if not "%%a"=="#" if not "%%a"=="# " set "%%a=%%b"
)

if /i "%~1"=="nobuild" set SKIP_BUILD=true
if /i "%~1"=="noinfra"  set SKIP_INFRA=true
if /i "%~2"=="nobuild" set SKIP_BUILD=true
if /i "%~2"=="noinfra"  set SKIP_INFRA=true

echo.
echo ============================================
echo    GoForge E-Commerce - Starting...
echo ============================================
echo.

:: Step 1: Docker
if "%SKIP_INFRA%"=="true" (
  echo [1/4] Skip Docker infrastructure
) else (
  echo [1/4] Starting Docker infrastructure...
  cd /d "%PROJECT_ROOT%"
  docker compose -f docker-compose-infra.yml up -d
  echo   Docker OK
)

:: Step 2: Build
if "%SKIP_BUILD%"=="true" (
  echo [2/4] Skip build
) else (
  echo [2/4] Building services...
  if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"
  cd /d "%PROJECT_ROOT%"

  for %%s in (user product order payment inventory cart promotion review logistics message search recommend file job seckill) do (
    echo   Building %%s-service...
    go build -o "%BIN_DIR%\%%s-service.exe" ".\cmd\%%s-service"
    if !errorlevel! neq 0 ( echo   FAILED & pause & exit /b 1 )
  )
  echo   Building api-gateway...
  go build -o "%BIN_DIR%\api-gateway.exe" ".\cmd\api-gateway"
  if !errorlevel! neq 0 ( echo   FAILED & pause & exit /b 1 )
  echo   Build OK
)

:: Step 3: Start Microservices
echo.
echo [3/4] Starting microservices...
cd /d "%PROJECT_ROOT%"
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"

:: Kill old processes
taskkill /F /IM user-service.exe /IM product-service.exe /IM order-service.exe /IM payment-service.exe /IM inventory-service.exe /IM cart-service.exe /IM promotion-service.exe /IM review-service.exe /IM logistics-service.exe /IM message-service.exe /IM search-service.exe /IM recommend-service.exe /IM file-service.exe /IM job-service.exe /IM seckill-service.exe /IM api-gateway.exe >nul 2>&1
timeout /t 2 /nobreak >nul

for %%s in ("user:8000:user" "product:8081:product" "order:8082:order" "payment:8083:payment" "inventory:8084:inventory" "cart:8085:cart" "promotion:8006:promotion" "review:8007:review" "logistics:8008:logistics" "message:8009:message" "search:8010:search" "recommend:8011:recommend" "file:8012:file" "job:8013:job" "seckill:8090:seckill") do (
  for /f "tokens=1,2,3 delims=:" %%a in (%%s) do (
    echo   Starting %%a-service...
    start /B "" "%BIN_DIR%\%%a-service.exe" -f "%CONFIG_DIR%\%%c-config.yaml" > "%LOG_DIR%\%%a-service.log" 2>&1
    timeout /t 1 /nobreak >nul
  )
)

echo   Waiting for services (20s)...
timeout /t 20 /nobreak >nul

:: Step 4: Gateway
echo.
echo [4/4] Starting API Gateway...
start /B "" "%BIN_DIR%\api-gateway.exe" -f "%CONFIG_DIR%\gateway.yaml" > "%LOG_DIR%\api-gateway.log" 2>&1
timeout /t 6 /nobreak >nul

echo.
echo ============================================
echo    GoForge is running!
echo    API  : http://localhost:8080/api/v1
echo    Docs : http://localhost:8095
echo ============================================
echo.

endlocal
