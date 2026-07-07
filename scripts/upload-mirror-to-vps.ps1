# Upload delivery_mirror.db to VPS (after pull_delivery_mirror.py on a VPN machine).
$ErrorActionPreference = 'Stop'

$RepoRoot = Split-Path -Parent $PSScriptRoot
$Mirror = Join-Path $RepoRoot 'backend\data\delivery_mirror.db'
if (-not (Test-Path $Mirror)) {
    Write-Error "Mirror not found: $Mirror - run pull_delivery_mirror.py first"
}

$host = [Environment]::GetEnvironmentVariable('INTERXION_SWI_HOST', 'User')
$user = [Environment]::GetEnvironmentVariable('INTERXION_SWI_USER', 'User')
if (-not $user) { $user = 'root' }
$hostkey = [Environment]::GetEnvironmentVariable('INTERXION_SWI_HOSTKEY', 'User')
$pass = [Environment]::GetEnvironmentVariable('INTERXION_SWI_PASS', 'User')
if (-not $host) { Write-Error 'Set INTERXION_SWI_HOST (User env)' }
if (-not $hostkey) { Write-Error 'Set INTERXION_SWI_HOSTKEY (User env)' }
if (-not $pass) { Write-Error 'Set INTERXION_SWI_PASS (User env)' }
$remote = "${user}@${host}"
$remotePath = '/opt/gallup-q14/data/delivery_mirror.db'

& 'C:\Program Files\PuTTY\pscp.exe' -batch -hostkey $hostkey -pw $pass $Mirror "${remote}:${remotePath}.new"
& 'C:\Program Files\PuTTY\plink.exe' -batch -hostkey $hostkey -ssh $remote -pw $pass "set -e; mv ${remotePath}.new ${remotePath}; chown gallup:gallup ${remotePath}; bash /opt/gallup-q14/scripts/vps-sync-from-mirror.sh; systemctl restart gallup-q14"

Write-Host 'Mirror uploaded and reference synced on VPS.'
