# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0] - 2025-12-05

### Added
- **Bidirectional Type Checking** (`kernel/check/` package - Phase 3 M5)
  - `Checker` struct with `Synth`, `Check`, and `CheckIsType` public API
  - Full bidirectional typing rules for all term constructors:
    - Synthesis for: `Var`, `Sort`, `Global`, `Pi`, `Sigma`, `Lam` (annotated), `App`, `Fst`, `Snd`, `Let`
    - Checking for: `Lam` (unannotated against Pi), `Pair` (against Sigma)
  - Source position tracking with `Span` type for precise error locations
  - Structured error types with `ErrorKind` categorization and detailed diagnostics:
    - `ErrUnboundVariable`, `ErrTypeMismatch`, `ErrNotAFunction`, `ErrNotAPair`, `ErrNotAType`, `ErrUnknownGlobal`, `ErrCannotInfer`
  - Global environment (`GlobalEnv`) with staged structure:
    - Axioms (type only), Definitions (type + body + transparency), Inductives (type + constructors), Primitives (built-in)
  - Built-in primitives: `Nat`, `zero`, `succ`, `natElim`, `Bool`, `true`, `false`, `boolElim`
  - Integration with existing `kernel/ctx`, `kernel/subst`, `internal/core`, and `internal/eval`
  - Comprehensive test suite (~24 tests + 2 benchmarks):
    - Identity function `λA.λx.x : Π(A:Type).A→A` (success criterion)
    - Composition function, Nat/Bool primitives, type formation, dependent pairs
    - Error case tests with span verification
    - Nil context handling, API coverage, ErrorKind tests

### Added
- **macOS DMG releases** (`.github/workflows/release.yml`)
  - New `build-dmg` job creates `.dmg` installers for macOS (amd64 and arm64)
  - DMGs built natively on macOS runner using `hdiutil`

### Fixed
- **Nil context handling** (`kernel/check/`)
  - Public API methods (`Synth`, `Check`, `CheckIsType`, `InferAndCheck`) now accept nil context
  - Nil context treated as empty context instead of causing panic
- **Removed unused `etaExpand` function** (`internal/core/conv.go`)
  - Dead code cleanup; eta equality uses `etaEqual` instead
- **CI toolchain version mismatch** (`.github/workflows/`)
  - Added `GOTOOLCHAIN: local` to all workflows to prevent Go from auto-downloading newer toolchain versions
  - Fixes version mismatch errors when GitHub Actions has older patch version than latest available

### Changed (Breaking)
- **Go version requirement bumped to 1.25**
  - Updated `go.mod` and all CI workflows from Go 1.22.x to Go 1.25.x
- **Generic constraint changed from `comparable` to `cmp.Ordered`**
  - `Hypergraph[V]`, `Edge[V]`, `Graph[V]` now require `V` to satisfy `cmp.Ordered`
  - Enables efficient native sorting without string conversion
  - Migration: Custom vertex types must be ordered (string, int, float, or underlying ordered type)

### Changed
- **Sorting performance improvements** (hypergraph package)
  - Replaced `fmt.Sprintf` comparisons with `slices.Sort()` and direct `<` comparison
  - Affects: `algorithms.go`, `incidence.go`, `serialize.go`, `transforms.go`
  - Eliminates quadratic string allocations during sorting

### Added
- **Comprehensive documentation**
  - Package-level docs for `eval` package explaining NbE algorithm
  - Algorithm documentation for `GreedyHittingSet`, `EnumerateMinimalTransversals`, `GreedyColoring`
  - Detailed godoc for `Eval`, `Apply`, `Reify` functions
  - Concurrency warning in hypergraph package docs
- **Comprehensive test coverage** (~25 new tests)
  - `hypergraph/edge_cases_test.go`: Empty graphs, vertex/edge removal, traversal edge cases
  - `kernel/subst/edge_cases_test.go`: Nil handling, zero shifts, all term types
  - `internal/eval/edge_cases_test.go`: Nil env/term, Let/Pi/Sigma eval, deep nesting
  - New benchmarks for eval and hypergraph operations

