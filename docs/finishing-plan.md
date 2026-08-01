# Finishing Plan

- Basis: [2026-08-01 revival baseline audit](audits/2026-08-01_revival-baseline.md)
- Starting commit: 41f5718 (v1.9.0)

Each milestone below is a set of small, separately committable changes. A
milestone is done when its exit criterion holds, not when its commits exist.
Defect numbers (D1 through D6) refer to the audit's defect register.

## Milestone 1: truthful surfaces

The first hour a newcomer spends with this repository should contain no
instruction that fails when followed literally.

- Fix `make build` to compile both `bin/hg` and `bin/hottgo` (D1).
- Replace `make kernel-selftest` with a target that runs something real.
  The honest equivalent of a kernel self-test already exists: loading every
  proof file in `examples/proofs` through `hottgo --load` exercises the
  parser, elaborator, tactics, and kernel together (D2).
- Correct the README module install command to the case-exact path
  `github.com/watchthelight/HypergraphGo` (D3).
- Align README, ROADMAP, and DESIGN universe descriptions with kernel
  behavior: the tower is predicative and non-cumulative, and lifting is
  explicit (docs half of D5).

Verification: run every Makefile target; run the corrected `go get` line in
a scratch module; grep the three docs for stale cumulativity claims.

Exit criterion: a fresh reader can follow the README install section and
the Makefile help text end to end without hitting a single broken command.

## Milestone 2: guarded quality gates

Green must mean something again.

- Add kernel tests that lock in universe rejection semantics: a `Type_0`
  inhabitant offered where `Type_1` is expected must fail, at the
  conversion layer and through `Checker.Check`, and the cumulative mode of
  `internal/core.Conv` must stay reachable behind its flag for future work
  (test half of D5).
- Restore the lint step in `go.yml` using a current golangci-lint v2
  release, after a clean local run against the existing `.golangci.yml`
  (D4). If a local run is impossible, land the workflow change only with
  evidence from a CI run on a doc-only push, never blind.

Verification: `go test ./kernel/...` including the new rejection tests;
one full CI run with lint enabled and passing.

Exit criterion: CI runs vet, import boundaries, lint, race tests, and
cubical tests on every push, and all of them gate.

## Milestone 3: proof-author experience

Scope corrected after re-testing D6: tactic failures inside the proof loop
already report `(line N)`. The remaining work is the tail of the pipeline.

- Attach the theorem's line to the three location-free failure modes in
  `executeTheorem`: incomplete proof, failed extraction, failed final
  re-check (D6).
- Add regression tests that pin line reporting for a failing tactic (the
  behavior that already works) and for an incomplete proof (the behavior
  this milestone adds).

Verification: new tests in `tactics/script`; a hand-broken proof file
reports a line for every failure mode through `hottgo --load`.

Exit criterion: no script failure mode anywhere in `tactics/script`
reports without a source location.

## Milestone 4: cut v1.9.1

Not attempted in the current pass; it is the horizon after milestones 1
through 3.

- Collect the fixes above into CHANGELOG.md.
- Verify goreleaser config against the two-binary build.
- Tag, release, and confirm the package pipelines (Homebrew, Scoop,
  Chocolatey, AUR, APT, Docker) pick up the version.

Exit criterion: `hottgo --version` from each package manager reports
1.9.1.
