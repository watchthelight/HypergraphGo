param(
    [Parameter(Mandatory=$true)]
    [string]$Version,
    [string]$Repo = "watchthelight/HypergraphGo"
)

$checksumsUrl = "https://github.com/$Repo/releases/download/v$Version/checksums.txt"

try {
    $checksums = Invoke-WebRequest -Uri $checksumsUrl -UseBasicParsing | Select-Object -ExpandProperty Content
} catch {
    Write-Error "Failed to download checksums.txt from $checksumsUrl"
    exit 1
}

$lines = $checksums -split "`n"
$checksum = $null
foreach ($line in $lines) {
    if ($line -match "hottgo_${Version}_Windows_amd64\.zip") {
        $checksum = ($line -split " ")[0]
        break
    }
}

if (-not $checksum) {
    Write-Error "Checksum for hottgo_${Version}_Windows_amd64.zip not found"
    exit 1
}

$templatePath = "packaging/chocolatey/tools/chocolateyinstall.ps1.tmpl"
$installScriptPath = "packaging/chocolatey/tools/chocolateyinstall.ps1"

$template = Get-Content $templatePath -Raw
$installScript = $template -replace '\$version\$', $Version -replace '\$sha256_amd64\$', $checksum
Set-Content $installScriptPath $installScript

# Update nuspec version
$nuspecPath = "packaging/chocolatey/hottgo.nuspec"
$nuspec = Get-Content $nuspecPath -Raw
$nuspec = $nuspec -replace '__REPLACE__', $Version
Set-Content $nuspecPath $nuspec

# Pack
choco pack $nuspecPath

# Push
$nupkg = "hottgo.$Version.nupkg"
if (Test-Path $nupkg) {
    choco push $nupkg --api-key $env:CHOCOLATEY_API_KEY
} else {
    Write-Error "Nupkg file $nupkg not found"
    exit 1
}
