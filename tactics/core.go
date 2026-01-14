// core.go implements core tactics: Intro, Apply, Exact, Split, Rewrite, etc.
//
// See doc.go for package overview.

package tactics

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/elab"
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

		// Solve the goal with Lam{..., Body: Meta{subgoalMetaID}}
		err := state.SolveGoalWithSubgoals(
			goal.ID,
			[]ast.Term{pi.B},
			[][]proofstate.Hypothesis{newHyps},
			func(metaIDs []elab.MetaID) ast.Term {
				return ast.Lam{
					Binder: varName,
					Ann:    pi.A,
					Body:   ast.Meta{ID: int(metaIDs[0])},
				}
			},
		)
		if err != nil {
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
			hypTy := hyps[hypIdx].Type
			// Shift the type to account for hypotheses added after this one.
			// When the hypothesis was stored, the context was smaller by (t.Ix + 1).
			// For example, x at de Bruijn 0 in [A, x] has type stored relative to [A],
			// so we need to shift by 1 to make it relative to [A, x].
			shift := t.Ix + 1
			hypTy = subst.Shift(shift, 0, hypTy)
			return hypTy, true
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
			// Shift the hypothesis type to account for hypotheses added after it.
			// The shift amount is len(hyps) - slice_index = de_bruijn_index + 1.
			shift := len(goal.Hypotheses) - i
			hypTy := subst.Shift(shift, 0, hyp.Type)
			hypTy = eval.EvalNBE(hypTy)

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

		// Solve the goal with App{term, Meta{argMetaID}}
		// The subgoal will be the argument type
		err := state.SolveGoalWithSubgoals(
			goal.ID,
			[]ast.Term{pi.A},
			nil, // Use same hypotheses
			func(metaIDs []elab.MetaID) ast.Term {
				return ast.App{
					T: term,
					U: ast.Meta{ID: int(metaIDs[0])},
				}
			},
		)
		if err != nil {
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

		// For snd, we need to extend the context with the first component
		// This is simplified - proper implementation would handle dependencies
		sndHyps := make([]proofstate.Hypothesis, len(goal.Hypotheses)+1)
		copy(sndHyps, goal.Hypotheses)
		sndHyps[len(goal.Hypotheses)] = proofstate.Hypothesis{
			Name: sigma.Binder,
			Type: sigma.A,
		}

		// Solve the goal with Pair{Meta{fstMetaID}, Meta{sndMetaID}}
		err := state.SolveGoalWithSubgoals(
			goal.ID,
			[]ast.Term{sigma.A, sigma.B},
			[][]proofstate.Hypothesis{nil, sndHyps}, // nil means use parent's hyps
			func(metaIDs []elab.MetaID) ast.Term {
				return ast.Pair{
					Fst: ast.Meta{ID: int(metaIDs[0])},
					Snd: ast.Meta{ID: int(metaIDs[1])},
				}
			},
		)
		if err != nil {
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

		// The base case goal type: goal[y↦x] (with x instead of y)
		// This is what we need to prove for the reflexivity case
		baseGoalTy := substTerm(goal.Type, id.Y, id.X)

		// Build the motive: C = λy. λp. goal[x↦y]
		// C : (y : A) -> Id A x y -> Type
		// We need to shift goal.Type under two binders and replace x with Var{1}
		goalShifted := subst.Shift(2, 0, goal.Type)
		xShifted := subst.Shift(2, 0, id.X)
		motiveBody := substTerm(goalShifted, xShifted, ast.Var{Ix: 1})
		motive := ast.Lam{
			Binder: "y",
			Ann:    id.A,
			Body: ast.Lam{
				Binder: "p",
				Ann:    ast.Id{A: subst.Shift(1, 0, id.A), X: subst.Shift(1, 0, id.X), Y: ast.Var{Ix: 0}},
				Body:   motiveBody,
			},
		}

		// Solve with J A C (Meta for base case) x y h
		err := state.SolveGoalWithSubgoals(
			goal.ID,
			[]ast.Term{baseGoalTy},
			nil, // Use same hypotheses
			func(metaIDs []elab.MetaID) ast.Term {
				return ast.J{
					A: id.A,
					C: motive,
					D: ast.Meta{ID: int(metaIDs[0])}, // Base case proof
					X: id.X,
					Y: id.Y,
					P: ast.Var{Ix: hypIdx}, // The hypothesis
				}
			},
		)
		if err != nil {
			return Fail(err)
		}

		return SuccessMsg(state, fmt.Sprintf("rewrote using %s", hypName))
	}
}

// RewriteRev uses h : Id A x y to rewrite y to x (reverse direction).
// This requires symmetry (sym h) to reverse the equality direction.
func RewriteRev(hypName string) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		hyp, hypIdx, ok := goal.LookupHypothesis(hypName)
		if !ok {
			return Failf("hypothesis %s not found", hypName)
		}

		hypTy := eval.EvalNBE(hyp.Type)

		id, ok := hypTy.(ast.Id)
		if !ok {
			return Failf("hypothesis %s is not an identity type", hypName)
		}

		// The base case goal type: goal[x↦y] (with y instead of x)
		baseGoalTy := substTerm(goal.Type, id.X, id.Y)

		// Build the motive for reverse direction: C = λx. λp. goal[y↦x]
		// C : (x : A) -> Id A x y -> Type
		goalShifted := subst.Shift(2, 0, goal.Type)
		yShifted := subst.Shift(2, 0, id.Y)
		motiveBody := substTerm(goalShifted, yShifted, ast.Var{Ix: 1})
		motive := ast.Lam{
			Binder: "x",
			Ann:    id.A,
			Body: ast.Lam{
				Binder: "p",
				Ann:    ast.Id{A: subst.Shift(1, 0, id.A), X: ast.Var{Ix: 0}, Y: subst.Shift(1, 0, id.Y)},
				Body:   motiveBody,
			},
		}

		// Solve with J A C (Meta for base case) y x (sym h)
		// For reverse rewriting, we use J on the reversed path
		err := state.SolveGoalWithSubgoals(
			goal.ID,
			[]ast.Term{baseGoalTy},
			nil,
			func(metaIDs []elab.MetaID) ast.Term {
				// Build symmetry: J A (λz.λ_. Id A z x) refl y x h
				symMotive := ast.Lam{
					Binder: "z",
					Ann:    id.A,
					Body: ast.Lam{
						Binder: "_",
						Ann:    ast.Id{A: subst.Shift(1, 0, id.A), X: subst.Shift(1, 0, id.X), Y: ast.Var{Ix: 0}},
						Body:   ast.Id{A: subst.Shift(2, 0, id.A), X: ast.Var{Ix: 1}, Y: subst.Shift(2, 0, id.X)},
					},
				}
				symProof := ast.J{
					A: id.A,
					C: symMotive,
					D: ast.Refl{A: id.A, X: id.X},
					X: id.X,
					Y: id.Y,
					P: ast.Var{Ix: hypIdx},
				}
				return ast.J{
					A: id.A,
					C: motive,
					D: ast.Meta{ID: int(metaIDs[0])},
					X: id.Y,
					Y: id.X,
					P: symProof,
				}
			},
		)
		if err != nil {
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

// Contradiction solves any goal using a hypothesis of type Empty.
// For a goal G with hypothesis h : Empty, produces emptyElim (λ_. G) h.
func Contradiction() Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		// Look for a hypothesis of type Empty
		for i := len(goal.Hypotheses) - 1; i >= 0; i-- {
			hyp := goal.Hypotheses[i]
			hypTy := eval.EvalNBE(hyp.Type)

			// Check if the hypothesis type is Empty
			if g, ok := hypTy.(ast.Global); ok && g.Name == "Empty" {
				// Found an Empty hypothesis
				// Build: emptyElim (λ_. G) h
				ix := len(goal.Hypotheses) - 1 - i
				hypVar := ast.Var{Ix: ix}

				// The motive is λ_. G (a function ignoring its Empty argument)
				// But we need to shift G under the lambda binder
				shiftedGoal := subst.Shift(1, 0, goal.Type)
				motive := ast.Lam{
					Binder: "_",
					Ann:    ast.Global{Name: "Empty"},
					Body:   shiftedGoal,
				}

				// Build emptyElim P h
				// emptyElim : (P : Empty → Type) → (e : Empty) → P e
				proofTerm := ast.App{
					T: ast.App{
						T: ast.Global{Name: "emptyElim"},
						U: motive,
					},
					U: hypVar,
				}

				if err := state.SolveGoal(goal.ID, proofTerm); err != nil {
					return Fail(err)
				}

				return SuccessMsg(state, fmt.Sprintf("contradiction from %s : Empty", hyp.Name))
			}
		}

		return Failf("no hypothesis of type Empty found")
	}
}

