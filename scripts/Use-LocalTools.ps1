$toolDirectories = @(
    (Join-Path $env:LOCALAPPDATA 'KlevLiga\tools\go\bin'),
    'C:\Program Files\Go\bin',
    'C:\Program Files\nodejs',
    (Join-Path $env:LOCALAPPDATA 'KlevLiga\tools\pgsql\bin'),
    (Join-Path $env:LOCALAPPDATA 'KlevLiga\tools\mingw64\bin')
) | Where-Object { Test-Path -LiteralPath $_ }
$env:Path = ($toolDirectories -join ';') + ';' + $env:Path
