#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<USAGE
Usage: $(basename "$0") [patch|minor|major|X.Y.Z]

Bumps VERSION file, creates tag vX.Y.Z, and pushes the tag.
USAGE
}

if [[ ${1:-} == "" ]]; then
  usage
  exit 2
fi

current="0.1.0"
if [[ -f VERSION ]]; then
  current=$(tr -d '\r\n' < VERSION)
fi

arg=$1
new=""

if [[ "$arg" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  new="$arg"
else
  IFS='.' read -r major minor patch <<<"$current"
  case "$arg" in
    patch)
      patch=$((patch+1))
      ;;
    minor)
      minor=$((minor+1))
      patch=0
      ;;
    major)
      major=$((major+1))
      minor=0
      patch=0
      ;;
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

# Commit change if needed
if ! git diff --quiet -- VERSION; then
  git add VERSION
  git commit -m "chore(release): v$new"
else
  echo "VERSION unchanged; skipping commit"
fi

# Create tag if not exists
if git rev-parse -q --verify "refs/tags/v$new" >/dev/null; then
  echo "Tag v$new already exists; skipping tag creation"
else
  git tag -a "v$new" -m "Release v$new"
fi

echo "Pushing tag v$new"
git push origin "v$new"

echo "Done."