// extractSumArgs extracts A and B from a Sum A B type.
// Returns (A, B, true) if the type is Sum A B, (nil, nil, false) otherwise.
func extractSumArgs(ty ast.Term) (ast.Term, ast.Term, bool) {
	// Sum A B is represented as App{App{Global{Sum}, A}, B}
	app1, ok := ty.(ast.App)
	if !ok {
		return nil, nil, false
	}
	app2, ok := app1.T.(ast.App)
	if !ok {
		return nil, nil, false
	}
	sum, ok := app2.T.(ast.Global)
	if !ok || sum.Name != "Sum" {
		return nil, nil, false
	}
	return app2.U, app1.U, true
}

// Left proves a Sum A B goal by providing a proof of A.
// For goal Sum A B, creates subgoal A and uses inl A B to construct the proof.
func Left() Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		goalTy := eval.EvalNBE(goal.Type)

		// Extract A and B from Sum A B
		a, b, ok := extractSumArgs(goalTy)
		if !ok {
			return Failf("goal is not a Sum type, got %v", goalTy)
		}

		// Solve the goal with inl A B ?meta
		err := state.SolveGoalWithSubgoals(
			goal.ID,
			[]ast.Term{a}, // Subgoal type is A
			nil,           // Use same hypotheses
			func(metaIDs []elab.MetaID) ast.Term {
				// Build: inl A B ?meta
				return ast.MkApps(
					ast.Global{Name: "inl"},
					a,
					b,
					ast.Meta{ID: int(metaIDs[0])},
				)
			},
		)
		if err != nil {
			return Fail(err)
		}

		return SuccessMsg(state, "using left injection (inl)")
	}
}

