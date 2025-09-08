Param(
  [Parameter(Mandatory = $true)]
  [string]$Bump
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Usage {
  Write-Host "Usage: scripts/release.ps1 [patch|minor|major|X.Y.Z]"
  Write-Host "Bumps VERSION file, creates tag vX.Y.Z, and pushes the tag."
}

if (-not $Bump) { Usage; exit 2 }

$current = '0.1.0'
if (Test-Path VERSION) {
  $current = (Get-Content VERSION -Raw).Trim()
}

if ($Bump -match '^[0-9]+\.[0-9]+\.[0-9]+$') {
  $new = $Bump
}
else {
  $parts = $current.Split('.')
  if ($parts.Count -lt 3) { throw "Invalid VERSION: $current" }
  $major = [int]$parts[0]
  $minor = [int]$parts[1]
  $patch = [int]$parts[2]
  switch ($Bump) {
    'patch' { $patch += 1 }
    'minor' { $minor += 1; $patch = 0 }
    'major' { $major += 1; $minor = 0; $patch = 0 }
    default { Usage; exit 2 }
  }
  $new = "$major.$minor.$patch"
}

Write-Host "Bumping version: $current -> $new"
Set-Content -Path VERSION -Value $new

# Commit change if needed
& git diff --quiet -- VERSION
if ($LASTEXITCODE -ne 0) {
  & git add VERSION
  & git commit -m "chore(release): v$new"
}
else {
  Write-Host "VERSION unchanged; skipping commit"
}

# Create tag if not exists
& git rev-parse -q --verify "refs/tags/v$new" | Out-Null
if ($LASTEXITCODE -ne 0) {
  & git tag -a "v$new" -m "Release v$new"
}
else {
  Write-Host "Tag v$new already exists; skipping tag creation"
}

Write-Host "Pushing tag v$new"
& git push origin "v$new"
Write-Host 'Done.'

