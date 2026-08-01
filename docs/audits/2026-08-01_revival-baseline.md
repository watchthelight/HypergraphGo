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

## Defect register

### D1: `make build` compiles only one of the two binaries

The `build` target produces `bin/hg` and nothing else. `cmd/hottgo` is the
type checker binary that the README documents at length, and no Makefile
target builds it. Anyone following the docs with `make build` ends up
without the tool most of the documentation is about.

Evidence: `Makefile` lines 3-4.

### D2: `make kernel-selftest` invokes a command that does not exist

The target runs `go run ./cmd/hg check --selftest`. The `hg` CLI has no
`check` command (it prints `hg: unknown command 'check'` and its help text),
and the string `selftest` appears nowhere in any Go source file in the tree.
The target has likely been broken since the command surface was reorganized.

Evidence: running the target; `grep -ri selftest --include='*.go'` returns
nothing.

### D3: README module install command fails

The README says `go get github.com/watchthelight/hypergraphgo`. Go module
paths are case-sensitive and the module declares itself as
`github.com/watchthelight/HypergraphGo`, so the documented command fails:

    module declares its path as: github.com/watchthelight/HypergraphGo
            but was required as: github.com/watchthelight/hypergraphgo

Evidence: reproduced with `go get` in a scratch module against
proxy.golang.org.

### D4: CI lint step disabled behind a stale TODO

`go.yml` carries a commented-out lint step with the note "Re-enable when
golangci-lint supports Go 1.25". Current golangci-lint v2 releases support
Go 1.25 and 1.26, and the repository's `.golangci.yml` is already written
for the v2 config format. Lint has been off while the code kept changing,
which is how drift accumulates unnoticed.

Evidence: `.github/workflows/go.yml`, commented block after the
import-boundary step.

### D5: docs claim cumulative universes; the kernel disables them

README ("Predicative cumulative tower"), ROADMAP line 107, and DESIGN line 11
all describe the universe hierarchy as cumulative. The kernel constructs
every `Checker` with `CumulativeUniv: false`, and the comment on
`NewChecker` explains why: path endpoint comparison needs exact equality,
and folding subtyping into conversion without a separate judgment would be
unsound. The conversion engine supports cumulativity behind a flag
(`internal/core/conv.go`), but nothing in the kernel turns it on. The tower
itself is real (`Type : Sort 1` synthesizes), so the false part is
specifically the claim that `Type_i` inhabitants are accepted at `Type_j`
for `j > i`.

Evidence: `kernel/check/check.go` lines 52-75; `internal/core/conv.go`
lines 28-54.

### D6: tactic errors drop source line numbers

Every statement in the proof-script AST carries a `Line` field, and the
comments in `tactics/script/ast.go` say it exists for error reporting. The
executor never uses it: failures come back as bare strings like
`unknown tactic: foo` or `exact requires a term argument`. A failing
300-line proof file reports no location at all, which turns script
debugging into bisection by hand.

Evidence: `tactics/script/ast.go` lines 53-75; `tactics/script/exec.go`
lines 310-399.
