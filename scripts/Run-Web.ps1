$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'Use-LocalTools.ps1')
Push-Location (Join-Path $PSScriptRoot '..\apps\web')
try {
    & npm.cmd run dev -- --host 127.0.0.1 --strictPort
    if ($LASTEXITCODE -ne 0) { throw 'Frontend exited with an error' }
} finally { Pop-Location }
