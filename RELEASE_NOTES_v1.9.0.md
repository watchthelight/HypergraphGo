# HoTTGo v1.9.0

Proof scripts, REPL proof mode, inductive tactics, implicit arguments,
and **374 verified example theorems**. This release consolidates Phase 9
(Standard Library & Proof Mode) and Phase 10 (Usability Improvements)
into a single minor version — both phases are additive and backwards
compatible.

## Highlights

### Proof scripts (`.htt`) and `--load`

Full tactic-script executor with a first-class file format:

```
Theorem pathp_refl : (Pi A Type (Pi x (Var 0) (PathP (Var 1) (Var 0) (Var 0))))
Proof
  intro A
  intro x
  exact (PathLam i (Var 0))
Qed
```

- `hottgo --load file.htt` verifies every `Theorem` in the file, reports
  `✓`/`✗` per theorem, and exits non-zero on any failure.
- `Definition name : TYPE := TERM` and `Axiom name : TYPE` items are
  processed in order so later theorems can reference earlier content.
- Comment syntax (`--`) and line-number tracking for errors.

### REPL proof mode

`hottgo` is now a real interactive prover, not just a checker.

- `:prove TYPE` / `:prove NAME : TYPE` — enter proof mode
- `:tactic NAME ...` (or just `NAME ...`) — apply tactics
- `:goal` / `:goals` / `:undo` / `:qed` / `:abort`
- `:history`, `:checkpoint NAME`, `:restore NAME`, `:checkpoints`
- Context-aware printing in goals: `(Pi n Nat (Id Nat n n))` instead of
  `(Id Nat (Var 0) (Var 0))`
- Dynamic prompt shows remaining goal count: `proof[N]>`
- `:set named on|off`, `:set verbose on|off`, `:settings`
- `:env`, `:print NAME`, `:search PATTERN` for environment inspection

### Standard library and inductive tactics

Built-in `Unit`, `Empty`, `Sum`, and `List` types with their eliminators,
plus inductive-aware tactics:

- `induction HYP` — induction on `Nat`/`List` (generates IH subgoals)
- `destruct HYP` — case analysis on `Sum`/`Bool`
- `cases HYP` — non-recursive case analysis on `Nat`/`List`/`Bool`/`Sum`
- `left`, `right`, `contradiction`, `constructor`, `exists`
- `symmetry`, `transitivity`, `ap`, `transport`, `path-app-at`
- `unfold NAME` to unfold definitions in the goal

Error messages now carry a structured `TacticError` with the current goal,
expected/actual types, and hints.

### Implicit arguments and surface inductives

- `{x : A}` Pi binders and `{x}` implicit lambdas
- Surface inductive syntax: `(Nat.zero)`, `(Nat.succ n)`, `(natElim ...)`
- Implicit applications elaborate through metavariables via Miller
  pattern unification.

### Performance

- NbE evaluation cache with pointer-identity keys and a configurable
  size bound.
- `ConvCached` / `ConvContext` / `ConvAllCached` for batched conversion
  checks that share a cache.
- AST substitution fast path for closed terms (`substClosed`,
  `substClosedExtension` for cubical), with early exit in `Shift`.

### Testing and CI

- **374 theorems** across 20 proof files in `examples/proofs/`,
  covering paths, funext, equivalences, univalence, HITs,
  truncation/h-levels, and Peano/groups as integration tests.
- Fuzz targets in CI (parser, script parser, hypergraph JSON).
- Import-boundary check (`scripts/check-imports.sh`) in CI to enforce
  the kernel purity rule.

## Breaking changes

None. All additions are backwards compatible with v1.8.x.

## Install / upgrade

```bash
# Homebrew (macOS/Linux)
brew update
brew install watchthelight/tap/hg

# Scoop (Windows)
scoop bucket add watchthelight https://github.com/watchthelight/scoop-bucket
scoop install hg

# APT (Debian/Ubuntu — focal/jammy/noble/bookworm)
curl -1sLf 'https://dl.cloudsmith.io/public/watchthelight/hottgo/setup.deb.sh' | sudo -E bash
sudo apt install hypergraphgo

# Chocolatey (Windows)
choco install hypergraphgo

# Docker (multi-arch)
docker pull ghcr.io/watchthelight/hypergraphgo:1.9.0

# Go module
go install github.com/watchthelight/HypergraphGo/cmd/hg@v1.9.0
go install github.com/watchthelight/HypergraphGo/cmd/hottgo@v1.9.0
```

## Verification

```bash
hg --version           # -> hg 1.9.0 (...)
hottgo --version       # -> hottgo 1.9.0 (...)

# Verify all 374 example theorems:
git clone --branch v1.9.0 https://github.com/watchthelight/HypergraphGo
cd HypergraphGo
go build -o hottgo ./cmd/hottgo
shopt -s globstar
for f in examples/proofs/**/*.htt; do ./hottgo --load "$f"; done

# Or try the REPL:
./hottgo
> :prove (Id Nat zero zero)
proof[1]> reflexivity
proof[0]> :qed
```

## Artifacts

Each release ships:

- `hg_1.9.0_{linux,darwin}_{amd64,arm64}.tar.gz` — contains `hg` and `hottgo`
- `hg_1.9.0_windows_{amd64,arm64}.zip` — contains `hg.exe` and `hottgo.exe`
- `hg_1.9.0_linux_{amd64,arm64}_musl.tar.gz` — fully static builds
- `hypergraphgo_1.9.0_{amd64,arm64}.deb`
- `hypergraphgo_1.9.0_{amd64,arm64}.rpm`
- `hg_1.9.0_darwin_{amd64,arm64}.dmg`
- `checksums.txt` (SHA-256 for every archive above)
- `ghcr.io/watchthelight/hypergraphgo:1.9.0` (linux/amd64, linux/arm64)

## Known limitations

- Universe cumulativity subtyping is implemented but currently disabled
  (`CumulativeUniv: false`) pending proper path-endpoint handling.
- Race-detector builds require CGO; the shipped binaries are
  `CGO_ENABLED=0`.
- Chocolatey package submissions often sit in "Waiting for maintainer
  review" on the community feed for a while after push — this is
  Chocolatey's moderation queue, not a publish failure.
