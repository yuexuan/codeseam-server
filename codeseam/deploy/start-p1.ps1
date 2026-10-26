[CmdletBinding()]
param([switch]$Build)
$ErrorActionPreference='Stop'
$repo=(Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$runtime=Join-Path $PSScriptRoot '.local'
New-Item -ItemType Directory -Force $runtime | Out-Null
$secretPath=Join-Path $runtime 'test-credentials.json'
if (-not (Test-Path -LiteralPath $secretPath)) {
    function New-Secret { return [Convert]::ToHexString([Security.Cryptography.RandomNumberGenerator]::GetBytes(24)) }
    [ordered]@{db_password=(New-Secret);admin_email='operator@codeseam.test';admin_password=(New-Secret);user_email='p1@codeseam.test';user_password=(New-Secret);jwt_secret=(New-Secret)} | ConvertTo-Json | Set-Content -LiteralPath $secretPath
    # Test-only credentials stay in this ignored directory with the current Windows user's ACL.
    & icacls.exe $runtime /inheritance:r /grant:r "${env:USERNAME}:(OI)(CI)F" /grant:r 'SYSTEM:(OI)(CI)F' | Out-Null
    if ($LASTEXITCODE -ne 0) {throw 'Cannot restrict runtime ACL'}
}
$secret=Get-Content -LiteralPath $secretPath -Raw | ConvertFrom-Json
$env:CODESEAM_DB_PASSWORD=$secret.db_password
& docker compose -p codeseam-p1 -f (Join-Path $PSScriptRoot 'compose.p1.yml') up -d --wait
if ($LASTEXITCODE -ne 0) {throw 'PostgreSQL/Redis failed'}
$go=Join-Path $env:LOCALAPPDATA 'CodeSeam/toolchains/go1.27.3/go/bin/go.exe'
$env:ZONEINFO=Join-Path (Split-Path (Split-Path $go)) 'lib/time/zoneinfo.zip'
$binary=Join-Path $repo 'backend/bin/codeseam-server.exe'
if ($Build -or -not (Test-Path -LiteralPath $binary)) {
    Push-Location (Join-Path $repo 'backend')
    try { $env:CGO_ENABLED='0'; & $go build -tags embed -trimpath -o $binary ./cmd/server; if ($LASTEXITCODE -ne 0) {throw 'Build failed'} } finally {Pop-Location}
}
$env:AUTO_SETUP='true';$env:DATA_DIR=Join-Path $runtime 'data'
New-Item -ItemType Directory -Force $env:DATA_DIR | Out-Null
$env:DATABASE_HOST='127.0.0.1';$env:DATABASE_PORT='15432';$env:DATABASE_USER='codeseam';$env:DATABASE_PASSWORD=$secret.db_password;$env:DATABASE_DBNAME='codeseam';$env:DATABASE_SSLMODE='disable'
$env:REDIS_HOST='127.0.0.1';$env:REDIS_PORT='16379'
$env:ADMIN_EMAIL=$secret.admin_email;$env:ADMIN_PASSWORD=$secret.admin_password;$env:JWT_SECRET=$secret.jwt_secret
$env:SERVER_HOST='127.0.0.1';$env:SERVER_PORT='18081';$env:SERVER_MODE='release';$env:RUN_MODE='standard';$env:TZ='Asia/Shanghai'
$existing=Get-NetTCPConnection -State Listen -LocalPort 18081 -ErrorAction SilentlyContinue
if ($existing) {throw '18081 already listening; inspect existing process before restart'}
$stamp=Get-Date -Format 'yyyyMMdd-HHmmss'
foreach($name in @('server.stdout.log','server.stderr.log')) {
    $log=Join-Path $runtime $name
    if(Test-Path -LiteralPath $log){Move-Item -LiteralPath $log -Destination (Join-Path $runtime "$stamp-$name")}
}
$p=Start-Process -FilePath $binary -WorkingDirectory (Join-Path $repo 'backend') -WindowStyle Hidden -PassThru -RedirectStandardOutput (Join-Path $runtime 'server.stdout.log') -RedirectStandardError (Join-Path $runtime 'server.stderr.log')
$p.Id | Set-Content (Join-Path $runtime 'server.pid')
Write-Host "CodeSeam main service PID $($p.Id), http://127.0.0.1:18081; test credentials: $secretPath"
