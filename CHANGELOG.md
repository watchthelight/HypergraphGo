# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

#### Phase 9: Standard Library & Inductive Tactics (M1-M3)

- **Standard library types** (`kernel/check/stdlib.go`, `kernel/check/stdlib_test.go`)
  - `Unit` type with `tt` constructor and `unitElim` eliminator
  - `Empty` type (no constructors) with `emptyElim` eliminator
  - `Sum` type with `inl`/`inr` constructors and `sumElim` eliminator
  - `List` type with `nil`/`cons` constructors and `listElim` eliminator
  - All types use `DeclareInductive` for proper eliminator generation
  - `NewGlobalEnvWithStdlib()` and `NewCheckerWithStdlib()` convenience functions

- **Inductive tactics** (`tactics/core.go`)
  - `Contradiction()` - proves any goal from Empty hypothesis via `emptyElim`
  - `Left()` - proves Sum goal by providing A witness (uses `inl`)
  - `Right()` - proves Sum goal by providing B witness (uses `inr`)
  - `Destruct(hypName)` - case analysis on Sum or Bool hypothesis
    - For Sum: creates two subgoals with `inl`/`inr` decomposition
    - For Bool: creates two subgoals for `true`/`false` cases
  - `Induction(hypName)` - induction on Nat or List hypothesis
    - For Nat: creates base case (zero) and step case (n, IH)
    - For List: creates nil case and cons case (x, xs, IH)
  - `Cases(hypName)` - non-recursive case analysis (no IH)
    - Works on Nat, List, Bool, and Sum
  - `Constructor()` - applies first applicable constructor
    - For Unit: applies `tt` (completes goal)
    - For Sum: applies `inl` (Left)
    - For List: applies `nil` (completes with empty list)
  - `Exists(witness)` - provides witness for Sigma goal
    - For Σ(x:A).B with witness w, creates subgoal B[w/x]

- **Performance optimization: NbE evaluation caching** (`internal/eval/cache.go`, `internal/eval/nbe_cached.go`)
  - `Cache` struct with memoization for evaluation results
  - Pointer identity-based cache keys for (term, env) pairs
  - Size-limited cache with configurable max entries (default 10,000)
  - Thread-safe with read/write mutex
  - `EvalCached`, `ApplyCached`, `NormalizeWithCache` convenience functions
  - `WithCache[T]` generic helper for scoped cache usage

- **Performance optimization: Cached conversion checking** (`internal/core/conv_cached.go`)
  - `ConvCached` - single conversion with fresh cache
  - `ConvContext` - reusable context for batch conversions
  - `ConvAllCached` - check multiple term pairs with shared cache

- **Performance benchmarks** (`internal/core/perf_bench_test.go`)
  - Church numeral benchmarks (depth 10, 50, 100)
  - Repeated conversion benchmarks
  - Shared subterm benchmarks
  - Alpha-equality benchmarks
  - Cached vs non-cached comparison benchmarks

- **Performance documentation** (`docs/perf.md`)
  - Benchmark results and analysis
  - API reference for caching functions
  - Guidelines for when caching provides benefit

- **Package documentation** (doc.go files for previously undocumented packages)
  - `internal/elab/doc.go` - elaboration pipeline overview
  - `internal/unify/doc.go` - Miller pattern unification
  - `internal/util/doc.go` - generic Set type
  - `internal/version/doc.go` - build information
  - `tactics/doc.go` - Ltac-style proof tactics
  - `tactics/proofstate/doc.go` - proof state management

- **Getting started tutorial** (`docs/getting-started-hottgo.md`)
  - Installation via package managers and from source
  - REPL usage with :eval and :synth commands
  - Type checking files from CLI
  - S-expression syntax reference
  - Go library usage examples
  - Proof tactics tutorial with code samples
  - Error interpretation guide

- **Contributor guardrails** (`CONTRIBUTING.md`)
  - Architecture & Boundaries section with import rules table
  - Where to add new features guide (tactics, syntax, types)
  - Testing Standards section with commands for unit, race, fuzz, and benchmark tests
  - Table-driven test example

- **Import boundary CI check** (`scripts/check-imports.sh`, `.github/workflows/go.yml`)
  - Automated verification of kernel import restrictions
  - Checks that kernel packages don't import parser or tactics
  - Checks that internal packages don't import cmd or tactics
  - Integrated into CI pipeline

- **Fuzz tests for robustness** (`internal/parser/fuzz_test.go`, `hypergraph/fuzz_test.go`)
  - `FuzzParseTerm`, `FuzzParseMultiple`, `FuzzMalformedInput` for S-expression parser
  - `FuzzLoadJSON`, `FuzzLoadJSONMalformed`, `FuzzJSONRoundTrip`, `FuzzHypergraphOperations`, `FuzzJSONSpecialCharacters` for hypergraph JSON loader
  - Comprehensive seed corpora including edge cases, Unicode, control characters, deeply nested structures

- **Property-based tests** (`internal/parser/property_test.go`, `hypergraph/property_test.go`)
  - Parse/format round-trip tests for HoTT terms (basic, cubical)
  - Format idempotency tests
  - Parse stability tests for whitespace and comment variations
  - Hypergraph add/remove vertex/edge invariants
  - Copy independence tests
  - JSON round-trip preservation tests
  - Edge membership and degree consistency tests

- **Extended unify package tests** (`internal/unify/unify_test.go`)
  - Cubical term zonking tests (Partial, System, Comp, HComp, Fill, Glue, GlueElem, Unglue, UA, UABeta, HITApp)
  - Occurs check tests for cubical terms
  - hasMeta tests for cubical terms
  - Coverage improved from 75.5% to 90.3%

### Fixed
- **Tactics proof term construction** (`tactics/core.go`, `tactics/proofstate/state.go`)
  - `Intro`, `Apply`, `Split`, `Rewrite` now properly construct proof terms via metavariable linking
  - Added `SolveGoalWithSubgoals` to link parent goals to child metavariables
  - `ExtractTerm()` now correctly assembles proof terms from tactic applications
  - `Rewrite` and `RewriteRev` now construct proper J terms for transport

- **HITApp alpha-equality** (`internal/eval/alpha_eq.go`)
  - Added missing `HITApp` case to `AlphaEq` function
  - Compares HITName, Ctor, Args, and IArgs for structural equality

- **Thread-safe fluent API** (`tactics/prover.go`)
  - Moved `lastError` from package-level global to instance field in `Prover` struct
  - Multiple concurrent provers no longer share error state

## [1.8.2] - 2026-01-09

### Fixed
- **Tactics soundness improvements** (`tactics/core.go`)
  - `Exact` tactic now type checks the provided term against the goal type
  - `Apply` tactic now verifies the function's codomain matches the goal
  - Added `inferTermType` helper for type inference from hypothesis context
  - Added `containsVar0` helper for dependent type detection

- **Unification cubical term support** (`internal/unify/unify.go`)
  - Added full cubical term support to `zonkTerm` function
  - Added cubical term support to `occurs` check
  - Added cubical term support to `hasMeta` detection
  - Supported term types: Interval, I0, I1, IVar, Face*, Partial, System, Comp, HComp, Fill, Glue, GlueElem, Unglue, UA, UABeta, HITApp
  - Fixed occurs check to handle meta-to-meta unification (`?X = ?X`)

- **Elaboration PathLam endpoint computation** (`internal/elab/elab.go`)
  - `synthPathLam` now correctly computes path endpoints
  - Endpoints computed by substituting i0 and i1 for the interval variable
  - Fixed: `<i> body` now has type `PathP (λi. T) body[i0/i] body[i1/i]`

### Changed
- Updated tactics tests to use type-correct goals (e.g., `Type₁` with `Exact(Type₀)`)

## [1.8.0] - 2026-01-08

### Added
- **Elaboration System** (`internal/elab/`)
  - Surface syntax with implicit arguments and holes (`surface.go`)
  - Metavariable store for type inference (`meta.go`)
  - Elaboration context extending kernel context (`context.go`)
  - Bidirectional elaboration algorithm (`elab.go`)
  - Zonking (metavariable substitution) (`zonk.go`)
  - Comprehensive test suite (`elab_test.go`) - 88.7% coverage

- **Unification** (`internal/unify/`)
  - Miller pattern unification algorithm
  - Pattern inversion with variable shifting
  - Occurs check for cycle detection
  - Constraint solving and deferred constraints
  - Zonk functions for metavariable substitution
  - Comprehensive test suite - 95.0% coverage

- **Surface Syntax Parser** (`internal/parser/surface.go`)
  - Implicit Pi types: `{x : A} -> B`
  - Implicit lambdas: `\{x}. body`
  - Holes: `_` (anonymous) and `?name` (named)
  - Implicit application: `f {arg}`
  - S-expression forms compatibility
  - FormatSurfaceTerm for pretty printing
  - Extended test suite - 86.5% coverage

- **Tactics System** (`tactics/`)
  - Proof state management (`proofstate/state.go`) - 97.5% coverage
  - Tactic type and result (`tactic.go`)
  - Tactic combinators (`combinators.go`):
    - `Seq`: Sequential composition
    - `OrElse`: Try first, fallback to second
    - `Try`: Succeed even on failure
    - `Repeat`: Apply until failure
    - `First`: First successful tactic
    - `All`: Apply to all goals
    - `Focus`: Apply to specific goal
    - `Progress`: Fail if no progress
    - `Complete`: Require full proof
  - Core tactics (`core.go`):
    - `Intro`, `IntroN`, `Intros`: Introduce hypotheses
    - `Exact`: Provide exact proof term
    - `Assumption`: Use hypothesis
    - `Apply`: Apply function to goal
    - `Reflexivity`: Prove reflexivity goals
    - `Split`: Split sigma/product types
    - `Simpl`: Normalize goal
    - `Rewrite`, `RewriteRev`: Rewrite with equality
    - `Trivial`, `Auto`: Automation
  - Prover Go API (`prover.go`):
    - `NewProver`: Create interactive prover
    - Fluent API methods for chaining
    - `Prove`, `MustProve`: Convenience functions
  - Comprehensive test suite - 90.9% coverage

### Tests
- Extensive test coverage for Phase 8 components:
  - `internal/elab`: 88.7% coverage with tests for zonk, meta store, context, and elaboration
  - `internal/unify`: 95.0% coverage with tests for all term types and edge cases
  - `tactics/proofstate`: 97.5% coverage with comprehensive proof state tests
  - `tactics`: 90.9% coverage with tests for combinators and core tactics
  - `internal/parser`: 86.5% coverage for surface syntax parsing

## [1.7.1] - 2026-01-08

### Added
- **"Because I love stats" badge section** in README
  - 24 dynamic badges across 5 rows showing project metrics
  - Row 1: version, coverage, tests, packages, lines of code
  - Row 2: files, go files, source files, test files, folders
  - Row 3: functions, structs, test funcs, benchmarks, dependencies (0!)
  - Row 4: kernel files, internal files, examples, workflows, CLI commands
  - Row 5: commits, releases, age, TODOs (0!)
- Extended `scripts/generate-badges.sh` with 19 new metric calculations
- Extended `.github/workflows/update-badges.yml` with 19 new badge uploads
- **README rewrite** for clarity and voice
  - Added HoTT kernel quickstart with working examples
  - Clarified "Why Go?" rationale
  - Added concrete line counts and architecture breakdown
  - Sharpened cooltt/redtt comparison
  - Expanded audience section