// Right proves a Sum A B goal by providing a proof of B.
// For goal Sum A B, creates subgoal B and uses inr A B to construct the proof.
func Right() Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		goalTy := eval.EvalNBE(goal.Type)

		// Extract A and B from Sum A B
		a, b, ok := extractSumArgs(goalTy)
		if !ok {
			return Failf("goal is not a Sum type, got %v", goalTy)
		}

		// Solve the goal with inr A B ?meta
		err := state.SolveGoalWithSubgoals(
			goal.ID,
			[]ast.Term{b}, // Subgoal type is B
			nil,           // Use same hypotheses
			func(metaIDs []elab.MetaID) ast.Term {
				// Build: inr A B ?meta
				return ast.MkApps(
					ast.Global{Name: "inr"},
					a,
					b,
					ast.Meta{ID: int(metaIDs[0])},
				)
			},
		)
		if err != nil {
			return Fail(err)
		}

		return SuccessMsg(state, "using right injection (inr)")
	}
}

// Destruct performs case analysis on a hypothesis.
// For h : Sum A B, creates two subgoals (one for inl, one for inr).
// For h : Bool, creates two subgoals (one for true, one for false).
func Destruct(hypName string) Tactic {
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

		// Try to destruct based on the type
		switch {
		case isSumType(hypTy):
			return destructSum(state, goal, *hyp, hypIdx, hypTy)
		case isBoolType(hypTy):
			return destructBool(state, goal, *hyp, hypIdx)
		default:
			return Failf("cannot destruct hypothesis %s of type %v", hypName, hypTy)
		}
	}
}

