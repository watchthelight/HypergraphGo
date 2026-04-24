# Releasing HoTTGo

Release automation is driven by the **`release: published`** GitHub event.
Pushing a tag alone does NOT fire the release workflows — you must also create
a GitHub Release for the tag. The helper scripts in `scripts/` do both.

## Prerequisites

- Push access to `main`
- GitHub CLI `gh` installed and authenticated (`gh auth status`)
- GoReleaser v2 installed locally (for `goreleaser check` and dry runs)
- Release secrets configured on the repo:
  - `TAP_GITHUB_TOKEN` — writes to `watchthelight/homebrew-tap`
  - `SCOOP_GITHUB_TOKEN` — writes to `watchthelight/scoop-bucket`
  - `CLOUDSMITH_TOKEN` — pushes `.deb` to Cloudsmith
  - `CHOCOLATEY_API_KEY` — pushes the `hypergraphgo` Chocolatey package
  - `GIST_TOKEN` + `BADGE_GIST_ID` (var) — badge gist updates
  - `GITHUB_TOKEN` is provided automatically

## Release Checklist

### 1. Prepare the release on `main`

- [ ] CI green on `main` (Go, CI Linux, CI Windows, CodeQL)
- [ ] Move items from `[Unreleased]` into `[X.Y.Z] - YYYY-MM-DD` in `CHANGELOG.md`
- [ ] Update `README.md` / `ROADMAP.md` version references if stale
- [ ] Write `RELEASE_NOTES_vX.Y.Z.md` (summary, highlights, breaking changes,
      install/upgrade, verification, artifacts, known limitations)
- [ ] Commit with `chore(release): prepare vX.Y.Z` and push to `main`
- [ ] Run `goreleaser check` locally — should pass

### 2. Cut the release

**Linux/macOS/WSL:**

```bash
./scripts/release.sh minor           # bumps VERSION based on bump keyword
./scripts/release.sh 1.9.0 --notes RELEASE_NOTES_v1.9.0.md   # explicit version
```

**Windows PowerShell:**

```powershell
./scripts/release.ps1 minor
./scripts/release.ps1 1.9.0 -NotesFile RELEASE_NOTES_v1.9.0.md
```

Each script:

1. Updates the `VERSION` file and commits `chore(release): vX.Y.Z` (if needed).
2. Pushes the current branch.
3. Creates the annotated tag `vX.Y.Z` and pushes it.
4. Creates the GitHub Release with `gh release create`, preferring
   `RELEASE_NOTES_vX.Y.Z.md` when present, otherwise auto-generated notes.

The `release: published` event fired by step 4 is what triggers every
downstream workflow.

### 3. Automated steps (triggered by `release: published`)

- `release.yml` → GoReleaser builds binaries, `.tar.gz`/`.zip` archives,
  musl archives, `.deb`/`.rpm` via nfpm, `checksums.txt`, and uploads all
  assets to the GitHub Release. Pushes formula to the Homebrew tap and
  manifest to the Scoop bucket.
- `release.yml` (`build-dmg` job) → builds macOS `.dmg` files from the
  darwin tarballs and uploads them to the release.
- `release.yml` (`publish-deb`) → pushes `.deb` files to Cloudsmith for
  `ubuntu/focal`, `ubuntu/jammy`, `ubuntu/noble`, `debian/bookworm`.
- `docker.yml` → builds multi-arch image (linux/amd64, linux/arm64) and
  pushes to `ghcr.io/watchthelight/hypergraphgo` with tags
  `X.Y.Z`, `X.Y`, `X`, and `latest` on stable releases.
- `chocolatey.yml` → waits for `checksums.txt`, regenerates the nuspec
  and install script, and publishes to community.chocolatey.org.
  (Skipped on prereleases.)

### 4. Watch workflows

```bash
gh run list --workflow release.yml
gh run list --workflow docker.yml
gh run list --workflow chocolatey.yml
gh run list --workflow update-badges.yml
```

### 5. Post-release verification

- [ ] GitHub Release page has all expected assets and matching checksums
- [ ] `docker pull ghcr.io/watchthelight/hypergraphgo:X.Y.Z` works
- [ ] `brew update && brew install watchthelight/tap/hg` works
- [ ] Scoop bucket manifest updated
- [ ] Cloudsmith repo has `.deb` for focal/jammy/noble/bookworm
- [ ] Chocolatey package page shows the new version (status may stay in
      "Submitted" / "Waiting for maintainer review" — normal)
- [ ] Badges gist refreshed by `update-badges.yml`

## Manual re-publish paths

If Cloudsmith publishing fails after artifacts are already on the GitHub
Release, re-run just the deb publish step:

```bash
gh workflow run release.yml --field tag=vX.Y.Z
```

This triggers the `publish-deb-manual` job (`workflow_dispatch` path) which
re-downloads the `.deb` artifacts from the existing release and pushes to
Cloudsmith.

If the whole release failed mid-way but the tag already exists, use:

```bash
gh workflow run release-manual.yml --field tag=vX.Y.Z
```

This checks out the existing tag and runs GoReleaser against it.

## Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features, backwards compatible
- **PATCH**: Bug fixes, backwards compatible

Pre-releases use `-rc.N` suffix (`v1.9.0-rc.1`). GoReleaser marks them as
prerelease automatically via `release.prerelease: auto`. Chocolatey publishing
is skipped for prereleases (see `.github/workflows/chocolatey.yml`).
