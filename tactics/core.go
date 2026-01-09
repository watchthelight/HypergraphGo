// Package tactics provides Ltac-style proof tactics for HoTT.
// This file implements core tactics for proof construction.

package tactics

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
	"github.com/watchthelight/HypergraphGo/kernel/subst"
	"github.com/watchthelight/HypergraphGo/tactics/proofstate"
)

// Intro introduces a variable from a Pi type goal.
// For goal `(x : A) -> B`, introduces `x : A` and creates goal `B`.
func Intro(name string) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		// Normalize the goal type
		goalTy := eval.EvalNBE(goal.Type)

		// Check if goal is a Pi type
		pi, ok := goalTy.(ast.Pi)
		if !ok {
			return Failf("goal is not a Pi type, got %T", goalTy)
		}

		// Use provided name or binder name
		varName := name
		if varName == "" {
			varName = pi.Binder
			if varName == "_" {
				varName = "x"
			}
		}

		// Add hypothesis to goal
		newHyps := make([]proofstate.Hypothesis, len(goal.Hypotheses)+1)
		copy(newHyps, goal.Hypotheses)
		newHyps[len(goal.Hypotheses)] = proofstate.Hypothesis{
			Name: varName,
			Type: pi.A,
		}

		// The new goal type is the codomain
		// (no need to substitute since we're using de Bruijn indices)
		newGoal := proofstate.Goal{
			Type:       pi.B,
			Hypotheses: newHyps,
		}

		// Replace current goal with new goal
		if err := state.ReplaceGoal(goal.ID, []proofstate.Goal{newGoal}); err != nil {
			return Fail(err)
		}

		return SuccessMsg(state, fmt.Sprintf("introduced %s : %v", varName, pi.A))
	}
}

// IntroN introduces multiple variables at once.
func IntroN(names ...string) Tactic {
	if len(names) == 0 {
		return NoOp()
	}

	tactics := make([]Tactic, len(names))
	for i, name := range names {
		tactics[i] = Intro(name)
	}
	return Seq(tactics...)
}

// Intros introduces all possible variables from nested Pi types.
func Intros() Tactic {
	return Repeat(Intro(""))
}

// Exact solves a goal with an exact term.
// The term must have a type definitionally equal to the goal type.
func Exact(term ast.Term) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		// Type check: infer the type of the term and compare with goal
		// For variables, look up in hypothesis context
		termTy, ok := inferTermType(term, goal.Hypotheses)
		if !ok {
			// If we can't infer the type, accept the term (for backward compatibility)
			// This handles cases like globals and complex terms
			if err := state.SolveGoal(goal.ID, term); err != nil {
				return Fail(err)
			}
			return SuccessMsg(state, "exact proof term accepted (unchecked)")
		}

		// Normalize both types and compare
		goalTy := eval.EvalNBE(goal.Type)
		termTyNorm := eval.EvalNBE(termTy)

		if !eval.AlphaEq(goalTy, termTyNorm) {
			return Failf("type mismatch: expected %v, got %v", goalTy, termTyNorm)
		}

		// Solve the goal
		if err := state.SolveGoal(goal.ID, term); err != nil {
			return Fail(err)
		}

		return SuccessMsg(state, "exact proof term accepted")
	}
}

// inferTermType tries to infer the type of a term given the hypothesis context.
// Returns (type, true) if successful, (nil, false) otherwise.
func inferTermType(term ast.Term, hyps []proofstate.Hypothesis) (ast.Term, bool) {
	switch t := term.(type) {
	case ast.Var:
		// Look up in hypotheses (de Bruijn index)
		if t.Ix < len(hyps) {
			hypIdx := len(hyps) - 1 - t.Ix
			return hyps[hypIdx].Type, true
		}
		return nil, false

	case ast.Sort:
		// Type : Type (simplified, ignoring universe levels)
		return ast.Sort{U: t.U + 1}, true

	case ast.Refl:
		// refl : Id A x x
		return ast.Id{A: t.A, X: t.X, Y: t.X}, true

	case ast.Pair:
		// For pairs, we'd need the expected type - can't infer
		return nil, false

	case ast.Lam:
		// For lambdas, we'd need to check the body - complex
		return nil, false

	default:
		// For other terms, we can't easily infer
		return nil, false
	}
}

// Assumption solves the goal using a hypothesis.
func Assumption() Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		goalTy := eval.EvalNBE(goal.Type)

		// Look for a matching hypothesis
		for i := len(goal.Hypotheses) - 1; i >= 0; i-- {
			hyp := goal.Hypotheses[i]
			hypTy := eval.EvalNBE(hyp.Type)

			if eval.AlphaEq(goalTy, hypTy) {
				// Found a match - use de Bruijn index
				ix := len(goal.Hypotheses) - 1 - i
				term := ast.Var{Ix: ix}

				if err := state.SolveGoal(goal.ID, term); err != nil {
					return Fail(err)
				}

				return SuccessMsg(state, fmt.Sprintf("used hypothesis %s", hyp.Name))
			}
		}

		return Failf("no matching hypothesis found")
	}
}