// isSumType checks if a type is Sum A B.
func isSumType(ty ast.Term) bool {
	_, _, ok := extractSumArgs(ty)
	return ok
}

// isBoolType checks if a type is Bool.
func isBoolType(ty ast.Term) bool {
	g, ok := ty.(ast.Global)
	return ok && g.Name == "Bool"
}

// destructSum performs case analysis on a Sum hypothesis.
// For h : Sum A B, creates two subgoals:
// - Goal[a/h] for the inl case (with a : A)
// - Goal[b/h] for the inr case (with b : B)
func destructSum(state *proofstate.ProofState, goal *proofstate.Goal, hyp proofstate.Hypothesis, hypIdx int, hypTy ast.Term) TacticResult {
	a, b, _ := extractSumArgs(hypTy)
	hypVarIdx := len(goal.Hypotheses) - 1 - hypIdx

	// Build new hypotheses for each branch
	// For inl: replace h : Sum A B with a : A
	// For inr: replace h : Sum A B with b : B
	inlHyps := make([]proofstate.Hypothesis, len(goal.Hypotheses))
	copy(inlHyps, goal.Hypotheses)
	inlHyps[hypIdx] = proofstate.Hypothesis{Name: "a_" + hyp.Name, Type: a}

	inrHyps := make([]proofstate.Hypothesis, len(goal.Hypotheses))
	copy(inrHyps, goal.Hypotheses)
	inrHyps[hypIdx] = proofstate.Hypothesis{Name: "b_" + hyp.Name, Type: b}

	// Build the proof term:
	// sumElim A B (λs. G) (λa. ?goal1) (λb. ?goal2) h
	// where G is the goal type with the hypothesis

	// The motive is λs. G where s is the Sum value
	// Since h is at position hypVarIdx, we need to shift G under the lambda
	// and replace the variable at hypVarIdx with Var{0}
	shiftedGoal := subst.Shift(1, 0, goal.Type)

	// Replace Var{hypVarIdx+1} with Var{0} in shiftedGoal
	// (The shift moved h from hypVarIdx to hypVarIdx+1)
	motive := ast.Lam{
		Binder: "_",
		Body:   shiftedGoal,
	}

	err := state.SolveGoalWithSubgoals(
		goal.ID,
		[]ast.Term{goal.Type, goal.Type}, // Both branches have the same goal type
		[][]proofstate.Hypothesis{inlHyps, inrHyps},
		func(metaIDs []elab.MetaID) ast.Term {
			// Build: sumElim A B motive (λa. ?goal1) (λb. ?goal2) h
			inlCase := ast.Lam{
				Binder: "a_" + hyp.Name,
				Ann:    a,
				Body:   ast.Meta{ID: int(metaIDs[0])},
			}
			inrCase := ast.Lam{
				Binder: "b_" + hyp.Name,
				Ann:    b,
				Body:   ast.Meta{ID: int(metaIDs[1])},
			}
			return ast.MkApps(
				ast.Global{Name: "sumElim"},
				a,
				b,
				motive,
				inlCase,
				inrCase,
				ast.Var{Ix: hypVarIdx},
			)
		},
	)
	if err != nil {
		return Fail(err)
	}

	return SuccessMsg(state, fmt.Sprintf("destructed %s : Sum into two cases", hyp.Name))
}

