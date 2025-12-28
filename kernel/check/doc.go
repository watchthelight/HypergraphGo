// Package check implements bidirectional type checking for the HoTT kernel.
//
// This package is part of the trusted kernel and provides the core type
// checking algorithm. It uses bidirectional typing with synthesis and
// checking modes to infer and verify types of terms.
//
// # Checker
//
// The [Checker] type performs type checking against a [GlobalEnv]:
//
//	globals := check.NewGlobalEnv()
//	globals.AddAxiom("Nat", ast.Sort{U: 0})
//	checker := check.NewChecker(globals)
//
// Create checkers with different configurations:
//
//   - [NewChecker] - standard checker
//   - [NewCheckerWithEta] - enables η-equality for Π and Σ types
//   - [NewCheckerWithPrimitives] - includes built-in Nat and Bool
//
// # Bidirectional Type Checking
//
// The algorithm operates in two modes:
//
// Synthesis (Synth): infers the type of a term from its structure.
// Works for variables, globals, applications, projections, sorts.
//
//	ty, err := checker.Synth(ctx, span, term)
//
// Checking (Check): verifies a term has an expected type.
// Works for lambdas, pairs, and terms that can also be synthesized.
//
//	err := checker.Check(ctx, span, term, expectedType)
//
// # Type Errors
//
// Type checking failures return [*TypeError] with:
//
//   - Span - source location for error reporting
//   - Kind - [ErrorKind] for programmatic handling
//   - Message - human-readable description
//   - Details - [ErrorDetails] with structured information
//
// Error kinds include:
//
//   - [ErrUnboundVariable] - variable index out of scope
//   - [ErrTypeMismatch] - type didn't match expected
//   - [ErrNotAFunction] - applied a non-function
//   - [ErrNotAPair] - projected from non-pair
//   - [ErrNotAType] - expected a type, got a term
//   - [ErrUnknownGlobal] - undefined global name
//   - [ErrCannotInfer] - cannot synthesize type (use Check mode)
//
// # Global Environment
//
// [GlobalEnv] stores global declarations:
//
//   - [Axiom] - uninterpreted type declarations
//   - [Definition] - defined constants with body
//   - [Inductive] - inductive type families with constructors
//   - [Primitive] - built-in operations
//
// Definitions have [Transparency] controlling unfolding:
//
//   - [Transparent] - always unfold in conversion
//   - [Opaque] - never unfold
//
// # Cubical Features
//
// The checker supports cubical type theory with interval context tracking:
//
//   - [Checker.ICtxDepth] - current depth of interval bindings
//   - [Checker.CheckIVar] - validate interval variable index
//   - [Checker.PushIVar] - extend interval context (returns cleanup func)
//
// Cubical terms (paths, transport, composition) are checked via bidir_cubical.go.
//
// # Inductive Types
//
// Inductive types are checked for strict positivity to ensure consistency.
// The eliminator (recursor) type is automatically derived. See:
//
//   - [GlobalEnv.AddInductive] - register inductive type
//   - positivity.go - strict positivity checking
//   - recursor.go - eliminator type synthesis
//
// # Definitional Equality
//
// Type checking uses NbE-based conversion checking from internal/core:
//
//   - Terms are normalized to weak head normal form
//   - Normalized terms are compared structurally
//   - Optional η-rules for functions and pairs
package check
