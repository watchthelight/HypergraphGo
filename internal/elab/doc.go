// Package elab provides elaboration from surface syntax to core terms.
//
// Elaboration is the process of transforming user-friendly surface syntax
// (with named variables, implicit arguments, and holes) into the kernel's
// core representation (with de Bruijn indices and no implicit arguments).
//
// # Overview
//
// The elaboration pipeline:
//
//  1. Parse surface syntax (STerm) from the parser package
//  2. Name resolution: resolve names to de Bruijn indices
//  3. Implicit insertion: fill in implicit arguments
//  4. Hole solving: create metavariables and solve via unification
//  5. Zonking: substitute solved metavariables
//  6. Output: kernel-ready ast.Term
//
// # Surface Syntax
//
// [STerm] is the interface for all surface terms. Surface syntax supports:
//
//   - Named variables: [SVar]
//   - Global references: [SGlobal]
//   - Universes: [SType]
//   - Pi types (explicit and implicit): [SPi], [SArrow]
//   - Sigma types: [SSigma], [SProd]
//   - Lambda abstractions: [SLam]
//   - Application: [SApp]
//   - Pairs and projections: [SPair], [SFst], [SSnd]
//   - Let bindings: [SLet]
//   - Holes: [SHole] (unnamed _) and [SNamedHole] (named ?x)
//   - Identity types: [SId], [SRefl], [SJ]
//   - Cubical types: [SPath], [SPathLam], [SPathApp], [STransport]
//
// # Elaboration Context
//
// [ElabCtx] maintains state during elaboration:
//
//   - Variable bindings for name resolution
//   - Metavariable store for holes and implicit arguments
//   - Type checking environment
//   - Source location for error messages
//
// Create a context with [NewElabCtx] or extend with [ElabCtx.Extend].
//
// # Metavariables
//
// [MetaStore] manages metavariables (placeholders for unknown terms).
// Each metavariable has:
//
//   - Unique [MetaID]
//   - Expected type
//   - Solution (once found)
//   - Source [Span] for error messages
//
// Metavariables are solved via unification in the unify package.
//
// # Zonking
//
// [Zonk] substitutes solved metavariables into a term. After unification,
// call Zonk to produce a metavariable-free term.
//
// # Error Handling
//
// [ElabError] contains source location and context for better error messages.
// Use [Elaborator.Elaborate] for type inference or [Elaborator.ElaborateCheck]
// for checking against an expected type.
//
// # Example
//
//	ctx := elab.NewElabCtx()
//	e := elab.NewElaborator()
//
//	// Elaborate a surface term
//	term, ty, err := e.Elaborate(ctx, surfaceTerm)
//	if err != nil {
//	    // Handle error with source location
//	}
//
//	// Zonk to substitute solved metavariables
//	finalTerm := elab.Zonk(ctx.Metas, term)
package elab