### Changed
- **Comprehensive diagram updates** (`DIAGRAMS.md`)
  - Added Higher Inductive Types (HITs) to master architecture diagrams
  - New Section 13: Higher Inductive Types with 8 mermaid diagrams:
    - HIT Architecture Overview (AST, eval, check, built-ins)
    - HITSpec Structure class diagram
    - HIT Evaluation Flow (endpoint reduction)
    - Built-in HITs (S1, Trunc, Susp, Int, Quot)
    - HIT Eliminator Type Construction
    - HIT Declaration Pipeline sequence diagram
    - RecursorInfo for HITs class diagram
  - Updated Type System Summary to include HITs section
  - Updated Computation Rules with HIT reduction rules
  - Updated GlobalEnv/Inductive class diagram with HIT fields
  - HIT components styled with purple (#8250df) to distinguish from cubical (green)
  - Updated Summary section with cubical and HIT coverage

### Tests

- **Resolve tests** (`internal/ast/resolve_test.go` - extended)
  - Added 40+ tests covering all RTerm cases: Sort, Pi, Sigma, Pair, Fst, Snd, Let, Id, Refl, J
  - Error path coverage for unbound variables in all subterms
  - Coverage for `Resolve` improved from 53.8% to 99.0%
  - Coverage for `internal/ast` improved to 98.8%

- **Cubical type synthesis tests** (`kernel/check/path_test.go` - extended)
  - GlueElem: `TestSynthGlueElem`, `TestSynthGlueElem_WithBranch`, `TestSynthGlueElem_InvalidFace`
  - Unglue: `TestSynthUnglue`, `TestSynthUnglue_FromGlueType`, `TestSynthUnglue_StuckCase`
  - UA: `TestSynthUA`, `TestSynthUA_UniverseMismatch`
  - UABeta: `TestSynthUABeta`, `TestSynthUABeta_NotSigma`
  - Composition operations: `TestSynthFill`, `TestSynthComp`, `TestSynthHComp`
  - Coverage for `synthGlueElem` improved from 0% to 83.3%
  - Coverage for `synthUnglue` improved from 0% to 77.8%
  - Coverage for `kernel/check` improved from 83.8% to 85.5%

- **Recursor buildRecursorCallWithIndices tests** (`internal/eval/recursor_test.go` - extended)
  - Integration tests via tryGenericRecursorReduction
  - Tests for multiple indices, out of bounds, negative position, incomplete metadata
  - Coverage for `buildRecursorCallWithIndices` improved from 36% to 92%
  - Coverage for `internal/eval` improved from 80.8% to 87.6%

- **NbE and equality tests** (`internal/eval/nbe_cubical_test.go`, `internal/eval/equality_test.go` - extended)
  - PathApply tests for VFill, VPath, VPathP, VHITPathCtor
  - EvalCubical tests for HITApp and UABeta
  - ValueEqual tests for VComp, VHComp, VFill, VGlue, VGlueElem, VUA, VUnglue, VHITPathCtor
  - Environment equality tests: envEqual, ienvEqual, closureEqual, faceValueEqual

- **EvalCubical comprehensive tests** (`internal/eval/nbe_cubical_test.go` - extended)
  - Glue/GlueElem: `TestEvalCubical_GlueEmptySystem`, `TestEvalCubical_GlueWithBranches`,
    `TestEvalCubical_GlueElemEmptySystem`, `TestEvalCubical_GlueElemWithBranches`
  - Unglue: `TestEvalCubical_UnglueNilTy`
  - HITApp: `TestEvalCubical_HITAppWithTermArgs`
  - Composition: `TestEvalCubical_CompWithFaceEq`, `TestEvalCubical_HCompWithFaceAnd`,
    `TestEvalCubical_FillWithFaceOr`
  - Univalence: `TestEvalCubical_UASimple`, `TestPathApply_VUAEndpoints`
  - PathP: `TestEvalCubical_PathPWithClosure`
  - PathApply: `TestPathApply_VFill` for fill @ i0/i1/neutral
  - tryEvalCubical: `TestTryEvalCubical_MoreCases` with additional term types
  - tryReifyCubical: `TestTryReifyCubical_MoreValues` with additional value types
  - Coverage for `EvalCubical` improved from 69.8% to 99.2%

- **Normalize edge case tests** (`internal/eval/eval_test.go` - extended)
  - `TestNormalize_FstNeutral`: Fst on non-pair (stuck case)
  - `TestNormalize_SndNeutral`: Snd on non-pair (stuck case)
  - `TestNormalize_Default`: Sort, Pi, Sigma, Lambda pass-through
  - `TestNormalize_NestedBeta`: Nested beta reductions
  - Coverage for `Normalize` improved to 100%

- Coverage for `internal/eval` improved from 88.6% to 91.5%

- **IShift edge case tests** (`kernel/subst/subst_cubical_test.go` - extended)
  - Face formulas as direct Term input: `TestIShift_FaceEq_AsDirectTerm_BelowCutoff`,
    `TestIShift_FaceEq_AsDirectTerm_AboveCutoff`, `TestIShift_FaceEq_AsDirectTerm_AtCutoff`,
    `TestIShift_FaceAnd_AsDirectTerm`, `TestIShift_FaceOr_AsDirectTerm`, `TestIShift_FaceTopBot_AsDirectTerm`
  - Negative shift: `TestIShift_NegativeShift`, `TestIShift_NegativeShift_Path`
  - Zero shift: `TestIShift_ZeroShift`, `TestIShift_ZeroShift_Complex`
  - Cutoff boundary: `TestIShift_IVar_ExactCutoff`, `TestIShift_IVar_OneBelowCutoff`
  - Empty structures: `TestIShift_System_Empty`, `TestIShift_Glue_EmptySystem`, `TestIShift_GlueElem_EmptySystem`
  - Nested structures: `TestIShift_NestedPathApp`, `TestIShift_NestedBindersCutoffPropagation`,
    `TestIShift_DeepNestedSystem`
  - Edge cases: `TestIShift_LargeShift`, `TestIShift_HITApp_EmptyArgs`
  - Coverage for `kernel/subst` improved from 91.7% to 93.6%

- **Recursor indexed inductive tests** (`kernel/check/recursor_test.go` - extended)
  - `extractUniverseLevel`: `TestExtractUniverseLevel_DirectSort`, `TestExtractUniverseLevel_SinglePi`,
    `TestExtractUniverseLevel_NestedPi`, `TestExtractUniverseLevel_DeeplyNestedPi`, `TestExtractUniverseLevel_Fallback`
  - `isRecursiveArgTypeMulti`: 10 tests for mutual inductive detection including higher-order cases
  - `buildAppliedInductiveFull`: `TestBuildAppliedInductiveFull_NoParams_NoIndices`,
    `TestBuildAppliedInductiveFull_OneParam_NoIndices`, `TestBuildAppliedInductiveFull_OneParam_OneIndex`,
    `TestBuildAppliedInductiveFull_TwoParams_TwoIndices`
  - `buildMotiveTypeFull`: `TestBuildMotiveTypeFull_NoIndices`, `TestBuildMotiveTypeFull_WithIndices`

- **Mutual positivity tests** (`kernel/check/positivity_test.go` - extended)
  - Tests for `checkArgTypePositivityMulti` covering all term types: Sigma, Lam, Pair, Fst/Snd, Let, Id, Refl, J, Var, Sort, App
  - Coverage for `checkArgTypePositivityMulti` improved from 22.4% to 74.1%

- **Identity type and J tests** (`kernel/check/id_test.go` - extended)
  - Error path tests for `synthJ`: A not a type, X/Y type mismatch, invalid motive C, invalid base case D, invalid proof P
  - Coverage for `synthJ` improved from 66.7% to 100%

- **Cubical type error path tests** (`kernel/check/path_test.go` - extended)
  - Path type formation: `TestPathTypeFormation_ErrorInA/X/Y`
  - PathP type formation: `TestPathPTypeFormation_ErrorInA/X/Y`
  - Partial type synthesis: `TestSynthPartial_ErrorInFace/A`
  - Coverage for `synthPath` improved from 62.5% to 100%
  - Coverage for `synthPathP` improved from 69.2% to 92.3%
  - Coverage for `synthPartial` improved from 66.7% to 100%

- **Parser error path tests** (`internal/parser/sexpr_test.go`, `sexpr_cubical_test.go` - extended)
  - FormatTerm: `TestFormatTerm_Nil`, `TestFormatTerm_LamWithoutAnn`, `TestFormatTerm_SortHigherLevel`
  - Normalize: `TestNormalize` with 6 cases
  - Parse error tests: `TestParseJ_Errors`, `TestParseSigma_Errors`, `TestParseLet_Errors`, `TestParseId_Errors`,
    `TestParseRefl_Errors`, `TestParsePair_Errors`, `TestParseGlobal_WithParen`, `TestParseSort_InvalidLevel`
  - Cubical parse errors: `TestParsePath_Errors`, `TestParsePathP_Errors`, `TestParsePathLam_Errors`,
    `TestParsePathApp_Errors`, `TestParseTransport_Errors`
  - Coverage for `internal/parser` improved from 87.9% to 94.6%
  - Coverage for `Normalize` improved from 0% to 100%

- **Recursor helper function tests** (`kernel/check/recursor_test.go` - continued)
  - `paramName`/`indexName`: name generation tests for parameter and index naming
  - `buildIHType`: `TestBuildIHType_NoIndices`, `TestBuildIHType_WithIndices`
  - `extractIndicesFromType`: 4 tests for index extraction including edge cases
  - `extractConstructorIndices`: `TestExtractConstructorIndices_Simple`
  - `shiftIndexExpr`: `TestShiftIndexExpr_Variable`, `TestShiftIndexExpr_Global`
  - Indexed inductives: `TestGenerateRecursorType_Vec`, `TestGenerateRecursorType_Fin`,
    `TestBuildCaseTypeFull_VecCons`
  - Coverage for `kernel/check` improved from 75.8% to 83.8%

### Added
- **internal/core package documentation** (`internal/core/doc.go` - new)
  - Documents definitional equality checking using NbE
  - Main functions: `Conv`, `AlphaEq`, `NewEnv`, `Extend`
  - Configuration: `ConvOptions` with `EnableEta` flag
  - Algorithm explanation: Eval → Reify → AlphaEq with optional η-expansion
  - Cubical type theory support documentation

### Tests
- **Cubical positivity tests** (`kernel/check/positivity_cubical_test.go` - new)
  - Path type positivity: `TestCheckPositivity_Path`, `TestCheckPositivity_PathP`,
    `TestCheckPositivity_PathLam`, `TestCheckPositivity_PathApp`
  - Transport and composition: `TestCheckPositivity_Transport`,
    `TestCheckPositivity_Comp`, `TestCheckPositivity_HComp`, `TestCheckPositivity_Fill`
  - Glue types: `TestCheckPositivity_Glue`, `TestCheckPositivity_GlueElem`,
    `TestCheckPositivity_Unglue`
  - Univalence: `TestCheckPositivity_UA`, `TestCheckPositivity_UABeta`
  - Partial types: `TestCheckPositivity_Partial`, `TestCheckPositivity_System`
  - Interval terms: `TestCheckPositivity_IntervalTerms`
  - Face formulas: `TestCheckPositivity_FaceFormulas`
  - Occurrence checking: `TestOccursIn_CubicalTerms` (57 subtests for all cubical terms)
  - Face helpers: `TestOccursInFace`, `TestCheckArgTypePositivityFace`
  - Coverage for `kernel/check` improved from 75.8% to 83.5%
- **IShift cubical tests** (`kernel/subst/subst_cubical_test.go` - extended)
  - Interval constants: `TestIShift_IntervalConstants` (I0, I1, Interval)
  - Interval variables: `TestIShift_IVar` with cutoff boundary conditions
  - Path types: `TestIShift_Path`, `TestIShift_PathP` (with interval binding),
    `TestIShift_PathLam` (with cutoff increment), `TestIShift_PathApp`
  - Transport: `TestIShift_Transport` with differential binding (A at cutoff+1, E at cutoff)
  - Composition: `TestIShift_Comp`, `TestIShift_HComp`, `TestIShift_Fill` with
    complex cutoff handling for interval-binding constructs
  - Partial types: `TestIShift_Partial`, `TestIShift_System` with multiple branches
  - Glue types: `TestIShift_Glue`, `TestIShift_GlueElem`, `TestIShift_Unglue`
  - Univalence: `TestIShift_UA`, `TestIShift_UABeta`
  - Coverage for `kernel/subst` improved from 84.3% to 91.7%
- **Indexed inductive tests** (`internal/eval/recursor_test.go` - extended)
  - Vec with metadata: `TestIndexedInductive_VecWithMetadata` with IndexArgPositions
  - Vec without metadata: `TestIndexedInductive_WithoutMetadata` (fallback heuristic)
  - Multiple indices: `TestIndexedInductive_MultipleIndices` (2-index type)
  - Multiple recursive args: `TestIndexedInductive_MultipleRecursiveArgs` (binary tree)
  - Partial metadata: `TestIndexedInductive_PartialMetadata` (incomplete positions)
  - Empty positions: `TestIndexedInductive_EmptyIndexArgPositions`

### Changed
- **docs/index.md updated** to reflect v1.7.0 status
  - Phase 7 (HITs) marked complete
  - Added Phase 8 (Elaboration and tactics) and Phase 9 (Standard library seed) to roadmap
  - Added HIT highlights: S¹, Trunc, Susp, Int, Quotients
  - Updated copyright year to 2025-2026

### Fixed
- **Badge workflow fixes** (`.github/workflows/update-badges.yml`)
  - Added `fetch-depth: 0` for full git history (fixes commits/age/releases count)
  - Pass `GH_TOKEN` to generate script for release count via gh CLI
  - Changed from `exuanbo/actions-deploy-gist` to `gh gist edit` for badge uploads
  - Added retry logic with 2s backoff for 409 Conflict errors
  - Added 1s delay between uploads to avoid API rate limits
- **Race condition in HIT boundary tests** (`internal/eval/nbe_hit_test.go`)
  - Removed `t.Parallel()` from 4 tests that modify global recursor registry
  - Tests: `TestLookupHITBoundaries_NoRecursor`, `TestLookupHITBoundaries_NonHIT`,
    `TestLookupHITBoundaries_WithPathCtor`, `TestLookupHITBoundaries_UnknownPathCtor`
- **Staticcheck SA4003 warning** (`kernel/check/path_test.go:2690`)
  - Removed impossible `sort.U < 0` check (uint cannot be negative)

### Tests
- **Cubical term printing coverage** (`internal/ast/print_cubical_test.go` - new)
  - Comprehensive tests for all 28 cubical term types
  - Interval types: `Interval`, `I0`, `I1`, `IVar` with various indices
  - Path types: `Path`, `PathP`, `PathLam` (with/without binder), `PathApp`, `Transport`
  - Face formulas: `FaceTop`, `FaceBot`, `FaceEq`, `FaceAnd`, `FaceOr`, nested combinations
  - Partial types: `Partial`, `System` (empty, single, multiple branches)
  - Composition: `Comp` (with/without binder), `HComp`, `Fill` (with/without binder)
  - Glue types: `Glue` (empty/single/multiple branches), `GlueElem`, `Unglue`
  - Univalence: `UA`, `UABeta`
  - HIT: `HITApp` (no args, term args, interval args, both)
  - Edge cases: nil face handling, complex nested cubical terms
  - Coverage improved from 38.8% to 86.1% for `internal/ast` package
- **HIT term structure tests** (`internal/ast/term_test.go` - extended)
  - `HITSpec.IsHIT()`: with/without path constructors, empty path constructors
  - `HITSpec.MaxLevel()`: no constructors, level-1, level-2, mixed levels
  - `PathConstructor` structure verification
  - `Boundary` structure verification
  - `HITApp` structure and Term interface compliance
  - `Constructor` structure verification
  - Full `HITSpec` integration test (Circle S1)
  - `MkApps` helper: empty, single arg, multiple args
  - `Sort.IsZeroLevel()` tests
  - Coverage improved to 87.8% for `internal/ast` package
- **Cubical S-expression parsing tests** (`internal/parser/sexpr_cubical_test.go` - new)
  - Interval atoms: `I`, `Interval`, `i0`, `i1`
  - `IVar` parsing with valid indices and error cases
  - `Path` and `PathP` type parsing with nested types
  - `PathLam` parsing (both `PathLam` and `<>` syntax)
  - `PathApp` parsing (both `PathApp` and `@` syntax)
  - `Transport` parsing with nested type families
  - `HITApp` parsing: no args, term args, interval args, both
  - `parseTermList` edge cases via HITApp
  - `formatCubicalTerm` for all cubical types
  - Round-trip tests for all cubical terms
  - Complex expressions: PathApp on PathLam, mixed cubical/non-cubical
  - Coverage improved from 52.5% to 79.9% for `internal/parser` package
- **Alpha-equality cubical tests** (`internal/core/conv_cubical_test.go` - extended)
  - HITApp tests: same, different HIT/ctor name, term args, iargs, full combinations
  - shiftTermExtension tests for all cubical types (Interval, I0, I1, IVar, Path, PathP, PathLam, PathApp, Transport)
  - Non-cubical term verification (shiftTermExtension returns false)
  - Cubical vs non-cubical type mismatch tests
  - Coverage improved from 59.2% to 70.4% for `internal/core` package
- **HITApp substitution tests** (`kernel/subst/subst_cubical_test.go` - extended)
  - IShift tests for HITApp: no shift at cutoff, shift at higher levels, nested IVar
  - ISubst tests for HITApp: substitution at i0/i1, variable substitution, nested terms
  - shiftExtension tests for HITApp: same at 0, shift at higher levels, nested structures
  - substExtension tests for HITApp: substitution of term variables, nested substitution
  - Additional tests for IShiftFace, faceToTerm, simplifyFaceAndAST, simplifyFaceOrAST
  - Coverage improved from 59.4% to 84.3% for `kernel/subst` package
- **HIT evaluation tests** (`internal/eval/nbe_hit_test.go` - new)
  - VHITPathCtor interface tests
  - BoundaryVal structure tests
  - evalHITApp tests: reduce at i0/i1, stuck at IVar, multiple IArgs, no boundaries
  - lookupHITBoundaries tests: no recursor, non-HIT, with path ctor, unknown ctor
  - tryHITPathReduction tests: unknown ctor, at i0/i1, with extra args, stuck cases, term args
- **Value equality tests** (`internal/eval/equality_test.go` - new)
  - alphaEqFace tests for nil, FaceTop, FaceBot, FaceEq, FaceAnd, FaceOr
  - ValueEqual tests for VSort, VGlobal, VPair, VId, VRefl, intervals, VIVar, VPath
  - ValueEqual tests for VNeutral, VLam, VPi, VSigma, VPathP, VPathLam, VTransport
  - ValueEqual tests for face values: VFaceTop, VFaceBot, VFaceEq, VFaceAnd, VFaceOr
  - ValueEqual tests for VPartial, VSystem
  - NeutralEqual and faceValueEqual tests
- **Pretty printing tests** (`internal/eval/pretty_test.go` - new)
  - SprintValue tests for all value types: VSort, VGlobal, VPair, VLam, VPi, VSigma
  - SprintValue tests for VId, VRefl, intervals, VPath, VPathP, VPathLam, VTransport
  - SprintValue tests for face values: VFaceTop, VFaceBot, VFaceEq, VFaceAnd, VFaceOr
  - SprintValue tests for VPartial, VSystem, VGlue, VGlueElem, VComp, VHComp
  - SprintValue tests for VHITPathCtor, VNeutral with spine
  - WriteFaceValue tests for nested formulas
  - Coverage improved from 66.6% to 80.8% for `internal/eval` package
- **HIT checker tests** (`kernel/check/hit_test.go` - extended)
  - Extended containsHIT tests for Lam, PathLam, Path endpoints, PathP, PathApp
  - New isPathResult tests for Path, PathP, App, non-path types
- **Conversion/alpha-equality tests** (`internal/core/conv_test.go` - extended)
  - `TestEnv_Extend`: environment extension with simple and complex terms
  - `TestShiftTerm_*`: comprehensive shiftTerm tests for all term types (Var, Lam, Pi, Sigma, Pair, Fst, Snd, Let, Id, Refl, J, App, nil)
  - `TestAlphaEq_EdgeCases`: nil handling, Sort, Var, Global, Pi, Sigma, Pair, Fst, Snd, Let, Id, Refl, J, App, cross-type comparisons
  - `TestEtaEqual_EdgeCases`: function eta (f = λx. f x), pair eta (p = (fst p, snd p)), negative cases
  - `TestEtaEqualPair_DirectCases`: valid eta, not a pair, wrong fst/snd
  - `TestEtaEqualFunction_DirectCases`: valid eta, not a lambda, wrong body/arg/func
  - `TestConv_NilEnvironment`: nil environment handling
  - Coverage improved from 70.4% to 87.6% for `internal/core` package
- **Cubical error constructors** (`kernel/check/errors_test.go` - extended)
  - `TestErrNotAPath`: path type error construction
  - `TestErrPathEndpointMismatch`: endpoint mismatch error
  - `TestErrUnboundIVar`: unbound interval variable error
- **Face checking tests** (`kernel/check/path_test.go` - extended)
  - `TestCheckFace_*`: FaceTop, FaceBot, FaceEq (bound/unbound), FaceAnd, FaceOr, nil
  - Invalid face nesting tests for FaceAnd/FaceOr
- **PathApp synthesis tests** (`kernel/check/path_test.go` - extended)
  - `TestSynthPathApp_I0Endpoint`, `TestSynthPathApp_I1Endpoint`, `TestSynthPathApp_IVar`
  - `TestSynthPathApp_NotAPath`: error case for non-path term
- **PathLam checking tests** (`kernel/check/path_test.go` - extended)
  - `TestCheckPathLam_ValidConstant`: checking against Path type
  - `TestCheckPathLam_WrongLeftEndpoint`, `TestCheckPathLam_WrongRightEndpoint`: endpoint mismatch errors
  - `TestCheckPathLam_AgainstPathP`: checking against PathP type
- **Cubical synthesis tests** (`kernel/check/path_test.go` - extended)
  - Interval, I0, I1, IVar (valid/unbound) synthesis
  - Face synthesis: FaceTop, FaceBot, FaceEq (valid/unbound), FaceAnd, FaceOr
  - Partial type synthesis
  - Transport synthesis
  - System synthesis (empty error, single branch)
- **Checker constructor test** (`kernel/check/check_test.go` - extended)
  - `TestNewCheckerWithPrimitives`: convenience constructor with primitive verification
  - Coverage improved from 73.7% to 75.8% for `kernel/check` package
- **CLI command edge case tests** (`cmd/hg/commands_test.go` - extended)
  - `TestCmdNew_InvalidOutputPath`: saveGraph error handling
  - `TestCmdCopy_InvalidOutputPath`: output file error handling
  - `TestCmdAddVertex_InvalidOutputPath`: output file error handling
  - `TestCmdAddEdge_InvalidOutputPath`: output file error handling
  - `TestCmdAddEdge_DuplicateEdge`: duplicate edge ID error handling
- **REPL edge case tests** (`cmd/hg/repl_test.go` - extended)
  - `TestExecuteReplCommand_EdgeCases`: missing argument errors for remove-edge, has-vertex, has-edge, degree, edge-size, bfs, dfs
  - Nonexistent vertex/edge error handling for degree, edge-size, dfs
  - `TestExecuteReplCommand_SaveError`: save with invalid path error
  - `TestExecuteReplCommand_LoadShorthand`: unknown colon command error
  - Coverage improved from 79.7% to 81.6% for `cmd/hg` package
- **Alpha-equality extension tests** (`internal/core/conv_cubical_test.go` - extended)
  - `TestAlphaEqExtension_TypeMismatch`: 24 test cases for cubical type vs non-cubical mismatches
  - `TestAlphaEqFace_TypeMismatch`: face formula type mismatch tests
  - `TestAlphaEqExtension_SameCubicalTypeMismatch`: same-type field mismatches
  - `TestAlphaEqExtension_DefaultCase`: default branch coverage
  - `TestAlphaEqExtension_GlueElemSystemLengthMismatch`: system branch count differences
  - `TestAlphaEqExtension_CrossCubicalMismatch`: cross-cubical type comparisons
  - Coverage for `alphaEqExtension` improved from 76% to 100%
  - Coverage for `alphaEqFace` improved from 78.9% to 94.7%
  - Coverage for `internal/core` improved from 87.6% to 99.1%
- **Positivity checker tests** (`kernel/check/positivity_test.go`, `positivity_cubical_test.go` - extended)
  - `TestCheckArgTypePositivity_AllPaths`: 12 test cases for J, Id, Refl, Pair, Let, etc.
  - `TestCheckArgTypePositivityMulti_AllPaths`: 7 test cases for mutual inductives
  - `TestCheckPositivity_CubicalErrorPaths`: 14 test cases for PathP, PathApp, Comp, HComp, Fill, Glue, UA, UABeta
  - `TestOccursInFace_AllBranches`: face formula occurrence tests
  - Coverage for `checkArgTypePositivity` improved from 65.5% to 92.7%
  - Coverage for `checkArgTypePositivityMulti` improved from 74.1% to 84.5%
  - Coverage for `kernel/check` improved from 88.4% to 90.2%
- **Cubical type checking tests** (`kernel/check/bidir_cubical_test.go` - new)
  - Glue type synthesis: `TestSynthGlue_UniverseLevelMismatch`, `TestSynthGlue_EquivError`, `TestSynthGlue_FaceError`, `TestSynthGlue_TypeError`, `TestSynthGlue_ANotType`
  - HComp synthesis: `TestSynthHComp_ANotType`, `TestSynthHComp_FaceError`, `TestSynthHComp_BaseError`, `TestSynthHComp_TubeError`, `TestSynthHComp_TubeBaseDisagreement`
  - System synthesis: `TestSynthSystem_FirstTermError`, `TestSynthSystem_SubsequentTermError`, `TestSynthSystem_SubsequentFaceError`, `TestCheckSystemAgreement_Disagreement`
  - Transport synthesis: `TestSynthTransport_ANotType`, `TestSynthTransport_EError`
  - Comp synthesis: `TestSynthComp_ANotType`, `TestSynthComp_FaceError`, `TestSynthComp_BaseError`, `TestSynthComp_TubeError`, `TestSynthComp_TubeBaseDisagreement`
  - Fill synthesis: `TestSynthFill_ANotType`, `TestSynthFill_FaceError`, `TestSynthFill_BaseError`, `TestSynthFill_TubeError`, `TestSynthFill_TubeBaseDisagreement`
  - PathApp synthesis: `TestSynthPathApp_NotPath`, `TestSynthPathApp_PathCaseRError`, `TestSynthPathApp_PathPCaseRError`, `TestSynthPathApp_PathPSynthError`
  - UA synthesis: `TestSynthUA_UniverseLevelMismatch`, `TestSynthUA_EquivError`, `TestSynthUA_BNotType`
  - Unglue synthesis: `TestSynthUnglue_WithTypeAnnotation`, `TestSynthUnglue_WithNonGlueAnnotation`, `TestSynthUnglue_SynthError`
  - PathLam checking: `TestCheckPathLam_PathCase`, `TestCheckPathLam_PathCaseBodyError`, `TestCheckPathLam_PathCaseLeftMismatch`, `TestCheckPathLam_PathCaseRightMismatch`, `TestCheckPathLam_NotPathType`, `TestCheckPathLam_PathPCaseBodyError`, `TestCheckPathLam_PathPCaseLeftMismatch`, `TestCheckPathLam_PathPCaseRightMismatch`
  - UABeta synthesis: `TestSynthUABeta_EquivError`, `TestSynthUABeta_ArgError`, `TestSynthUABeta_WithSigmaPiEquiv`
  - GlueElem synthesis: `TestSynthGlueElem_BaseError`, `TestSynthGlueElem_FaceError`, `TestSynthGlueElem_TermError`
  - Face checking: `TestCheckFace_DefaultCase`, `TestCheckFace_FaceOrRightError`
  - faceIsBot: `TestFaceIsBot_AllCases` (10 test cases), `TestFaceIsBot_DefaultCase`
  - collectFaceEqs: `TestCollectFaceEqs` (6 test cases)
  - PathLam synthesis: `TestSynthPathLam_BodyError`
  - 20+ functions improved to 100% coverage including: synthPathApp, synthTransport, checkPathLam, checkSystemAgreement, faceIsBot, collectFaceEqs, synthComp, synthHComp, synthFill, synthGlue, synthGlueElem
  - Coverage for `kernel/check` improved from 90.2% to 93.1%
- **NbE coverage tests** (`internal/eval/nbe_test.go` - extended)
  - `TestMakeNeutralGlobal`: exported helper function coverage
  - Value reification: `TestReifyAt_VPi`, `TestReifyAt_VSigma`, `TestReifyAt_VId`, `TestReifyAt_VRefl`
  - Neutral reification: `TestReifyNeutralWithReifier_FstSndCases`, `TestReifyNeutralWithReifier_JCase`, `TestReifyNeutralWithReifier_FstSndInsufficientSpine`, `TestReifyNeutralWithReifier_JInsufficientSpine`
  - Debug mode: `TestNBE_DebugMode`, `TestNBE_DebugModePanic`, `TestNBE_ReifyErrorDebugModePanic`
  - Eliminator stuck cases: `TestNBE_NatElimNonConstructorTarget`, `TestNBE_BoolElimNonConstructorTarget`
  - Term evaluation: `TestEval_LetTerm`, `TestEval_IdTerm`, `TestEval_ReflTerm`, `TestEval_PiTerm`, `TestEval_SigmaTerm`
  - Environment: `TestEnv_LookupOutOfBounds`
  - Application: `TestApply_VLamClosure`, `TestApply_ToNeutral`
  - Projections: `TestFst_Snd_OnNeutral`
  - Coverage for `evalError` improved from 66.7% to 100%
  - Coverage for `reifyError` improved from 66.7% to 100%
  - Coverage for `MakeNeutralGlobal` improved from 0% to 100%
  - Coverage for `Eval` improved from 88.4% to 97.7%
  - Coverage for `reifyAt` improved from 74.2% to 96.8%
  - Coverage for `reifyNeutralWithReifier` improved to 100%
  - Coverage for `internal/eval` improved from 87.6% to 88.6%

## [1.7.0] - 2026-01-04

### Fixed
- **Composition agreement checks now conditional and enforced** (`kernel/check/bidir_cubical.go`)
  - `synthComp`, `synthHComp`, `synthFill` now check if φ[i0/i] is satisfiable before applying agreement checks
  - When φ[i0/i] ≠ ⊥, tube[i0/i] must equal base - returns error on mismatch
  - Previously these checks were non-fatal (used `_ = c.conv(...)`)

- **Resolved all golangci-lint issues** (162 issues total)
  - Fixed unchecked `fs.Parse()` errors in all cmd/hg command functions
  - Fixed unchecked file Close/Remove errors with proper error handling patterns
  - Fixed type assertion in `hypergraph/serialize.go` by using typed variable
  - Removed unused helper functions in test files (`nbe_test.go`, `subst_cubical_test.go`)
  - Fixed De Morgan's law issue in `algorithms_test.go`
  - Updated `.golangci.yml` to golangci-lint v2 format with proper exclusions

### Changed
- **CI/CD improvements** (`.github/workflows/`)
  - Updated `codecov/codecov-action` from v3 to v5 in `go.yml`
  - Fixed `publish-deb` job in `release.yml` to download release artifacts before pushing to Cloudsmith
  - Added race detection and cubical tests to `ci-linux.yml`
  - Added cubical tests to `ci-windows.yml`

- **Removed unused sentinel errors** (`hypergraph/errors.go`)
  - Removed `ErrUnknownEdge` and `ErrUnknownVertex` - these were defined but never used
  - The API design uses silent no-ops for missing items (consistent with existing behavior)

- **Refactored reifyNeutral to reduce code duplication** (`internal/eval/nbe.go`, `nbe_cubical.go`)
  - Extracted `reifySpine` helper for applying spine arguments
  - Extracted `reifyNeutralWithReifier` for shared fst/snd/J handling
  - `reifyNeutralCubicalAt` now uses shared helper with cubical-specific "@" case
  - Reduced ~80 lines of duplicated code

### Added
- **Higher Inductive Types (HITs)** - Phase 7 implementation
  - **AST extensions** (`internal/ast/term_hit.go` - new)
    - `PathConstructor` type for path-level constructors (with level and boundaries)
    - `Boundary` type for interval endpoint values
    - `HITApp` term for path constructor application to intervals
    - `HITSpec` type for HIT declarations
    - `Constructor` type for point constructors
  - **Evaluation support** (`internal/eval/nbe_hit.go` - new)
    - `VHITPathCtor` value type for evaluated path constructors
    - `BoundaryVal` for evaluated boundary values
    - `evalHITApp` for HIT application evaluation
    - `lookupHITBoundaries` for boundary retrieval
    - Path constructors reduce at interval endpoints (loop @ i0 → base)
  - **Type checker extensions** (`kernel/check/hit.go` - new)
    - `DeclareHIT` function for validating and registering HITs
    - `validatePathConstructor` for path constructor well-formedness
    - `checkHITPositivity` for strict positivity of HITs
    - `GenerateHITRecursorType` for eliminator type generation
    - `buildPathCaseType` for path case types (PathP)
  - **Recursor extensions** (`internal/eval/recursor.go`)
    - `PathConstructorInfo` and `BoundarySpec` types
    - `IsHIT` and `PathCtors` fields in `RecursorInfo`
    - `tryHITPathReduction` for HIT eliminator reduction
    - Proper handling of path constructor elimination
  - **Parser support** (`internal/parser/sexpr_cubical.go`)
    - `parseHITApp` for parsing HIT applications
    - `parseTermList` helper for argument lists
    - `formatCubicalTerm` extended for HITApp output
  - **Built-in HITs** (`kernel/check/env_hit.go` - new)
    - **Circle (S1)**: base point and loop path
    - **Truncation (Trunc)**: propositional truncation with inc and squash
    - **Suspension (Susp)**: north, south points with merid path
    - **Integers (Int)**: pos, neg constructors with zeroPath
    - **Set Quotient (Quot)**: quot constructor with eq path
  - **Extended Inductive struct** (`kernel/check/env.go`)
    - `PathCtors` field for path constructors
    - `IsHIT` flag to mark Higher Inductive Types
    - `MaxLevel` for highest path constructor dimension
  - **Comprehensive test suite** (`kernel/check/hit_test.go` - new)
    - Tests for DeclareHIT, path constructor validation
    - Tests for HIT recursor info building
    - Tests for built-in HITs (S1, Trunc, Susp, Int, Quot)
    - Tests for VHITPathCtor evaluation and reduction

- **Dependabot configuration** (`.github/dependabot.yml`)
  - Weekly updates for GitHub Actions (`ci` prefix)
  - Weekly updates for Go modules (`deps` prefix)

- **Expanded Makefile targets**
  - `build` - Build the hg binary to `bin/hg`
  - `clean` - Remove build artifacts
  - `test` - Run tests
  - `test-race` - Run tests with race detection
  - `test-cubical` - Run cubical feature tests
  - `coverage` - Generate HTML coverage report
  - `help` - List available targets
  - Added `.PHONY` declarations for all targets

- **Raw identity types** (`internal/ast/raw.go`, `internal/ast/resolve.go`)
  - RId, RRefl, RJ raw term types for identity types
  - Resolver cases to convert raw identity terms to core terms
  - Test coverage in `internal/ast/raw_test.go`

- **Primal method** (`hypergraph/transforms.go`)
  - Added `Primal()` method as synonym for `TwoSection()` (was documented but not implemented)

- **Alpha-equality module** (`internal/eval/alpha_eq.go` - new)
  - `AlphaEq(a, b ast.Term) bool` for proper alpha-equality checking
  - Handles all core term types: Var, Global, Sort, Lam, App, Pi, Sigma, Pair, Fst, Snd, Let, Id, Refl, J
  - Handles all cubical term types: Interval, I0, I1, IVar, Path, PathP, PathLam, PathApp, Transport
  - Handles face formulas: FaceTop, FaceBot, FaceEq, FaceAnd, FaceOr
  - Handles partial types, composition, glue, and univalence terms
  - Binder names are correctly ignored (de Bruijn indices used)

- **Alpha-equality test coverage** (`internal/eval/alpha_eq_test.go` - new)
  - Comprehensive tests for all term types
  - Tests for binder name irrelevance
  - Tests for structural differences

- **Pretty-print support for all 22 cubical Value types** (`internal/eval/pretty.go`)
  - `writeValue()`: VI0, VI1, VIVar, VPath, VPathP, VPathLam, VTransport, VFaceTop, VFaceBot, VFaceEq, VFaceAnd, VFaceOr, VPartial, VSystem, VComp, VHComp, VFill, VGlue, VGlueElem, VUnglue, VUA, VUABeta
  - `ValueEqual()`: Structural equality for all cubical Value types
  - `valueTypeName()`: Type name strings for all cubical Value types
  - `writeFaceValue()`: Helper for face formula pretty-printing
  - `faceValueEqual()`, `iClosureEqual()`, `ienvEqual()`: Equality helpers

- **Pretty-print cubical test coverage** (`internal/eval/pretty_cubical_test.go` - new)
  - `TestSprintValue_Cubical`: Tests for all 22 cubical Value types
  - `TestValueEqual_Cubical`: Equality tests for cubical Values
  - `TestValueTypeName_Cubical`: Type name tests
  - `TestFaceValueEqual`, `TestWriteFaceValue`: Face formula tests

- **Cubical type checker test coverage** (`kernel/check/ictx_test.go` - new, `kernel/check/path_test.go` - extended)
  - **ICtx tests**: Deep nesting (5/10 levels), defer cleanup, nested defers, boundary conditions (negative indices, exact boundaries), complex push/pop patterns (interleaved, double pop), context creation/destruction, immutability chain verification
  - **Face formula tests**: faceIsBot edge cases (nested contradictions, three variables, Or of bots, Or of contradictions, And with bot, deeply nested), isContradictoryFaceAnd (different variables, same constraint, triple nested, with FaceTop), collectFaceEqs (deeply nested, Or doesn't collect, Top/Bot return empty, mixed constraints)
  - **Path type tests**: Nested PathLam with multiple interval variables, PathApp on nested PathLam, PathApp at IVar, PathP with constant type family, PathLam endpoint computation, PathApp beta reduction at i0/i1, Path with same endpoints, PathLam constant body
  - **Composition tests**: Comp with contradictory face, HComp with FaceOr, Transport constant type identity, Transport non-constant stuck, Fill at i0/i1 endpoints, Comp with nested FaceAnd, HComp base type verification, Comp result type A[i1/i]
  - **Glue/UA tests**: Glue empty system, multiple branches, FaceBot branch, GlueElem evaluation, Unglue round-trip, UA endpoints at i0/i1, UA type checking, Glue with FaceTop reduces, Glue with FaceBot doesn't reduce

### Fixed
- **collectSpine O(n^2) performance** (`internal/ast/print.go`)
  - Fixed slice prepending that caused quadratic time complexity
  - Now collects forward and reverses at end for O(n) performance

- **GreedyHittingSet determinism** (`hypergraph/algorithms.go`)
  - Fixed non-deterministic results due to map iteration order
  - Now sorts vertices before iteration for reproducible results

- **Nil type validation in ctx.Extend** (`kernel/ctx/ctx.go`)
  - Added nil check that panics with "ctx.Extend: nil type" message
  - Catches programming errors early before they propagate

- **Removed custom itoa reimplementation** (`kernel/check/span.go`, `kernel/check/errors.go`)
  - Replaced 18-line custom `itoa` function with `strconv.Itoa`
  - Uses standard library instead of reimplementing integer-to-string conversion
  - Import added: `"strconv"` in both files

- **Incorrect alpha-equality using string comparison** (`internal/eval/nbe_cubical.go:898`)
  - Previously used `ast.Sprint(a) == ast.Sprint(b)` which is incorrect for alpha-equality
  - Now uses proper structural comparison via `AlphaEq(a, b)` with de Bruijn indices
  - New `alpha_eq.go` module handles all AST term types including cubical extensions

- **EvalFill always returning stuck value** (`internal/eval/nbe_cubical.go:750-753`)
  - Fill with face=⊤ now reduces to `VPathLam{Body: tube}` (path lambda over tube)
  - PathApply on VFill now computes: at i0 returns base, at i1 returns comp result
  - Stuck VFill only returned when face is not definitionally true

- **EvalUABeta not computing transport rule** (`internal/eval/nbe_cubical.go:870-880`)
  - Now properly computes: `transport (ua e) a = e.fst a` when e is a concrete pair
  - Returns `VUABeta{Equiv, Arg}` when equivalence is neutral (stuck term)
  - Preserves semantic structure instead of reducing through Fst/Apply on neutrals
  - Enables transport along univalence paths to actually compute

- **:quit safety mechanism losing unsaved state** (`cmd/hg/repl.go:83-89`)
  - Previously cleared `modified` flag on first quit, losing track of unsaved changes
  - Now uses separate `quitConfirmed` flag to track two-stage quit without losing state
  - If user runs another command between two :quits, confirmation resets

- **:new safety mechanism losing unsaved state** (`cmd/hg/repl.go:127-137`)
  - Same fix as :quit: uses `newConfirmed` flag instead of clearing `modified`
  - Confirmation resets if user runs other commands

- **No signal handling for Ctrl+C in hg REPL** (`cmd/hg/repl.go:53-65`)
  - Added SIGINT/SIGTERM handlers that warn about unsaved changes
  - First Ctrl+C with unsaved changes shows warning
  - Second Ctrl+C force exits with code 130

- **Non-atomic file writes in hg** (`cmd/hg/io.go:19-42`)
  - File saves now write to temp file first, then atomic rename
  - Prevents data corruption on write failure or interrupt
  - Temp file cleaned up on error

- **Scanner buffer too small and errors unchecked in hg REPL** (`cmd/hg/repl.go:67-68, 89-91`)
  - Increased buffer from 64KB to 1MB max line size
  - Added `scanner.Err()` check after scan loop to report I/O errors

### Tests
- **Hypergraph test coverage** (`hypergraph/hypergraph_test.go`, `hypergraph/algorithms_test.go`)
  - Added `TestEdgeMembers_ExistingEdge`, `TestEdgeMembers_NonExistentEdge`
  - Added `TestCopy_DeepCopySemantics`, `TestCopy_Independence`, `TestCopy_EmptyHypergraph`
  - Added `TestAddEdge_DuplicateVertices`, `TestAddEdge_AllDuplicates`
  - Added `TestVertexDegree_NonExistentVertex`, `TestEdgeSize_NonExistentEdge`, `TestVertexDegree_MultipleEdges`
  - Added `TestGreedyHittingSet_Deterministic` - verifies deterministic output
  - Coverage improved to 97.5%

- **Comprehensive kernel/ctx test coverage** (`kernel/ctx/ctx_test.go`)
  - `TestLen`: Empty context, single Extend, multiple Extends
  - `TestDrop`: Empty context, restores previous state, value semantics verification
  - `TestLookupVarNegativeIndex`: Negative index returns false
  - `TestLookupVarEmptyContext`: Empty context returns false for any index
  - `TestChainedExtendDrop`: Build up and drop one at a time, drop from empty stays empty
  - `TestExtendNilTypePanics`: Validates nil type panic behavior
  - Coverage improved from 50% to 100%

- **REPL robustness tests** (`cmd/hg/repl_test.go` - extended)
  - `TestConfirmationFlagReset`: Verifies quit/new confirmation flags reset when other commands run
  - `TestAtomicSave`: Verifies temp file cleanup after successful atomic save

### Fixed
- **CLI docstring accuracy** (`cmd/hottgo/main.go`)
  - Removed stale "(TODO)" from REPL usage comment - the REPL is fully implemented

- **REPL error handling** (`cmd/hottgo/main.go`)
  - Added `scanner.Err()` check after REPL loop to report I/O errors instead of silently ignoring them

### Added
- **CLI test coverage** (`cmd/hottgo/main_test.go` - new)
  - doCheck tests: valid file, missing file, parse error, type error, empty file
  - doEval tests: valid expressions, parse errors
  - doSynth tests: valid expressions, parse errors, type errors
  - REPL tests: :eval, :synth, :quit commands, plain expressions, empty lines, EOF handling
  - Version flag handling

- **Comprehensive test coverage** (17 new/extended test files, ~7300 lines)
  - **cmd/hg/io_test.go** (new): loadGraph/saveGraph with valid files, missing files, invalid JSON, round-trip verification
  - **cmd/hg/commands_test.go** (new): All 14 CLI commands (info, new, validate, add-vertex, remove-vertex, has-vertex, add-edge, remove-edge, has-edge, vertices, edges, degree, edge-size, copy) with flag parsing, error paths, output verification
  - **cmd/hg/main_test.go** (new): Subcommand routing, usage output, help command entries, global flags documentation
  - **cmd/hg/repl_test.go** (new): REPL state machine, :load/:save/:new/:info/:quit/:help commands, modified flag behavior, unsaved changes warnings
  - **cmd/hg/transforms_test.go** (new): dual, two-section, line-graph commands with output validation, edge cases
  - **cmd/hg/algorithms_test.go** (new): hitting-set, transversals, coloring, incidence commands with correctness checks
  - **cmd/hg/traversal_test.go** (new): bfs, dfs, components commands with connectivity verification
  - **hypergraph/algorithms_test.go** (new): GreedyHittingSet, EnumerateMinimalTransversals, GreedyColoring with correctness verification, edge cases, cutoffs
  - **hypergraph/transforms_test.go** (new): Dual (round-trip, incidence preservation), TwoSection (clique formation, deduplication), LineGraph (intersection detection, star/chain structures)
  - **hypergraph/incidence_test.go** (new): IncidenceMatrix COO format, index stability, bounds checking, reconstruction, row/column sums
  - **kernel/check/positivity_test.go** (extended): Multiple recursive args, Let/Fst/Snd/Pair/Refl/J in constructors, mutual positivity (Even/Odd, three-way), PositivityError formatting
  - **kernel/check/bidir_test.go** (new): Complex application chains, Id/Refl/J synthesis, universe level max, unannotated lambda/pair checking, deep context, Ensure* helpers
  - **kernel/check/errors_test.go** (new): All error constructors, TypeError formatting, ErrorDetails interface, complex type mismatches, Span edge cases
  - **internal/eval/recursor_test.go** (new): Recursor registry (register, lookup, clear, overwrite), RecursorInfo/ConstructorInfo structures, concurrent access safety
  - **internal/ast/print_test.go** (new): Sprint for all term types, empty binder handling, collectSpine, nested structures, output verification
  - **internal/eval/nbe_cubical_test.go** (extended): Edge case tests for cubical evaluation - unknown term types, variable/interval out-of-bounds lookup, face formula deep nesting, PathApply edge cases (VUA endpoints, stuck values), transport with interval-using body, composition fallback to transport, Glue/GlueElem branch filtering, reification edge cases (high levels, unknown types), neutral reification with spine (fst/snd/J/@), J eliminator evaluation, System branches, alpha equality, isConstantFamily, closure capture

- **Hypergraph CLI (hg)** (`cmd/hg/` - Full implementation)
  - Subcommand-based CLI for hypergraph operations
  - **Core commands**: `info`, `new`, `validate`, `add-vertex`, `remove-vertex`,
    `has-vertex`, `add-edge`, `remove-edge`, `has-edge`, `vertices`, `edges`,
    `degree`, `edge-size`, `copy`
  - **Transform commands**: `dual`, `two-section`, `line-graph`
  - **Traversal commands**: `bfs`, `dfs`, `components`
  - **Algorithm commands**: `hitting-set`, `transversals`, `coloring`, `incidence`
  - **Interactive REPL mode**: Load/save files, all operations available interactively
  - **Per-command help**: `hg help <command>` for detailed usage
  - Uses `internal/version` for consistent version output

- **Hypergraph EdgeMembers method** (`hypergraph/hypergraph.go`)
  - New public method to retrieve member vertices of an edge by ID
  - Returns nil if edge doesn't exist

## [1.6.1] - 2025-12-24

### Fixed
- **Chocolatey package review issues** (`packaging/chocolatey/hypergraphgo.nuspec`)
  - Removed unnecessary `chocolatey` dependency from nuspec
  - Excluded `.tmpl` template file from package (explicitly list only `.ps1` scripts)
  - Added `workflow_dispatch` trigger to chocolatey workflow for manual re-publishing

- **Alpha-equality completeness for cubical types** (`internal/core/conv_cubical.go`)
  - Added alpha-equality cases for Face formulas: `FaceTop`, `FaceBot`, `FaceEq`, `FaceAnd`, `FaceOr`
  - Added alpha-equality cases for Partial types: `Partial`, `System`
  - Added alpha-equality cases for Composition: `Comp`, `HComp`, `Fill`
  - Added alpha-equality cases for Glue types: `Glue`, `GlueElem`, `Unglue`
  - Added alpha-equality cases for Univalence: `UA`, `UABeta`
  - New `alphaEqFace` helper function for face formula equality
  - Comprehensive test suite in `internal/core/conv_cubical_test.go`

- **Type synthesis improvements** (`kernel/check/bidir_cubical.go`)
  - `synthUABeta` now properly extracts target type B from `Equiv A B` structure
  - `synthComp`, `synthHComp`, `synthFill` now check tube types against type family
  - System branch agreement checking via `checkSystemAgreement`
  - Face formula satisfiability checking with `faceIsBot` and `isContradictoryFaceAnd`

### Added
- **Composition tests** (`kernel/check/path_test.go`)
  - `TestCompFaceSatisfied` / `TestCompFaceEmpty` - comp reduction rules
  - `TestHCompFaceSatisfied` / `TestHCompFaceEmpty` - hcomp reduction rules
  - `TestFillEvaluation` - fill endpoint behavior
  - `TestCompTypeCheck` / `TestHCompTypeCheckWithBot` / `TestFillTypeCheck` - type checking
  - `TestTransportUAComputes` - verifies `transport (ua e) a = e.fst a`
  - `TestUABetaReification` - UABeta value reification

### Changed
- **Revamped architecture diagrams** (`DIAGRAMS.md`)
  - Complete rewrite of master architecture diagram with 3-column layout (more square/aesthetic)
  - Added detailed component architecture diagram showing all internal functions
  - Added type system summary diagram (MLTT + Cubical + Inductive)
  - Added computation rules diagram covering all reduction rules
  - Cubical-specific components highlighted with green borders (`#2da44e`)
  - Removed all background fills for mermaid.live compatibility (stroke-only styling)
  - All diagrams now accurate to current codebase including Phase 7 cubical features

## [1.6.0] - 2025-12-08

### Added
- **Computational Univalence** (Phase 7: Complete cubical univalence implementation)
  - **Univalence axiom (ua)** - `ua A B e : Path Type A B`
    - `UA` and `UABeta` AST types
    - Computation: `(ua e) @ i0 = A`, `(ua e) @ i1 = B`
    - Intermediate: `(ua e) @ i = Glue B [(i=0) ↦ (A, e)]`
  - **Glue types** - `Glue A [φ ↦ (T, e)] : Type`
    - `Glue`, `GlueElem`, `Unglue` AST types
    - Computation: `Glue A [⊤ ↦ (T, e)] = T`, `glue [⊤ ↦ t] a = t`
  - **Composition operations** - `comp^i A [φ ↦ u] a₀ : A[i1/i]`
    - `Comp`, `HComp`, `Fill` AST types with `IClosure` for interval environments
    - Computation: `comp^i A [⊤ ↦ u] a₀ → u[i1/i]`, `hcomp A [⊥ ↦ _] a₀ → a₀`
  - **Face formulas and Partial types** - `Partial φ A`, `System`
    - Face formulas: `FaceTop`, `FaceBot`, `FaceEq`, `FaceAnd`, `FaceOr`
    - Face simplification: `(i=0)∧(i=1) → ⊥`, `(i=0)∨(i=1) → ⊤`
  - Full NbE support with value types for all cubical constructs
  - Type synthesis, substitution, and positivity checking
  - Comprehensive test suite for all phases

- **Mutual inductive types** (`kernel/check/env.go`, `kernel/check/positivity.go`)
  - `MutualInductiveSpec` struct for specifying types in a mutual block
  - `DeclareMutual` API for declaring mutually recursive types (e.g., Even/Odd)
  - `Inductive.MutualGroup` field tracks other types in the mutual block
  - `DeclareInductive` refactored to call `DeclareMutual` for backward compatibility
  - `CheckMutualPositivity` validates strict positivity across all types in the block
  - Separate eliminators generated per type (simpler than joint eliminators)
  - Cross-type recursion must be handled explicitly in case functions

### Design Notes
- **Separate eliminators**: Each type in a mutual block gets its own eliminator
  - `evenElim : (P : even -> Type) -> P zero -> ((o : odd) -> P (succOdd o)) -> even -> P e`
  - `oddElim : (Q : odd -> Type) -> ((e : even) -> Q (succ e)) -> odd -> Q o`
  - Cross-type args (e.g., `o : odd` in `succOdd`) pass through without IH
  - For full mutual induction, compose eliminators in case functions

### Tests
- `TestMutualInductive_EvenOdd` - basic mutual declaration
- `TestMutualInductive_SingleIsSameAsDeclareInductive` - backward compatibility
- `TestMutualInductive_Positivity_Reject` - rejects negative occurrences
- `TestMutualInductive_Positivity_Accept` - accepts positive occurrences
- `TestMutualInductive_Reduction` - verifies separate eliminator reduction
- `TestMutualInductive_SameTypeRecursion` - same-type recursion still gets IH in mutual blocks
- `TestMutualInductive_NestedNegative` - rejects deeply nested negative occurrences
- `TestMutualInductive_DoublyNegativeIsPositive` - documents strict positivity (no double-flip)
- `TestMutualInductive_SymmetricNegative` - symmetric checking across mutual types

### Changed
- **Cubical types always enabled** (Phase 7.1: Merge Cubical into Main)
  - Removed `//go:build cubical` tags from all cubical files
  - Deleted stub files (`*_ext.go`, `*_nocubical.go`)
  - Path types, interval type, and transport now available in default build
  - No longer need `-tags cubical` for cubical features

### Fixed
- **golangci-lint cleanup** (30 issues resolved)
  - S1016 staticcheck: Use type conversions instead of struct literals in `internal/ast/resolve.go`, `kernel/check/bidir.go`
  - ST1023 staticcheck: Omit redundant type in declaration in `kernel/check/recursor_test.go`
  - Removed unused `sigma` function from `internal/core/conv_test.go`
  - Removed unused `cubicalEnabled` variable from `internal/parser/sexpr.go` and `sexpr_cubical.go`
  - Fixed unchecked error returns in `examples/basic/main.go` and `hypergraph/edge_cases_test.go`

## [1.5.8] - 2025-12-06

### Added
- **Parameterized inductive types** (`kernel/check/`)
  - `Inductive` struct extended with `NumParams` and `ParamTypes` fields
  - `extractPiChain` extracts all Pi arguments from inductive type
  - `validateInductiveType` validates Pi chains ending in Sort
  - `validateConstructorResult` validates constructor returns applied inductive with correct args
  - `extractAppArgs` collects arguments from application chains

- **Indexed inductive types** (`kernel/check/`)
  - `Inductive` struct extended with `NumIndices` and `IndexTypes` fields
  - `analyzeParamsAndIndices` determines which type args are parameters vs indices by analyzing constructor result types
  - Parameters are uniform (always variables), indices can vary (arbitrary expressions)
  - Example: `Vec : Type -> Nat -> Type` has 1 param (A), 1 index (n)

- **Parameterized/indexed recursor generation** (`kernel/check/recursor.go`)
  - `GenerateRecursorType` generates eliminators for parameterized and indexed inductives
  - `buildAppliedInductiveFull` builds `T params... indices...` with correct de Bruijn indices
  - `buildMotiveTypeFull` builds `P : (indices...) -> T params indices -> Type_j` for indexed motives
  - `buildCaseTypeFull` handles index values in case result types and IH types
  - `extractConstructorIndices` extracts index expressions from constructor results
  - `buildIHType` builds IH types with correct index arguments
  - `indexName` generates names for index binders (n, m, k, ...)

- **Parameterized/indexed recursor reduction** (`internal/eval/recursor.go`)
  - `tryGenericRecursorReduction` correctly handles params and indices in argument layout
  - Constructor spine validation accounts for `NumParams + NumArgs`
  - `buildRecursorInfo` extracts `NumParams` and `NumIndices` from `Inductive`
  - `buildRecursorCallWithIndices` extracts IH indices from constructor args (fixes indexed IH construction)

- **IndexArgPositions metadata for robust IH construction** (`internal/eval/recursor.go`, `kernel/check/env.go`)
  - `ConstructorInfo.IndexArgPositions` maps each recursive arg to its index arg positions in the constructor
  - `computeIndexArgPositions` analyzes recursive arg types at declaration time to extract index positions
  - Uses De Bruijn analysis: for data arg at position j, a Var{V} in its type refers to position j-1-V
  - Precomputed metadata used only when COMPLETE (all indices are variable references)
  - Fallback heuristic retained for backwards compatibility and computed index expressions

### Fixed
- **Partial index metadata bug** (`internal/eval/recursor.go:191-224`)
  - Fixed: `buildRecursorCallWithIndices` now only uses `IndexArgPositions` when metadata is complete
  - Previously, partial metadata (some indices are Vars, others are computed expressions) would
    apply only the partial set without falling back, producing incorrect IH calls
  - Now correctly falls back to heuristic when `len(IndexArgPositions[recArgIdx]) != NumIndices`

### Known Limitations
- **Computed index expressions**: Indices that are computed expressions (e.g., `succ n` in `Vec A (succ n)`)
  are not captured in `IndexArgPositions` metadata and rely on the fallback heuristic
- The fallback heuristic assumes indices precede the recursive arg, which may not hold for all
  indexed inductives with computed index expressions

### Tests
- **Parameterized List tests** (`kernel/check/integration_test.go`)
  - `TestEndToEnd_ParameterizedList`: List declaration, NumParams extraction, eliminator structure
  - `TestEndToEnd_ParameterizedListReduction`: listElim reduction on `nil Nat` and `cons Nat`

- **Indexed Vec tests** (`kernel/check/integration_test.go`)
  - `TestEndToEnd_IndexedVec`: Vec declaration, NumParams/NumIndices extraction, eliminator structure
  - `TestEndToEnd_IndexedVecReduction`: vecElim reduction on `vnil Nat` and `vcons Nat`
  - `TestEndToEnd_IndexArgPositionsMetadata`: verifies IndexArgPositions[2]=[0] for vcons (xs uses n)
  - `TestEndToEnd_NestedVecReduction`: length-2 vector reduction exercising recursive IH construction
  - `TestEndToEnd_ComputedIndexFallback`: verifies metadata is incomplete for `Stepped` inductive with computed index `succ n`

- **Property tests** (`kernel/check/env_test.go`)
  - `TestProperty_IndexArgPositionsCompleteness`: verifies invariant that IndexArgPositions entries are either complete (len == NumIndices) or absent, never partially filled

## [1.5.7] - 2025-12-06

### Added
- **Higher-order recursive detection** (`kernel/check/recursor.go`)
  - Extended `isRecursiveArgType` to detect Pi types with inductive in codomain
  - Correctly identifies `(A -> T)` as recursive when T is the inductive type
  - Uses `OccursIn` for robust detection in nested codomains (e.g., `A -> List T`)

- **Parameterized/indexed inductive infrastructure** (`internal/eval/recursor.go`)
  - Extended `RecursorInfo` with `NumParams` and `NumIndices` fields
  - Updated `tryGenericRecursorReduction` to correctly locate scrutinee when parameters/indices present
  - Updated `buildRecursorInfo` to initialize new fields (currently 0 for non-parameterized)

### Tests
- **Higher-order recursive detection tests** (`kernel/check/recursor_test.go`)
  - `TestIsRecursiveArgType/Higher-order:_A_->_T`: Pi with inductive in codomain
  - `TestIsRecursiveArgType/Higher-order:_A_->_B_->_T`: Nested Pi with inductive
  - `TestIsRecursiveArgType/Higher-order:_A_->_List_T_(applied)`: Applied type in codomain
  - `TestIsRecursiveArgType/Not_higher-order:_A_->_B_(no_T_in_codomain)`: Negative case
  - `TestIsRecursiveArgType/Not_recursive_even_with_T_in_domain`: Domain-only occurrence
- **buildCaseType verification tests** (`kernel/check/recursor_test.go`)
  - `TestBuildCaseType_Nat`: succ case type with IH
  - `TestBuildCaseType_List`: cons case with mixed recursive/non-recursive args
  - `TestBuildCaseType_Tree`: branch case with two recursive args
- **Concurrent registry tests** (`internal/eval/nbe_test.go`)
  - `TestRecursorRegistry_Concurrent`: Thread safety with race detector
  - `TestRecursorInfo_NumParams`: NumParams field verification

## [1.5.6] - 2025-12-06

### Tests
- **Extended integration tests** (`kernel/check/integration_test.go`)
  - `TestEndToEnd_List`: listElim reduces on nil and cons (demonstrates generic recursor)
  - `TestEndToEnd_Tree`: treeElim with multiple recursive args (branch case)
  - `TestEndToEnd_NestedRecursion`: Deep nested recursive calls (msucc (msucc (msucc mzero)))

## [1.5.5] - 2025-12-06

### Added
- **Full constructor type validation** (`kernel/check/env.go`)
  - Uses Checker API (`CheckIsType`) to validate constructor argument types
  - Temporary axiom pattern for self-referential validation

- **Generic recursor reduction** (`internal/eval/recursor.go`)
  - New recursor registry for user-defined inductives
  - `RegisterRecursor` called automatically by `DeclareInductive`
  - `tryGenericRecursorReduction` handles arbitrary recursors
  - Proper IH construction for recursive arguments
  - Built-in natElim/boolElim preserved for backwards compatibility

- **Positivity checker cubical extension** (`kernel/check/positivity_cubical.go`, `positivity_ext.go`)
  - Proper positivity checking for Path, PathP, PathLam, PathApp, Transport
  - Build-tag aware extension pattern for cubical types

### Fixed
- **buildCaseType de Bruijn indices** (`kernel/check/recursor.go`)
  - Uses `subst.Shift` for proper index adjustment when IH binders are interleaved
  - Track IH count for correct shifting

### Tests
- **Integration tests** (`kernel/check/integration_test.go`)
  - `TestEndToEnd_DeclareAndEvaluate`: Full pipeline from declaration to NbE reduction
  - `TestEndToEnd_CustomNatLike`: Custom recursive inductive with mzero/msucc
  - `TestEndToEnd_PositivityRejection`: Negative occurrences properly rejected
  - `TestEndToEnd_IllFormedConstructor`: Unknown types in constructors rejected
  - `TestEndToEnd_RecursorTypeStructure`: Generated eliminator structure verified
- **Generic recursor tests** (`internal/eval/nbe_test.go`)
  - `TestNBE_GenericRecursor`: Unit-like inductive reduction
  - `TestNBE_GenericRecursorWithRecursiveArg`: Nat-like recursive inductive
  - `TestNBE_GenericRecursorNotRegistered`: Unregistered eliminator stays stuck
- **Extended positivity tests** (`kernel/check/positivity_test.go`)
  - `TestCheckPositivity_DoubleDomain`: Nested domain polarity handling
  - `TestCheckPositivity_Sigma`: Sigma type component checking
  - `TestCheckPositivity_Id`: Identity type component checking
- **Constructor validation tests** (`kernel/check/env_test.go`)
  - `TestDeclareInductive_IllFormedConstructor`: Unknown types detected

## [1.5.4] - 2025-12-06

### Fixed
- **DeclareInductive validation** (`kernel/check/env.go`)
  - Now validates inductive type is a Sort before accepting
  - Generates and registers eliminator in GlobalEnv automatically
  - Added `InductiveError` and `ValidationError` types for clear diagnostics

- **GenerateRecursorType universe handling** (`kernel/check/recursor.go`)
  - Now extracts universe level from inductive's type via `extractUniverseLevel()`
  - Motive `P : T -> Type_j` uses correct universe j instead of hardcoded Type0

- **buildCaseType de Bruijn indices** (`kernel/check/recursor.go`)
  - Rewrote with clearer forward-pass algorithm
  - Explicit depth tracking for correct variable indices
  - Removed ad-hoc index arithmetic that could lead to off-by-one errors

- **CheckPositivity conservative handling** (`kernel/check/positivity.go`)
  - Unknown AST node types now checked conservatively using `OccursIn`
  - Rejects if inductive occurs in unknown node at negative position

### Tests
- **Validation and eliminator registration tests** (`kernel/check/env_test.go`)

## [1.5.3] - 2025-12-06

### Added
- **Phase 5: Inductives infrastructure** (`kernel/check/`, `internal/eval/`)
  - **Positivity checker** (`positivity.go`)
    - `CheckPositivity` validates strict positivity for inductive definitions
    - Polarity tracking with proper handling of negative positions
    - Prevents logical inconsistencies from non-well-founded types
    - Rejects nested negative occurrences (e.g., `((T -> A) -> B) -> T`)
  - **Inductive validation** (`env.go`)
    - `DeclareInductive` validates and adds inductives with full checking
    - Constructor result type validation
    - Positivity checking integrated into declaration pipeline
  - **Recursor generation** (`recursor.go`)
    - `GenerateRecursorType` creates eliminator types for inductives
    - Handles zero-arg, single-arg, and recursive constructors
    - Proper de Bruijn index calculation for motive and cases
    - `GenerateRecursorTypeSimple` falls back to hand-crafted types for Nat/Bool
  - **NbE recursor reduction** (`nbe.go`)
    - `natElim` computation rules: reduces on `zero` and `succ n`
    - `boolElim` computation rules: reduces on `true` and `false`
    - Recursive reduction for nested constructors (e.g., `succ (succ zero)`)
    - Stuck terms remain neutral for open/variable scrutinees

- **S-expression parser** (`internal/parser/`)
  - New package for parsing S-expression term syntax
  - Supports all core AST types (Pi, Lam, App, Sigma, Pair, Id, Refl, J, etc.)
  - Cubical types supported via build tag (`-tags cubical`)
  - Round-trip formatting via `FormatTerm()`
  - Helper functions: `ParseTerm`, `ParseMultiple`, `MustParse`

- **CLI connected** (`cmd/hottgo/`)
  - `hottgo --version`: Print version info
  - `hottgo --check FILE`: Type-check a file of S-expression terms
  - `hottgo --eval EXPR`: Evaluate an S-expression term
  - `hottgo --synth EXPR`: Synthesize the type of an S-expression term
  - Interactive REPL mode with `:eval`, `:synth`, `:quit` commands

- **Inductive type documentation** (`docs/rules/inductive.md`)
  - Formation, introduction, and elimination rules
  - Strict positivity requirements with examples
  - Recursor generation schema
  - Computation rules for Nat, Bool, List

- **NewCheckerWithPrimitives** (`kernel/check/check.go`)
  - Convenience constructor for type checker with Nat and Bool

### Changed
- **IVar validation strictness** (`kernel/check/check.go`)
  - `CheckIVar` now rejects interval variables outside path context
  - Previously returned `true` when `ictx == nil`; now returns `false`
  - IVar is only valid inside PathLam scopes

### Tests
- **Positivity checker tests** (`kernel/check/positivity_test.go`)
- **Recursor generation tests** (`kernel/check/recursor_test.go`)
- **NbE recursor reduction tests** (`internal/eval/nbe_test.go`)
- **S-expression parser tests** (`internal/parser/sexpr_test.go`)

### CI/CD
- **Added cubical tests to CI** (`.github/workflows/go.yml`)
  - New step runs `go test -tags cubical -race ./...`
  - Catches regressions in conditionally-compiled cubical type theory code
- **Added golangci-lint to CI** (`.github/workflows/go.yml`)
  - Uses `golangci/golangci-lint-action@v4`
  - Runs linters configured in `.golangci.yml`

### Fixed
- **CRITICAL: AlphaEq lambda annotations** (`internal/core/conv.go`)
  - Fixed bug where lambda annotations were not compared in alpha-equality
  - Previously `λ(x:Nat).x` was incorrectly considered equal to `λ(x:Bool).x`
  - Now properly compares annotations: both nil, or both non-nil and alpha-equal
  - Added regression test `TestAlphaEq_LamAnnotations`

- **CRITICAL: J eliminator reification in NbE** (`internal/eval/nbe.go`)
  - Fixed stuck J terms reifying as nested `App` nodes instead of `ast.J`
  - Added "J" case in `reifyNeutralAt` to properly reconstruct J terms
  - J with spine `[a, c, d, x, y, p]` now correctly reifies to `ast.J{...}`
  - Added regression tests `TestNBE_StuckJReification`, `TestNBE_JComputation`

- **HIGH: synthPathLam cubical evaluation** (`kernel/check/bidir_cubical.go`)
  - Fixed use of wrong evaluation function (`EvalNBE` vs `EvalCubical`)
  - Added `normalizeCubical` helper for proper cubical term normalization
  - PathP endpoints now correctly normalized as AST terms

- **HIGH: VIVar reification** (`internal/eval/nbe_cubical.go`)
  - Fixed incorrect formula in `tryReifyCubical` for interval variables
  - Simplified to use level directly (correct for non-cubical reify context)
  - For proper handling, callers should use `ReifyCubicalAt` directly

- **MEDIUM: J reification in cubical NbE** (`internal/eval/nbe_cubical.go`)
  - Added "J" case in `reifyNeutralCubicalAt` for stuck J terms
  - Mirrors the fix in standard NbE but uses `ReifyCubicalAt`

- **LOW: Pretty printing gaps** (`internal/eval/pretty.go`)
  - Added `VId` and `VRefl` cases in `writeValue`, `ValueEqual`, `valueTypeName`
  - Identity type values now print correctly for debugging

### Security
- **AUDIT: Full codebase audit completed** (`AUDIT_REPORT_FULL.md`)
  - Comprehensive audit of HoTT kernel, NbE, typechecker, and hypergraph library
  - All tests pass including race detector and cubical build tag
  - Coverage: 59.8% overall, 81.8% for typechecker
  - See detailed findings at `/AUDIT_REPORT_FULL.md`
- **AUDIT: Critical bug found in AlphaEq** (now fixed - see above)
  - Lambda annotations not compared in definitional equality
  - Could allow type soundness violations with annotated lambdas
  - See detailed audit report at `/AUDIT_REPORT.md`
- **AUDIT: NbE correctness issues found** (now fixed - see above)
  - Missing J eliminator reification in both standard and cubical NbE
  - Missing VId/VRefl handling in pretty printing functions
  - See detailed audit report at `/nbe_audit_report.md`

### Known Issues (from full audit - see `AUDIT_REPORT_FULL.md`)
- **CRITICAL: synthVar shifting** - Re-analyzed: shift IS correct (audit finding incorrect)
- **HIGH: IEnv.Lookup bounds** - Analyzed: round-trips correctly, no fix needed
- ~~HIGH: IVar validation~~ - **FIXED in this release**
- ~~MEDIUM: Silent fallbacks~~ - **FIXED in this release** (see below)

### Audit Response (fixes from full codebase audit)
- **HIGH: isConstantFamily false positives** (`internal/eval/nbe_cubical.go:277-290`)
  - Fixed `isConstantFamily` to use proper ilevel for reification
  - Changed from `ReifyCubicalAt(0, 0, ...)` to `ReifyCubicalAt(level, c.IEnv.ILen()+1, ...)`
  - Prevents false positives when comparing type families at interval endpoints

- **HIGH: IVar bounds validation** (`kernel/check/bidir_cubical.go`)
  - Added `CheckIVar` method to `Checker` for interval variable validation
  - Added `PushIVar` for scoped interval context management
  - Invalid interval variables now rejected with `ErrUnboundVariable` error
  - Added `errUnboundIVar` error constructor

- **HIGH: PathLam interval context** (`kernel/check/bidir_cubical.go:131-148`)
  - `synthPathLam` now extends interval context before synthesizing body
  - `checkPathLam` also extends interval context for correct IVar validation
  - Uses `PushIVar()/defer popIVar()` pattern for clean scope management

- **MEDIUM: VPathP in PathApply** (`internal/eval/nbe_cubical.go:252-294`)
  - PathApply now handles VPathP and VPath values for endpoint access
  - `PathP @ i0` returns left endpoint X, `PathP @ i1` returns right endpoint Y
  - Neutral interval variables remain stuck as before

- **MEDIUM: Silent fallbacks** (`internal/eval/nbe.go`)
  - Added `DebugMode` flag (set via `HOTTGO_DEBUG=1` env var)
  - Added `evalError` and `reifyError` helpers for consistent error handling
  - Error values now prefixed with `error:` for easier debugging
  - In debug mode, internal errors panic instead of returning fallbacks
  - Updated: nil term, unknown term type, bad_app, reify_error

### Added
- **Cubical Path Types** (Phase 4 M6b - gated by `cubical` build tag)
  - New AST nodes: `Interval`, `I0`, `I1`, `IVar`, `Path`, `PathP`, `PathLam`, `PathApp`, `Transport` (`internal/ast/term_cubical.go`)
  - Path formation: `Path A x y : Type_i` and `PathP A x y : Type_j`
  - Path introduction: `<i> t : PathP (λi. A) t[i0/i] t[i1/i]`
  - Path elimination: `p @ r : A[r/i]` with beta rules `(<i> t) @ i0 --> t[i0/i]`
  - Transport: `transport A e : A[i1/i]` with constant reduction
  - Separate interval de Bruijn index space with `IShift`, `ISubst` (`kernel/subst/subst_cubical.go`)
  - NbE values: `VI0`, `VI1`, `VIVar`, `VPath`, `VPathP`, `VPathLam`, `VTransport` (`internal/eval/nbe_cubical.go`)
  - Interval closures and environment: `IClosure`, `IEnv`
  - Cubical evaluation: `EvalCubical`, `PathApply`, `EvalTransport`, `ReifyCubicalAt`
  - Conversion checking: `alphaEqExtension`, `shiftTermExtension` (`internal/core/conv_cubical.go`)
  - Type checking: `synthPath`, `synthPathP`, `synthPathLam`, `synthPathApp`, `synthTransport` (`kernel/check/bidir_cubical.go`)
  - Pretty printing: S-expression output for cubical terms (`internal/ast/print_cubical.go`)
  - Build tag gating: `//go:build cubical` on all cubical files
  - Extension hooks in base files for conditional cubical support
- **NbE reification bug tests** (`internal/eval/reify_bug_test.go`)
  - `TestReifyFstWithSpineGt1`, `TestReifySndWithSpineGt1` for projection bugs
  - `TestReifyNestedPi`, `TestReifyNestedPiThroughEval` for level tracking bugs
- **Cubical test suite** (`kernel/check/path_test.go`)
  - `TestPathTypeFormation`, `TestPathPTypeFormation`, `TestPathLamIntro`
  - `TestPathAppBetaI0`, `TestPathAppBetaI1`
  - `TestTransportConstant`, `TestTransportTyping`
  - `TestReflAsPath`, `TestISubst`, `TestIShift`
  - `TestCubicalPrinting`, `TestAlphaEqCubical`
- **Documentation** (`docs/rules/path.md`)
  - Complete typing rules for cubical path types
  - Interval type, path formation, introduction, elimination
  - Transport and computation rules
  - Implementation notes with AST and NbE details

### Fixed
- **NbE reification bugs** (`internal/eval/nbe.go`)
  - Fixed level tracking in Pi/Sigma type reification for nested dependent types
  - Fresh variables now use level-indexed indices with proper de Bruijn conversion
  - Nested `Π(A:Type). Π(x:A). A` now correctly reifies to `Pi{Sort{0}, Pi{Var{0}, Var{1}}}`
- **fst/snd projection with spine > 1** (`internal/eval/nbe.go`)
  - Fixed `reifyNeutral` to correctly handle fst/snd neutrals with applied arguments
  - `(fst p) arg` now reifies to `App{Fst{p}, arg}` instead of `App{App{Global{fst}, p}, arg}`

### Changed
- **Documentation for evalJ** (`internal/eval/nbe.go`)
  - Added explanation of why `x == y` check is not needed (typing invariant guarantees it)

## [1.5.0] - 2025-12-05

### Added
- **Martin-Lof Identity Types** (`internal/ast/term.go`, `kernel/check/bidir.go` - Phase 4 M6a)
  - New AST nodes: `Id`, `Refl`, `J` for identity types
  - Formation rule: `Id A x y : Type_i` where `A : Type_i` and `x, y : A`
  - Introduction rule: `refl A x : Id A x x`
  - Elimination rule: `J A C d x y p : C y p` (path induction)
  - Computation rule: `J A C d x x (refl A x) --> d`
- **NbE support** (`internal/eval/nbe.go`)
  - New value types: `VId`, `VRefl`
  - J reduction in semantic domain
- **Substitution support** (`kernel/subst/subst.go`)
  - `Shift` and `Subst` cases for Id, Refl, J
- **Conversion support** (`internal/core/conv.go`)
  - `AlphaEq` and `shiftTerm` cases for identity types
- **Pretty printing** (`internal/ast/print.go`)
  - S-expression printing for Id, Refl, J
- **Documentation** (`docs/rules/id.md`)
  - Complete typing rules for identity types
  - Derived operations: transport, symmetry, transitivity, congruence
- **Test suite** (`kernel/check/id_test.go`)
  - TestIdTypeFormation, TestRefl, TestJComputation
  - TestJTyping, TestTransport (success criterion)
  - TestSymmetry, TestIdTypeMismatch, TestReflTypeMismatch
- **Architecture diagrams** (`DIAGRAMS.md`)
  - 12 Mermaid diagrams covering kernel architecture
  - Term/Value type hierarchies, NbE pipeline, type checking flow

## [1.4.0] - 2025-12-05

### Fixed
- **Homebrew formula** (`Formula/hg.rb`)
  - Updated to v1.4.0 with correct SHA256 checksums for all platforms
- **Winget manifest** (`packaging/winget/watchthelight.hg.yaml`)
  - Updated to v1.4.0 with correct SHA256 checksums for Windows amd64/arm64
- **Chocolatey uninstall script** (`packaging/chocolatey/tools/chocolateyuninstall.ps1`)
  - Fixed binary name from `hottgo.exe` to `hg.exe`
- **GoReleaser configuration** (`.goreleaser.yaml`)
  - Filled brew tap settings: `watchthelight/homebrew-tap`
  - Filled scoop bucket settings: `watchthelight/scoop-bucket`

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
- **Mermaid diagram agent** (`.claude/commands/mermaid.md`)
  - Claude Code slash command for generating Mermaid diagrams
  - Supports flowcharts, class diagrams, sequence diagrams, state diagrams
  - Codebase-aware: understands HypergraphGo package structure and key types
  - Usage: `/mermaid <description of desired diagram>`
- **Platform packaging guide** (`docs/PACKAGING.md`)
  - Comprehensive guide for Winget, Chocolatey, RPM, Homebrew, musl/Alpine, static binaries
  - Includes copy-pasteable manifests with placeholders
  - GitHub Actions workflow examples (manual and GoReleaser)
  - Maintainer checklist for release processes
- **Homebrew formula** (`Formula/hg.rb`)
  - Multi-arch support for macOS (amd64/arm64) and Linux (amd64/arm64)
  - Ready for use with a Homebrew tap
- **Winget manifest** (`packaging/winget/watchthelight.hg.yaml`)
  - Windows Package Manager support for amd64 and arm64
  - Ready for submission to microsoft/winget-pkgs
- **RPM packaging** (`packaging/rpm/hypergraphgo.spec`, `.goreleaser.yaml`)
  - RPM spec file for Fedora/RHEL/CentOS
  - GoReleaser nfpms integration for automated RPM builds
- **musl/Alpine static builds** (`.goreleaser.yaml`)
  - Fully static Linux binaries with netgo and osusergo tags
  - Artifacts: `hg_{{VERSION}}_linux_{{ARCH}}_musl.tar.gz`

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