// destructBool performs case analysis on a Bool hypothesis.
// For h : Bool, creates two subgoals:
// - Goal for the true case
// - Goal for the false case
func destructBool(state *proofstate.ProofState, goal *proofstate.Goal, hyp proofstate.Hypothesis, hypIdx int) TacticResult {
	hypVarIdx := len(goal.Hypotheses) - 1 - hypIdx

	// The motive is λb. G where b is the Bool value
	shiftedGoal := subst.Shift(1, 0, goal.Type)
	motive := ast.Lam{
		Binder: "_",
		Body:   shiftedGoal,
	}

	err := state.SolveGoalWithSubgoals(
		goal.ID,
		[]ast.Term{goal.Type, goal.Type}, // Both branches have the same goal type
		nil, // Use same hypotheses for both branches
		func(metaIDs []elab.MetaID) ast.Term {
			// Build: boolElim motive ?true_case ?false_case h
			return ast.MkApps(
				ast.Global{Name: "boolElim"},
				motive,
				ast.Meta{ID: int(metaIDs[0])}, // true case
				ast.Meta{ID: int(metaIDs[1])}, // false case
				ast.Var{Ix: hypVarIdx},
			)
		},
	)
	if err != nil {
		return Fail(err)
	}

	return SuccessMsg(state, fmt.Sprintf("destructed %s : Bool into two cases", hyp.Name))
}

// Induction performs induction on a hypothesis.
// For h : Nat, creates base case (zero) and step case (succ n with IH).
// For h : List A, creates nil case and cons case (with element, tail, and IH).
func Induction(hypName string) Tactic {
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

		// Try to do induction based on the type
		switch {
		case isNatType(hypTy):
			return inductionNat(state, goal, *hyp, hypIdx)
		case isListType(hypTy):
			return inductionList(state, goal, *hyp, hypIdx, hypTy)
		default:
			return Failf("cannot perform induction on hypothesis %s of type %v", hypName, hypTy)
		}
	}
}

// isNatType checks if a type is Nat.
func isNatType(ty ast.Term) bool {
	g, ok := ty.(ast.Global)
	return ok && g.Name == "Nat"
}

// isListType checks if a type is List A.
func isListType(ty ast.Term) bool {
	_, ok := extractListArg(ty)
	return ok
}

// extractListArg extracts A from List A.
// Returns (A, true) if the type is List A, (nil, false) otherwise.
func extractListArg(ty ast.Term) (ast.Term, bool) {
	app, ok := ty.(ast.App)
	if !ok {
		return nil, false
	}
	list, ok := app.T.(ast.Global)
	if !ok || list.Name != "List" {
		return nil, false
	}
	return app.U, true
}

// inductionNat performs induction on a Nat hypothesis.
// Creates two subgoals:
// - Base case: Goal with h = zero
// - Step case: (n : Nat) → Goal[n/h] → Goal[succ n/h]
func inductionNat(state *proofstate.ProofState, goal *proofstate.Goal, hyp proofstate.Hypothesis, hypIdx int) TacticResult {
	hypVarIdx := len(goal.Hypotheses) - 1 - hypIdx
	nat := ast.Global{Name: "Nat"}

	// The motive is λn. G where G is the goal with h replaced by n
	shiftedGoal := subst.Shift(1, 0, goal.Type)
	motive := ast.Lam{
		Binder: "_",
		Body:   shiftedGoal,
	}

	// Build hypotheses for step case: add n : Nat and ih : Goal[n/h]
	stepHyps := make([]proofstate.Hypothesis, len(goal.Hypotheses)+2)
	copy(stepHyps, goal.Hypotheses)
	stepHyps[len(goal.Hypotheses)] = proofstate.Hypothesis{Name: "n", Type: nat}
	stepHyps[len(goal.Hypotheses)+1] = proofstate.Hypothesis{Name: "ih", Type: goal.Type}

	// Step goal type: Goal[succ n/h]
	// This is simplified - we use the same goal type since we're not doing substitution
	stepGoalType := goal.Type

	err := state.SolveGoalWithSubgoals(
		goal.ID,
		[]ast.Term{goal.Type, stepGoalType}, // Base and step goal types
		[][]proofstate.Hypothesis{nil, stepHyps},
		func(metaIDs []elab.MetaID) ast.Term {
			// Build: natElim motive ?base (λn ih. ?step) h
			stepCase := ast.Lam{
				Binder: "n",
				Ann:    nat,
				Body: ast.Lam{
					Binder: "ih",
					Body:   ast.Meta{ID: int(metaIDs[1])},
				},
			}
			return ast.MkApps(
				ast.Global{Name: "natElim"},
				motive,
				ast.Meta{ID: int(metaIDs[0])}, // base case
				stepCase,
				ast.Var{Ix: hypVarIdx},
			)
		},
	)
	if err != nil {
		return Fail(err)
	}

	return SuccessMsg(state, fmt.Sprintf("induction on %s : Nat", hyp.Name))
}

