package script

import (
	"fmt"
	"strings"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/parser"
	"github.com/watchthelight/HypergraphGo/kernel/check"
	"github.com/watchthelight/HypergraphGo/tactics"
	"github.com/watchthelight/HypergraphGo/tactics/proofstate"
)

// ExecError represents an error during script execution.
type ExecError struct {
	Theorem string
	Line    int
	Message string
}

func (e *ExecError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("theorem %s (line %d): %s", e.Theorem, e.Line, e.Message)
	}
	return fmt.Sprintf("theorem %s: %s", e.Theorem, e.Message)
}

// Result represents the result of executing a script.
type Result struct {
	Theorems []TheoremResult
}

// TheoremResult represents the result of proving a single theorem.
type TheoremResult struct {
	Name      string
	Type      ast.Term
	ProofTerm ast.Term
	Success   bool
	Error     error
}

// Execute runs a script and returns the results.
func Execute(script *Script, checker *check.Checker) *Result {
	result := &Result{
		Theorems: make([]TheoremResult, len(script.Theorems)),
	}

	for i, thm := range script.Theorems {
		result.Theorems[i] = executeTheorem(thm, checker)
	}

	return result
}

// executeTheorem executes a single theorem proof.
func executeTheorem(thm Theorem, checker *check.Checker) TheoremResult {
	result := TheoremResult{
		Name: thm.Name,
		Type: thm.Type,
	}

	// Verify the goal type is valid
	_, checkErr := checker.Synth(nil, check.Span{}, thm.Type)
	if checkErr != nil {
		result.Error = &ExecError{
			Theorem: thm.Name,
			Message: fmt.Sprintf("invalid goal type: %v", checkErr),
		}
		return result
	}

	// Create proof state
	state := proofstate.NewProofState(thm.Type, nil)

	// Apply each tactic
	for _, cmd := range thm.Proof {
		tactic, err := parseTactic(cmd.Name, cmd.Args)
		if err != nil {
			result.Error = &ExecError{
				Theorem: thm.Name,
				Line:    cmd.Line,
				Message: err.Error(),
			}
			return result
		}

		tacticResult := tactic(state)
		if !tacticResult.IsSuccess() {
			result.Error = &ExecError{
				Theorem: thm.Name,
				Line:    cmd.Line,
				Message: fmt.Sprintf("tactic '%s' failed: %v", cmd.Name, tacticResult.Err),
			}
			return result
		}
	}

	// Check if proof is complete
	if !state.IsComplete() {
		result.Error = &ExecError{
			Theorem: thm.Name,
			Message: fmt.Sprintf("proof incomplete: %d goals remaining", state.GoalCount()),
		}
		return result
	}

	// Extract proof term
	proofTerm, err := state.ExtractTerm()
	if err != nil {
		result.Error = &ExecError{
			Theorem: thm.Name,
			Message: fmt.Sprintf("extraction failed: %v", err),
		}
		return result
	}

	// Type check the proof term
	checkErr = checker.Check(nil, check.NoSpan(), proofTerm, thm.Type)
	if checkErr != nil {
		result.Error = &ExecError{
			Theorem: thm.Name,
			Message: fmt.Sprintf("type check failed: %v", checkErr),
		}
		return result
	}

	result.ProofTerm = proofTerm
	result.Success = true
	return result
}

// parseTactic converts a tactic name and args to a Tactic function.
func parseTactic(name string, args []string) (tactics.Tactic, error) {
	switch name {
	case "intro":
		binder := ""
		if len(args) > 0 {
			binder = args[0]
		}
		return tactics.Intro(binder), nil

	case "intros":
		return tactics.Intros(), nil

	case "exact":
		if len(args) < 1 {
			return nil, fmt.Errorf("exact requires a term argument")
		}
		term, err := parser.ParseTerm(strings.Join(args, " "))
		if err != nil {
			return nil, fmt.Errorf("parsing exact argument: %w", err)
		}
		return tactics.Exact(term), nil

	case "assumption":
		return tactics.Assumption(), nil

	case "reflexivity", "refl":
		return tactics.Reflexivity(), nil

	case "split":
		return tactics.Split(), nil

	case "left":
		return tactics.Left(), nil

	case "right":
		return tactics.Right(), nil

	case "destruct":
		if len(args) < 1 {
			return nil, fmt.Errorf("destruct requires a hypothesis name")
		}
		return tactics.Destruct(args[0]), nil

	case "induction":
		if len(args) < 1 {
			return nil, fmt.Errorf("induction requires a hypothesis name")
		}
		return tactics.Induction(args[0]), nil

	case "cases":
		if len(args) < 1 {
			return nil, fmt.Errorf("cases requires a hypothesis name")
		}
		return tactics.Cases(args[0]), nil

	case "constructor":
		return tactics.Constructor(), nil

	case "exists":
		if len(args) < 1 {
			return nil, fmt.Errorf("exists requires a witness term")
		}
		term, err := parser.ParseTerm(strings.Join(args, " "))
		if err != nil {
			return nil, fmt.Errorf("parsing exists argument: %w", err)
		}
		return tactics.Exists(term), nil

	case "contradiction":
		return tactics.Contradiction(), nil

	case "rewrite":
		if len(args) < 1 {
			return nil, fmt.Errorf("rewrite requires a hypothesis name")
		}
		return tactics.Rewrite(args[0]), nil

	case "simpl":
		return tactics.Simpl(), nil

	case "trivial":
		return tactics.Trivial(), nil

	case "auto":
		return tactics.Auto(), nil

	case "apply":
		if len(args) < 1 {
			return nil, fmt.Errorf("apply requires a term argument")
		}
		term, err := parser.ParseTerm(strings.Join(args, " "))
		if err != nil {
			return nil, fmt.Errorf("parsing apply argument: %w", err)
		}
		return tactics.Apply(term), nil

	default:
		return nil, fmt.Errorf("unknown tactic: %s", name)
	}
}
