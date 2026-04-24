Param(
  [Parameter(Mandatory = $true)]
  [string]$Bump,

  [Parameter(Mandatory = $false)]
  [string]$NotesFile
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Usage {
  Write-Host "Usage: scripts/release.ps1 [patch|minor|major|X.Y.Z] [-NotesFile FILE]"
  Write-Host ""
  Write-Host "Bumps VERSION file, creates annotated tag vX.Y.Z, pushes it, and"
  Write-Host "publishes a GitHub Release so release-published workflows fire."
  Write-Host ""
  Write-Host "Requires: git, gh (authenticated)."
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

& git diff --quiet -- VERSION
if ($LASTEXITCODE -ne 0) {
  & git add VERSION
  & git commit -m "chore(release): v$new"
}
else {
  Write-Host "VERSION unchanged; skipping commit"
}

# Push the current branch first so the tag points at a pushed commit
$branch = (& git rev-parse --abbrev-ref HEAD).Trim()
& git push origin $branch

& git rev-parse -q --verify "refs/tags/v$new" | Out-Null
if ($LASTEXITCODE -ne 0) {
  & git tag -a "v$new" -m "Release v$new"
}
else {
  Write-Host "Tag v$new already exists; skipping tag creation"
}

Write-Host "Pushing tag v$new"
& git push origin "v$new"

$gh = Get-Command gh -ErrorAction SilentlyContinue
if (-not $gh) {
  Write-Warning "gh CLI not found. Tag pushed but GitHub Release NOT created. Run: gh release create v$new --title v$new --notes-file RELEASE_NOTES_v$new.md"
  exit 0
}

$title = "v$new"
if ($NotesFile) {
  Write-Host "Creating GitHub Release v$new with notes from $NotesFile"
  & gh release create "v$new" --title $title --notes-file $NotesFile
}
elseif (Test-Path "RELEASE_NOTES_v$new.md") {
  Write-Host "Creating GitHub Release v$new with notes from RELEASE_NOTES_v$new.md"
  & gh release create "v$new" --title $title --notes-file "RELEASE_NOTES_v$new.md"
}
else {
  Write-Host "Creating GitHub Release v$new with auto-generated notes"
  & gh release create "v$new" --title $title --generate-notes
}

Write-Host 'Done. Watch workflows:'
Write-Host '  gh run list --workflow release.yml'
Write-Host '  gh run list --workflow docker.yml'
Write-Host '  gh run list --workflow chocolatey.yml'
