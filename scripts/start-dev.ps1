# Запуск MVP в режиме разработки (два окна)
param(
    [string]$AdminPassword = "change-me"
)

$root = Split-Path -Parent $PSScriptRoot

Write-Host "Starting backend on :8080..."
Start-Process powershell -ArgumentList @(
    "-NoExit", "-Command",
    "cd '$root\backend'; `$env:ADMIN_PASSWORD='$AdminPassword'; `$env:CORS_ORIGIN='http://localhost:5173'; go run ./cmd/server"
)

Start-Sleep -Seconds 2

Write-Host "Starting frontend on :5173..."
Start-Process powershell -ArgumentList @(
    "-NoExit", "-Command",
    "cd '$root\frontend'; npm run dev"
)

Write-Host ""
Write-Host "Survey:  http://localhost:5173/survey"
Write-Host "Admin:   http://localhost:5173/admin  (password: $AdminPassword)"
Write-Host "Docs:    $root\docs\"
