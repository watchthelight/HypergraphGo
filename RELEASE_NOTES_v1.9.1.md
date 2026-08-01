# HoTTGo v1.9.1

Patch release. v1.9.0 shipped a healthy kernel surrounded by surfaces that
had drifted: build targets that skipped a binary, install commands that
failed as written, documentation that promised more than the kernel checks,
and a CI lint gate that had been off for months. This release repairs those
surfaces, sharpens proof-script diagnostics, and tightens the release
pipeline itself. No kernel semantics changed.

## Highlights

### Truthful build and install surfaces

- `make build` now produces both `bin/hg` and `bin/hottgo`. It previously
  built only `hg`, even though `hottgo` is the tool most of the
  documentation covers.
- `make kernel-selftest` verifies all 20 proof files in `examples/proofs`
  through `hottgo --load`. The old recipe invoked `hg check --selftest`, a
  command that never existed.
- The documented Go module command uses the case-exact path
  `go get github.com/watchthelight/HypergraphGo`. The lowercase form fails
  with a module path mismatch.

### Documentation matches the kernel

- README, ROADMAP, and DESIGN describe the universe hierarchy as it is:
  predicative and non-cumulative, with explicit lifting. A `Type0`
  inhabitant is not accepted at `Type1`.
- The documentation landing page now leads with the kernel: what HoTTGo
  is, how to try it, and where the hypergraph library fits. It previously
  still described v1.7.0.

### Better proof-script diagnostics

Incomplete proofs, failed proof-term extractions, and failed final
re-checks now report the theorem's source line. Tactic failures inside the
proof loop already carried line numbers; these three tail failures were the
only locations without one. New regression tests pin both behaviors.

### Correctness locked by tests

Four new kernel tests pin universe semantics: the tower step
(`Sort 0 : Sort 1`) checks, sort lifting and inhabitant lifting are
rejected, and the conversion engine's cumulative flag keeps its documented
directional meaning for future subtyping work.

### Build and CI

- The golangci-lint gate is back in the Go workflow after a clean local
  run; the tactics package was brought to zero findings first.
- The Codecov step names its input correctly (`files`), clearing a
  persistent workflow warning.
- macOS disk images now contain both binaries; the DMG job previously
  packaged only `hg`.
- The release workflow pins the GoReleaser 2.17 line that this release was
  rehearsed with, instead of floating on `latest`.

## Breaking changes

None. No exported Go API changed, the proof-script format is unchanged,
and all 374 example theorems verify with this release exactly as they did
with v1.9.0. Error messages for incomplete proofs and tail failures now
include a `(line N)` prefix, which only affects output that previously had
no location.

## Install / upgrade

```bash
# Homebrew (macOS/Linux)
brew update && brew install watchthelight/tap/hg

# Scoop (Windows)
scoop update && scoop install hg

# APT (Debian/Ubuntu - focal/jammy/noble/bookworm)
sudo apt update && sudo apt install hypergraphgo

# Chocolatey (Windows)
choco upgrade hypergraphgo

# Docker (multi-arch)
docker pull ghcr.io/watchthelight/hypergraphgo:1.9.1

# Go module
go get github.com/watchthelight/HypergraphGo@v1.9.1
```

## Verification

```bash
hg --version       # hg 1.9.1 (<commit>, <date>)
hottgo --version   # hottgo 1.9.1 (<commit>, <date>)

# Verify the example proof suite (374 theorems across 20 files):
for f in examples/proofs/**/*.htt; do hottgo --load "$f"; done

# Verify a download against checksums.txt:
sha256sum -c checksums.txt --ignore-missing
```

## Artifacts

Each archive contains both binaries plus LICENSE.md and README.md:

- `hg_1.9.1_linux_amd64.tar.gz`, `hg_1.9.1_linux_arm64.tar.gz`
- `hg_1.9.1_linux_amd64_musl.tar.gz`, `hg_1.9.1_linux_arm64_musl.tar.gz`
- `hg_1.9.1_darwin_amd64.tar.gz`, `hg_1.9.1_darwin_arm64.tar.gz`
- `hg_1.9.1_windows_amd64.zip`, `hg_1.9.1_windows_arm64.zip`
- `hg_1.9.1_darwin_amd64.dmg`, `hg_1.9.1_darwin_arm64.dmg`
- `hypergraphgo` `.deb` and `.rpm` packages (amd64, arm64)
- `checksums.txt`
- Docker: `ghcr.io/watchthelight/hypergraphgo:1.9.1` (linux/amd64, linux/arm64)

## Known limitations

- Universe cumulativity is not implemented as subtyping. The tower is
  predicative and non-cumulative; lifting is explicit. This is a design
  decision documented in DESIGN.md, not an omission.
- The AUR `PKGBUILD` in `packaging/arch` updates through a separate manual
  flow and may lag this release until that pass happens.
- Chocolatey moderation can hold the package in review after publication;
  the version appears once moderation completes.
