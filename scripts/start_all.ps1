Write-Host "🚀 Starting Nebula Cluster..." -ForegroundColor Cyan

Write-Host "Checking Redis..."
if (!(docker ps -q -f name=nebula-redis)) {
    Write-Host "Starting Redis Container..." -ForegroundColor Yellow
    docker run -d --name nebula-redis -p 6379:6379 redis:alpine
} else {
    Write-Host "Redis is already running." -ForegroundColor Green
}

Write-Host "Starting Worker 1 (Port 9090)..." -ForegroundColor Green
Start-Process powershell -ArgumentList "go run cmd/worker/main.go -port 9090"

Write-Host "Starting Worker 2 (Port 9091)..." -ForegroundColor Green
Start-Process powershell -ArgumentList "go run cmd/worker/main.go -port 9091"

Write-Host "Starting Gateway (Port 3000)..." -ForegroundColor Magenta
Write-Host "Access Dashboard at http://localhost:3000"
go run cmd/gateway/main.go