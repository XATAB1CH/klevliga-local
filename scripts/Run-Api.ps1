$ErrorActionPreference = 'Stop'
. (Join-Path $PSScriptRoot 'Use-LocalTools.ps1')
Push-Location (Join-Path $PSScriptRoot '..\services\api')
try {
    & go run ./cmd/klevliga -c ./config.yaml
    if ($LASTEXITCODE -ne 0) { throw 'API exited with an error' }
} finally { Pop-Location }