// inductionList performs induction on a List A hypothesis.
// Creates two subgoals:
// - Nil case: Goal with h = nil A
// - Cons case: (x : A) → (xs : List A) → Goal[xs/h] → Goal[cons A x xs/h]
func inductionList(state *proofstate.ProofState, goal *proofstate.Goal, hyp proofstate.Hypothesis, hypIdx int, hypTy ast.Term) TacticResult {
	a, _ := extractListArg(hypTy)
	hypVarIdx := len(goal.Hypotheses) - 1 - hypIdx
	listA := hypTy

	// The motive is λl. G where G is the goal with h replaced by l
	shiftedGoal := subst.Shift(1, 0, goal.Type)
	motive := ast.Lam{
		Binder: "_",
		Body:   shiftedGoal,
	}

	// Build hypotheses for cons case: add x : A, xs : List A, ih : Goal[xs/h]
	consHyps := make([]proofstate.Hypothesis, len(goal.Hypotheses)+3)
	copy(consHyps, goal.Hypotheses)
	consHyps[len(goal.Hypotheses)] = proofstate.Hypothesis{Name: "x", Type: a}
	consHyps[len(goal.Hypotheses)+1] = proofstate.Hypothesis{Name: "xs", Type: listA}
	consHyps[len(goal.Hypotheses)+2] = proofstate.Hypothesis{Name: "ih", Type: goal.Type}

	err := state.SolveGoalWithSubgoals(
		goal.ID,
		[]ast.Term{goal.Type, goal.Type}, // Nil and cons goal types
		[][]proofstate.Hypothesis{nil, consHyps},
		func(metaIDs []elab.MetaID) ast.Term {
			// Build: listElim A motive ?nil (λx xs ih. ?cons) h
			consCase := ast.Lam{
				Binder: "x",
				Ann:    a,
				Body: ast.Lam{
					Binder: "xs",
					Ann:    listA,
					Body: ast.Lam{
						Binder: "ih",
						Body:   ast.Meta{ID: int(metaIDs[1])},
					},
				},
			}
			return ast.MkApps(
				ast.Global{Name: "listElim"},
				a,
				motive,
				ast.Meta{ID: int(metaIDs[0])}, // nil case
				consCase,
				ast.Var{Ix: hypVarIdx},
			)
		},
	)
	if err != nil {
		return Fail(err)
	}

	return SuccessMsg(state, fmt.Sprintf("induction on %s : List", hyp.Name))
}

// Cases performs non-recursive case analysis on a hypothesis.
// Unlike Induction, Cases does not introduce an induction hypothesis.
// For h : Nat, creates zero case and succ n case (no IH).
// For h : List A, creates nil case and cons x xs case (no IH).
// For h : Bool, creates true and false cases.
// For h : Sum A B, creates inl and inr cases.
func Cases(hypName string) Tactic {
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

		// Try to do case analysis based on the type
		switch {
		case isNatType(hypTy):
			return casesNat(state, goal, *hyp, hypIdx)
		case isListType(hypTy):
			return casesList(state, goal, *hyp, hypIdx, hypTy)
		case isBoolType(hypTy):
			return destructBool(state, goal, *hyp, hypIdx) // Same as destruct for Bool
		case isSumType(hypTy):
			return destructSum(state, goal, *hyp, hypIdx, hypTy) // Same as destruct for Sum
		default:
			return Failf("cannot perform case analysis on hypothesis %s of type %v", hypName, hypTy)
		}
	}
}

