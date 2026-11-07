param(
    [ValidatePattern('^[a-zA-Z0-9.-]+$')][string]$Domain = 'klevliga.ru',
    [ValidatePattern('^[a-zA-Z0-9_-]+$')][string]$DkimSelector
)
$ErrorActionPreference = 'Stop'
$queries = @(
    @{ Name = $Domain; Type = 'NS' },
    @{ Name = $Domain; Type = 'A' },
    @{ Name = $Domain; Type = 'AAAA' },
    @{ Name = ('www.' + $Domain); Type = 'CNAME' },
    @{ Name = $Domain; Type = 'MX' },
    @{ Name = $Domain; Type = 'TXT' },
    @{ Name = ('_dmarc.' + $Domain); Type = 'TXT' }
)
if ($DkimSelector) { $queries += @{ Name = ($DkimSelector + '._domainkey.' + $Domain); Type = 'TXT' } }
foreach ($query in $queries) {
    $uri = 'https://dns.google/resolve?name=' + $query.Name + '&type=' + $query.Type
    try {
        $result = Invoke-RestMethod -Uri $uri -TimeoutSec 15
        $status = switch ([int]$result.Status) { 0 { 'OK' }; 3 { 'NXDOMAIN' }; default { 'DNS error ' + $result.Status } }
        $answers = @($result.Answer | ForEach-Object { $_.data }) -join ' | '
        [pscustomobject]@{ Name = $query.Name; Type = $query.Type; Status = $status; Value = $answers }
    } catch {
        [pscustomobject]@{ Name = $query.Name; Type = $query.Type; Status = 'Resolver unavailable'; Value = '' }
    }
}
