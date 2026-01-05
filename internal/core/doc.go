// Package core implements definitional equality checking (conversion) for the HoTT kernel.
//
// This package provides the conversion checking algorithm that determines if two terms
// are definitionally equal. This is essential for type checking, as types are compared
// for equality modulo computation (beta reduction, eta rules, etc.).
//
// # Conversion Checking
//
// The [Conv] function reports whether two terms are definitionally equal under a
// typing environment:
//
//	env := core.NewEnv()
//	equal := core.Conv(env, term1, term2, ConvOptions{EnableEta: true})
//
// The algorithm uses Normalization by Evaluation (NbE):
//
//  1. Evaluate both terms to semantic values using internal/eval
//  2. Reify values back to normal forms
//  3. Apply η-expansion if enabled
//  4. Compare normal forms structurally via [AlphaEq]
//
// # Options
//
// [ConvOptions] controls conversion behavior:
//
//   - EnableEta: Enable η-equality for functions (Π) and pairs (Σ)
//
// # Alpha-Equivalence
//
// [AlphaEq] compares two terms for structural equality modulo de Bruijn indices.
// Since variables use de Bruijn indices, alpha-equivalence is simply structural
// comparison. This is used as the final step in conversion checking.
//
// # Eta-Equality
//
// When [ConvOptions.EnableEta] is true, the converter recognizes eta-equal terms:
//
//   - Functions: f ≡ \x. f x (eta rule for Π types)
//   - Pairs: p ≡ (fst p, snd p) (eta rule for Σ types)
//
// # Environment
//
// [Env] represents a typing environment mapping de Bruijn indices to values.
// It wraps [eval.Env] and is extended with terms using [Env.Extend].
//
// # Cubical Type Theory
//
// The package supports cubical type theory extensions including:
//
//   - Interval types: [ast.Interval], [ast.I0], [ast.I1], [ast.IVar]
//   - Paths: [ast.Path], [ast.PathP], [ast.PathLam], [ast.PathApp]
//   - Transport: [ast.Transport]
//   - Composition: [ast.Comp], [ast.HComp], [ast.Fill]
//   - Face formulas: [ast.FaceTop], [ast.FaceBot], [ast.FaceEq], [ast.FaceAnd], [ast.FaceOr]
//   - Partial types: [ast.Partial], [ast.System]
//   - Glue types: [ast.Glue], [ast.GlueElem], [ast.Unglue]
//   - Univalence: [ast.UA], [ast.UABeta]
//   - Higher inductive types: [ast.HITApp]
//
// Cubical equality is handled via extension functions (alphaEqExtension,
// shiftTermExtension) that delegate to appropriate handlers for cubical terms.
//
// # Legacy API
//
// [ConvLegacy] provides backward compatibility with code using [EtaFlags].
// New code should use [Conv] with [ConvOptions] instead.
package core
