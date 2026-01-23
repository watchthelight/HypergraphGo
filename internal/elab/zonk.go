// Package elab provides elaboration from surface syntax to core terms.
// This file implements zonking - substituting solved metavariables with their solutions.

package elab

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// Zonk substitutes all solved metavariables in a term with their solutions.
// Unsolved metavariables are left in place.
func Zonk(metas *MetaStore, t ast.Term) ast.Term {
	z := &zonker{metas: metas}
	return z.zonkTerm(t)
}

// ZonkFull substitutes all metavariables and errors if any remain unsolved.
func ZonkFull(metas *MetaStore, t ast.Term) (ast.Term, error) {
	result := Zonk(metas, t)
	if HasMeta(result) {
		return result, fmt.Errorf("unsolved metavariables remain")
	}
	return result, nil
}

// zonker performs metavariable substitution.
type zonker struct {
	metas *MetaStore
}

// zonkTerm recursively substitutes metavariables with their solutions.
func (z *zonker) zonkTerm(t ast.Term) ast.Term {
	if t == nil {
		return nil
	}

	switch tt := t.(type) {
	case ast.Meta:
		if sol, ok := z.metas.GetSolution(MetaID(tt.ID)); ok {
			// Recursively zonk the solution
			result := z.zonkTerm(sol)
			// Apply arguments to solution (implicit since these are meta args)
			for _, arg := range tt.Args {
				result = ast.App{T: result, U: z.zonkTerm(arg), Implicit: true}
			}
			return result
		}
		// Keep meta but zonk arguments
		newArgs := make([]ast.Term, len(tt.Args))
		for i, arg := range tt.Args {
			newArgs[i] = z.zonkTerm(arg)
		}
		return ast.Meta{ID: tt.ID, Args: newArgs}

	case ast.Var, ast.Global, ast.Sort:
		return t

	case ast.Pi:
		return ast.Pi{
			Binder:   tt.Binder,
			A:        z.zonkTerm(tt.A),
			B:        z.zonkTerm(tt.B),
			Implicit: tt.Implicit,
		}

	case ast.Lam:
		return ast.Lam{
			Binder:   tt.Binder,
			Ann:      z.zonkTerm(tt.Ann),
			Body:     z.zonkTerm(tt.Body),
			Implicit: tt.Implicit,
		}

	case ast.App:
		return ast.App{
			T:        z.zonkTerm(tt.T),
			U:        z.zonkTerm(tt.U),
			Implicit: tt.Implicit,
		}

	case ast.Sigma:
		return ast.Sigma{
			Binder: tt.Binder,
			A:      z.zonkTerm(tt.A),
			B:      z.zonkTerm(tt.B),
		}

	case ast.Pair:
		return ast.Pair{
			Fst: z.zonkTerm(tt.Fst),
			Snd: z.zonkTerm(tt.Snd),
		}

	case ast.Fst:
		return ast.Fst{P: z.zonkTerm(tt.P)}

	case ast.Snd:
		return ast.Snd{P: z.zonkTerm(tt.P)}

	case ast.Let:
		return ast.Let{
			Binder: tt.Binder,
			Ann:    z.zonkTerm(tt.Ann),
			Val:    z.zonkTerm(tt.Val),
			Body:   z.zonkTerm(tt.Body),
		}

	case ast.Id:
		return ast.Id{
			A: z.zonkTerm(tt.A),
			X: z.zonkTerm(tt.X),
			Y: z.zonkTerm(tt.Y),
		}

	case ast.Refl:
		return ast.Refl{
			A: z.zonkTerm(tt.A),
			X: z.zonkTerm(tt.X),
		}

	case ast.J:
		return ast.J{
			A: z.zonkTerm(tt.A),
			C: z.zonkTerm(tt.C),
			D: z.zonkTerm(tt.D),
			X: z.zonkTerm(tt.X),
			Y: z.zonkTerm(tt.Y),
			P: z.zonkTerm(tt.P),
		}

	// Path types
	case ast.Path:
		return ast.Path{
			A: z.zonkTerm(tt.A),
			X: z.zonkTerm(tt.X),
			Y: z.zonkTerm(tt.Y),
		}

	case ast.PathP:
		return ast.PathP{
			A: z.zonkTerm(tt.A),
			X: z.zonkTerm(tt.X),
			Y: z.zonkTerm(tt.Y),
		}

	case ast.PathLam:
		return ast.PathLam{
			Binder: tt.Binder,
			Body:   z.zonkTerm(tt.Body),
		}

	case ast.PathApp:
		return ast.PathApp{
			P: z.zonkTerm(tt.P),
			R: z.zonkTerm(tt.R),
		}

	case ast.Transport:
		return ast.Transport{
			A: z.zonkTerm(tt.A),
			E: z.zonkTerm(tt.E),
		}

	// Interval endpoints and variables
	case ast.Interval, ast.I0, ast.I1, ast.IVar:
		return t

	default:
		// Unknown term type, return as-is
		return t
	}
}

