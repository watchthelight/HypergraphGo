# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
