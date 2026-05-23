# ============================================
# One-click microservice bootstrap script
# Usage:
#   .\scripts\bootstrap.ps1                     # Full: Docker + build + all services + gateway
#   .\scripts\bootstrap.ps1 -SkipBuild           # Skip compilation
#   .\scripts\bootstrap.ps1 -SkipInfra            # Skip Docker (containers already running)
#   .\scripts\bootstrap.ps1 -NoGateway            # Don't start API gateway
#
# Or via Make:
#   make bootstrap
# ============================================
param(
    [switch]$SkipBuild = $false,
    [switch]$SkipInfra = $false,
    [switch]$NoGateway = $false
)

$ErrorActionPreference = "Stop"
$projectRoot = $PSScriptRoot | Split-Path -Parent

# ---------- Service definitions ----------
# Core gRPC services (all must be running before gateway)
$allServices = @(
    @{Name="user-service";      Config="configs/dev/user-config.yaml";       Port=8000},
    @{Name="product-service";   Config="configs/dev/product-config.yaml";    Port=8081},
    @{Name="order-service";     Config="configs/dev/order-config.yaml";      Port=8082},
    @{Name="payment-service";   Config="configs/dev/payment-config.yaml";    Port=8083},
    @{Name="inventory-service"; Config="configs/dev/inventory-config.yaml";  Port=8084},
    @{Name="cart-service";      Config="configs/dev/cart-config.yaml";       Port=8085},
    @{Name="promotion-service"; Config="configs/dev/promotion-config.yaml";  Port=8006},
    @{Name="review-service";    Config="configs/dev/review-config.yaml";     Port=8007},
    @{Name="logistics-service"; Config="configs/dev/logistics-config.yaml";  Port=8008},
    @{Name="message-service";   Config="configs/dev/message-config.yaml";    Port=8009},
    @{Name="search-service";    Config="configs/dev/search-config.yaml";     Port=8010},
    @{Name="recommend-service"; Config="configs/dev/recommend-config.yaml";  Port=8011},
    @{Name="file-service";      Config="configs/dev/file-config.yaml";       Port=8012},
    @{Name="job-service";       Config="configs/dev/job-config.yaml";        Port=8013},
    @{Name="seckill-service";   Config="configs/dev/seckill-config.yaml";    Port=8090}
)

# Infrastructure ports to check
$infraServices = @(
    @{Name="Redis"; Host="127.0.0.1"; Port=6379},
    @{Name="etcd";  Host="127.0.0.1"; Port=2379},
    @{Name="Mongo"; Host="127.0.0.1"; Port=27017},
    @{Name="Kafka"; Host="127.0.0.1"; Port=9092}
)

# ==============================================
Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Microservice Bootstrap" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "  Config: Build=$(-not $SkipBuild) Infra=$(-not $SkipInfra) Gateway=$(-not $NoGateway)"
Write-Host ""

# ==============================================
# Step 1 + 2: Docker + Infrastructure
# ==============================================
if (-not $SkipInfra) {
    Write-Host "[1/5] Checking Docker..." -ForegroundColor Yellow

    $dockerOk = $false
    try { docker ps 2>&1 | Out-Null; if ($LASTEXITCODE -eq 0) { $dockerOk = $true } } catch { }

    if (-not $dockerOk) {
        Write-Host "  Starting Docker Desktop..." -ForegroundColor Gray
        Start-Process "C:\Program Files\Docker\Docker\Docker Desktop.exe" -ErrorAction SilentlyContinue
        $wait = 0
        while ($wait -lt 60) {
            try { docker ps 2>&1 | Out-Null; if ($LASTEXITCODE -eq 0) { $dockerOk = $true; break } } catch { }
            Write-Host "  Waiting... ($wait s)" -ForegroundColor Gray
            Start-Sleep -Seconds 5
            $wait += 5
        }
        if (-not $dockerOk) {
            Write-Host "  ERROR: Docker Desktop failed to start" -ForegroundColor Red
            exit 1
        }
    }
    Write-Host "  Docker is ready" -ForegroundColor Green

    Write-Host ""
    Write-Host "[2/5] Starting infrastructure containers..." -ForegroundColor Yellow
    Push-Location $projectRoot

    docker compose -f docker-compose-infra.yml down --remove-orphans 2>$null
    docker rm -f infra-redis infra-etcd infra-etcd-keeper infra-mongodb infra-elasticsearch infra-prometheus infra-zookeeper infra-kafka infra-kafka-ui 2>$null

    docker compose -f docker-compose-infra.yml up -d
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ERROR: docker compose failed" -ForegroundColor Red
        Pop-Location
        exit 1
    }

    Write-Host "  Waiting for infrastructure..." -ForegroundColor Gray
    foreach ($svc in $infraServices) {
        $ready = $false
        for ($i = 0; $i -lt 30; $i++) {
            try {
                $tcp = New-Object Net.Sockets.TcpClient
                $tcp.Connect($svc.Host, $svc.Port)
                $tcp.Close()
                $ready = $true
                break
            } catch { Start-Sleep -Seconds 2 }
        }
        $icon = if ($ready) { "[+]" } else { "[!]" }
        $color = if ($ready) { "Green" } else { "Yellow" }
        Write-Host "  $icon $($svc.Name) ($($svc.Host):$($svc.Port))" -ForegroundColor $color
    }
    Pop-Location
} else {
    Write-Host "[1/5] Skipping Docker" -ForegroundColor Yellow
    Write-Host "[2/5] Skipping infrastructure" -ForegroundColor Yellow
}

