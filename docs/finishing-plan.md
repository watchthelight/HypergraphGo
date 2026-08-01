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
