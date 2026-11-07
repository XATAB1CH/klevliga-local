param(
    [ValidateSet('Local', 'Public')][string]$Mode = 'Local',
    [switch]$ValidateOnly,
    [switch]$Reload,
    [switch]$SkipBuild
)
$ErrorActionPreference = 'Stop'
if ($ValidateOnly -and $Reload) { throw 'Choose either ValidateOnly or Reload' }
. (Join-Path $PSScriptRoot 'Use-LocalTools.ps1')
$projectRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$caddyCommand = Get-Command caddy.exe -ErrorAction SilentlyContinue
if (!$caddyCommand) {
    $caddyLink = Join-Path $env:LOCALAPPDATA 'Microsoft\WinGet\Links\caddy.exe'
    if (Test-Path -LiteralPath $caddyLink) { $caddyCommand = Get-Item -LiteralPath $caddyLink }
}
if (!$caddyCommand) {
    $caddyPackages = Get-ChildItem -LiteralPath (Join-Path $env:LOCALAPPDATA 'Microsoft\WinGet\Packages') -Directory -Filter 'CaddyServer.Caddy_*' -ErrorAction SilentlyContinue
    foreach ($package in $caddyPackages) {
        $candidate = Join-Path $package.FullName 'caddy.exe'
        if (Test-Path -LiteralPath $candidate) { $caddyCommand = Get-Item -LiteralPath $candidate; break }
    }
}
if (!$caddyCommand) { throw 'Install Caddy: winget install --id CaddyServer.Caddy --exact' }
$caddyPath = $caddyCommand.FullName
if (!$caddyPath) { $caddyPath = $caddyCommand.Source }
$env:KLEVLIGA_WEB_ROOT = (Join-Path $projectRoot 'apps\web\dist').Replace('\', '/')
if (!$env:KLEVLIGA_API_UPSTREAM) { $env:KLEVLIGA_API_UPSTREAM = '127.0.0.1:8080' }
if (!$ValidateOnly -and !$Reload -and !$SkipBuild) {
    Push-Location (Join-Path $projectRoot 'apps\web')
    try {
        & npm.cmd run build
        if ($LASTEXITCODE -ne 0) { throw 'Frontend build failed' }
    } finally { Pop-Location }
}
if (!$ValidateOnly -and !(Test-Path -LiteralPath (Join-Path $env:KLEVLIGA_WEB_ROOT 'index.html'))) {
    throw 'Build apps/web before starting the gateway'
}
$configName = if ($Mode -eq 'Public') { 'Caddyfile' } else { 'Caddyfile.local' }
$gatewayConfig = Join-Path $projectRoot ('infra\' + $configName)
Push-Location $projectRoot
try {
    if ($ValidateOnly) {
        & $caddyPath validate --config $gatewayConfig --adapter caddyfile
    } elseif ($Reload) {
        & $caddyPath reload --config $gatewayConfig --adapter caddyfile
    } else {
        Write-Output ('Gateway mode: ' + $Mode + '. Stop with Ctrl+C.')
        & $caddyPath run --config $gatewayConfig --adapter caddyfile
    }
    if ($LASTEXITCODE -ne 0) { throw 'Caddy exited with an error' }
} finally { Pop-Location }
