// Package ctx provides the typing context for the HoTT kernel.
//
// A typing context (often written Γ in type theory) is an ordered sequence of
// variable bindings that tracks which variables are in scope and their types
// during type checking. This package implements contexts as telescopes, where
// later bindings may depend on earlier ones.
//
// # Context Structure
//
// The context is represented as a slice of bindings, where each binding pairs
// a name with a type. The slice is ordered from oldest to newest binding:
//
//	Ctx{Tele: []{A: Type, x: A, y: A}}
//
// This internal ordering is the reverse of de Bruijn indices: the most recent
// binding (y in the example) has de Bruijn index 0, while the oldest (A) has
// the highest index.
//
// # Variable Lookup
//
// LookupVar translates between de Bruijn indices and the internal slice order:
//
//	ctx.LookupVar(0)  -- returns type of most recent binding (y: A)
//	ctx.LookupVar(2)  -- returns type of oldest binding (A: Type)
//
// This convention matches the de Bruijn representation where index 0 refers
// to the innermost (most recently bound) variable.
//
// # Context Extension
//
// When type checking enters a binder (lambda, pi, let, etc.), the context is
// extended with the new binding:
//
//	ctx.Extend("x", typeOfX)
//
// The Extend method mutates the context by appending to the telescope. When
// exiting a binder scope, use Drop to restore the previous context:
//
//	shortened := ctx.Drop()
//
// Note that Drop returns a new context (immutable operation) while Extend
// modifies the receiver (mutable operation). This asymmetry reflects typical
// usage: extending on entry, then discarding the extension on exit.
//
// # Telescope Dependencies
//
// In a telescope, types may refer to earlier bindings. For example, in:
//
//	(A : Type) (x : A) (p : Id A x x)
//
// The type A appears in the binding for x, and both A and x appear in the
// binding for p. The kernel/subst package handles the necessary index
// adjustments when these types are used in different contexts.
//
// # Usage in Type Checking
//
// The context flows through bidirectional type checking:
//
//   - synth(ctx, term) synthesizes a type for term in context ctx
//   - check(ctx, term, type) verifies term has type in context ctx
//
// When checking under a binder, the context is extended with the bound variable,
// and types of later bindings may need shifting via kernel/subst.Shift.
//
// # References
//
//   - Martin-Löf, P. (1984). Intuitionistic Type Theory. Bibliopolis.
//   - The Univalent Foundations Program (2013). Homotopy Type Theory:
//     Univalent Foundations of Mathematics. §A.2 (Contexts and Telescopes).
package ctx
