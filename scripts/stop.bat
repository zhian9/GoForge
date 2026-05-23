@echo off
echo Stopping all services...

taskkill /F /IM user-service.exe /IM product-service.exe /IM order-service.exe /IM payment-service.exe /IM inventory-service.exe /IM cart-service.exe /IM promotion-service.exe /IM review-service.exe /IM logistics-service.exe /IM message-service.exe /IM search-service.exe /IM recommend-service.exe /IM file-service.exe /IM job-service.exe /IM seckill-service.exe /IM api-gateway.exe >nul 2>&1

echo All services stopped.
