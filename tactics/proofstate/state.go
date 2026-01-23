// state.go implements ProofState, Goal, and Hypothesis types.
//
// See doc.go for package overview.

package proofstate

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/elab"
	"github.com/watchthelight/HypergraphGo/internal/parser"
)

// GoalID uniquely identifies a goal within a proof state.
type GoalID int

// Hypothesis represents a local binding in the goal context.
type Hypothesis struct {
	Name string   // Variable name
	Type ast.Term // Type of the hypothesis
}

// Goal represents a single proof obligation.
type Goal struct {
	ID         GoalID       // Unique identifier
	Type       ast.Term     // What we need to prove
	Hypotheses []Hypothesis // Local context
	MetaID     elab.MetaID  // Associated metavariable
}

// ProofState holds the current state of an interactive proof.
type ProofState struct {
	// Goals are the current proof obligations (first is focused)
	Goals []Goal

	// Metas is the metavariable store for the proof
	Metas *elab.MetaStore

	// nextGoalID is the next available goal ID
	nextGoalID GoalID

	// history stores previous states for undo
	history []*ProofState
}

// NewProofState creates a new proof state with a single goal.
func NewProofState(goalType ast.Term, hypotheses []Hypothesis) *ProofState {
	metas := elab.NewMetaStore()

	// Create a metavariable for the goal
	metaID := metas.Fresh(goalType, nil, elab.NoSpan)

	goal := Goal{
		ID:         0,
		Type:       goalType,
		Hypotheses: hypotheses,
		MetaID:     metaID,
	}

	return &ProofState{
		Goals:      []Goal{goal},
		Metas:      metas,
		nextGoalID: 1,
	}
}

// CurrentGoal returns the focused goal, or nil if there are no goals.
func (p *ProofState) CurrentGoal() *Goal {
	if len(p.Goals) == 0 {
		return nil
	}
	return &p.Goals[0]
}

// GoalCount returns the number of remaining goals.
func (p *ProofState) GoalCount() int {
	return len(p.Goals)
}

// IsComplete returns true if all goals have been solved.
func (p *ProofState) IsComplete() bool {
	return len(p.Goals) == 0
}

// Focus switches the focused goal to the one with the given ID.
func (p *ProofState) Focus(id GoalID) error {
	for i, g := range p.Goals {
		if g.ID == id {
			// Move goal to front
			p.Goals[0], p.Goals[i] = p.Goals[i], p.Goals[0]
			return nil
		}
	}
	return fmt.Errorf("goal %d not found", id)
}

// AddGoal adds a new goal to the proof state.
func (p *ProofState) AddGoal(goalType ast.Term, hypotheses []Hypothesis) GoalID {
	metaID := p.Metas.Fresh(goalType, nil, elab.NoSpan)

	goal := Goal{
		ID:         p.nextGoalID,
		Type:       goalType,
		Hypotheses: hypotheses,
		MetaID:     metaID,
	}

	p.nextGoalID++
	p.Goals = append(p.Goals, goal)
	return goal.ID
}

