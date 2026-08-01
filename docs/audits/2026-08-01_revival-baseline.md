# Revival Baseline Audit

- Date: 2026-08-01
- Commit: 41f57189642a93d956f252ac655c7ea415877694 (v1.9.0)
- Toolchain: go1.26.3 windows/amd64

## Purpose

Nothing was pushed to this repository between 2026-06-21 and this audit. Before
new work lands, this document records what builds, tests, and verifies at the
commit above, so the finishing work that follows starts from a measured floor
instead of a remembered one.

## Commands and results

All commands ran from a fresh clone on Windows 11 with Go 1.26.3.

| Command | Result |
|---------|--------|
| `go build ./...` | pass |
| `go vet ./...` | pass |
| `go test ./...` | pass: 16 packages with tests, no failures |
| `go test -tags cubical ./...` | pass |
| `go test -race ./...` | did not run locally; see race note below |
| `bash scripts/check-imports.sh` | pass: all import boundary checks |
| `hottgo --load` over `examples/proofs` | 20 of 20 files verify |
| `golangci-lint run` | not runnable: binary absent here and the CI lint step is disabled (defect D4) |

Race note: this Windows machine has a C compiler whose install path breaks
go's cgo invocation (the path truncates at the space in `C:\Program Files`),
and `-race` requires cgo. The `go.yml` workflow runs the race suite on
ubuntu-latest, so race coverage continues in CI.

The proof suite matches the README's claim exactly: 20 `.htt` files containing
374 `Theorem` declarations, and every file verifies through `hottgo --load`.

## What the results say

The core is in better shape than the six-week silence suggests. Every package
builds and tests clean, the cubical feature set passes under its build tag, and
all 374 example theorems check. The remaining defects sit in the surfaces
around the kernel: build targets, install instructions, disabled lint, doc
claims, and error reporting. The next sections list them with evidence.
