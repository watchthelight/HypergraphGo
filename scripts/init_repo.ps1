Param()

$ErrorActionPreference = "Stop"

if (-not $env:watchthelight) { throw "Set watchthelight environment variable to your GitHub username/org" }
$hypergraphgo = if ($env:hypergraphgo) { $env:hypergraphgo } else { "hypergraph-go" }
$visibility = if ($env:VISIBILITY) { $env:VISIBILITY } else { "public" }

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
  throw "GitHub CLI 'gh' not found. Install it first."
}

# Optional auth via GH_TOKEN
if ($env:GH_TOKEN) {
  $env:GH_TOKEN | gh auth login --with-token | Out-Null
}

# Replace module placeholders
if (Test-Path "go.mod") {
  (Get-Content "go.mod") -replace 'github.com/watchthelight/HypergraphGo', ("github.com/$env:watchthelight/$hypergraphgo") | Set-Content "go.mod"
}
$files = Get-ChildItem -Recurse -File | Where-Object { Select-String -Path $_.FullName -Pattern 'github.com/watchthelight/HypergraphGo' -Quiet }
foreach ($f in $files) {
  (Get-Content $f.FullName) -replace 'github.com/watchthelight/HypergraphGo', ("github.com/$env:watchthelight/$hypergraphgo") | Set-Content $f.FullName
}

git init
git config user.name "watchthelight"
git config user.email "admin@watchthelight.org"
git add .
git commit -m "feat: initial hypergraph library"
git branch -M main

gh hypergraphgo create "$env:watchthelight/$hypergraphgo" --source=. --remote=origin --$visibility --push

git tag v0.1.0
git push --tags

Write-Host "hypergraphgo created: https://github.com/$env:watchthelight/$hypergraphgo"