// SolveGoal marks a goal as solved with the given term.
func (p *ProofState) SolveGoal(id GoalID, solution ast.Term) error {
	for i, g := range p.Goals {
		if g.ID == id {
			// Solve the metavariable
			if err := p.Metas.Solve(g.MetaID, solution); err != nil {
				return err
			}
			// Remove the goal
			p.Goals = append(p.Goals[:i], p.Goals[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("goal %d not found", id)
}

// ReplaceGoal replaces a goal with new subgoals.
// DEPRECATED: Use SolveGoalWithSubgoals instead for proper proof term construction.
func (p *ProofState) ReplaceGoal(id GoalID, newGoals []Goal) error {
	for i, g := range p.Goals {
		if g.ID == id {
			// Assign IDs to new goals
			for j := range newGoals {
				newGoals[j].ID = p.nextGoalID
				newGoals[j].MetaID = p.Metas.Fresh(newGoals[j].Type, nil, elab.NoSpan)
				p.nextGoalID++
			}
			// Replace the old goal with new goals
			p.Goals = append(p.Goals[:i], append(newGoals, p.Goals[i+1:]...)...)
			return nil
		}
	}
	return fmt.Errorf("goal %d not found", id)
}

// SolveGoalWithSubgoals solves a goal with a term containing metavariable references
// to new subgoals. This properly links parent goals to child goals for proof extraction.
// The termBuilder receives the MetaIDs of the created subgoals and returns the proof term.
func (p *ProofState) SolveGoalWithSubgoals(id GoalID, subgoalTypes []ast.Term, subgoalHyps [][]Hypothesis, termBuilder func([]elab.MetaID) ast.Term) error {
	// Find the goal
	var goalIdx int = -1
	var goal Goal
	for i, g := range p.Goals {
		if g.ID == id {
			goalIdx = i
			goal = g
			break
		}
	}
	if goalIdx == -1 {
		return fmt.Errorf("goal %d not found", id)
	}

	// Create subgoals and collect their MetaIDs
	metaIDs := make([]elab.MetaID, len(subgoalTypes))
	newGoals := make([]Goal, len(subgoalTypes))
	for i, ty := range subgoalTypes {
		metaID := p.Metas.Fresh(ty, nil, elab.NoSpan)
		metaIDs[i] = metaID

		hyps := goal.Hypotheses
		if i < len(subgoalHyps) && subgoalHyps[i] != nil {
			hyps = subgoalHyps[i]
		}

		newGoals[i] = Goal{
			ID:         p.nextGoalID,
			Type:       ty,
			Hypotheses: hyps,
			MetaID:     metaID,
		}
		p.nextGoalID++
	}

	// Build the proof term with metavariable references
	proofTerm := termBuilder(metaIDs)

	// Solve the parent goal's metavariable
	if err := p.Metas.Solve(goal.MetaID, proofTerm); err != nil {
		return err
	}

	// Remove the old goal and add new subgoals
	p.Goals = append(p.Goals[:goalIdx], append(newGoals, p.Goals[goalIdx+1:]...)...)

	return nil
}

// Clone creates a deep copy of the proof state.
func (p *ProofState) Clone() *ProofState {
	goals := make([]Goal, len(p.Goals))
	for i, g := range p.Goals {
		hyps := make([]Hypothesis, len(g.Hypotheses))
		copy(hyps, g.Hypotheses)
		goals[i] = Goal{
			ID:         g.ID,
			Type:       g.Type,
			Hypotheses: hyps,
			MetaID:     g.MetaID,
		}
	}

	return &ProofState{
		Goals:      goals,
		Metas:      p.Metas.Clone(),
		nextGoalID: p.nextGoalID,
		history:    nil, // Don't clone history
	}
}

// SaveState saves the current state for undo.
func (p *ProofState) SaveState() {
	p.history = append(p.history, p.Clone())
}

// Undo restores the previous state, returning true if successful.
func (p *ProofState) Undo() bool {
	if len(p.history) == 0 {
		return false
	}

	last := p.history[len(p.history)-1]
	p.history = p.history[:len(p.history)-1]

	p.Goals = last.Goals
	p.Metas = last.Metas
	p.nextGoalID = last.nextGoalID

	return true
}

// ExtractTerm extracts the proof term once all goals are solved.
// Returns an error if there are unsolved goals.
func (p *ProofState) ExtractTerm() (ast.Term, error) {
	if !p.IsComplete() {
		return nil, fmt.Errorf("cannot extract term: %d goals remaining", len(p.Goals))
	}

	// Get the solution for the original goal (metavariable 0)
	sol, ok := p.Metas.GetSolution(0)
	if !ok {
		return nil, fmt.Errorf("no solution found for root goal")
	}

	// Zonk to substitute all metavariables
	return elab.ZonkFull(p.Metas, sol)
}

// FormatGoal returns a string representation of a goal.
// Uses context-aware printing to show variable names instead of de Bruijn indices.
func (p *ProofState) FormatGoal(g *Goal) string {
	result := ""

	// Build context from hypotheses for context-aware printing
	ctx := make([]string, 0, len(g.Hypotheses))
	for _, h := range g.Hypotheses {
		ctx = append(ctx, h.Name)
	}

	// Format hypotheses - each hypothesis can only reference earlier hypotheses
	for i, h := range g.Hypotheses {
		hypCtx := ctx[:i] // Context for this hypothesis (only earlier hypotheses)
		result += fmt.Sprintf("  %s : %s\n", h.Name, parser.FormatTermWithContext(h.Type, hypCtx))
	}

	// Format goal type with full context
	result += fmt.Sprintf("  ========================\n")
	result += fmt.Sprintf("  %s\n", parser.FormatTermWithContext(g.Type, ctx))

	return result
}

// FormatState returns a string representation of the proof state.
func (p *ProofState) FormatState() string {
	if p.IsComplete() {
		return "No more goals.\n"
	}

	result := fmt.Sprintf("%d goal(s)\n\n", len(p.Goals))

	for i, g := range p.Goals {
		if i == 0 {
			result += fmt.Sprintf("Goal %d (focused):\n", g.ID)
		} else {
			result += fmt.Sprintf("Goal %d:\n", g.ID)
		}
		result += p.FormatGoal(&g)
		result += "\n"
	}

	return result
}

// GetGoal retrieves a goal by ID.
func (p *ProofState) GetGoal(id GoalID) (*Goal, bool) {
	for i := range p.Goals {
		if p.Goals[i].ID == id {
			return &p.Goals[i], true
		}
	}
	return nil, false
}

// LookupHypothesis finds a hypothesis by name in a goal.
func (g *Goal) LookupHypothesis(name string) (*Hypothesis, int, bool) {
	for i := len(g.Hypotheses) - 1; i >= 0; i-- {
		if g.Hypotheses[i].Name == name {
			return &g.Hypotheses[i], len(g.Hypotheses) - 1 - i, true
		}
	}
	return nil, 0, false
}

// AddHypothesis adds a hypothesis to a goal.
func (g *Goal) AddHypothesis(name string, ty ast.Term) {
	g.Hypotheses = append(g.Hypotheses, Hypothesis{Name: name, Type: ty})
}
