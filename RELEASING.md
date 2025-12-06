# Releasing HoTTGo

Releases are automated via [GoReleaser](https://goreleaser.com/). This document covers the release process.

## Prerequisites

- Push access to `main`
- GoReleaser installed (for local testing)
- Access to release secrets (for maintainers)

## Release Checklist

### 1. Prepare the Release

- [ ] Ensure `main` is clean and passing CI
- [ ] Review `CHANGELOG.md` — move items from `[Unreleased]` to the new version
- [ ] Update version references if needed

### 2. Create the Release

**Linux/macOS/WSL:**

```bash
./scripts/release.sh patch   # or minor | major | 1.2.3
```

**Windows PowerShell:**

```powershell
./scripts/release.ps1 patch  # or minor | major | 1.2.3
```

This script:
1. Updates the `VERSION` file
2. Creates git tag `vX.Y.Z`
3. Pushes the tag to `origin`

### 3. Automated Steps

When the tag is pushed, GitHub Actions will:
1. Run GoReleaser (`.goreleaser.yaml`)
2. Build binaries for all platforms
3. Create GitHub Release with artifacts
4. Build and upload:
   - `.tar.gz` / `.zip` archives
   - `.deb` and `.rpm` packages
   - musl static builds
   - macOS DMG installers

### 4. Post-Release

- [ ] Verify GitHub Release page
- [ ] Verify package manager updates (Homebrew, AUR, Chocolatey)
- [ ] Update docs site if needed

## GoReleaser Configuration

See [`.goreleaser.yaml`](.goreleaser.yaml) for:
- Build matrix (linux, darwin, windows × amd64, arm64)
- Archive formats
- Package manager integrations (brew, scoop, nfpms)
- Changelog generation

## Manual Release (Emergency)

If automation fails:

```bash
# Build locally
goreleaser release --clean --skip=publish

# Then manually upload to GitHub Releases
```

## Versioning

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, backwards compatible

Pre-releases use `-rc.N` suffix: `v1.6.0-rc.1`
