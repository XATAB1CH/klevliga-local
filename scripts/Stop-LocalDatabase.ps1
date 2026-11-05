$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'Use-LocalTools.ps1')
$dataDirectory = Join-Path $env:LOCALAPPDATA 'KlevLiga\postgres\data'
if (Test-Path -LiteralPath (Join-Path $dataDirectory 'PG_VERSION')) {
    & pg_ctl.exe -D $dataDirectory -m fast -w stop
    if ($LASTEXITCODE -ne 0) { throw 'PostgreSQL could not be stopped' }
}