// casesNat performs non-recursive case analysis on a Nat hypothesis.
// Creates two subgoals:
// - Zero case: Goal with h = zero
// - Succ case: (n : Nat) → Goal[succ n/h] (no IH)
func casesNat(state *proofstate.ProofState, goal *proofstate.Goal, hyp proofstate.Hypothesis, hypIdx int) TacticResult {
	hypVarIdx := len(goal.Hypotheses) - 1 - hypIdx
	nat := ast.Global{Name: "Nat"}

	// The motive is λn. G
	shiftedGoal := subst.Shift(1, 0, goal.Type)
	motive := ast.Lam{
		Binder: "_",
		Body:   shiftedGoal,
	}

	// Build hypotheses for succ case: add n : Nat (no IH)
	succHyps := make([]proofstate.Hypothesis, len(goal.Hypotheses)+1)
	copy(succHyps, goal.Hypotheses)
	succHyps[len(goal.Hypotheses)] = proofstate.Hypothesis{Name: "n", Type: nat}

	err := state.SolveGoalWithSubgoals(
		goal.ID,
		[]ast.Term{goal.Type, goal.Type}, // Zero and succ goal types
		[][]proofstate.Hypothesis{nil, succHyps},
		func(metaIDs []elab.MetaID) ast.Term {
			// Build: natElim motive ?zero (λn _. ?succ) h
			// We use natElim but ignore the IH by binding to _
			succCase := ast.Lam{
				Binder: "n",
				Ann:    nat,
				Body: ast.Lam{
					Binder: "_", // IH is ignored
					Body:   ast.Meta{ID: int(metaIDs[1])},
				},
			}
			return ast.MkApps(
				ast.Global{Name: "natElim"},
				motive,
				ast.Meta{ID: int(metaIDs[0])}, // zero case
				succCase,
				ast.Var{Ix: hypVarIdx},
			)
		},
	)
	if err != nil {
		return Fail(err)
	}

	return SuccessMsg(state, fmt.Sprintf("case analysis on %s : Nat", hyp.Name))
}

// casesList performs non-recursive case analysis on a List A hypothesis.
// Creates two subgoals:
// - Nil case: Goal with h = nil A
// - Cons case: (x : A) → (xs : List A) → Goal[cons A x xs/h] (no IH)
func casesList(state *proofstate.ProofState, goal *proofstate.Goal, hyp proofstate.Hypothesis, hypIdx int, hypTy ast.Term) TacticResult {
	a, _ := extractListArg(hypTy)
	hypVarIdx := len(goal.Hypotheses) - 1 - hypIdx
	listA := hypTy

	// The motive is λl. G
	shiftedGoal := subst.Shift(1, 0, goal.Type)
	motive := ast.Lam{
		Binder: "_",
		Body:   shiftedGoal,
	}

	// Build hypotheses for cons case: add x : A, xs : List A (no IH)
	consHyps := make([]proofstate.Hypothesis, len(goal.Hypotheses)+2)
	copy(consHyps, goal.Hypotheses)
	consHyps[len(goal.Hypotheses)] = proofstate.Hypothesis{Name: "x", Type: a}
	consHyps[len(goal.Hypotheses)+1] = proofstate.Hypothesis{Name: "xs", Type: listA}

	err := state.SolveGoalWithSubgoals(
		goal.ID,
		[]ast.Term{goal.Type, goal.Type}, // Nil and cons goal types
		[][]proofstate.Hypothesis{nil, consHyps},
		func(metaIDs []elab.MetaID) ast.Term {
			// Build: listElim A motive ?nil (λx xs _. ?cons) h
			// We use listElim but ignore the IH by binding to _
			consCase := ast.Lam{
				Binder: "x",
				Ann:    a,
				Body: ast.Lam{
					Binder: "xs",
					Ann:    listA,
					Body: ast.Lam{
						Binder: "_", // IH is ignored
						Body:   ast.Meta{ID: int(metaIDs[1])},
					},
				},
			}
			return ast.MkApps(
				ast.Global{Name: "listElim"},
				a,
				motive,
				ast.Meta{ID: int(metaIDs[0])}, // nil case
				consCase,
				ast.Var{Ix: hypVarIdx},
			)
		},
	)
	if err != nil {
		return Fail(err)
	}

	return SuccessMsg(state, fmt.Sprintf("case analysis on %s : List", hyp.Name))
}

