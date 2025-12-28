# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
