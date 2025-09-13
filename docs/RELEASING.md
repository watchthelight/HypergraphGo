# Releasing

We use **Semantic Versioning**. Until v1, treat the API as evolving (v0.x.y).

## One-time setup
- Ensure `go.mod` module path is `github.com/watchthelight/HypergraphGo`.
- Ensure CI is green.

## Cut a release
```bash
# Update changelog in the commit (optional)
git switch -c release/v0.1.0
# bump README status if needed, commit, push, open PR, merge

# Tag and push:
git tag -a v0.1.0 -m "v0.1.0: first pre-release (CLI: hottgo)"
git push origin v0.1.0
GitHub Actions runs release.yml, builds multi-arch binaries, uploads checksums and an SBOM, and publishes a GitHub Release.

Install methods
Stable binaries: download from Releases (checksums provided)

Go users: go install github.com/watchthelight/HypergraphGo/cmd/hg@latest (produces hottgo)
