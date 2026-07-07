# Register Windows scheduled task: Delivery mirror on the 1st of each month at 08:00.
# VPN must be available at run time (or run delivery-monthly-sync.ps1 manually when connected).
$ErrorActionPreference = 'Stop'

$TaskName = 'GallupQ14-DeliveryMirrorMonthly'
$ScriptPath = Join-Path $PSScriptRoot 'delivery-monthly-sync.ps1'
$RepoRoot = Split-Path -Parent $PSScriptRoot

$TaskCmd = "powershell.exe -NoProfile -ExecutionPolicy Bypass -File `"$ScriptPath`""

schtasks /Delete /TN $TaskName /F 2>$null | Out-Null
schtasks /Create /TN $TaskName /TR $TaskCmd /SC MONTHLY /D 1 /ST 08:00 /RL LIMITED /F

Write-Host "Registered: $TaskName (1st of month, 08:00)"
Write-Host "Manual: schtasks /Run /TN $TaskName"
Write-Host "Or: powershell -File `"$ScriptPath`""
