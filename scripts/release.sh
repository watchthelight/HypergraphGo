#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<USAGE
Usage: $(basename "$0") [patch|minor|major|X.Y.Z] [--notes FILE]

Bumps VERSION file, creates annotated tag vX.Y.Z, pushes it, and publishes a
GitHub Release so release-published workflows (GoReleaser, Docker, Chocolatey,
Cloudsmith, DMG) fire.

Requires: git, gh (authenticated). Optional: --notes FILE to pass a release
notes file to "gh release create".
USAGE
}

if [[ ${1:-} == "" || ${1:-} == "-h" || ${1:-} == "--help" ]]; then
  usage
  exit 2
fi

arg=$1
shift || true

notes_file=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --notes)
      notes_file=${2:-}
      shift 2
      ;;
    *)
      echo "Unknown flag: $1" >&2
      usage
      exit 2
      ;;
  esac
done

current="0.1.0"
if [[ -f VERSION ]]; then
  current=$(tr -d '\r\n' < VERSION)
fi

new=""
if [[ "$arg" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  new="$arg"
else
  IFS='.' read -r major minor patch <<<"$current"
  case "$arg" in
    patch) patch=$((patch+1)) ;;
    minor) minor=$((minor+1)); patch=0 ;;
    major) major=$((major+1)); minor=0; patch=0 ;;
    *)
      echo "Unknown bump: $arg" >&2
      usage
      exit 2
      ;;
  esac
  new="$major.$minor.$patch"
fi

echo "Bumping version: $current -> $new"
printf "%s\n" "$new" > VERSION

if ! git diff --quiet -- VERSION; then
  git add VERSION
  git commit -m "chore(release): v$new"
else
  echo "VERSION unchanged; skipping commit"
fi

# Push the current branch first so the tag points at a pushed commit
current_branch=$(git rev-parse --abbrev-ref HEAD)
git push origin "$current_branch"

if git rev-parse -q --verify "refs/tags/v$new" >/dev/null; then
  echo "Tag v$new already exists; skipping tag creation"
else
  git tag -a "v$new" -m "Release v$new"
fi

echo "Pushing tag v$new"
git push origin "v$new"

if ! command -v gh >/dev/null 2>&1; then
  cat <<WARN
WARNING: gh CLI not found. Tag was pushed, but the GitHub Release was not
created, so release-published workflows (GoReleaser, Docker, Chocolatey,
Cloudsmith, DMG) will NOT fire automatically.

Create the release manually:
    gh release create v$new --title "v$new" --notes-file RELEASE_NOTES_v$new.md
WARN
  exit 0
fi

title="v$new"
if [[ -n "$notes_file" ]]; then
  echo "Creating GitHub Release v$new with notes from $notes_file"
  gh release create "v$new" --title "$title" --notes-file "$notes_file"
elif [[ -f "RELEASE_NOTES_v$new.md" ]]; then
  echo "Creating GitHub Release v$new with notes from RELEASE_NOTES_v$new.md"
  gh release create "v$new" --title "$title" --notes-file "RELEASE_NOTES_v$new.md"
else
  echo "Creating GitHub Release v$new with auto-generated notes"
  gh release create "v$new" --title "$title" --generate-notes
fi

echo "Done. Watch workflows:"
echo "  gh run list --workflow release.yml"
echo "  gh run list --workflow docker.yml"
echo "  gh run list --workflow chocolatey.yml"
