param([ValidateRange(1024, 65535)][int]$Port = 5434)
$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'Use-LocalTools.ps1')
$projectRoot = Split-Path -Parent $PSScriptRoot
$databaseRoot = Join-Path $env:LOCALAPPDATA 'KlevLiga\postgres'
$dataDirectory = Join-Path $databaseRoot 'data'
$settingsPath = Join-Path $databaseRoot 'connection.json'
New-Item -ItemType Directory -Force -Path $databaseRoot | Out-Null

if (!(Test-Path -LiteralPath $settingsPath)) {
    if (Test-Path -LiteralPath (Join-Path $dataDirectory 'PG_VERSION')) {
        throw 'An existing cluster has no connection settings. Restore connection.json; the cluster was not modified.'
    }
    $randomBytes = New-Object byte[] 32
    $random = [Security.Cryptography.RandomNumberGenerator]::Create()
    try { $random.GetBytes($randomBytes) } finally { $random.Dispose() }
    $settings = @{ username = 'klevliga'; password = [Convert]::ToBase64String($randomBytes); port = $Port }
    [IO.File]::WriteAllText($settingsPath, ($settings | ConvertTo-Json), (New-Object Text.UTF8Encoding $false))
}
$settings = Get-Content -Raw -Encoding UTF8 -LiteralPath $settingsPath | ConvertFrom-Json
if ($Port -ne $settings.port) { throw 'Use the port recorded in the existing local connection settings.' }

if (!(Test-Path -LiteralPath (Join-Path $dataDirectory 'PG_VERSION'))) {
    $passwordPath = Join-Path $databaseRoot 'init-password.tmp'
    try {
        [IO.File]::WriteAllText($passwordPath, $settings.password, (New-Object Text.UTF8Encoding $false))
        & initdb.exe -D $dataDirectory -U $settings.username --encoding=UTF8 --locale=C --locale-provider=icu --icu-locale=ru-RU --auth=scram-sha-256 --pwfile=$passwordPath
        if ($LASTEXITCODE -ne 0) { throw 'initdb failed' }
    } finally {
        if (Test-Path -LiteralPath $passwordPath) { Remove-Item -LiteralPath $passwordPath }
    }
}

& pg_ctl.exe -D $dataDirectory status *> $null
if ($LASTEXITCODE -ne 0) {
    $logPath = Join-Path $databaseRoot 'postgres.log'
    & pg_ctl.exe -D $dataDirectory -l $logPath -o "-h 127.0.0.1 -p $Port" -w start
    if ($LASTEXITCODE -ne 0) { throw "PostgreSQL failed to start. See $logPath" }
}

$previousPassword = $env:PGPASSWORD
try {
    $env:PGPASSWORD = $settings.password
    foreach ($database in @('klevliga', 'klevliga_test')) {
        $exists = & psql.exe -h 127.0.0.1 -p $Port -U $settings.username -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = '$database'"
        if ($LASTEXITCODE -ne 0) { throw 'Cannot connect to the local PostgreSQL cluster' }
        if ($exists -ne '1') {
            & createdb.exe -h 127.0.0.1 -p $Port -U $settings.username $database
            if ($LASTEXITCODE -ne 0) { throw "Cannot create $database" }
        }
    }
} finally { $env:PGPASSWORD = $previousPassword }

$configPath = Join-Path $projectRoot 'services\api\config.yaml'
if (!(Test-Path -LiteralPath $configPath)) {
    $localConfig = @"
server:
  http:
    address: "127.0.0.1:8080"
connections:
  postgres:
    main:
      host: 127.0.0.1
      port: $Port
      user: $($settings.username)
      password: "$($settings.password)"
      dbname: klevliga
auth:
  session_ttl: 720h
  cookie_name: klevliga_session
  cookie_secure: false
  cookie_domain: ""
logger:
  title: klevliga-api
  level: debug
  pretty: true
"@
    [IO.File]::WriteAllText($configPath, $localConfig, (New-Object Text.UTF8Encoding $false))
}
Write-Output "PostgreSQL is ready at 127.0.0.1:$Port (klevliga, klevliga_test)."
