// Package subst implements de Bruijn variable substitution and shifting operations.
//
// This package provides the core variable manipulation operations for the HoTT
// kernel, following the formal treatment in Pierce's "Types and Programming
// Languages" (TAPL) §6.2.1-6.2.2. These operations are foundational for
// correctly handling variable binding during type checking and evaluation.
//
// # De Bruijn Indices
//
// The kernel uses de Bruijn indices for variable representation, where each
// variable is represented by a natural number indicating how many binders
// separate the variable from its binding site. Index 0 refers to the innermost
// (most recently bound) variable:
//
//	λ. λ. 1 0   -- outer λ is 1, inner λ is 0
//
// This nameless representation eliminates variable capture during substitution
// and simplifies alpha-equivalence checking.
//
// # Core Operations
//
// The package provides two fundamental operations:
//
// Shift adjusts variable indices when moving terms across binders. When a term
// enters a new binding context, its free variables must be shifted up to account
// for the new binder:
//
//	Shift(d, cutoff, t)
//	  - d: amount to shift (positive = up, negative = down)
//	  - cutoff: only shift variables with index >= cutoff
//	  - t: term to transform
//
// Subst performs capture-avoiding substitution, replacing a variable with a term
// while correctly adjusting indices:
//
//	Subst(j, s, t)
//	  - j: index of variable to replace
//	  - s: term to substitute
//	  - t: term in which to perform substitution
//
// When substituting under binders, Subst automatically shifts the substituted
// term to prevent variable capture.
//
// # Cubical Extension
//
// For cubical type theory, the package extends these operations to handle
// interval variables, which form a separate namespace from term variables:
//
//	IShift(d, cutoff, t) - shifts interval variables in t
//	ISubst(j, s, t)      - substitutes interval term s for interval variable j
//
// Interval variables appear in path types, path abstractions, and composition
// operations. Forms that bind interval variables (PathP, PathLam, Transport,
// Comp, Fill) increment the cutoff when recursing into their bodies.
//
// # Face Formula Operations
//
// The package also provides operations for face formulas used in cubical types:
//
//	IShiftFace(d, cutoff, f) - shifts interval variables in face formula f
//	ISubstFace(j, s, f)      - substitutes interval term s in face formula f
//
// Face formulas represent constraints on interval variables (e.g., i = 0, i = 1)
// and are used in partial types and systems. Substitution in face formulas
// performs simplification: when substituting i0 into (i = 0), the result is ⊤;
// when substituting i0 into (i = 1), the result is ⊥.
//
// # References
//
//   - Pierce, B. C. (2002). Types and Programming Languages. MIT Press. §6.2
//   - Cohen, C., Coquand, T., Huber, S., & Mörtberg, A. (2018). Cubical Type Theory:
//     A Constructive Interpretation of the Univalence Axiom. TYPES 2015.
package subst
