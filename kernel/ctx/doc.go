// Package ctx provides typing context management for the HoTT kernel.
//
// This package is part of the trusted kernel boundary. It manages the
// typing context (also called a telescope) that tracks variable bindings
// during type checking.
//
// # De Bruijn Indices
//
// Variables use de Bruijn indices where 0 refers to the most recently
// bound variable. For example, in the context [x:A, y:B, z:C]:
//
//   - Index 0 refers to z (most recent)
//   - Index 1 refers to y
//   - Index 2 refers to x (oldest)
//
// # Context Type
//
// [Ctx] holds a telescope of bindings:
//
//	ctx := &ctx.Ctx{}
//	ctx.Extend("x", typeA)
//	ctx.Extend("y", typeB)
//
// # Operations
//
//   - [Ctx.Extend] - adds a new binding (mutates)
//   - [Ctx.LookupVar] - retrieves type by de Bruijn index
//   - [Ctx.Drop] - returns context without most recent binding (immutable)
//   - [Ctx.Len] - number of bindings
//
// # Binding
//
// Each [Binding] contains:
//
//   - Name - variable name (for pretty-printing only)
//   - Ty - the variable's type as an [ast.Term]
//
// The name is purely for display; the kernel operates on indices.
//
// # Thread Safety
//
// Ctx is NOT safe for concurrent use. The Extend method mutates the
// context in place. Use separate contexts for concurrent operations.
//
// # Example
//
// Type checking a lambda term (λx:A. body):
//
//	// Before entering the lambda body:
//	ctx.Extend("x", domainType)
//
//	// Check body against codomain
//	err := checker.Check(ctx, span, body, codomainType)
//
//	// Alternatively, use Drop for immutable operations:
//	innerCtx := Ctx{Tele: append(ctx.Tele, Binding{Name: "x", Ty: domain})}
package ctx
