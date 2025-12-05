# Platform Packaging Guide

This guide covers packaging HypergraphGo for various platforms and package managers.

## Static Binaries (GitHub Releases)

Pre-built binaries for all platforms. No package manager required—download, extract, run.

**Why this matters**: Users on any platform can install without admin rights or package manager setup.

**Files**: Handled by `.goreleaser.yaml` (already configured).

**Artifacts produced**:
```
hg_{{VERSION}}_linux_amd64.tar.gz
hg_{{VERSION}}_linux_arm64.tar.gz
hg_{{VERSION}}_darwin_amd64.tar.gz
hg_{{VERSION}}_darwin_arm64.tar.gz
hg_{{VERSION}}_windows_amd64.zip
hg_{{VERSION}}_windows_arm64.zip
checksums.txt
```

**Publish**: Automatic on `git push origin vX.Y.Z`. GoReleaser uploads to GitHub Releases.

**Security**: SHA-256 checksums in `checksums.txt`. Consider GPG-signing releases for enterprise users.

---

## Homebrew (macOS & Linux)

The standard package manager for macOS. Also works on Linux via Linuxbrew.

**Why this matters**: Most macOS developers expect `brew install`.

**Files**: `Formula/hg.rb` (for a tap) or PR to `homebrew-core`.

**Location**: Create `Formula/hg.rb` in repo root, or maintain a separate tap repo.

### Minimal Formula

```ruby
# Formula/hg.rb
class Hg < Formula
  desc "Hypergraph & HoTT tooling in Go"
  homepage "https://github.com/watchthelight/HypergraphGo"
  version "{{VERSION}}"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/watchthelight/HypergraphGo/releases/download/v{{VERSION}}/hg_{{VERSION}}_darwin_arm64.tar.gz"
      sha256 "{{SHA256_DARWIN_ARM64}}"
    else
      url "https://github.com/watchthelight/HypergraphGo/releases/download/v{{VERSION}}/hg_{{VERSION}}_darwin_amd64.tar.gz"
      sha256 "{{SHA256_DARWIN_AMD64}}"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/watchthelight/HypergraphGo/releases/download/v{{VERSION}}/hg_{{VERSION}}_linux_arm64.tar.gz"
      sha256 "{{SHA256_LINUX_ARM64}}"
    else
      url "https://github.com/watchthelight/HypergraphGo/releases/download/v{{VERSION}}/hg_{{VERSION}}_linux_amd64.tar.gz"
      sha256 "{{SHA256_LINUX_AMD64}}"
    end
  end

  def install
    bin.install "hg"
  end

  test do
    system "#{bin}/hg", "-version"
  end
end
```

**Publish**:
- **Tap** (recommended initially): Create `watchthelight/homebrew-tap` repo, add formula, users run `brew tap watchthelight/tap && brew install hg`.
- **homebrew-core**: Submit PR once project has significant usage. Requires 50+ GitHub stars and active maintenance.

**Security**: Homebrew verifies SHA-256. No additional signing required for taps.

---

## Chocolatey (Windows)

The most popular Windows package manager. Installs via `choco install hypergraphgo`.

**Why this matters**: Windows developers expect Chocolatey for CLI tools.

**Files** (already exist):
```
packaging/chocolatey/
├── hypergraphgo.nuspec
└── tools/
    ├── chocolateyinstall.ps1.tmpl
    └── chocolateyuninstall.ps1
```

### Install Script

```powershell
# packaging/chocolatey/tools/chocolateyinstall.ps1
$ErrorActionPreference = 'Stop'

$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$version = '{{VERSION}}'

# Windows amd64 only (arm64 when Chocolatey supports it)
$url = "https://github.com/watchthelight/HypergraphGo/releases/download/v$version/hg_${version}_windows_amd64.zip"
$checksum = '{{SHA256_WINDOWS_AMD64}}'

$packageArgs = @{
  packageName   = $env:ChocolateyPackageName
  unzipLocation = $toolsDir
  url           = $url
  checksum      = $checksum
  checksumType  = 'sha256'
}

Install-ChocolateyZipPackage @packageArgs
```

**Publish**:
```bash
cd packaging/chocolatey
choco pack
choco push hypergraphgo.{{VERSION}}.nupkg --source https://push.chocolatey.org/ --api-key $CHOCO_API_KEY
```

