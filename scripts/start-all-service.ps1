# 启动所有微服务脚本 (Windows PowerShell)
# 使用方法: .\\scripts\\start-all-services.ps1

param(
    [switch]$Build = $false,
    [switch]$Gateway = $false
)

$ErrorActionPreference = "Stop"

# 服务配置列表 (服务名, 配置文件, 端口)
$services = @(
    @{Name="user-service"; Config="configs/dev/user-config.yaml"; Port=8080},
    @{Name="product-service"; Config="configs/dev/product-config.yaml"; Port=8081},
    @{Name="inventory-service"; Config="configs/dev/inventory-config.yaml"; Port=8084},
    @{Name="cart-service"; Config="configs/dev/cart-config.yaml"; Port=8085},
    @{Name="order-service"; Config="configs/dev/order-config.yaml"; Port=8082},
    @{Name="payment-service"; Config="configs/dev/payment-config.yaml"; Port=8083}
)

# 扩展服务（可选）
$optionalServices = @(
    @{Name="promotion-service"; Config="configs/dev/promotion-config.yaml"; Port=8086},
    @{Name="review-service"; Config="configs/dev/review-config.yaml"; Port=8087},
    @{Name="logistics-service"; Config="configs/dev/logistics-config.yaml"; Port=8088},
    @{Name="message-service"; Config="configs/dev/message-config.yaml"; Port=8089},
    @{Name="search-service"; Config="configs/dev/search-config.yaml"; Port=8090},
    @{Name="recommend-service"; Config="configs/dev/recommend-config.yaml"; Port=8091},
    @{Name="file-service"; Config="configs/dev/file-config.yaml"; Port=8092},
    @{Name="job-service"; Config="configs/dev/job-config.yaml"; Port=8093}
)

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "启动所有微服务" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# 检查配置文件
Write-Host "检查配置文件..." -ForegroundColor Yellow
foreach ($service in $services) {
    if (-not (Test-Path $service.Config)) {
        Write-Host "错误: 配置文件不存在: $($service.Config)" -ForegroundColor Red
        Write-Host "请先复制示例文件: $($service.Config).example -> $($service.Config)" -ForegroundColor Yellow
        exit 1
    }
}
Write-Host "配置文件检查通过" -ForegroundColor Green
Write-Host ""

# 编译服务（如果需要）
if ($Build) {
    Write-Host "编译所有服务..." -ForegroundColor Yellow
    foreach ($service in $services) {
        Write-Host "编译 $($service.Name)..." -ForegroundColor Yellow
        $exePath = "bin\\$($service.Name).exe"
        if (-not (Test-Path $exePath)) {
            go build -o $exePath "./cmd/$($service.Name)"
            if ($LASTEXITCODE -ne 0) {
                Write-Host "编译失败: $($service.Name)" -ForegroundColor Red
                exit 1
            }
        }
    }
    Write-Host "编译完成" -ForegroundColor Green
    Write-Host ""
}

# 创建日志目录
if (-not (Test-Path "logs")) {
    New-Item -ItemType Directory -Path "logs" | Out-Null
}

# 启动服务
Write-Host "启动服务..." -ForegroundColor Yellow
$jobs = @()

foreach ($service in $services) {
    Write-Host "启动 $($service.Name) (端口: $($service.Port))..." -ForegroundColor Cyan

    $exePath = "bin\\$($service.Name).exe"
    if (Test-Path $exePath) {
        # 使用编译后的可执行文件
        $job = Start-Process -FilePath $exePath -ArgumentList "-f", $service.Config -PassThru -NoNewWindow
    } else {
        # 直接运行 Go 程序
        $job = Start-Process -FilePath "go" -ArgumentList "run", "cmd/$($service.Name)/main.go", "-f", $service.Config -PassThru -NoNewWindow
    }

    $jobs += $job
    Start-Sleep -Seconds 2  # 等待服务启动
}

# 启动 API 网关（如果指定）
if ($Gateway) {
    Write-Host ""
    Write-Host "启动 API 网关..." -ForegroundColor Cyan
    $gatewayConfig = "configs/dev/gateway.yaml"
    if (Test-Path $gatewayConfig) {
        $exePath = "bin\\api-gateway.exe"
        if (Test-Path $exePath) {
            $job = Start-Process -FilePath $exePath -ArgumentList "-f", $gatewayConfig -PassThru -NoNewWindow
        } else {
            $job = Start-Process -FilePath "go" -ArgumentList "run", "cmd/api-gateway/main.go", "-f", $gatewayConfig -PassThru -NoNewWindow
        }
        $jobs += $job
    } else {
        Write-Host "警告: 网关配置文件不存在: $gatewayConfig" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host "所有服务已启动" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""
Write-Host "服务列表:" -ForegroundColor Cyan
foreach ($service in $services) {
    Write-Host "  - $($service.Name): http://localhost:$($service.Port)" -ForegroundColor White
}
if ($Gateway) {
    Write-Host "  - api-gateway: http://localhost:8080" -ForegroundColor White
}
Write-Host ""
Write-Host "按 Ctrl+C 停止所有服务" -ForegroundColor Yellow

# 等待用户中断
try {
    while ($true) {
        Start-Sleep -Seconds 1
    }
} finally {
    Write-Host ""
    Write-Host "正在停止所有服务..." -ForegroundColor Yellow
    foreach ($job in $jobs) {
        if (-not $job.HasExited) {
            Stop-Process -Id $job.Id -Force -ErrorAction SilentlyContinue
        }
    }
    Write-Host "所有服务已停止" -ForegroundColor Green
}
