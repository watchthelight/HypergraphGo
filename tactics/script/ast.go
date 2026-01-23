// Package script provides parsing and execution of HoTT tactic scripts.
//
// Script files (.htt) contain definitions, axioms, and theorem declarations:
//
//	-- HoTT Tactic Script
//	-- Comments start with --
//
//	-- Simple definition
//	Definition id : (Pi A Type (Pi _ (Var 0) (Var 1))) := (Lam A Type (Lam x (Var 0) (Var 0)))
//
//	-- Axiom (postulated)
//	Axiom funext : (Pi A Type (Pi B Type ...))
//
//	-- Theorem with proof
//	Theorem id_refl : (Id (Pi A Type (Pi _ A A)) id id)
//	Proof
//	  reflexivity
//	Qed
//
// Scripts are parsed into an AST and executed against the proof engine.
package script

import "github.com/watchthelight/HypergraphGo/internal/ast"

// ItemKind represents the type of script item.
type ItemKind int

const (
	ItemDefinition ItemKind = iota // Definition name : TYPE := TERM
	ItemAxiom                      // Axiom name : TYPE
	ItemTheorem                    // Theorem name : TYPE  Proof ... Qed
)

// Script represents a tactic script file containing definitions, axioms, and theorems.
type Script struct {
	Items    []Item    // All script items in order
	Theorems []Theorem // For backwards compatibility - just the theorems
}

// Item represents any top-level script item.
type Item struct {
	Kind       ItemKind
	Definition *Definition // If Kind == ItemDefinition
	Axiom      *Axiom      // If Kind == ItemAxiom
	Theorem    *Theorem    // If Kind == ItemTheorem
}

// Definition represents a term definition with type and body.
type Definition struct {
	Name string   // Definition name
	Type ast.Term // Type annotation
	Body ast.Term // Definition body
	Line int      // Source line number
}

// Axiom represents a postulated constant with only a type.
type Axiom struct {
	Name string   // Axiom name
	Type ast.Term // Type
	Line int      // Source line number
}

// Theorem represents a theorem declaration with its proof.
type Theorem struct {
	Name  string      // Theorem name (e.g., "id")
	Type  ast.Term    // Goal type
	Proof []TacticCmd // Tactic commands
	Line  int         // Source line number
}

// TacticCmd represents a single tactic command.
type TacticCmd struct {
	Name string   // Tactic name (e.g., "intro", "assumption")
	Args []string // Arguments (e.g., ["A"] for "intro A")
	Line int      // Source line number for error reporting
}