**Security**: Chocolatey verifies SHA-256. Packages are moderated before public listing.

---

## Winget (Windows Package Manager)

Microsoft's built-in package manager (Windows 10+). Installs via `winget install watchthelight.hg`.

**Why this matters**: Pre-installed on modern Windows—no Chocolatey setup needed.

**Files**:
```
packaging/winget/
└── watchthelight.hg.yaml
```

### Manifest (Multi-arch)

```yaml
# packaging/winget/watchthelight.hg.yaml
PackageIdentifier: watchthelight.hg
PackageVersion: "{{VERSION}}"
PackageLocale: en-US
Publisher: watchthelight
PackageName: HypergraphGo
License: MIT
ShortDescription: Hypergraph & HoTT tooling in Go
PackageUrl: https://github.com/watchthelight/HypergraphGo
Installers:
  - Architecture: x64
    InstallerUrl: https://github.com/watchthelight/HypergraphGo/releases/download/v{{VERSION}}/hg_{{VERSION}}_windows_amd64.zip
    InstallerSha256: {{SHA256_WINDOWS_AMD64}}
    InstallerType: zip
    NestedInstallerType: portable
    NestedInstallerFiles:
      - RelativeFilePath: hg.exe
        PortableCommandAlias: hg
  - Architecture: arm64
    InstallerUrl: https://github.com/watchthelight/HypergraphGo/releases/download/v{{VERSION}}/hg_{{VERSION}}_windows_arm64.zip
    InstallerSha256: {{SHA256_WINDOWS_ARM64}}
    InstallerType: zip
    NestedInstallerType: portable
    NestedInstallerFiles:
      - RelativeFilePath: hg.exe
        PortableCommandAlias: hg
ManifestType: singleton
ManifestVersion: 1.6.0
```

**Publish**: Fork [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs), add manifest under `manifests/w/watchthelight/hg/{{VERSION}}/`, submit PR.

**Security**: Winget verifies SHA-256. Microsoft reviews PRs before merging.

---

## RPM (Red Hat / Fedora)

Native package format for RHEL, Fedora, CentOS, Rocky Linux. Installs via `dnf install hypergraphgo`.

**Why this matters**: Enterprise Linux users expect RPM packages.

**Files**:
```
packaging/rpm/
└── hypergraphgo.spec
```

### Spec File

```spec
# packaging/rpm/hypergraphgo.spec
Name:           hypergraphgo
Version:        {{VERSION}}
Release:        1%{?dist}
Summary:        Hypergraph & HoTT tooling in Go
License:        MIT
URL:            https://github.com/watchthelight/HypergraphGo
Source0:        https://github.com/watchthelight/HypergraphGo/releases/download/v%{version}/hg_%{version}_linux_amd64.tar.gz

%description
CLI tool for hypergraph manipulation and HoTT type theory.

%prep
%setup -q -c

%install
install -Dm755 hg %{buildroot}%{_bindir}/hg

%files
%{_bindir}/hg
%license LICENSE
%doc README.md

%changelog
* %(date "+%a %b %d %Y") watchthelight <admin@watchthelight.org> - {{VERSION}}-1
- Release {{VERSION}}
```

