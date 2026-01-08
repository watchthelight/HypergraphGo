// Package tactics provides Ltac-style proof tactics for HoTT.
// This file provides the Go API for interactive proof construction.

package tactics

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/tactics/proofstate"
)

// Prover manages an interactive proof session.
type Prover struct {
	state *proofstate.ProofState
}

// NewProver creates a new prover for the given goal type.
func NewProver(goalType ast.Term) *Prover {
	return &Prover{
		state: proofstate.NewProofState(goalType, nil),
	}
}

// NewProverWithHyps creates a new prover with initial hypotheses.
func NewProverWithHyps(goalType ast.Term, hyps []proofstate.Hypothesis) *Prover {
	return &Prover{
		state: proofstate.NewProofState(goalType, hyps),
	}
}

// Apply runs a tactic on the current proof state.
func (p *Prover) Apply(t Tactic) error {
	result := RunTactic(p.state, t)
	if !result.IsSuccess() {
		return result.Err
	}
	p.state = result.State
	return nil
}

// Goals returns all current goals.
func (p *Prover) Goals() []proofstate.Goal {
	return p.state.Goals
}

// CurrentGoal returns the focused goal.
func (p *Prover) CurrentGoal() *proofstate.Goal {
	return p.state.CurrentGoal()
}

// GoalCount returns the number of remaining goals.
func (p *Prover) GoalCount() int {
	return p.state.GoalCount()
}

// Done returns true if all goals have been solved.
func (p *Prover) Done() bool {
	return p.state.IsComplete()
}

// Extract returns the completed proof term.
func (p *Prover) Extract() (ast.Term, error) {
	return p.state.ExtractTerm()
}

// Undo undoes the last tactic.
func (p *Prover) Undo() bool {
	return p.state.Undo()
}

// Focus switches to a specific goal.
func (p *Prover) Focus(id proofstate.GoalID) error {
	return p.state.Focus(id)
}

// FormatState returns a string representation of the current state.
func (p *Prover) FormatState() string {
	return p.state.FormatState()
}

// State returns the underlying proof state.
func (p *Prover) State() *proofstate.ProofState {
	return p.state
}

// Fluent API methods that return the prover for chaining

// Intro_ applies intro with error tracking.
func (p *Prover) Intro_(name string) *Prover {
	if err := p.Apply(Intro(name)); err != nil {
		p.setError(err)
	}
	return p
}

// IntroN_ applies introN with error tracking.
func (p *Prover) IntroN_(names ...string) *Prover {
	if err := p.Apply(IntroN(names...)); err != nil {
		p.setError(err)
	}
	return p
}

// Intros_ applies intros with error tracking.
func (p *Prover) Intros_() *Prover {
	if err := p.Apply(Intros()); err != nil {
		p.setError(err)
	}
	return p
}

// Exact_ applies exact with error tracking.
func (p *Prover) Exact_(term ast.Term) *Prover {
	if err := p.Apply(Exact(term)); err != nil {
		p.setError(err)
	}
	return p
}

// Assumption_ applies assumption with error tracking.
func (p *Prover) Assumption_() *Prover {
	if err := p.Apply(Assumption()); err != nil {
		p.setError(err)
	}
	return p
}

// Reflexivity_ applies reflexivity with error tracking.
func (p *Prover) Reflexivity_() *Prover {
	if err := p.Apply(Reflexivity()); err != nil {
		p.setError(err)
	}
	return p
}

// Split_ applies split with error tracking.
func (p *Prover) Split_() *Prover {
	if err := p.Apply(Split()); err != nil {
		p.setError(err)
	}
	return p
}

// Simpl_ applies simpl with error tracking.
func (p *Prover) Simpl_() *Prover {
	if err := p.Apply(Simpl()); err != nil {
		p.setError(err)
	}
	return p
}

// Rewrite_ applies rewrite with error tracking.
func (p *Prover) Rewrite_(hypName string) *Prover {
	if err := p.Apply(Rewrite(hypName)); err != nil {
		p.setError(err)
	}
	return p
}

// RewriteRev_ applies rewrite_rev with error tracking.
func (p *Prover) RewriteRev_(hypName string) *Prover {
	if err := p.Apply(RewriteRev(hypName)); err != nil {
		p.setError(err)
	}
	return p
}

// Trivial_ applies trivial with error tracking.
func (p *Prover) Trivial_() *Prover {
	if err := p.Apply(Trivial()); err != nil {
		p.setError(err)
	}
	return p
}

// Auto_ applies auto with error tracking.
func (p *Prover) Auto_() *Prover {
	if err := p.Apply(Auto()); err != nil {
		p.setError(err)
	}
	return p
}

// lastError stores the last error from fluent API methods
var lastError error

// setError stores an error for later retrieval.
func (p *Prover) setError(err error) {
	lastError = err
}

// Error returns any error from the last fluent API call.
func (p *Prover) Error() error {
	err := lastError
	lastError = nil
	return err
}

// Prove is a convenience function that creates a prover and runs tactics.
// It returns the extracted proof term on success.
func Prove(goalType ast.Term, tactics ...Tactic) (ast.Term, error) {
	prover := NewProver(goalType)

	for _, t := range tactics {
		if err := prover.Apply(t); err != nil {
			return nil, fmt.Errorf("tactic failed: %w", err)
		}
	}

	if !prover.Done() {
		return nil, fmt.Errorf("proof incomplete: %d goals remaining", prover.GoalCount())
	}

	return prover.Extract()
}

// MustProve is like Prove but panics on failure.
func MustProve(goalType ast.Term, tactics ...Tactic) ast.Term {
	term, err := Prove(goalType, tactics...)
	if err != nil {
		panic(err)
	}
	return term
}

// ProveSeq proves a goal using a sequence of tactics.
func ProveSeq(goalType ast.Term, tactic Tactic) (ast.Term, error) {
	return Prove(goalType, tactic)
}
