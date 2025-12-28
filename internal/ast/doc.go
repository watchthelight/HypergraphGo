// Package ast defines the abstract syntax tree for the HoTT kernel.
//
// This package provides the core term representation used throughout the
// type checker and evaluator. All terms use de Bruijn indices for variable
// binding, which eliminates name capture issues during substitution.
//
// # Term Interface
//
// All syntax nodes implement the [Term] interface via a private isCoreTerm()
// method. This includes both standard type-theoretic constructs and cubical
// type theory extensions.
//
// # Base Types
//
//   - [Sort] - universes Type^n where n is a [Level]
//   - [Var] - de Bruijn indexed variables (0 = innermost binder)
//   - [Global] - references to named global definitions
//
// # Function Types (Π)
//
//   - [Pi] - dependent function types (x:A) → B
//   - [Lam] - lambda abstractions λx.t
//   - [App] - function application (t u)
//
// # Dependent Pairs (Σ)
//
//   - [Sigma] - dependent pair types (x:A) × B
//   - [Pair] - pair constructors (a, b)
//   - [Fst], [Snd] - projections
//
// # Let Bindings
//
//   - [Let] - let x = v : A in body
//
// # Identity Types (Martin-Löf)
//
//   - [Id] - propositional equality Id A x y
//   - [Refl] - reflexivity proof refl A x : Id A x x
//   - [J] - path induction eliminator
//
// # Cubical Extensions
//
// Cubical type theory extensions are defined in term_cubical.go:
//
//   - Interval: [Interval], [I0], [I1], [IVar]
//   - Paths: [Path], [PathP], [PathLam], [PathApp]
//   - Transport: [Transport]
//   - Composition: [Comp], [HComp], [Fill]
//   - Faces: [Face], [FaceTop], [FaceBot], [FaceEq], [FaceAnd], [FaceOr]
//   - Partial: [Partial], [System]
//   - Glue: [Glue], [GlueElem], [Unglue]
//   - Univalence: [UA], [UABeta]
//
// # De Bruijn Indices
//
// Variables use de Bruijn indices where 0 refers to the innermost binder.
// For example, in λx.λy.x, the variable x has index 1 (skipping y at 0).
// Binder names are preserved only for pretty-printing; the kernel operates
// solely on indices.
//
// # Helper Functions
//
//   - [Sprint] - pretty-prints a term to a string
//   - [MkApps] - builds left-associative application chains
//   - [Sort.IsZeroLevel] - checks if a sort is Type₀
//
// # Raw Terms
//
// The raw.go file defines surface syntax terms (RVar, RPi, etc.) that use
// named variables. These are elaborated to core terms during parsing.
package ast
