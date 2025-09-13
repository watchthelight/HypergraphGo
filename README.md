# HypergraphGo

[![CI](https://github.com/watchthelight/HypergraphGo/actions/workflows/ci.yml/badge.svg)](https://github.com/watchthelight/HypergraphGo/actions/workflows/ci.yml)
![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-informational)
![Platforms](https://img.shields.io/badge/CI-ubuntu%20%7C%20windows-success)

A work-in-progress implementation of a **native HoTT kernel in Go**. This project began as a hypergraph library and is expanding into a full proof kernel with dependent types.

---

## Project Status

- ✅ **Phase 1:** Core AST, raw AST, resolver, pretty-printer, ctx + capture-avoiding subst with tests
- ✅ **CI:** Cross-OS matrix (Ubuntu + Windows) running `go vet` and `go test`
- 🚧 **Phase 2:** Definitional equality via NbE; normalization and α-equivalence; tests and golden NFs
- ⏭ Future: inductives, paths, univalence, HITs, elaborator, stdlib

See [`TODO.txt`](./TODO.txt) for the detailed roadmap.

---

## Repo Layout

internal/ast/ # Core & Raw AST, printer, resolver, tests
internal/core/ # Definitional equality (Conv) and tests
internal/eval/ # NbE evaluator + tests/bench
internal/util/ # Utilities
kernel/ctx/ # Bindings/ctx utilities + tests
kernel/subst/ # Shift/Subst (de Bruijn) + tests
cmd/hg/ # CLI entry point (WIP)
docs/ # DESIGN.md, CONTRIBUTING.md

---

## Screenshots

*Placeholder for demo screenshots of the HoTT kernel in action, CLI output, etc.*

---

## Install

**Option A: Go users**
```bash
go install github.com/watchthelight/HypergraphGo/cmd/hg@latest
hottgo -version
```

**Option B: Download a release**  
Grab the latest binaries from the [Releases](https://github.com/watchthelight/HypergraphGo/releases) page. Each archive includes a checksum and SBOM.

**Versioning:** we use SemVer. Pre-1.0 versions may contain breaking changes in minor bumps.

---

## Building & Testing

You’ll need **Go 1.22+**.

```bash
git clone https://github.com/watchthelight/HypergraphGo.git
cd HypergraphGo
go test ./...
```

Useful variants:

```bash
# Vet + test with coverage output
go vet ./... && go test ./... -count=1 -v -coverprofile=coverage.out
go tool cover -func=coverage.out   # show totals
go tool cover -html=coverage.out   # open HTML report
```

CI publishes `coverage.out` as a build artifact for each run.

---

## Contributing

Keep the kernel (`internal/*`, `kernel/*`) panic-free and minimal.

Any sugar/tactics live outside and are checked again at the kernel boundary.

See [`docs/CONTRIBUTING.md`](docs/CONTRIBUTING.md) for invariants and hygiene.

---

## License

MIT. See LICENSE.
