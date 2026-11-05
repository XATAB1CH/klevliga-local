$wingetPackages = Join-Path $env:LOCALAPPDATA 'Microsoft\WinGet\Packages'
$compilerDirectories = if (Test-Path -LiteralPath $wingetPackages) {
    Get-ChildItem -LiteralPath $wingetPackages -Directory -Filter 'BrechtSanders.WinLibs*' |
        ForEach-Object { Join-Path $_.FullName 'mingw64\bin' }
}
$toolDirectories = @(
    (Join-Path $env:LOCALAPPDATA 'KlevLiga\tools\go\bin'),
    'C:\Program Files\Go\bin',
    'C:\Program Files\nodejs',
    (Join-Path $env:LOCALAPPDATA 'KlevLiga\tools\pgsql\bin'),
    (Join-Path $env:LOCALAPPDATA 'KlevLiga\tools\mingw64\bin')
    $compilerDirectories
) | Where-Object { Test-Path -LiteralPath $_ }
$env:Path = ($toolDirectories -join ';') + ';' + $env:Path
