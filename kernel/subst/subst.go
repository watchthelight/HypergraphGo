package subst

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// maxFreeVar returns the maximum free variable index in a term, or -1 if closed.
// This is used to optimize shifting and substitution.
func maxFreeVar(t ast.Term) int {
	return maxFreeVarUnder(t, 0)
}

// maxFreeVarUnder returns the maximum free variable index adjusted for depth.
func maxFreeVarUnder(t ast.Term, depth int) int {
	if t == nil {
		return -1
	}
	switch tm := t.(type) {
	case ast.Var:
		if tm.Ix >= depth {
			return tm.Ix - depth
		}
		return -1 // Bound variable
	case ast.Sort, ast.Global:
		return -1
	case ast.Pi:
		return max2(maxFreeVarUnder(tm.A, depth), maxFreeVarUnder(tm.B, depth+1))
	case ast.Lam:
		a := -1
		if tm.Ann != nil {
			a = maxFreeVarUnder(tm.Ann, depth)
		}
		return max2(a, maxFreeVarUnder(tm.Body, depth+1))
	case ast.App:
		return max2(maxFreeVarUnder(tm.T, depth), maxFreeVarUnder(tm.U, depth))
	case ast.Sigma:
		return max2(maxFreeVarUnder(tm.A, depth), maxFreeVarUnder(tm.B, depth+1))
	case ast.Pair:
		return max2(maxFreeVarUnder(tm.Fst, depth), maxFreeVarUnder(tm.Snd, depth))
	case ast.Fst:
		return maxFreeVarUnder(tm.P, depth)
	case ast.Snd:
		return maxFreeVarUnder(tm.P, depth)
	case ast.Let:
		ann := -1
		if tm.Ann != nil {
			ann = maxFreeVarUnder(tm.Ann, depth)
		}
		return max2(max2(ann, maxFreeVarUnder(tm.Val, depth)), maxFreeVarUnder(tm.Body, depth+1))
	case ast.Id:
		return max2(max2(maxFreeVarUnder(tm.A, depth), maxFreeVarUnder(tm.X, depth)), maxFreeVarUnder(tm.Y, depth))
	case ast.Refl:
		return max2(maxFreeVarUnder(tm.A, depth), maxFreeVarUnder(tm.X, depth))
	case ast.J:
		m := maxFreeVarUnder(tm.A, depth)
		m = max2(m, maxFreeVarUnder(tm.C, depth))
		m = max2(m, maxFreeVarUnder(tm.D, depth))
		m = max2(m, maxFreeVarUnder(tm.X, depth))
		m = max2(m, maxFreeVarUnder(tm.Y, depth))
		m = max2(m, maxFreeVarUnder(tm.P, depth))
		return m
	default:
		// For extension types, assume potentially unbounded - don't optimize
		return 1000000 // Large number to disable optimization for unknown types
	}
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Shift shifts all free de Bruijn variables >= cutoff by d positions.
// Follows TAPL §6.2.1.
func Shift(d, cutoff int, t ast.Term) ast.Term {
	if t == nil {
		return nil
	}
	// Optimization: shifting by 0 is identity
	if d == 0 {
		return t
	}
	// Optimization: if no free variables >= cutoff, no changes needed
	if maxFreeVar(t) < cutoff {
		return t
	}
	switch tm := t.(type) {
	case ast.Var:
		if tm.Ix >= cutoff {
			return ast.Var{Ix: tm.Ix + d}
		}
		return tm
	case ast.Sort:
		return tm
	case ast.Global:
		return tm
	case ast.Pi:
		return ast.Pi{
			Binder: tm.Binder,
			A:      Shift(d, cutoff, tm.A),
			B:      Shift(d, cutoff+1, tm.B),
		}
	case ast.Lam:
		return ast.Lam{
			Binder: tm.Binder,
			Ann:    Shift(d, cutoff, tm.Ann),
			Body:   Shift(d, cutoff+1, tm.Body),
		}
	case ast.App:
		return ast.App{
			T: Shift(d, cutoff, tm.T),
			U: Shift(d, cutoff, tm.U),
		}
	case ast.Sigma:
		return ast.Sigma{
			Binder: tm.Binder,
			A:      Shift(d, cutoff, tm.A),
			B:      Shift(d, cutoff+1, tm.B),
		}
	case ast.Pair:
		return ast.Pair{
			Fst: Shift(d, cutoff, tm.Fst),
			Snd: Shift(d, cutoff, tm.Snd),
		}
	case ast.Fst:
		return ast.Fst{P: Shift(d, cutoff, tm.P)}
	case ast.Snd:
		return ast.Snd{P: Shift(d, cutoff, tm.P)}
	case ast.Let:
		return ast.Let{
			Binder: tm.Binder,
			Ann:    Shift(d, cutoff, tm.Ann),
			Val:    Shift(d, cutoff, tm.Val),
			Body:   Shift(d, cutoff+1, tm.Body),
		}
	case ast.Id:
		return ast.Id{
			A: Shift(d, cutoff, tm.A),
			X: Shift(d, cutoff, tm.X),
			Y: Shift(d, cutoff, tm.Y),
		}
	case ast.Refl:
		return ast.Refl{
			A: Shift(d, cutoff, tm.A),
			X: Shift(d, cutoff, tm.X),
		}
	case ast.J:
		return ast.J{
			A: Shift(d, cutoff, tm.A),
			C: Shift(d, cutoff, tm.C),
			D: Shift(d, cutoff, tm.D),
			X: Shift(d, cutoff, tm.X),
			Y: Shift(d, cutoff, tm.Y),
			P: Shift(d, cutoff, tm.P),
		}
	default:
		// Try extension handlers (e.g., cubical terms when built with -tags cubical)
		if result, ok := shiftExtension(d, cutoff, t); ok {
			return result
		}
		// Unknown term types are returned unchanged (treated as terminals)
		return t
	}
}

// IsClosed returns true if the term has no free variables.
func IsClosed(t ast.Term) bool {
	return maxFreeVar(t) < 0
}

// Subst substitutes s for variable j in t, adjusting indices.
// Follows TAPL §6.2.2.
func Subst(j int, s ast.Term, t ast.Term) ast.Term {
	if t == nil {
		return nil
	}
	// Optimization: if s is closed, use the faster substClosed which avoids shifting
	if IsClosed(s) {
		return substClosed(j, s, t)
	}
	switch tm := t.(type) {
	case ast.Var:
		if tm.Ix == j {
			return s
		} else if tm.Ix > j {
			return ast.Var{Ix: tm.Ix - 1}
		}
		return tm
	case ast.Sort:
		return tm
	case ast.Global:
		return tm
	case ast.Pi:
		return ast.Pi{
			Binder: tm.Binder,
			A:      Subst(j, s, tm.A),
			B:      Subst(j+1, Shift(1, 0, s), tm.B),
		}
	case ast.Lam:
		return ast.Lam{
			Binder: tm.Binder,
			Ann:    Subst(j, s, tm.Ann),
			Body:   Subst(j+1, Shift(1, 0, s), tm.Body),
		}
	case ast.App:
		return ast.App{
			T: Subst(j, s, tm.T),
			U: Subst(j, s, tm.U),
		}
	case ast.Sigma:
		return ast.Sigma{
			Binder: tm.Binder,
			A:      Subst(j, s, tm.A),
			B:      Subst(j+1, Shift(1, 0, s), tm.B),
		}
	case ast.Pair:
		return ast.Pair{
			Fst: Subst(j, s, tm.Fst),
			Snd: Subst(j, s, tm.Snd),
		}
	case ast.Fst:
		return ast.Fst{P: Subst(j, s, tm.P)}
	case ast.Snd:
		return ast.Snd{P: Subst(j, s, tm.P)}
	case ast.Let:
		return ast.Let{
			Binder: tm.Binder,
			Ann:    Subst(j, s, tm.Ann),
			Val:    Subst(j, s, tm.Val),
			Body:   Subst(j+1, Shift(1, 0, s), tm.Body),
		}
	case ast.Id:
		return ast.Id{
			A: Subst(j, s, tm.A),
			X: Subst(j, s, tm.X),
			Y: Subst(j, s, tm.Y),
		}
	case ast.Refl:
		return ast.Refl{
			A: Subst(j, s, tm.A),
			X: Subst(j, s, tm.X),
		}
	case ast.J:
		return ast.J{
			A: Subst(j, s, tm.A),
			C: Subst(j, s, tm.C),
			D: Subst(j, s, tm.D),
			X: Subst(j, s, tm.X),
			Y: Subst(j, s, tm.Y),
			P: Subst(j, s, tm.P),
		}
	default:
		// Try extension handlers (e.g., cubical terms when built with -tags cubical)
		if result, ok := substExtension(j, s, t); ok {
			return result
		}
		// Unknown term types are returned unchanged (treated as terminals)
		return t
	}
}

// substClosed is an optimized substitution for when s is a closed term.
// Since s has no free variables, we don't need to shift it when going under binders.
func substClosed(j int, s ast.Term, t ast.Term) ast.Term {
	if t == nil {
		return nil
	}
	switch tm := t.(type) {
	case ast.Var:
		if tm.Ix == j {
			return s
		} else if tm.Ix > j {
			return ast.Var{Ix: tm.Ix - 1}
		}
		return tm
	case ast.Sort:
		return tm
	case ast.Global:
		return tm
	case ast.Pi:
		return ast.Pi{
			Binder:   tm.Binder,
			A:        substClosed(j, s, tm.A),
			B:        substClosed(j+1, s, tm.B), // No shift needed!
			Implicit: tm.Implicit,
		}
	case ast.Lam:
		return ast.Lam{
			Binder:   tm.Binder,
			Ann:      substClosed(j, s, tm.Ann),
			Body:     substClosed(j+1, s, tm.Body), // No shift needed!
			Implicit: tm.Implicit,
		}
	case ast.App:
		return ast.App{
			T:        substClosed(j, s, tm.T),
			U:        substClosed(j, s, tm.U),
			Implicit: tm.Implicit,
		}
	case ast.Sigma:
		return ast.Sigma{
			Binder: tm.Binder,
			A:      substClosed(j, s, tm.A),
			B:      substClosed(j+1, s, tm.B), // No shift needed!
		}
	case ast.Pair:
		return ast.Pair{
			Fst: substClosed(j, s, tm.Fst),
			Snd: substClosed(j, s, tm.Snd),
		}
	case ast.Fst:
		return ast.Fst{P: substClosed(j, s, tm.P)}
	case ast.Snd:
		return ast.Snd{P: substClosed(j, s, tm.P)}
	case ast.Let:
		return ast.Let{
			Binder: tm.Binder,
			Ann:    substClosed(j, s, tm.Ann),
			Val:    substClosed(j, s, tm.Val),
			Body:   substClosed(j+1, s, tm.Body), // No shift needed!
		}
	case ast.Id:
		return ast.Id{
			A: substClosed(j, s, tm.A),
			X: substClosed(j, s, tm.X),
			Y: substClosed(j, s, tm.Y),
		}
	case ast.Refl:
		return ast.Refl{
			A: substClosed(j, s, tm.A),
			X: substClosed(j, s, tm.X),
		}
	case ast.J:
		return ast.J{
			A: substClosed(j, s, tm.A),
			C: substClosed(j, s, tm.C),
			D: substClosed(j, s, tm.D),
			X: substClosed(j, s, tm.X),
			Y: substClosed(j, s, tm.Y),
			P: substClosed(j, s, tm.P),
		}
	default:
		// For extension types, use the extension handler directly
		// We pass s unchanged since it's closed (no shifting needed)
		if result, ok := substClosedExtension(j, s, t); ok {
			return result
		}
		// Unknown term types are returned unchanged
		return t
	}
}
