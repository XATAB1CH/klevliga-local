param([switch]$Race)
$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'Use-LocalTools.ps1')
$settingsPath = Join-Path $env:LOCALAPPDATA 'KlevLiga\postgres\connection.json'
$previousDatabaseURL = $env:KLEVLIGA_TEST_DATABASE_URL
try {
    if (!$env:KLEVLIGA_TEST_DATABASE_URL) {
        $settings = Get-Content -Raw -Encoding UTF8 -LiteralPath $settingsPath | ConvertFrom-Json
        $password = [Uri]::EscapeDataString($settings.password)
        $env:KLEVLIGA_TEST_DATABASE_URL = "postgres://$($settings.username):$password@127.0.0.1:$($settings.port)/klevliga_test?sslmode=disable"
    }
    Push-Location (Join-Path $PSScriptRoot '..\services\api')
    try {
        $arguments = @('test', '-count=1', '-tags=integration')
        if ($Race) { $arguments += '-race' }
        & go @arguments ./...
        if ($LASTEXITCODE -ne 0) { throw 'API tests failed' }
    } finally { Pop-Location }
} finally { $env:KLEVLIGA_TEST_DATABASE_URL = $previousDatabaseURL }
