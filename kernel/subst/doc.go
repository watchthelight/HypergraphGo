// Package subst provides de Bruijn variable shifting and substitution.
//
// This package is part of the trusted kernel and implements the fundamental
// operations on de Bruijn indexed terms. The algorithms follow Types and
// Programming Languages (TAPL) §6.2.
//
// # De Bruijn Indices
//
// Variables are represented as indices rather than names:
//
//   - Index 0 refers to the innermost bound variable
//   - Index k refers to the variable bound by the k-th enclosing binder
//   - Free variables have indices >= context length
//
// # Shifting
//
// [Shift] adjusts variable indices when opening/closing binders:
//
//	// Shift free variables by 1 (entering a binder)
//	shifted := subst.Shift(1, 0, term)
//
//	// Shift variables >= cutoff by d
//	shifted := subst.Shift(d, cutoff, term)
//
// The cutoff parameter tracks binding depth - variables below the cutoff
// are bound and unchanged; variables at or above are shifted.
//
// # Substitution
//
// [Subst] replaces a variable with a term:
//
//	// Replace variable 0 with s in t
//	result := subst.Subst(0, s, t)
//
// Substitution automatically shifts the substituted term when entering
// binders to maintain correct indices.
//
// # Binder Handling
//
// Both Shift and Subst increment the cutoff/index when entering binders:
//
//   - Pi, Lam, Sigma, Let - bind one variable
//   - Body positions have cutoff+1 or j+1
//
// This ensures bound variables are never modified.
//
// # Interval Variables (Cubical)
//
// Cubical type theory uses a separate namespace for interval variables:
//
//   - [IShift] - shifts interval variables (IVar indices)
//   - [ISubst] - substitutes interval terms for interval variables
//
// These operate independently of term variable operations.
//
// # Face Formulas
//
// Face formula versions handle the Face types:
//
//   - [ShiftFace] - shifts term variables in face formulas
//   - [SubstFace] - substitutes in face formulas
//   - [IShiftFace] - shifts interval variables in face formulas
//   - [ISubstFace] - substitutes interval terms in face formulas
//
// # Extension Points
//
// Unknown term types are handled via extension functions:
//
//   - shiftExtension - for cubical and other extended terms
//   - substExtension - corresponding substitution
//
// This allows the base implementation to support additional term types.
//
// # TAPL Reference
//
// The implementation follows Pierce's "Types and Programming Languages":
//
//   - §6.2.1 - Shifting definition
//   - §6.2.2 - Substitution definition
//   - §6.2.3 - Properties and lemmas
package subst
