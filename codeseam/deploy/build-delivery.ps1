[CmdletBinding()]
param([string]$Installer='')
$ErrorActionPreference='Stop'
$repo=(Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$dest=Join-Path $PSScriptRoot 'dist'
if(-not $Installer){$Installer=Join-Path $repo '../codeseam-client/release/CodeSeam_0.5.0_x64-setup.exe'}
if(-not(Test-Path -LiteralPath $Installer)){throw 'Build the client installer first or pass -Installer'}
New-Item -ItemType Directory -Force $dest | Out-Null
Copy-Item -LiteralPath $Installer -Destination (Join-Path $dest 'CodeSeam-windows-x64-setup.exe')
$prior= $env:VITE_WINDOWS_DOWNLOAD_URL
Push-Location (Join-Path $repo 'codeseam/portal')
try {
    $env:VITE_WINDOWS_DOWNLOAD_URL='/downloads/codeseam-windows.exe'
    & npm.cmd run build
    if($LASTEXITCODE -ne 0){throw 'Portal production build failed'}
} finally {Pop-Location; $env:VITE_WINDOWS_DOWNLOAD_URL=$prior}
Write-Output 'Portal and installer prepared. Set CODESEAM_PORTAL_DIR and CODESEAM_WINDOWS_INSTALLER on the existing Go service.'
