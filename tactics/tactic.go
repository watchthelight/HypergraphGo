// tactic.go defines the Tactic type and TacticResult.
//
// See doc.go for package overview.

package tactics

import (
	"fmt"
	"strings"

	"github.com/watchthelight/HypergraphGo/internal/parser"
	"github.com/watchthelight/HypergraphGo/tactics/proofstate"
)

// TacticError represents a tactic failure with rich context information.
type TacticError struct {
	// Message is the primary error message
	Message string

	// Goal is the current goal (if available)
	Goal *proofstate.Goal

	// ExpectedType is the expected type (for type mismatch errors)
	ExpectedType string

	// ActualType is the actual type found
	ActualType string

	// Hints are suggestions for fixing the error
	Hints []string
}

// Error implements the error interface.
func (e *TacticError) Error() string {
	var sb strings.Builder
	sb.WriteString("error: ")
	sb.WriteString(e.Message)

	if e.ExpectedType != "" || e.ActualType != "" {
		sb.WriteString("\n")
		if e.ExpectedType != "" {
			sb.WriteString(fmt.Sprintf("  Expected: %s\n", e.ExpectedType))
		}
		if e.ActualType != "" {
			sb.WriteString(fmt.Sprintf("  Got: %s\n", e.ActualType))
		}
	}

	if e.Goal != nil && len(e.Goal.Hypotheses) > 0 {
		sb.WriteString("\n  Current hypotheses:\n")
		for _, h := range e.Goal.Hypotheses {
			sb.WriteString(fmt.Sprintf("    %s : %s\n", h.Name, parser.FormatTerm(h.Type)))
		}
	}

	if len(e.Hints) > 0 {
		sb.WriteString("\n  Hints:\n")
		for _, hint := range e.Hints {
			sb.WriteString(fmt.Sprintf("    - %s\n", hint))
		}
	}

	return sb.String()
}

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

// FailWithContext creates a failed tactic result with rich context.
func FailWithContext(msg string, goal *proofstate.Goal, hints ...string) TacticResult {
	return TacticResult{Err: &TacticError{
		Message: msg,
		Goal:    goal,
		Hints:   hints,
	}}
}

// FailTypeMismatch creates a type mismatch error with context and hints.
func FailTypeMismatch(goal *proofstate.Goal, expected, actual string, hints ...string) TacticResult {
	return TacticResult{Err: &TacticError{
		Message:      "type mismatch",
		Goal:         goal,
		ExpectedType: expected,
		ActualType:   actual,
		Hints:        hints,
	}}
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
