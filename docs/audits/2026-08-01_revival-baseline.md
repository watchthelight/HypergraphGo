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

### D6: three proof failure modes report without source lines

Corrected 2026-08-01, same day: the first version of this entry claimed the
executor never uses the AST's `Line` fields. That was wrong. Parse failures
and failing tactics inside the proof loop do come back as
`theorem name (line N): message`, verified by running a broken script
through `hottgo --load`.

The real gap is narrower. The three failure modes after the tactic loop in
`executeTheorem` construct their errors without `Line`: an incomplete proof
(`proof incomplete: 1 goals remaining`), a failed proof-term extraction,
and a failure of the final kernel re-check. All three print with no
location. Incomplete proofs are the most common authoring mistake, so the
omission still hurts, just less broadly than first written.

Evidence: `tactics/script/exec.go`, the `ExecError` values built after the
tactic loop omit `Line` while the two inside the loop set it; reproduced
with a deliberately incomplete theorem through `hottgo --load`.

## Identity

Three names circulate for one project. The repository is `HypergraphGo`. The
README banner, title, and roadmap all say `HoTTGo`. The binaries are `hg` and
`hottgo`. Package identifiers split the same way: Chocolatey and AUR publish
`hypergraphgo`, the Scoop bucket and the Cloudsmith APT repo are named
`hottgo`, and the Homebrew tap installs `hg`.

Code volume settles the question of what the project is. At this commit,
counting tests: `kernel/` is 25.3K lines, `internal/` is 34.5K, `tactics/`
is 8.5K, and `hypergraph/` is 4.2K. The hypergraph library the project
started as is about six percent of it.

Recommendation: HoTTGo is the identity. Present the repository as a cubical
type theory kernel first, and document the hypergraph library as the
component the project grew out of and still ships. Do not rename the
repository or the module path now: `github.com/watchthelight/HypergraphGo`
is baked into every packaging pipeline, badge, and downstream `go.mod`, and
a rename would break all of them in one stroke. Revisit only alongside a
major version, and treat D3 as the reminder that even letter case in module
paths is load-bearing.

## Audience

The README's own list holds up against what the code provides: PL
researchers who want a cubical kernel without an elaborator wrapped around
it, tool builders who need an embeddable type theory backend, HoTT learners
reading the internals, and Go developers curious about dependent types. The
common thread is people who want the layer below a proof assistant. Serving
them means keeping the kernel small, readable, and honest about its
boundaries, which is also what the defect list above mostly violates.

## North star

The README already states it: small enough to read, complete enough to use.
Concretely, a cubical type theory kernel where univalence computes, HITs
reduce, the kernel re-checks everything that crosses its boundary, and the
whole thing ships as one static binary. Growth should deepen trust in that
kernel (tests, docs that match behavior, sharper errors) rather than widen
scope toward being a proof assistant.

## Major risks

Technical:

- Lint has been off since spring (D4). Cheap drift accumulates invisibly.
- The cumulativity gap (D5) sits at the soundness boundary. Docs promising
  more than the kernel checks is the worst place in the project to
  overclaim.
- Race coverage exists only in CI on Linux; local Windows development
  cannot run `-race` (cgo toolchain path issue).
- Proof-script debugging without line numbers (D6) gets worse as example
  suites grow.

Product:

- The name split taxes every funnel. A reader lands on HoTTGo branding,
  installs a package called `hypergraphgo`, and receives a binary named
  `hg`. Each hop sheds users.
- Six weeks of silence with green tests reads as abandonment from outside.
  Small visible commits are the fix, and they are cheap.
- Single-maintainer history: 306 of roughly 417 commits are one identity.
  The docs and tests are the only onboarding that exists.

## Relationship between HypergraphGo and HoTTGo

They are one project and should stay that way. The hypergraph library is
the origin, works, has tests, and costs little to keep. The kernel is the
point. Keep both CLIs (`hg` for hypergraph operations, `hottgo` for the
kernel), keep one repository and one module path, and make the
documentation stop implying two projects. The identity work in this audit
plus the D1/D3 fixes get most of the way there without breaking a single
downstream consumer.
