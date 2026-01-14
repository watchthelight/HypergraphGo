// Package script provides parsing and execution of HoTT tactic scripts.
//
// Script files (.htt) contain theorem declarations with their proofs:
//
//	-- HoTT Tactic Script
//	-- Comments start with --
//
//	Theorem id : (A : Type) -> A -> A
//	Proof
//	  intro A
//	  intro x
//	  assumption
//	Qed
//
// Scripts are parsed into an AST and executed against the proof engine.
package script

import "github.com/watchthelight/HypergraphGo/internal/ast"

// Script represents a tactic script file containing theorems.
type Script struct {
	Theorems []Theorem
}

// Theorem represents a theorem declaration with its proof.
type Theorem struct {
	Name  string      // Theorem name (e.g., "id")
	Type  ast.Term    // Goal type
	Proof []TacticCmd // Tactic commands
}

// TacticCmd represents a single tactic command.
type TacticCmd struct {
	Name string   // Tactic name (e.g., "intro", "assumption")
	Args []string // Arguments (e.g., ["A"] for "intro A")
	Line int      // Source line number for error reporting
}
