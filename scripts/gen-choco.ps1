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
$zipName = $null
foreach ($line in $lines) {
    if ($line -match "_windows_amd64\.zip$") {
        $parts = $line -split " "
        $checksum = $parts[0]
        $zipName = $parts[1]
        break
    }
}

if (-not $checksum -or -not $zipName) {
    Write-Error "Checksum for Windows amd64 zip not found in checksums.txt"
    exit 1
}

$templatePath = "packaging/chocolatey/tools/chocolateyinstall.ps1.tmpl"
$installScriptPath = "packaging/chocolatey/tools/chocolateyinstall.ps1"

$template = Get-Content $templatePath -Raw
$installScript = $template -replace '\$version\$', $Version -replace '\$sha256_amd64\$', $checksum -replace '\$zipName\$', $zipName
Set-Content $installScriptPath $installScript

# Update nuspec version
$nuspecPath = "packaging/chocolatey/hg.nuspec"
$nuspec = Get-Content $nuspecPath -Raw
$nuspec = $nuspec -replace '__REPLACE__', $Version
Set-Content $nuspecPath $nuspec

# Pack
choco pack $nuspecPath

# Push
$nupkg = "hg.$Version.nupkg"
if (Test-Path $nupkg) {
    choco push $nupkg --api-key $env:CHOCOLATEY_API_KEY
} else {
    Write-Error "Nupkg file $nupkg not found"
    exit 1
}
