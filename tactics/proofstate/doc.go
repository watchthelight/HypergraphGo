// Package proofstate provides proof state management for the tactics system.
//
// A proof state tracks the current proof obligations (goals) and solved
// metavariables. Tactics transform proof states by solving goals, creating
// new subgoals, or manipulating the goal focus.
//
// # Proof State
//
// [ProofState] contains:
//
//   - Goals: list of proof obligations (first is focused)
//   - Metas: metavariable store tracking solutions
//   - History: stack of previous states for undo
//
// Create a new proof state with [NewProofState]:
//
//	state := proofstate.NewProofState(goalType, nil)
//
// # Goals
//
// Each [Goal] represents a proof obligation:
//
//   - ID: unique identifier ([GoalID])
//   - Type: what we need to prove
//   - Hypotheses: local context bindings
//   - MetaID: associated metavariable for proof term
//
// # Hypotheses
//
// [Hypothesis] represents a local binding in the goal context:
//
//	type Hypothesis struct {
//	    Name string   // Variable name
//	    Type ast.Term // Type of the hypothesis
//	}
//
// Tactics like Intro add hypotheses; tactics like Exact use them.
//
// # State Operations
//
// Query state:
//
//   - [ProofState.CurrentGoal] - get the focused goal
//   - [ProofState.GoalCount] - count remaining goals
//   - [ProofState.IsComplete] - check if all goals are solved
//   - [ProofState.GetGoal] - get a goal by ID
//
// Modify state:
//
//   - [ProofState.AddGoal] - add a new goal
//   - [ProofState.RemoveGoal] - remove a solved goal
//   - [ProofState.Focus] - switch focused goal
//   - [ProofState.SolveGoal] - solve with a term
//   - [ProofState.SolveGoalWithSubgoals] - solve creating new subgoals
//
// Undo support:
//
//   - [ProofState.SaveState] - snapshot current state
//   - [ProofState.Undo] - restore previous state
//   - [ProofState.Clone] - deep copy for backtracking
//
// # Proof Extraction
//
// When all goals are solved, extract the proof term:
//
//	term, err := state.ExtractTerm()
//
// This zonks all solved metavariables to produce a complete term.
//
// # Display
//
// [ProofState.FormatState] produces a human-readable representation
// of the current goals and hypotheses for interactive use.
package proofstate
