# ============================================
# Check status of all microservices
# Usage: .\scripts\status.ps1
# ============================================

$services = @(
    @{Name="user-service";      Port=8000},
    @{Name="product-service";   Port=8081},
    @{Name="order-service";     Port=8082},
    @{Name="payment-service";   Port=8083},
    @{Name="inventory-service"; Port=8084},
    @{Name="cart-service";      Port=8085},
    @{Name="promotion-service"; Port=8006},
    @{Name="review-service";    Port=8007},
    @{Name="logistics-service"; Port=8008},
    @{Name="message-service";   Port=8009},
    @{Name="search-service";    Port=8010},
    @{Name="recommend-service"; Port=8011},
    @{Name="file-service";      Port=8012},
    @{Name="job-service";       Port=8013},
    @{Name="seckill-service";   Port=8090}
)

$infra = @(
    @{Name="Redis";  Port=6379},
    @{Name="etcd";   Port=2379},
    @{Name="Mongo";  Port=27017},
    @{Name="Kafka";  Port=9092}
)

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Service Status" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Infrastructure
Write-Host "Infrastructure:" -ForegroundColor White
foreach ($svc in $infra) {
    try {
        $tcp = New-Object Net.Sockets.TcpClient
        $tcp.Connect("127.0.0.1", $svc.Port)
        $tcp.Close()
        Write-Host "  [+] $($svc.Name) :$($svc.Port)" -ForegroundColor Green
    } catch {
        Write-Host "  [-] $($svc.Name) :$($svc.Port) DOWN" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "Microservices:" -ForegroundColor White
$up = 0
$down = 0
foreach ($svc in $services) {
    try {
        $tcp = New-Object Net.Sockets.TcpClient
        $tcp.Connect("127.0.0.1", $svc.Port)
        $tcp.Close()
        Write-Host "  [+] $($svc.Name)" -ForegroundColor Green -NoNewline
        Write-Host "  localhost:$($svc.Port)"
        $up++
    } catch {
        Write-Host "  [-] $($svc.Name)" -ForegroundColor Red -NoNewline
        Write-Host "  localhost:$($svc.Port) DOWN"
        $down++
    }
}

# Gateway
Write-Host ""
Write-Host "Gateway:" -ForegroundColor White
try {
    $tcp = New-Object Net.Sockets.TcpClient
    $tcp.Connect("127.0.0.1", 8080)
    $tcp.Close()
    Write-Host "  [+] api-gateway :8080" -ForegroundColor Green
    try {
        $code = (Invoke-WebRequest -Uri "http://localhost:8080/api/v1/products" -UseBasicParsing -TimeoutSec 3).StatusCode
        Write-Host "  [+] HTTP check: $code" -ForegroundColor Green
    } catch {
        Write-Host "  [!] HTTP check failed" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  [-] api-gateway :8080 DOWN" -ForegroundColor Red
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Total: $up running, $down down" -ForegroundColor $(if ($down -eq 0) { "Green" } else { "Red" })
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
