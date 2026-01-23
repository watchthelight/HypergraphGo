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
	ItemKind string // "definition", "axiom", or "theorem"
	Name     string
	Line     int
	Message  string
}

func (e *ExecError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s %s (line %d): %s", e.ItemKind, e.Name, e.Line, e.Message)
	}
	return fmt.Sprintf("%s %s: %s", e.ItemKind, e.Name, e.Message)
}

// Result represents the result of executing a script.
type Result struct {
	Items    []ItemResult    // All items in order
	Theorems []TheoremResult // For backward compatibility
}

// ItemResult represents the result of processing a script item.
type ItemResult struct {
	Kind       ItemKind
	Name       string
	Type       ast.Term
	Success    bool
	Error      error
	Definition *DefinitionResult // If Kind == ItemDefinition
	Axiom      *AxiomResult      // If Kind == ItemAxiom
	Theorem    *TheoremResult    // If Kind == ItemTheorem
}

// DefinitionResult represents the result of processing a definition.
type DefinitionResult struct {
	Name    string
	Type    ast.Term
	Body    ast.Term
	Success bool
	Error   error
}

// AxiomResult represents the result of processing an axiom.
type AxiomResult struct {
	Name    string
	Type    ast.Term
	Success bool
	Error   error
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
// Definitions and axioms are added to the checker's GlobalEnv so later items can reference them.
// Theorems are also added as definitions after successful proof.
func Execute(script *Script, checker *check.Checker) *Result {
	result := &Result{
		Items:    make([]ItemResult, 0, len(script.Items)),
		Theorems: make([]TheoremResult, 0, len(script.Theorems)),
	}

	globals := checker.Globals()

	for _, item := range script.Items {
		switch item.Kind {
		case ItemDefinition:
			defResult := executeDefinition(item.Definition, checker, globals)
			result.Items = append(result.Items, ItemResult{
				Kind:       ItemDefinition,
				Name:       item.Definition.Name,
				Type:       item.Definition.Type,
				Success:    defResult.Success,
				Error:      defResult.Error,
				Definition: defResult,
			})

		case ItemAxiom:
			axResult := executeAxiom(item.Axiom, checker, globals)
			result.Items = append(result.Items, ItemResult{
				Kind:    ItemAxiom,
				Name:    item.Axiom.Name,
				Type:    item.Axiom.Type,
				Success: axResult.Success,
				Error:   axResult.Error,
				Axiom:   axResult,
			})

		case ItemTheorem:
			thmResult := executeTheorem(*item.Theorem, checker, globals)
			result.Items = append(result.Items, ItemResult{
				Kind:    ItemTheorem,
				Name:    item.Theorem.Name,
				Type:    item.Theorem.Type,
				Success: thmResult.Success,
				Error:   thmResult.Error,
				Theorem: &thmResult,
			})
			result.Theorems = append(result.Theorems, thmResult)
		}
	}

	// Handle backward compatibility: if there are no Items but there are Theorems
	// (old-style script with only Theorems), process them directly
	if len(script.Items) == 0 && len(script.Theorems) > 0 {
		for _, thm := range script.Theorems {
			thmResult := executeTheorem(thm, checker, globals)
			result.Items = append(result.Items, ItemResult{
				Kind:    ItemTheorem,
				Name:    thm.Name,
				Type:    thm.Type,
				Success: thmResult.Success,
				Error:   thmResult.Error,
				Theorem: &thmResult,
			})
			result.Theorems = append(result.Theorems, thmResult)
		}
	}

	return result
}

// executeDefinition type-checks and adds a definition to the environment.
func executeDefinition(def *Definition, checker *check.Checker, globals *check.GlobalEnv) *DefinitionResult {
	result := &DefinitionResult{
		Name: def.Name,
		Type: def.Type,
		Body: def.Body,
	}

	// Verify the type is valid
	_, checkErr := checker.Synth(nil, check.Span{}, def.Type)
	if checkErr != nil {
		result.Error = &ExecError{
			ItemKind: "definition",
			Name:     def.Name,
			Line:     def.Line,
			Message:  fmt.Sprintf("invalid type: %v", checkErr),
		}
		return result
	}

	// Check that the body has the declared type
	checkErr = checker.Check(nil, check.Span{}, def.Body, def.Type)
	if checkErr != nil {
		result.Error = &ExecError{
			ItemKind: "definition",
			Name:     def.Name,
			Line:     def.Line,
			Message:  fmt.Sprintf("body type mismatch: %v", checkErr),
		}
		return result
	}

	// Add to global environment
	globals.AddDefinition(def.Name, def.Type, def.Body, check.Transparent)

	result.Success = true
	return result
}

// executeAxiom type-checks and adds an axiom to the environment.
func executeAxiom(ax *Axiom, checker *check.Checker, globals *check.GlobalEnv) *AxiomResult {
	result := &AxiomResult{
		Name: ax.Name,
		Type: ax.Type,
	}

	// Verify the type is valid
	_, checkErr := checker.Synth(nil, check.Span{}, ax.Type)
	if checkErr != nil {
		result.Error = &ExecError{
			ItemKind: "axiom",
			Name:     ax.Name,
			Line:     ax.Line,
			Message:  fmt.Sprintf("invalid type: %v", checkErr),
		}
		return result
	}

	// Add to global environment
	globals.AddAxiom(ax.Name, ax.Type)

	result.Success = true
	return result
}

// executeTheorem executes a single theorem proof.
// On success, adds the theorem as a definition to the global environment.
func executeTheorem(thm Theorem, checker *check.Checker, globals *check.GlobalEnv) TheoremResult {
	result := TheoremResult{
		Name: thm.Name,
		Type: thm.Type,
	}

	// Verify the goal type is valid
	_, checkErr := checker.Synth(nil, check.Span{}, thm.Type)
	if checkErr != nil {
		result.Error = &ExecError{
			ItemKind: "theorem",
			Name:     thm.Name,
			Line:     thm.Line,
			Message:  fmt.Sprintf("invalid goal type: %v", checkErr),
		}
		return result
	}

	// Create proof state
	state := proofstate.NewProofState(thm.Type, nil)

	// Apply each tactic
	for _, cmd := range thm.Proof {
		tactic, err := parseTactic(cmd.Name, cmd.Args, globals)
		if err != nil {
			result.Error = &ExecError{
				ItemKind: "theorem",
				Name:     thm.Name,
				Line:     cmd.Line,
				Message:  err.Error(),
			}
			return result
		}

		tacticResult := tactic(state)
		if !tacticResult.IsSuccess() {
			result.Error = &ExecError{
				ItemKind: "theorem",
				Name:     thm.Name,
				Line:     cmd.Line,
				Message:  fmt.Sprintf("tactic '%s' failed: %v", cmd.Name, tacticResult.Err),
			}
			return result
		}
	}

	// Check if proof is complete
	if !state.IsComplete() {
		result.Error = &ExecError{
			ItemKind: "theorem",
			Name:     thm.Name,
			Message:  fmt.Sprintf("proof incomplete: %d goals remaining", state.GoalCount()),
		}
		return result
	}

	// Extract proof term
	proofTerm, err := state.ExtractTerm()
	if err != nil {
		result.Error = &ExecError{
			ItemKind: "theorem",
			Name:     thm.Name,
			Message:  fmt.Sprintf("extraction failed: %v", err),
		}
		return result
	}

	// Type check the proof term
	checkErr = checker.Check(nil, check.NoSpan(), proofTerm, thm.Type)
	if checkErr != nil {
		result.Error = &ExecError{
			ItemKind: "theorem",
			Name:     thm.Name,
			Message:  fmt.Sprintf("type check failed: %v", checkErr),
		}
		return result
	}

	// Add theorem to global environment so later theorems can reference it
	globals.AddDefinition(thm.Name, thm.Type, proofTerm, check.Opaque)

	result.ProofTerm = proofTerm
	result.Success = true
	return result
}

// parseTactic converts a tactic name and args to a Tactic function.
func parseTactic(name string, args []string, globals *check.GlobalEnv) (tactics.Tactic, error) {
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

	case "unfold":
		if len(args) < 1 {
			return nil, fmt.Errorf("unfold requires a definition name")
		}
		return tactics.UnfoldWith(globals)(args[0]), nil

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
