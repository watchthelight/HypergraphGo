// Package unify implements unification for type inference and elaboration.
//
// The unification algorithm solves constraints of the form `t₁ = t₂` by
// finding substitutions for metavariables that make the terms equal.
//
// # Overview
//
// This package implements Miller pattern unification, which handles
// metavariables applied to distinct bound variables (patterns like `?X x y`
// where x and y are distinct). Pattern unification is decidable and has
// most-general solutions when they exist.
//
// # Constraints
//
// [Constraint] represents an equality constraint between two terms:
//
//	c := Constraint{LHS: term1, RHS: term2}
//
// The unifier collects constraints and processes them, attempting to find
// a consistent assignment of terms to metavariables.
//
// # Unifier
//
// [Unifier] processes constraints and produces solutions:
//
//	u := unify.NewUnifier()
//	u.AddConstraint(t1, t2)
//	result := u.Solve()
//
//	for id, solution := range result.Solved {
//	    // metavariable ?id = solution
//	}
//
// # Results
//
// [UnifyResult] contains:
//
//   - Solved: map from metavariable IDs to their solutions
//   - Unsolved: constraints that couldn't be solved (stuck)
//   - Errors: unification failures (conflicting constraints)
//
// # Algorithm
//
// The unifier normalizes terms and proceeds by cases:
//
//   - Identical terms: constraint is satisfied
//   - Metavariable on one side: attempt to solve (occurs check, pattern check)
//   - Structural terms: decompose and recurse on subterms
//   - Stuck terms: defer to unsolved list
//
// # Zonking
//
// After solving, use [zonkTerm] to substitute solutions into terms.
// This replaces metavariable references with their solved values.
//
// # Cubical Support
//
// The unifier handles cubical type theory terms including:
//
//   - Interval: I0, I1, IVar
//   - Faces: FaceTop, FaceBot, FaceEq, FaceAnd, FaceOr
//   - Partial types: Partial, System
//   - Composition: Comp, HComp, Fill
//   - Glue types: Glue, GlueElem, Unglue
//   - Univalence: UA, UABeta
//   - Higher inductive types: HITApp
//
// # Limitations
//
// Miller pattern unification cannot solve all constraints. Non-pattern
// constraints (like `?X (f y) = t` where the argument is not a variable)
// are deferred to the unsolved list. Higher-order unification is
// undecidable in general.
package unify