# ==============================================
# Step 3: Build all services
# ==============================================
Write-Host ""
if (-not $SkipBuild) {
    Write-Host "[3/5] Building services..." -ForegroundColor Yellow

    if (-not (Test-Path "$projectRoot/bin")) {
        New-Item -ItemType Directory -Path "$projectRoot/bin" | Out-Null
    }

    $allNames = $allServices | ForEach-Object { $_.Name }
    $buildList = $allNames
    if (-not $NoGateway) { $buildList += "api-gateway" }

    Push-Location $projectRoot
    foreach ($name in $buildList) {
        Write-Host "  Building $name..." -ForegroundColor Gray
        go build -o "bin/$name.exe" "./cmd/$name"
        if ($LASTEXITCODE -ne 0) {
            Write-Host "  BUILD FAILED: $name" -ForegroundColor Red
            Pop-Location
            exit 1
        }
    }
    Pop-Location
    Write-Host "  All services built ($($buildList.Count) total)" -ForegroundColor Green
} else {
    Write-Host "[3/5] Skipping build" -ForegroundColor Yellow
}

# ==============================================
# Step 4: Start all microservices
# ==============================================
Write-Host ""
Write-Host "[4/5] Starting microservices..." -ForegroundColor Yellow

Push-Location $projectRoot
if (-not (Test-Path "logs")) { New-Item -ItemType Directory -Path "logs" | Out-Null }

# Kill all old processes
Write-Host "  Stopping old processes..." -ForegroundColor Gray
$allNames = $allServices | ForEach-Object { "$($_.Name).exe" }
Get-Process -Name $allNames -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

# Also kill gateway if we'll restart it
if (-not $NoGateway) {
    Get-Process -Name "api-gateway" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
}
Start-Sleep -Seconds 2

# Start each service, waiting for port
$startedOk = @()
$startedFail = @()
foreach ($svc in $allServices) {
    Write-Host "  Starting $($svc.Name)..." -ForegroundColor Gray
    $logPath = "logs/$($svc.Name).log"
    $exePath = "bin/$($svc.Name).exe"

    if (Test-Path $exePath) {
        Start-Process -FilePath "cmd.exe" -ArgumentList "/c", "`"$exePath`" -f `"$($svc.Config)`" > `"$logPath`" 2>&1" -NoNewWindow
    } else {
        Start-Process -FilePath "cmd.exe" -ArgumentList "/c", "go run `"cmd/$($svc.Name)/main.go`" -f `"$($svc.Config)`" > `"$logPath`" 2>&1" -NoNewWindow
    }

    # Wait for port to be actually bound
    $online = $false
    for ($i = 0; $i -lt 15; $i++) {
        Start-Sleep -Seconds 1
        try {
            $tcp = New-Object Net.Sockets.TcpClient
            $tcp.Connect("127.0.0.1", $svc.Port)
            $tcp.Close()
            $online = $true
            break
        } catch { }
    }

    if ($online) {
        Write-Host "    $($svc.Name) :$($svc.Port) OK" -ForegroundColor Green
        $startedOk += $svc
    } else {
        Write-Host "    $($svc.Name) :$($svc.Port) TIMEOUT (see logs/$($svc.Name).log)" -ForegroundColor Red
        $startedFail += $svc
    }
}