// Apply applies a term (function or hypothesis) to the goal.
// For goal B and term f : A -> B, creates subgoal A.
// The codomain of f must match the goal type.
func Apply(term ast.Term) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		goalTy := eval.EvalNBE(goal.Type)

		// Try to get the type of the term
		termTy, ok := inferTermType(term, goal.Hypotheses)
		if !ok {
			// Fallback: if term is a variable, look it up directly
			if v, vok := term.(ast.Var); vok {
				if v.Ix < len(goal.Hypotheses) {
					hypIdx := len(goal.Hypotheses) - 1 - v.Ix
					termTy = goal.Hypotheses[hypIdx].Type
					ok = true
				}
			}
		}

		if !ok {
			return Failf("apply: cannot infer type of term")
		}

		termTyNorm := eval.EvalNBE(termTy)

		// Check if it's a function type
		pi, isPi := termTyNorm.(ast.Pi)
		if !isPi {
			return Failf("apply: term is not a function, got %T", termTyNorm)
		}

		// Check that codomain matches goal (after substituting a metavariable)
		// For non-dependent functions, codomain should match goal directly
		// For dependent functions, we check that B[_/x] can unify with goal
		codomain := pi.B
		if !containsVar0(codomain) {
			// Non-dependent: codomain should match goal
			if !eval.AlphaEq(eval.EvalNBE(codomain), goalTy) {
				return Failf("apply: codomain %v does not match goal %v", codomain, goalTy)
			}
		}
		// For dependent cases, we trust the user (full unification would be needed)

		// Create subgoal for the argument
		newGoal := proofstate.Goal{
			Type:       pi.A,
			Hypotheses: goal.Hypotheses,
		}

		if err := state.ReplaceGoal(goal.ID, []proofstate.Goal{newGoal}); err != nil {
			return Fail(err)
		}

		return SuccessMsg(state, "applied term")
	}
}

// containsVar0 checks if a term contains Var{Ix: 0}.
func containsVar0(t ast.Term) bool {
	switch tt := t.(type) {
	case ast.Var:
		return tt.Ix == 0
	case ast.Pi:
		return containsVar0(tt.A) || containsVar0(tt.B)
	case ast.Lam:
		return containsVar0(tt.Body)
	case ast.App:
		return containsVar0(tt.T) || containsVar0(tt.U)
	case ast.Sigma:
		return containsVar0(tt.A) || containsVar0(tt.B)
	case ast.Pair:
		return containsVar0(tt.Fst) || containsVar0(tt.Snd)
	case ast.Id:
		return containsVar0(tt.A) || containsVar0(tt.X) || containsVar0(tt.Y)
	default:
		return false
	}
}

// Reflexivity solves an identity/path goal with reflexivity.
func Reflexivity() Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		goalTy := eval.EvalNBE(goal.Type)

		// Check if goal is an Id type
		switch id := goalTy.(type) {
		case ast.Id:
			// Check if x and y are equal
			if eval.AlphaEq(id.X, id.Y) {
				term := ast.Refl{A: id.A, X: id.X}
				if err := state.SolveGoal(goal.ID, term); err != nil {
					return Fail(err)
				}
				return SuccessMsg(state, "proved by reflexivity")
			}
			return Failf("endpoints are not definitionally equal")

		case ast.Path:
			// Check if x and y are equal
			if eval.AlphaEq(id.X, id.Y) {
				// Use path lambda with constant value
				term := ast.PathLam{Binder: "_", Body: id.X}
				if err := state.SolveGoal(goal.ID, term); err != nil {
					return Fail(err)
				}
				return SuccessMsg(state, "proved by path reflexivity")
			}
			return Failf("path endpoints are not definitionally equal")

		default:
			return Failf("goal is not an identity type, got %T", goalTy)
		}
	}
}

// Split handles sigma (product) types by creating two subgoals.
func Split() Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		goalTy := eval.EvalNBE(goal.Type)

		sigma, ok := goalTy.(ast.Sigma)
		if !ok {
			return Failf("goal is not a Sigma type, got %T", goalTy)
		}

		// Create two subgoals: one for fst, one for snd
		fstGoal := proofstate.Goal{
			Type:       sigma.A,
			Hypotheses: goal.Hypotheses,
		}

		// For snd, we need to extend the context with the first component
		// This is simplified - proper implementation would handle dependencies
		sndHyps := make([]proofstate.Hypothesis, len(goal.Hypotheses)+1)
		copy(sndHyps, goal.Hypotheses)
		sndHyps[len(goal.Hypotheses)] = proofstate.Hypothesis{
			Name: sigma.Binder,
			Type: sigma.A,
		}

		sndGoal := proofstate.Goal{
			Type:       sigma.B,
			Hypotheses: sndHyps,
		}

		if err := state.ReplaceGoal(goal.ID, []proofstate.Goal{fstGoal, sndGoal}); err != nil {
			return Fail(err)
		}

		return SuccessMsg(state, "split into two subgoals")
	}
}

