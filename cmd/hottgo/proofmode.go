// proofmode.go implements interactive proof mode for the REPL.
package main

import (
	"fmt"
	"strings"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/parser"
	"github.com/watchthelight/HypergraphGo/kernel/check"
	"github.com/watchthelight/HypergraphGo/tactics"
	"github.com/watchthelight/HypergraphGo/tactics/proofstate"
)

// TacticHistoryEntry records an applied tactic.
type TacticHistoryEntry struct {
	Name    string   // Tactic name
	Args    []string // Arguments
	Message string   // Result message
}

// Checkpoint stores a named proof state snapshot.
type Checkpoint struct {
	State   *proofstate.ProofState
	History []TacticHistoryEntry
}

// ProofMode manages interactive proof construction.
type ProofMode struct {
	state       *proofstate.ProofState
	checker     *check.Checker
	goalTy      ast.Term
	theoremName string // Optional name for the theorem

	// History of applied tactics
	tacticHistory []TacticHistoryEntry

	// Named checkpoints for save/restore
	checkpoints map[string]*Checkpoint
}

// NewProofMode creates a new proof mode for the given goal type.
func NewProofMode(goalTy ast.Term, checker *check.Checker) *ProofMode {
	return &ProofMode{
		state:       proofstate.NewProofState(goalTy, nil),
		checker:     checker,
		goalTy:      goalTy,
		checkpoints: make(map[string]*Checkpoint),
	}
}

// NewProofModeNamed creates a new proof mode with a theorem name.
func NewProofModeNamed(name string, goalTy ast.Term, checker *check.Checker) *ProofMode {
	return &ProofMode{
		state:       proofstate.NewProofState(goalTy, nil),
		checker:     checker,
		goalTy:      goalTy,
		theoremName: name,
		checkpoints: make(map[string]*Checkpoint),
	}
}

// TheoremName returns the name of the theorem being proved (may be empty).
func (pm *ProofMode) TheoremName() string {
	return pm.theoremName
}

// GoalType returns the goal type.
func (pm *ProofMode) GoalType() ast.Term {
	return pm.goalTy
}

// IsComplete returns true if the proof is complete.
func (pm *ProofMode) IsComplete() bool {
	return pm.state.IsComplete()
}

// CurrentGoal returns the current goal, or nil if complete.
func (pm *ProofMode) CurrentGoal() *proofstate.Goal {
	return pm.state.CurrentGoal()
}

// GoalCount returns the number of remaining goals.
func (pm *ProofMode) GoalCount() int {
	return pm.state.GoalCount()
}

// FormatCurrentGoal returns a formatted string of the current goal.
func (pm *ProofMode) FormatCurrentGoal() string {
	goal := pm.state.CurrentGoal()
	if goal == nil {
		return "No more goals."
	}
	return pm.state.FormatGoal(goal)
}

// FormatAllGoals returns a formatted string of all goals.
func (pm *ProofMode) FormatAllGoals() string {
	return pm.state.FormatState()
}

// ApplyTactic applies a tactic by name with arguments.
func (pm *ProofMode) ApplyTactic(name string, args []string) (string, error) {
	pm.state.SaveState()

	tactic, err := pm.parseTactic(name, args)
	if err != nil {
		return "", err
	}

	result := tactic(pm.state)
	if !result.IsSuccess() {
		pm.state.Undo()
		return "", result.Err
	}

	msg := result.Message
	if msg == "" {
		msg = fmt.Sprintf("applied %s", name)
	}

	// Record in history
	pm.tacticHistory = append(pm.tacticHistory, TacticHistoryEntry{
		Name:    name,
		Args:    args,
		Message: msg,
	})

	return msg, nil
}

// parseTactic converts a tactic name and args to a Tactic function.
func (pm *ProofMode) parseTactic(name string, args []string) (tactics.Tactic, error) {
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

	case "symmetry", "sym":
		if len(args) < 1 {
			return nil, fmt.Errorf("symmetry requires a hypothesis name")
		}
		return tactics.Symmetry(args[0]), nil

	case "transitivity", "trans":
		if len(args) < 2 {
			return nil, fmt.Errorf("transitivity requires two hypothesis names")
		}
		return tactics.Transitivity(args[0], args[1]), nil

	case "ap", "congruence":
		if len(args) < 2 {
			return nil, fmt.Errorf("ap requires a function term and hypothesis name")
		}
		// Parse all but last arg as function term, last arg is hypothesis name
		hypName := args[len(args)-1]
		funcStr := strings.Join(args[:len(args)-1], " ")
		funcTerm, err := parser.ParseTerm(funcStr)
		if err != nil {
			return nil, fmt.Errorf("parsing ap function: %w", err)
		}
		return tactics.Ap(funcTerm, hypName), nil

	case "transport":
		if len(args) < 2 {
			return nil, fmt.Errorf("transport requires path and term hypothesis names")
		}
		return tactics.Transport(args[0], args[1]), nil

	default:
		return nil, fmt.Errorf("unknown tactic: %s", name)
	}
}

// Undo reverts the last tactic application.
func (pm *ProofMode) Undo() bool {
	if pm.state.Undo() {
		// Remove from history as well
		if len(pm.tacticHistory) > 0 {
			pm.tacticHistory = pm.tacticHistory[:len(pm.tacticHistory)-1]
		}
		return true
	}
	return false
}

// History returns the tactic history.
func (pm *ProofMode) History() []TacticHistoryEntry {
	return pm.tacticHistory
}

// SaveCheckpoint saves the current state with a name.
func (pm *ProofMode) SaveCheckpoint(name string) {
	// Make a copy of history
	histCopy := make([]TacticHistoryEntry, len(pm.tacticHistory))
	copy(histCopy, pm.tacticHistory)

	pm.checkpoints[name] = &Checkpoint{
		State:   pm.state.Clone(),
		History: histCopy,
	}
}

// RestoreCheckpoint restores a previously saved checkpoint.
func (pm *ProofMode) RestoreCheckpoint(name string) error {
	cp, ok := pm.checkpoints[name]
	if !ok {
		return fmt.Errorf("checkpoint '%s' not found", name)
	}

	pm.state = cp.State.Clone()
	pm.tacticHistory = make([]TacticHistoryEntry, len(cp.History))
	copy(pm.tacticHistory, cp.History)
	return nil
}

// ListCheckpoints returns the names of all checkpoints.
func (pm *ProofMode) ListCheckpoints() []string {
	names := make([]string, 0, len(pm.checkpoints))
	for name := range pm.checkpoints {
		names = append(names, name)
	}
	return names
}

// Extract returns the proof term if complete.
func (pm *ProofMode) Extract() (ast.Term, error) {
	return pm.state.ExtractTerm()
}

// TypeCheck verifies the extracted proof term against the goal type.
func (pm *ProofMode) TypeCheck() error {
	term, err := pm.Extract()
	if err != nil {
		return err
	}

	checkErr := pm.checker.Check(nil, check.NoSpan(), term, pm.goalTy)
	if checkErr != nil {
		return fmt.Errorf("type check failed: %v", checkErr)
	}
	return nil
}