### Fixed
- **Removed panics in substitution functions** (`kernel/subst/subst.go`)
  - `Shift` and `Subst` now return unknown term types unchanged instead of panicking
  - Improves robustness when encountering unexpected AST nodes
- **Nil environment handling in Eval** (`internal/eval/nbe.go`)
  - `Eval` now handles nil environment gracefully by using empty environment
  - Prevents nil pointer dereference when called without environment
- **Edge case in GreedyHittingSet** (`hypergraph/algorithms.go`)
  - Fixed potential use of uninitialized vertex when no vertices have positive degree
  - Changed condition from `maxDeg == 0` to `maxDeg <= 0` to handle empty graphs
- **Pretty printer efficiency** (`internal/eval/pretty.go`)
  - Optimized `writeNeutral` to avoid string reallocation when adding parentheses
- **Unused parameter cleanup** (`internal/ast/print.go`)
  - Removed unused depth parameter from internal `write` function

## [1.2.0] - 2024-12-19

### Added
- **Normalization by Evaluation (NbE) skeleton** in `internal/eval`
  - Semantic domain with Values, Closures, Neutrals, and Environments
  - WHNF + spine representation for stuck computations
  - Reify/reflect infrastructure for Value ↔ Term conversion
  - Closure-based evaluation with de Bruijn environments
- **Definitional equality via NbE** with optional η for Π/Σ
  - `core.Conv(env, t, u, opts)` API for conversion checking
  - `ConvOptions{EnableEta bool}` feature flag (defaults to OFF)
  - η-equality for functions: `f ≡ \x. f x`
  - η-equality for pairs: `p ≡ (fst p, snd p)`
  - Environment support with `core.Env` wrapper
- **Conversion checker** at `core.Conv` with `ConvOptions`
  - Beta reduction normalization: `(\x. x) y ⇓ y`
  - Projection normalization: `fst (pair a b) ⇓ a`, `snd (pair a b) ⇓ b`
  - Sophisticated η-equality checking with proper de Bruijn handling
- **Expanded test suite and benchmarks**
  - 22 new NbE tests covering beta, projections, neutrals, complex terms
  - 15 conversion tests covering β/η equality, error handling, legacy API
  - 9 performance benchmarks showing ~108 ns/op for simple conversions
  - Comprehensive table-driven tests for all conversion scenarios
- **HoTT kernel Phase 2 completion**
  - M3: NbE skeleton with semantic domain
  - M4: Definitional equality checker with η-rules

### Changed
- **Updated README roadmap** to reflect Phase 2 completion
  - Added roadmap progress table showing Phases 0-2 complete
  - Added latest release section highlighting v1.2.0 features
  - Enhanced overview with HoTT kernel description
- **Enhanced core AST integration**
  - NbE reuses existing `ast.Term` constructors
  - Maintained kernel boundary (no Value types leak)
  - Preserved existing API compatibility

### Fixed
- **Minor determinism issues** in normalization tests
  - All tests now produce consistent, deterministic output
  - CI-friendly test execution with stable results
- **Error handling improvements**
  - No panics in kernel paths, graceful error handling throughout
  - Proper nil term handling in conversion checker
  - Robust environment management

### Performance
- **Microbenchmarks under target latency**
  - BenchmarkConv_Simple: ~108 ns/op
  - Fast normalization for beta reductions and projections
  - Efficient η-equality checking when enabled
- **Memory efficiency**
  - Closure-based evaluation minimizes copying
  - Spine representation for neutral terms
  - Environment sharing for performance

### Technical Details
- **Standard library only**: No external dependencies
- **Kernel fence maintained**: Internal types stay internal
- **De Bruijn indices**: Consistent variable representation
- **WHNF strategy**: Weak head normal form with spine application
- **Feature flags**: η-equality behind compile-time options

## [1.1.0] - Previous Release
- Hypergraph algorithms and CLI tools
- Basic AST and parsing infrastructure
- Initial project structure

## [1.0.0] - Initial Release
- Core hypergraph data structures
- Basic operations and transforms
- CLI interface foundation
