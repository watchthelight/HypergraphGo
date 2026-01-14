// combinators.go implements tactic combinators: Seq, OrElse, Try, Repeat, etc.
//
// See doc.go for package overview.

package tactics

import (
	"github.com/watchthelight/HypergraphGo/tactics/proofstate"
)

// Seq sequences multiple tactics, running them in order.
// Fails if any tactic fails.
func Seq(tactics ...Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		current := state
		for _, t := range tactics {
			result := t(current)
			if !result.IsSuccess() {
				return result
			}
			current = result.State
		}
		return Success(current)
	}
}

// OrElse tries the first tactic; if it fails, tries the second.
func OrElse(t1, t2 Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		// Clone state to try first tactic
		clone := state.Clone()
		result := t1(clone)
		if result.IsSuccess() {
			// Copy back the successful state
			*state = *result.State
			return Success(state)
		}
		// First failed, try second on original state
		return t2(state)
	}
}

// Try tries a tactic; succeeds even if the tactic fails.
func Try(t Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		clone := state.Clone()
		result := t(clone)
		if result.IsSuccess() {
			*state = *result.State
			return Success(state)
		}
		// Failure is okay, return original state
		return Success(state)
	}
}

// Repeat applies a tactic repeatedly until it fails.
// Always succeeds (possibly with no changes).
func Repeat(t Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		current := state
		for {
			clone := current.Clone()
			result := t(clone)
			if !result.IsSuccess() {
				break
			}
			*current = *result.State
		}
		return Success(current)
	}
}

// RepeatN applies a tactic at most n times.
func RepeatN(n int, t Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		current := state
		for i := 0; i < n; i++ {
			clone := current.Clone()
			result := t(clone)
			if !result.IsSuccess() {
				break
			}
			*current = *result.State
		}
		return Success(current)
	}
}

// First tries each tactic in order until one succeeds.
func First(tactics ...Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		for _, t := range tactics {
			clone := state.Clone()
			result := t(clone)
			if result.IsSuccess() {
				*state = *result.State
				return Success(state)
			}
		}
		return Failf("all tactics failed")
	}
}

// All applies a tactic to all goals.
func All(t Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		// Collect goal IDs first to avoid mutation issues
		goalIDs := make([]proofstate.GoalID, len(state.Goals))
		for i, g := range state.Goals {
			goalIDs[i] = g.ID
		}

		for _, id := range goalIDs {
			// Focus on this goal
			if err := state.Focus(id); err != nil {
				continue // Goal might have been solved
			}
			result := t(state)
			if !result.IsSuccess() {
				return result
			}
			state = result.State
		}
		return Success(state)
	}
}

// Focus applies a tactic to a specific goal by ID.
func Focus(id proofstate.GoalID, t Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		// Remember current focus
		var oldFocus proofstate.GoalID
		if current := state.CurrentGoal(); current != nil {
			oldFocus = current.ID
		}

		// Focus on target goal
		if err := state.Focus(id); err != nil {
			return Fail(err)
		}

		// Apply tactic
		result := t(state)

		// Restore focus if possible
		if result.IsSuccess() && oldFocus != id {
			_ = result.State.Focus(oldFocus)
		}

		return result
	}
}

// IfThenElse applies t1 if condition succeeds, otherwise t2.
func IfThenElse(cond, t1, t2 Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		clone := state.Clone()
		condResult := cond(clone)
		if condResult.IsSuccess() {
			return t1(state)
		}
		return t2(state)
	}
}

// Progress applies a tactic and fails if it doesn't make progress.
func Progress(t Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		initialCount := state.GoalCount()
		result := t(state)
		if !result.IsSuccess() {
			return result
		}
		if result.State.GoalCount() >= initialCount {
			// Check if any metavariables were solved
			// For now, just check goal count
			return Failf("tactic made no progress")
		}
		return result
	}
}

// Once applies a tactic exactly once (useful with Repeat).
func Once(t Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		return t(state)
	}
}

// Complete requires a tactic to completely solve all goals.
func Complete(t Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		result := t(state)
		if !result.IsSuccess() {
			return result
		}
		if !result.State.IsComplete() {
			return Failf("tactic did not complete proof: %d goals remaining", result.State.GoalCount())
		}
		return result
	}
}

// Do applies a tactic n times, failing if it can't be applied n times.
func Do(n int, t Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		current := state
		for i := 0; i < n; i++ {
			result := t(current)
			if !result.IsSuccess() {
				return Failf("tactic failed on iteration %d: %v", i+1, result.Err)
			}
			current = result.State
		}
		return Success(current)
	}
}

// Guard checks a condition before applying a tactic.
func Guard(check func(*proofstate.ProofState) bool, t Tactic) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		if !check(state) {
			return Failf("guard condition failed")
		}
		return t(state)
	}
}

// HasGoals is a predicate for Guard that checks if there are goals.
func HasGoals(state *proofstate.ProofState) bool {
	return state.GoalCount() > 0
}

// IsFinished is a predicate for Guard that checks if proof is complete.
func IsFinished(state *proofstate.ProofState) bool {
	return state.IsComplete()
}