// HasMeta checks if a term contains any metavariables.
func HasMeta(t ast.Term) bool {
	if t == nil {
		return false
	}

	switch tt := t.(type) {
	case ast.Meta:
		return true

	case ast.Var, ast.Global, ast.Sort, ast.Interval, ast.I0, ast.I1, ast.IVar:
		return false

	case ast.Pi:
		return HasMeta(tt.A) || HasMeta(tt.B)

	case ast.Lam:
		return HasMeta(tt.Ann) || HasMeta(tt.Body)

	case ast.App:
		return HasMeta(tt.T) || HasMeta(tt.U)

	case ast.Sigma:
		return HasMeta(tt.A) || HasMeta(tt.B)

	case ast.Pair:
		return HasMeta(tt.Fst) || HasMeta(tt.Snd)

	case ast.Fst:
		return HasMeta(tt.P)

	case ast.Snd:
		return HasMeta(tt.P)

	case ast.Let:
		return HasMeta(tt.Ann) || HasMeta(tt.Val) || HasMeta(tt.Body)

	case ast.Id:
		return HasMeta(tt.A) || HasMeta(tt.X) || HasMeta(tt.Y)

	case ast.Refl:
		return HasMeta(tt.A) || HasMeta(tt.X)

	case ast.J:
		return HasMeta(tt.A) || HasMeta(tt.C) || HasMeta(tt.D) ||
			HasMeta(tt.X) || HasMeta(tt.Y) || HasMeta(tt.P)

	case ast.Path:
		return HasMeta(tt.A) || HasMeta(tt.X) || HasMeta(tt.Y)

	case ast.PathP:
		return HasMeta(tt.A) || HasMeta(tt.X) || HasMeta(tt.Y)

	case ast.PathLam:
		return HasMeta(tt.Body)

	case ast.PathApp:
		return HasMeta(tt.P) || HasMeta(tt.R)

	case ast.Transport:
		return HasMeta(tt.A) || HasMeta(tt.E)

	default:
		return false
	}
}

// ZonkType zonks a type term (same as Zonk, but semantically for types).
func ZonkType(metas *MetaStore, t ast.Term) ast.Term {
	return Zonk(metas, t)
}

// ZonkCtx zonks all types in an elaboration context.
func ZonkCtx(metas *MetaStore, ctx *ElabCtx) *ElabCtx {
	if ctx == nil {
		return nil
	}

	z := &zonker{metas: metas}

	// Create a new context with zonked types
	result := NewElabCtx()
	result.Metas = ctx.Metas
	result.Globals = ctx.Globals

	// Zonk all bindings
	for _, b := range ctx.Bindings {
		zonkedTy := z.zonkTerm(b.Type)
		var zonkedDef ast.Term
		if b.Def != nil {
			zonkedDef = z.zonkTerm(b.Def)
		}
		newBinding := ElabBinding{
			Name:  b.Name,
			Type:  zonkedTy,
			Icity: b.Icity,
			Def:   zonkedDef,
		}
		result.Bindings = append(result.Bindings, newBinding)
	}

	// Copy implicit bindings
	for k, v := range ctx.IBindings {
		result.IBindings[k] = v
	}

	return result
}

// ReportUnsolvedMetas returns an error describing all unsolved metavariables.
func ReportUnsolvedMetas(metas *MetaStore) error {
	unsolved := metas.Unsolved()
	if len(unsolved) == 0 {
		return nil
	}

	msg := fmt.Sprintf("%d unsolved metavariable(s):\n", len(unsolved))
	for _, id := range unsolved {
		entry := metas.MustLookup(id)
		if entry.Name != "" {
			msg += fmt.Sprintf("  ?%s : %v at %v\n", entry.Name, entry.Type, entry.Span)
		} else {
			msg += fmt.Sprintf("  ?%d : %v at %v\n", id, entry.Type, entry.Span)
		}
	}
	return fmt.Errorf("%s", msg)
}

// CollectMetas collects all metavariable IDs in a term.
func CollectMetas(t ast.Term) []MetaID {
	collector := &metaCollector{seen: make(map[MetaID]bool)}
	collector.collect(t)
	return collector.result
}

type metaCollector struct {
	seen   map[MetaID]bool
	result []MetaID
}

func (c *metaCollector) collect(t ast.Term) {
	if t == nil {
		return
	}

	switch tt := t.(type) {
	case ast.Meta:
		id := MetaID(tt.ID)
		if !c.seen[id] {
			c.seen[id] = true
			c.result = append(c.result, id)
		}
		for _, arg := range tt.Args {
			c.collect(arg)
		}

	case ast.Var, ast.Global, ast.Sort, ast.Interval, ast.I0, ast.I1, ast.IVar:
		// No metas

	case ast.Pi:
		c.collect(tt.A)
		c.collect(tt.B)

	case ast.Lam:
		c.collect(tt.Ann)
		c.collect(tt.Body)

	case ast.App:
		c.collect(tt.T)
		c.collect(tt.U)

	case ast.Sigma:
		c.collect(tt.A)
		c.collect(tt.B)

	case ast.Pair:
		c.collect(tt.Fst)
		c.collect(tt.Snd)

	case ast.Fst:
		c.collect(tt.P)

	case ast.Snd:
		c.collect(tt.P)

	case ast.Let:
		c.collect(tt.Ann)
		c.collect(tt.Val)
		c.collect(tt.Body)

	case ast.Id:
		c.collect(tt.A)
		c.collect(tt.X)
		c.collect(tt.Y)

	case ast.Refl:
		c.collect(tt.A)
		c.collect(tt.X)

	case ast.J:
		c.collect(tt.A)
		c.collect(tt.C)
		c.collect(tt.D)
		c.collect(tt.X)
		c.collect(tt.Y)
		c.collect(tt.P)

	case ast.Path:
		c.collect(tt.A)
		c.collect(tt.X)
		c.collect(tt.Y)

	case ast.PathP:
		c.collect(tt.A)
		c.collect(tt.X)
		c.collect(tt.Y)

	case ast.PathLam:
		c.collect(tt.Body)

	case ast.PathApp:
		c.collect(tt.P)
		c.collect(tt.R)

	case ast.Transport:
		c.collect(tt.A)
		c.collect(tt.E)
	}
}
