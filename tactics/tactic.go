// Package tactics provides Ltac-style proof tactics for HoTT.
package tactics

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/tactics/proofstate"
)

// TacticResult represents the outcome of applying a tactic.
type TacticResult struct {
	// State is the new proof state after the tactic
	State *proofstate.ProofState

	// Err is set if the tactic failed
	Err error

	// Message is an optional message to display
	Message string
}

// Success creates a successful tactic result.
func Success(state *proofstate.ProofState) TacticResult {
	return TacticResult{State: state, Err: nil}
}

// SuccessMsg creates a successful tactic result with a message.
func SuccessMsg(state *proofstate.ProofState, msg string) TacticResult {
	return TacticResult{State: state, Err: nil, Message: msg}
}

// Fail creates a failed tactic result.
func Fail(err error) TacticResult {
	return TacticResult{Err: err}
}

// Failf creates a failed tactic result with a formatted message.
func Failf(format string, args ...any) TacticResult {
	return TacticResult{Err: fmt.Errorf(format, args...)}
}

// IsSuccess returns true if the result represents success.
func (r TacticResult) IsSuccess() bool {
	return r.Err == nil
}

// Tactic is a function that transforms a proof state.
type Tactic func(*proofstate.ProofState) TacticResult

// RunTactic runs a tactic on a proof state, saving the state for undo.
func RunTactic(state *proofstate.ProofState, t Tactic) TacticResult {
	if state == nil {
		return Fail(fmt.Errorf("nil proof state"))
	}

	// Save state for undo
	state.SaveState()

	// Apply the tactic
	result := t(state)

	if result.Err != nil {
		// Undo on failure
		state.Undo()
	}

	return result
}

// NoOp is a tactic that does nothing and succeeds.
func NoOp() Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		return Success(state)
	}
}

// FailWith is a tactic that always fails with the given error.
func FailWith(err error) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		return Fail(err)
	}
}

// FailWithMsg is a tactic that always fails with the given message.
func FailWithMsg(msg string) Tactic {
	return func(state *proofstate.ProofState) TacticResult {
		return Fail(fmt.Errorf("%s", msg))
	}
}