Write-Host ""
Write-Host "  Running: $($startedOk.Count)/$($allServices.Count) services" -ForegroundColor $(if ($startedFail.Count -eq 0) { "Green" } else { "Yellow" })

# ==============================================
# Step 5: Start API Gateway (LAST, after all upstreams ready)
# ==============================================
Write-Host ""
if (-not $NoGateway) {
    Write-Host "[5/5] Starting API Gateway..." -ForegroundColor Yellow

    # Gateway needs all upstream services to be available
    if ($startedFail.Count -eq 0) {
        $gwLog = "logs/api-gateway.log"
        $gwExe = "bin/api-gateway.exe"

        if (-not $SkipBuild) {
            Write-Host "  Building api-gateway..." -ForegroundColor Gray
            Push-Location $projectRoot
            go build -o $gwExe ./cmd/api-gateway
            Pop-Location
        }

        Write-Host "  Starting gateway..." -ForegroundColor Gray
        if (Test-Path $gwExe) {
            Start-Process -FilePath "cmd.exe" -ArgumentList "/c", "`"$gwExe`" -f configs/dev/gateway.yaml > `"$gwLog`" 2>&1" -NoNewWindow
        } else {
            Start-Process -FilePath "cmd.exe" -ArgumentList "/c", "go run cmd/api-gateway/main.go -f configs/dev/gateway.yaml > `"$gwLog`" 2>&1" -NoNewWindow
        }

        # Wait for gateway port
        Start-Sleep -Seconds 5
        $gwOnline = $false
        for ($i = 0; $i -lt 10; $i++) {
            try {
                $tcp = New-Object Net.Sockets.TcpClient
                $tcp.Connect("127.0.0.1", 8080)
                $tcp.Close()
                $gwOnline = $true
                break
            } catch { Start-Sleep -Seconds 1 }
        }

        if ($gwOnline) {
            Write-Host "  Gateway :8080 OK" -ForegroundColor Green
            Write-Host "  Swagger :8095" -ForegroundColor Green
        } else {
            Write-Host "  Gateway startup may have failed, check logs/api-gateway.log" -ForegroundColor Yellow
        }
    } else {
        Write-Host "[5/5] Gateway SKIPPED - some upstreams failed to start:" -ForegroundColor Red
        foreach ($svc in $startedFail) {
            Write-Host "  - $($svc.Name) :$($svc.Port)" -ForegroundColor Red
        }
        Write-Host "  Fix the issues above, then run: make start-services" -ForegroundColor Yellow
    }
} else {
    Write-Host "[5/5] Gateway SKIPPED (--no-gateway)" -ForegroundColor Yellow
}

# ==============================================
# Summary
# ==============================================
Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Startup Complete" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Microservices:" -ForegroundColor White
foreach ($svc in $allServices) {
    $status = if ($svc.Name -in $startedOk.Name) { "+" } else { "-" }
    $color = if ($svc.Name -in $startedOk.Name) { "Green" } else { "Red" }
    Write-Host "  [$status] $($svc.Name)" -ForegroundColor $color -NoNewline
    Write-Host "  localhost:$($svc.Port)"
}

Write-Host ""
Write-Host "Infrastructure:" -ForegroundColor White
Write-Host "  Redis:   localhost:6379" -ForegroundColor White
Write-Host "  etcd:    localhost:2379  (keeper: http://localhost:8089)" -ForegroundColor White
Write-Host "  MongoDB: localhost:27017" -ForegroundColor White
Write-Host "  Kafka:   localhost:9092  (UI: http://localhost:18090)" -ForegroundColor White
Write-Host "  ES:      localhost:9200" -ForegroundColor White
Write-Host "  Prom:    http://localhost:9090" -ForegroundColor White

if (-not $NoGateway) {
    Write-Host ""
    Write-Host "API Entry:" -ForegroundColor White
    Write-Host "  Gateway:  http://localhost:8080/api/v1/..." -ForegroundColor Green
    Write-Host "  Swagger:  http://localhost:8095/" -ForegroundColor Green
}

Write-Host ""
Write-Host "Logs: $projectRoot\logs\" -ForegroundColor Gray
Write-Host ""

Pop-Location