// Constructor applies the first applicable constructor for the goal type.
// - For Unit: applies tt (completes the goal)
// - For Sum A B: applies inl (creates subgoal A) - use Right() for inr
// - For List A: applies nil (completes with empty list)
func Constructor() Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		goalTy := eval.EvalNBE(goal.Type)

		// Check what type the goal is and apply appropriate constructor
		switch {
		case isUnitType(goalTy):
			return constructorUnit(state, goal)
		case isSumType(goalTy):
			// Apply Left (inl) as first constructor
			return Left()(state)
		case isListType(goalTy):
			return constructorNil(state, goal, goalTy)
		default:
			return Failf("cannot apply constructor to goal of type %v", goalTy)
		}
	}
}

// isUnitType checks if a type is Unit.
func isUnitType(ty ast.Term) bool {
	g, ok := ty.(ast.Global)
	return ok && g.Name == "Unit"
}

// constructorUnit solves a Unit goal with tt.
func constructorUnit(state *proofstate.ProofState, goal *proofstate.Goal) TacticResult {
	proofTerm := ast.Global{Name: "tt"}
	if err := state.SolveGoal(goal.ID, proofTerm); err != nil {
		return Fail(err)
	}
	return SuccessMsg(state, "applied constructor tt")
}

// constructorNil solves a List A goal with nil A.
func constructorNil(state *proofstate.ProofState, goal *proofstate.Goal, goalTy ast.Term) TacticResult {
	a, _ := extractListArg(goalTy)
	// nil : (A : Type) → List A
	// So nil A : List A
	proofTerm := ast.App{
		T: ast.Global{Name: "nil"},
		U: a,
	}
	if err := state.SolveGoal(goal.ID, proofTerm); err != nil {
		return Fail(err)
	}
	return SuccessMsg(state, "applied constructor nil")
}

// Exists provides a witness for a Sigma (existential) type goal.
// For goal Σ(x:A).B, given witness w : A, creates subgoal B[w/x].
// This is like Split but with an explicit witness for the first component.
func Exists(witness ast.Term) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		goal := state.CurrentGoal()
		if goal == nil {
			return Failf("no current goal")
		}

		// Check that the goal is a Sigma type
		sigma, ok := goal.Type.(ast.Sigma)
		if !ok {
			// Also check after evaluation
			evaled := eval.EvalNBE(goal.Type)
			sigma, ok = evaled.(ast.Sigma)
			if !ok {
				return Failf("Exists: goal is not a Sigma type, got %T", goal.Type)
			}
		}

		// The second component type is B with x substituted by witness
		// B is under one binder, so we substitute Var{0} with witness
		secondTy := subst.Subst(0, witness, sigma.B)

		// Solve the goal with (witness, ?meta)
		err := state.SolveGoalWithSubgoals(
			goal.ID,
			[]ast.Term{secondTy}, // Subgoal type is B[witness/x]
			nil,                  // Use same hypotheses
			func(metaIDs []elab.MetaID) ast.Term {
				// Build: (witness, ?meta)
				return ast.Pair{
					Fst: witness,
					Snd: ast.Meta{ID: int(metaIDs[0])},
				}
			},
		)
		if err != nil {
			return Fail(err)
		}

		return SuccessMsg(state, "provided witness for existential")
	}
}