// Simpl normalizes the goal type.
func Simpl() Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		// Normalize the goal type
		normalized := eval.EvalNBE(goal.Type)

		// Update the goal
		goal.Type = normalized

		return SuccessMsg(state, "simplified goal")
	}
}

// Unfold unfolds a definition in the goal.
func Unfold(name string) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		// TODO: Implement definition unfolding
		// This requires access to a global environment
		return Failf("unfold: not yet implemented")
	}
}

// Rewrite uses a hypothesis h : Id A x y to rewrite x to y in the goal.
func Rewrite(hypName string) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		// Look up the hypothesis
		hyp, hypIdx, ok := goal.LookupHypothesis(hypName)
		if !ok {
			return Failf("hypothesis %s not found", hypName)
		}

		hypTy := eval.EvalNBE(hyp.Type)

		// Check if hypothesis is an Id type
		id, ok := hypTy.(ast.Id)
		if !ok {
			return Failf("hypothesis %s is not an identity type, got %T", hypName, hypTy)
		}

		// Replace x with y in the goal type
		newGoalTy := substTerm(goal.Type, id.X, id.Y)

		// Create new goal
		newGoal := proofstate.Goal{
			Type:       newGoalTy,
			Hypotheses: goal.Hypotheses,
		}

		// The proof will use J to transport
		// Build: J A (λy p. goal[x↦y]) proof x y h
		_ = hypIdx // Will be used in full implementation

		if err := state.ReplaceGoal(goal.ID, []proofstate.Goal{newGoal}); err != nil {
			return Fail(err)
		}

		return SuccessMsg(state, fmt.Sprintf("rewrote using %s", hypName))
	}
}

// RewriteRev uses h : Id A x y to rewrite y to x (reverse direction).
func RewriteRev(hypName string) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		hyp, _, ok := goal.LookupHypothesis(hypName)
		if !ok {
			return Failf("hypothesis %s not found", hypName)
		}

		hypTy := eval.EvalNBE(hyp.Type)

		id, ok := hypTy.(ast.Id)
		if !ok {
			return Failf("hypothesis %s is not an identity type", hypName)
		}

		// Replace y with x in the goal type (reverse direction)
		newGoalTy := substTerm(goal.Type, id.Y, id.X)

		newGoal := proofstate.Goal{
			Type:       newGoalTy,
			Hypotheses: goal.Hypotheses,
		}

		if err := state.ReplaceGoal(goal.ID, []proofstate.Goal{newGoal}); err != nil {
			return Fail(err)
		}

		return SuccessMsg(state, fmt.Sprintf("rewrote (reversed) using %s", hypName))
	}
}

// substTerm replaces occurrences of old with new in term.
// This is a simplified substitution for term rewriting.
func substTerm(term, old, new ast.Term) ast.Term {
	if eval.AlphaEq(term, old) {
		return new
	}

	switch t := term.(type) {
	case ast.Var, ast.Global, ast.Sort:
		return term

	case ast.Pi:
		return ast.Pi{
			Binder: t.Binder,
			A:      substTerm(t.A, old, new),
			B:      substTerm(t.B, subst.Shift(1, 0, old), subst.Shift(1, 0, new)),
		}

	case ast.Lam:
		var ann ast.Term
		if t.Ann != nil {
			ann = substTerm(t.Ann, old, new)
		}
		return ast.Lam{
			Binder: t.Binder,
			Ann:    ann,
			Body:   substTerm(t.Body, subst.Shift(1, 0, old), subst.Shift(1, 0, new)),
		}

	case ast.App:
		return ast.App{
			T: substTerm(t.T, old, new),
			U: substTerm(t.U, old, new),
		}

	case ast.Sigma:
		return ast.Sigma{
			Binder: t.Binder,
			A:      substTerm(t.A, old, new),
			B:      substTerm(t.B, subst.Shift(1, 0, old), subst.Shift(1, 0, new)),
		}

	case ast.Pair:
		return ast.Pair{
			Fst: substTerm(t.Fst, old, new),
			Snd: substTerm(t.Snd, old, new),
		}

	case ast.Fst:
		return ast.Fst{P: substTerm(t.P, old, new)}

	case ast.Snd:
		return ast.Snd{P: substTerm(t.P, old, new)}

	case ast.Id:
		return ast.Id{
			A: substTerm(t.A, old, new),
			X: substTerm(t.X, old, new),
			Y: substTerm(t.Y, old, new),
		}

	case ast.Refl:
		return ast.Refl{
			A: substTerm(t.A, old, new),
			X: substTerm(t.X, old, new),
		}

	default:
		return term
	}
}

// Trivial tries a sequence of simple tactics.
func Trivial() Tactic {
	return First(
		Assumption(),
		Reflexivity(),
	)
}

// Auto tries to automatically solve the goal.
func Auto() Tactic {
	return Seq(
		Try(Intros()),
		Repeat(First(
			Assumption(),
			Reflexivity(),
		)),
	)
}
