# Upload delivery_reference_seed.sql to VPS (after sync from local mirror on VPN machine).
$ErrorActionPreference = 'Stop'

$RepoRoot = Split-Path -Parent $PSScriptRoot
$Seed = Join-Path $RepoRoot 'scripts\delivery_reference_seed.sql'
$ApplyScript = Join-Path $RepoRoot 'scripts\apply_delivery_reference.sh'
if (-not (Test-Path $Seed)) {
    Write-Error "Seed not found: $Seed - run sync_delivery_reference.py and export_delivery_reference_sql.py first"
}

function Get-DeployEnv([string]$Name) {
    if (Test-Path "Env:$Name") { return (Get-Item "Env:$Name").Value }
    return [Environment]::GetEnvironmentVariable($Name, 'User')
}

$vpsHost = Get-DeployEnv 'INTERXION_SWI_HOST'
$user = Get-DeployEnv 'INTERXION_SWI_USER'
if (-not $user) { $user = 'root' }
$hostkey = Get-DeployEnv 'INTERXION_SWI_HOSTKEY'
$pass = Get-DeployEnv 'INTERXION_SWI_PASS'
if (-not $vpsHost) { Write-Error 'Set INTERXION_SWI_HOST (User env)' }
if (-not $hostkey) { Write-Error 'Set INTERXION_SWI_HOSTKEY (User env)' }
if (-not $pass) { Write-Error 'Set INTERXION_SWI_PASS (User env)' }
$remote = "${user}@${vpsHost}"
$remoteSeed = '/opt/gallup-q14/scripts/delivery_reference_seed.sql'
$remoteApply = '/opt/gallup-q14/scripts/apply_delivery_reference.sh'

& 'C:\Program Files\PuTTY\pscp.exe' -batch -hostkey $hostkey -pw $pass $Seed "${remote}:${remoteSeed}.new"
& 'C:\Program Files\PuTTY\pscp.exe' -batch -hostkey $hostkey -pw $pass $ApplyScript "${remote}:${remoteApply}.new"
& 'C:\Program Files\PuTTY\plink.exe' -batch -hostkey $hostkey -ssh $remote -pw $pass "set -e; mv ${remoteSeed}.new ${remoteSeed}; mv ${remoteApply}.new ${remoteApply}; chmod +x ${remoteApply}; chown gallup:gallup ${remoteSeed}; rm -f /opt/gallup-q14/data/delivery_mirror.db; bash ${remoteApply}; systemctl restart gallup-q14"

Write-Host 'Reference seed uploaded and applied on VPS (mirror removed if present).'
