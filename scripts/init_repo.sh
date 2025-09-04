#!/usr/bin/env bash
set -euo pipefail

: "${watchthelight:?Set watchthelight env var to your GitHub username/org}"
hypergraphgo="${hypergraphgo:-hypergraph-go}"
VISIBILITY="${VISIBILITY:-public}"

if ! command -v gh >/dev/null 2>&1; then
  echo "GitHub CLI 'gh' not found. Install it first." >&2
  exit 1
fi

# Optional auth via GH_TOKEN
if [[ -n "${GH_TOKEN:-}" ]]; then
  printf "%s" "$GH_TOKEN" | gh auth login --with-token
fi

# Replace module placeholders
if [[ -f go.mod ]]; then
  sed -i.bak "s#github.com/watchthelight/hypergraphgo#github.com/${watchthelight}/${hypergraphgo}#g" go.mod
  rm -f go.mod.bak
fi
# Update any import paths containing the placeholder
grep -rl "github.com/watchthelight/hypergraphgo" . | xargs sed -i.bak "s#github.com/watchthelight/hypergraphgo#github.com/${watchthelight}/${hypergraphgo}#g" || true
find . -name "*.bak" -delete

git init
git config user.name "watchthelight"
git config user.email "admin@watchthelight.org"
git add .
git commit -m "feat: initial hypergraph library"
git branch -M main

# Create hypergraphgo and push
gh hypergraphgo create "${watchthelight}/${hypergraphgo}" --source=. --remote=origin --${VISIBILITY} --push

git tag v0.1.0
git push --tags

echo "hypergraphgo created: https://github.com/${watchthelight}/${hypergraphgo}"
