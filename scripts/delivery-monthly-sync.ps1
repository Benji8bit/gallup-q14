# Monthly Delivery mirror refresh (requires corporate VPN).
# Pulls from PostgreSQL into backend/data/delivery_mirror.db, then updates app DB + seed export.
$ErrorActionPreference = 'Stop'

$RepoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $RepoRoot

if (-not $env:DELIVERY_SAPIENS_DB_HOST) {
    $env:DELIVERY_SAPIENS_DB_HOST = [Environment]::GetEnvironmentVariable('DELIVERY_SAPIENS_DB_HOST', 'User')
}
if (-not $env:DELIVERY_SAPIENS_DB_HOST) {
    Write-Error 'DELIVERY_SAPIENS_DB_HOST is not set. Connect VPN and set User env var, then retry.'
}

$env:DB_PATH = Join-Path $RepoRoot 'backend\data\gallup-q14.db'

Write-Host '=== Pull Delivery mirror (VPN) ==='
python (Join-Path $PSScriptRoot 'pull_delivery_mirror.py')
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host '=== Sync reference into app SQLite ==='
python (Join-Path $PSScriptRoot 'sync_delivery_reference.py')
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host '=== Export seed for VPS ==='
python (Join-Path $PSScriptRoot 'export_delivery_reference_sql.py')
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host '=== Upload reference seed to VPS ==='
powershell -ExecutionPolicy Bypass -File (Join-Path $PSScriptRoot 'upload-reference-to-vps.ps1')