**Publish**:
- **COPR** (recommended): Create project at [copr.fedorainfracloud.org](https://copr.fedorainfracloud.org), upload `.spec`, users add repo via `dnf copr enable watchthelight/hypergraphgo`.
- **Packagecloud**: `package_cloud push watchthelight/hypergraphgo/el/8 hypergraphgo-{{VERSION}}.rpm`

**Security**: Sign packages with GPG. Add public key to repo metadata.

---

## musl / Alpine Static Builds

Fully static binaries using musl libc. Run on Alpine Linux and minimal containers without glibc.

**Why this matters**: Essential for Docker/Kubernetes users building minimal images.

**Build**: Add musl target to `.goreleaser.yaml`:

```yaml
# Add to builds section in .goreleaser.yaml
builds:
  - id: hg-musl
    main: ./cmd/hg
    binary: hg
    env:
      - CGO_ENABLED=0
    goos: [linux]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w -extldflags "-static"
    tags:
      - netgo
      - osusergo

archives:
  - id: musl
    builds: [hg-musl]
    format: tar.gz
    name_template: "hg_{{ .Version }}_linux_{{ .Arch }}_musl"
```

**Artifacts**:
```
hg_{{VERSION}}_linux_amd64_musl.tar.gz
hg_{{VERSION}}_linux_arm64_musl.tar.gz
```

**Publish**: Automatic with GoReleaser. For Alpine's official repos, submit to [aports](https://gitlab.alpinelinux.org/alpine/aports).

**Security**: Same checksums as other builds. Static linking eliminates shared library vulnerabilities.

---

## GitHub Actions: Manual Build Workflow

If not using GoReleaser, here's a minimal workflow:

```yaml
# .github/workflows/release-manual-build.yml
name: Build Release
on:
  push:
    tags: ['v*']

jobs:
  build:
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            goos: linux
            goarch: amd64
            ext: tar.gz
          - os: ubuntu-latest
            goos: linux
            goarch: arm64
            ext: tar.gz
          - os: macos-latest
            goos: darwin
            goarch: amd64
            ext: tar.gz
          - os: macos-latest
            goos: darwin
            goarch: arm64
            ext: tar.gz
          - os: windows-latest
            goos: windows
            goarch: amd64
            ext: zip
          - os: windows-latest
            goos: windows
            goarch: arm64
            ext: zip
    runs-on: ${{ matrix.os }}
    env:
      GOTOOLCHAIN: local
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          CGO_ENABLED: 0
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          go build -ldflags="-s -w" -o hg${{ matrix.goos == 'windows' && '.exe' || '' }} ./cmd/hg

      - name: Package
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          NAME="hg_${VERSION}_${{ matrix.goos }}_${{ matrix.goarch }}"
          if [ "${{ matrix.ext }}" = "zip" ]; then
            zip "${NAME}.zip" hg.exe LICENSE.md README.md
          else
            tar -czvf "${NAME}.tar.gz" hg LICENSE.md README.md
          fi
        shell: bash

      - uses: actions/upload-artifact@v4
        with:
          name: hg-${{ matrix.goos }}-${{ matrix.goarch }}
          path: hg_*.${{ matrix.ext }}

  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/download-artifact@v4
        with:
          path: dist
          merge-multiple: true

      - name: Generate checksums
        run: |
          cd dist
          sha256sum hg_* > checksums.txt

      - name: Upload to release
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          TAG=${GITHUB_REF#refs/tags/}
          gh release upload "$TAG" dist/* --clobber
```

### GoReleaser Alternative (Recommended)

GoReleaser handles cross-compilation, packaging, checksums, and publishing in one tool. The existing `.goreleaser.yaml` already produces all artifacts. To add RPM support:

```yaml
# Add to .goreleaser.yaml
nfpms:
  - id: rpm
    formats: [rpm]
    package_name: hypergraphgo
    vendor: watchthelight
    maintainer: "watchthelight <admin@watchthelight.org>"
    description: "Hypergraph & HoTT tooling in Go"
    license: MIT
```

---

## Maintainer Checklist

### Files to add/update
- [ ] `Formula/hg.rb` — Homebrew formula
- [ ] `packaging/winget/watchthelight.hg.yaml` — Winget manifest
- [ ] `packaging/rpm/hypergraphgo.spec` — RPM spec
- [ ] Update `.goreleaser.yaml` — Add musl builds, RPM via nfpms

### CI/Release steps
- [ ] Verify `go test ./...` passes
- [ ] Tag release: `git tag -a vX.Y.Z -m "vX.Y.Z"`
- [ ] Push tag: `git push origin vX.Y.Z`
- [ ] Verify GoReleaser artifacts on GitHub Releases
- [ ] Check `checksums.txt` uploaded

### Package registry submissions
- [ ] **Homebrew tap**: Push formula to `watchthelight/homebrew-tap`
- [ ] **Chocolatey**: Run `choco pack && choco push`
- [ ] **Winget**: PR to `microsoft/winget-pkgs`
- [ ] **COPR/RPM**: Upload spec or trigger build
- [ ] **AUR**: Update PKGBUILD, push to AUR (existing `packaging/aur/`)

### Security
- [ ] Verify SHA-256 checksums match artifacts
- [ ] (Optional) GPG-sign tags: `git tag -s vX.Y.Z`
- [ ] (Optional) GPG-sign release artifacts for enterprise users
- [ ] (Optional) Windows: Code-sign `.exe` with Authenticode certificate
