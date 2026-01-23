// Command hottgo is the HoTT kernel CLI.
//
// Usage:
//
//	hottgo --version           Print version info
//	hottgo --check FILE        Type-check a file of S-expression terms
//	hottgo --eval EXPR         Evaluate an S-expression term
//	hottgo --synth EXPR        Synthesize the type of an S-expression term
//	hottgo --load FILE         Load and verify a tactic script (.htt)
//	hottgo                     Start interactive REPL
//
// REPL Commands:
//
//	:eval EXPR                 Evaluate an expression
//	:synth EXPR                Synthesize the type of an expression
//	:define NAME TYPE TERM     Define a new constant
//	:axiom NAME TYPE           Postulate an axiom
//	:prove TYPE                Start proof mode with goal TYPE
//	:prove NAME : TYPE         Start proof mode with named theorem
//	:quit                      Exit the REPL
//
// Proof Mode Commands (when in proof mode):
//
//	:tactic NAME [ARGS]        Apply a tactic
//	:goal                      Show current goal
//	:goals                     Show all goals
//	:undo                      Undo last tactic
//	:qed                       Extract and verify proof, add to environment
//	:abort                     Exit proof mode without completing
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/watchthelight/HypergraphGo/internal/eval"
	"github.com/watchthelight/HypergraphGo/internal/parser"
	"github.com/watchthelight/HypergraphGo/internal/version"
	"github.com/watchthelight/HypergraphGo/kernel/check"
	"github.com/watchthelight/HypergraphGo/tactics/proofstate"
	"github.com/watchthelight/HypergraphGo/tactics/script"
)

func main() {
	ver := flag.Bool("version", false, "print version and exit")
	checkFile := flag.String("check", "", "file to type-check")
	evalExpr := flag.String("eval", "", "S-expression term to evaluate")
	synthExpr := flag.String("synth", "", "S-expression term to synthesize type")
	loadScript := flag.String("load", "", "tactic script file to load and verify")
	flag.Parse()

	if *ver {
		fmt.Printf("hottgo %s (%s, %s)\n", version.Version, version.Commit, version.Date)
		return
	}

	if *checkFile != "" {
		if err := doCheck(*checkFile); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *evalExpr != "" {
		if err := doEval(*evalExpr); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *synthExpr != "" {
		if err := doSynth(*synthExpr); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *loadScript != "" {
		if err := doLoad(*loadScript); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// REPL mode
	fmt.Println("hottgo - HoTT Kernel REPL")
	fmt.Println("Commands: :eval EXPR, :synth EXPR, :prove TYPE, :quit")
	fmt.Println("Type :help for more information")
	fmt.Println()
	repl()
}

func doCheck(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	terms, err := parser.ParseMultiple(string(data))
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	checker := check.NewCheckerWithPrimitives()
	for i, term := range terms {
		ty, checkErr := checker.Synth(nil, check.Span{}, term)
		if checkErr != nil {
			return fmt.Errorf("term %d: %v", i+1, checkErr)
		}
		fmt.Printf("term %d : %s\n", i+1, parser.FormatTerm(ty))
	}

	return nil
}

func doEval(expr string) error {
	term, err := parser.ParseTerm(expr)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	result := eval.EvalNBE(term)
	fmt.Println(parser.FormatTerm(result))
	return nil
}

func doSynth(expr string) error {
	term, err := parser.ParseTerm(expr)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	checker := check.NewCheckerWithPrimitives()
	ty, checkErr := checker.Synth(nil, check.Span{}, term)
	if checkErr != nil {
		return fmt.Errorf("type error: %v", checkErr)
	}

	fmt.Printf("%s : %s\n", parser.FormatTerm(term), parser.FormatTerm(ty))
	return nil
}

func doLoad(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	scr, err := script.ParseString(string(data))
	if err != nil {
		return fmt.Errorf("parsing script: %w", err)
	}

	if len(scr.Theorems) == 0 {
		fmt.Println("No theorems in script.")
		return nil
	}

	checker := check.NewCheckerWithStdlib()
	result := script.Execute(scr, checker)

	// Report results
	successCount := 0
	for _, thm := range result.Theorems {
		if thm.Success {
			fmt.Printf("✓ %s : %s\n", thm.Name, parser.FormatTerm(thm.Type))
			successCount++
		} else {
			fmt.Printf("✗ %s : %s\n", thm.Name, parser.FormatTerm(thm.Type))
			fmt.Printf("  Error: %v\n", thm.Error)
		}
	}

	fmt.Printf("\n%d/%d theorems verified.\n", successCount, len(result.Theorems))

	if successCount < len(result.Theorems) {
		return fmt.Errorf("some theorems failed")
	}

	return nil
}

// replSettings holds configurable display options.
type replSettings struct {
	namedVariables bool // Show named variables instead of de Bruijn indices
	verbose        bool // Show verbose output
}

// replState holds the state for the REPL session.
type replState struct {
	checker      *check.Checker
	proofMode    *ProofMode
	theoremCount int // For generating anonymous theorem names
	settings     replSettings
}

func repl() {
	state := &replState{
		checker: check.NewCheckerWithStdlib(),
		settings: replSettings{
			namedVariables: true, // Default: show named variables in proof mode
			verbose:        false,
		},
	}
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Show different prompt based on mode
		if state.proofMode != nil {
			fmt.Printf("proof[%d]> ", state.proofMode.GoalCount())
		} else {
			fmt.Print("> ")
		}

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if line == ":quit" || line == ":q" {
			if state.proofMode != nil {
				fmt.Println("Aborting proof mode.")
				state.proofMode = nil
			}
			break
		}

		if line == ":help" || line == ":h" {
			printHelp(state.proofMode != nil)
			continue
		}

		if line == ":clear" || line == ":cls" {
			// Clear terminal using ANSI escape codes
			fmt.Print("\033[H\033[2J")
			continue
		}

		// Handle proof mode commands
		if state.proofMode != nil {
			handled := handleProofModeCommand(state, line)
			if handled {
				continue
			}
		}

		// Handle :prove command to enter proof mode
		if strings.HasPrefix(line, ":prove ") {
			rest := strings.TrimPrefix(line, ":prove ")
			handleProveCommand(state, rest)
			continue
		}

		// Handle :define command
		if strings.HasPrefix(line, ":define ") {
			rest := strings.TrimPrefix(line, ":define ")
			handleDefineCommand(state, rest)
			continue
		}

		// Handle :axiom command
		if strings.HasPrefix(line, ":axiom ") {
			rest := strings.TrimPrefix(line, ":axiom ")
			handleAxiomCommand(state, rest)
			continue
		}

		if line == ":examples" {
			printExamples()
			continue
		}

		if line == ":tutorial" {
			printTutorial()
			continue
		}

		if strings.HasPrefix(line, ":set ") {
			handleSetCommand(state, strings.TrimPrefix(line, ":set "))
			continue
		}

		if line == ":settings" {
			printSettings(state)
			continue
		}

		if line == ":env" || line == ":environment" {
			handleEnvCommand(state, "")
			continue
		}

		if strings.HasPrefix(line, ":env ") {
			handleEnvCommand(state, strings.TrimPrefix(line, ":env "))
			continue
		}

		if strings.HasPrefix(line, ":print ") {
			handlePrintCommand(state, strings.TrimPrefix(line, ":print "))
			continue
		}

		if strings.HasPrefix(line, ":search ") {
			handleSearchCommand(state, strings.TrimPrefix(line, ":search "))
			continue
		}

		if strings.HasPrefix(line, ":eval ") {
			expr := strings.TrimPrefix(line, ":eval ")
			if err := doEval(expr); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
			continue
		}

		if strings.HasPrefix(line, ":synth ") {
			expr := strings.TrimPrefix(line, ":synth ")
			term, err := parser.ParseTerm(expr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
				continue
			}
			ty, checkErr := state.checker.Synth(nil, check.Span{}, term)
			if checkErr != nil {
				fmt.Fprintf(os.Stderr, "type error: %v\n", checkErr)
				continue
			}
			fmt.Printf("%s : %s\n", parser.FormatTerm(term), parser.FormatTerm(ty))
			continue
		}

		// Default: try to synth the expression
		term, err := parser.ParseTerm(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
			continue
		}
		ty, checkErr := state.checker.Synth(nil, check.Span{}, term)
		if checkErr != nil {
			fmt.Fprintf(os.Stderr, "type error: %v\n", checkErr)
			continue
		}
		fmt.Printf("%s\n", parser.FormatTerm(ty))
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
	}
}

// handleProveCommand parses ":prove TYPE" or ":prove NAME : TYPE".
func handleProveCommand(state *replState, rest string) {
	var name string
	var typeStr string

	// Check if there's a "NAME : TYPE" pattern
	if colonIdx := strings.Index(rest, ":"); colonIdx > 0 {
		possibleName := strings.TrimSpace(rest[:colonIdx])
		// Only treat it as a named theorem if the name doesn't contain spaces
		// and doesn't start with '('
		if !strings.Contains(possibleName, " ") && !strings.HasPrefix(possibleName, "(") {
			name = possibleName
			typeStr = strings.TrimSpace(rest[colonIdx+1:])
		} else {
			typeStr = rest
		}
	} else {
		typeStr = rest
	}

	goalTy, err := parser.ParseTerm(typeStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		return
	}

	// Verify it's a valid type
	_, checkErr := state.checker.Synth(nil, check.Span{}, goalTy)
	if checkErr != nil {
		fmt.Fprintf(os.Stderr, "type error: %v\n", checkErr)
		return
	}

	if name != "" {
		state.proofMode = NewProofModeNamed(name, goalTy, state.checker)
		fmt.Printf("Entering proof mode for theorem '%s'.\n", name)
	} else {
		state.proofMode = NewProofMode(goalTy, state.checker)
		fmt.Println("Entering proof mode.")
	}
	fmt.Println(state.proofMode.FormatCurrentGoal())
}

// handleDefineCommand parses ":define NAME TYPE TERM" and adds a definition.
func handleDefineCommand(state *replState, rest string) {
	// Parse: NAME TYPE TERM
	// First token is the name
	parts := strings.Fields(rest)
	if len(parts) < 3 {
		fmt.Fprintln(os.Stderr, "usage: :define NAME TYPE TERM")
		return
	}

	name := parts[0]

	// Need to find where TYPE ends and TERM begins
	// This is tricky because both are S-expressions
	// Let's require them to be space-separated top-level terms
	restAfterName := strings.TrimPrefix(rest, name)
	restAfterName = strings.TrimSpace(restAfterName)

	// Parse the type (first complete S-expression)
	typeStr, termStr, err := splitTwoTerms(restAfterName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	defType, err := parser.ParseTerm(typeStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error in type: %v\n", err)
		return
	}

	defBody, err := parser.ParseTerm(termStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error in body: %v\n", err)
		return
	}

	// Verify the type is valid
	_, checkErr := state.checker.Synth(nil, check.Span{}, defType)
	if checkErr != nil {
		fmt.Fprintf(os.Stderr, "type error: %v\n", checkErr)
		return
	}

	// Check that the body has the declared type
	checkErr = state.checker.Check(nil, check.Span{}, defBody, defType)
	if checkErr != nil {
		fmt.Fprintf(os.Stderr, "body type mismatch: %v\n", checkErr)
		return
	}

	// Add to global environment
	state.checker.Globals().AddDefinition(name, defType, defBody, check.Transparent)
	fmt.Printf("Defined %s : %s\n", name, parser.FormatTerm(defType))
}

// handleAxiomCommand parses ":axiom NAME TYPE" and adds an axiom.
func handleAxiomCommand(state *replState, rest string) {
	// Parse: NAME TYPE
	parts := strings.Fields(rest)
	if len(parts) < 2 {
		fmt.Fprintln(os.Stderr, "usage: :axiom NAME TYPE")
		return
	}

	name := parts[0]
	typeStr := strings.TrimPrefix(rest, name)
	typeStr = strings.TrimSpace(typeStr)

	axType, err := parser.ParseTerm(typeStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		return
	}

	// Verify the type is valid
	_, checkErr := state.checker.Synth(nil, check.Span{}, axType)
	if checkErr != nil {
		fmt.Fprintf(os.Stderr, "type error: %v\n", checkErr)
		return
	}

	// Add to global environment
	state.checker.Globals().AddAxiom(name, axType)
	fmt.Printf("Axiom %s : %s\n", name, parser.FormatTerm(axType))
}

// splitTwoTerms splits a string into two S-expression terms.
func splitTwoTerms(s string) (string, string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", "", fmt.Errorf("expected two terms")
	}

	// Find the end of the first term
	var firstEnd int
	if s[0] == '(' {
		// Find matching closing paren
		depth := 0
		for i, c := range s {
			if c == '(' {
				depth++
			} else if c == ')' {
				depth--
				if depth == 0 {
					firstEnd = i + 1
					break
				}
			}
		}
		if depth != 0 {
			return "", "", fmt.Errorf("unbalanced parentheses in type")
		}
	} else {
		// Simple atom - find first whitespace
		idx := strings.IndexFunc(s, func(r rune) bool {
			return r == ' ' || r == '\t' || r == '\n'
		})
		if idx == -1 {
			return "", "", fmt.Errorf("expected two terms, got one")
		}
		firstEnd = idx
	}

	first := strings.TrimSpace(s[:firstEnd])
	second := strings.TrimSpace(s[firstEnd:])
	if second == "" {
		return "", "", fmt.Errorf("expected two terms, got one")
	}

	return first, second, nil
}

// handleProofModeCommand processes proof mode specific commands.
// Returns true if the command was handled.
func handleProofModeCommand(state *replState, line string) bool {
	pm := state.proofMode

	switch {
	case line == ":goal" || line == ":g":
		fmt.Println(pm.FormatCurrentGoal())
		return true

	case line == ":goals":
		fmt.Println(pm.FormatAllGoals())
		return true

	case line == ":undo" || line == ":u":
		if pm.Undo() {
			fmt.Println("Undone.")
			fmt.Println(pm.FormatCurrentGoal())
		} else {
			fmt.Println("Nothing to undo.")
		}
		return true

	case line == ":qed":
		if !pm.IsComplete() {
			fmt.Fprintf(os.Stderr, "Proof not complete. %d goals remaining.\n", pm.GoalCount())
			return true
		}
		term, err := pm.Extract()
		if err != nil {
			fmt.Fprintf(os.Stderr, "extraction error: %v\n", err)
			return true
		}
		if err := pm.TypeCheck(); err != nil {
			fmt.Fprintf(os.Stderr, "type check failed: %v\n", err)
			return true
		}

		// Generate theorem name if not provided
		thmName := pm.TheoremName()
		if thmName == "" {
			state.theoremCount++
			thmName = fmt.Sprintf("anon_%d", state.theoremCount)
		}

		// Add theorem to global environment
		state.checker.Globals().AddDefinition(thmName, pm.GoalType(), term, check.Opaque)

		fmt.Println("Proof complete!")
		fmt.Printf("Added theorem: %s : %s\n", thmName, parser.FormatTerm(pm.GoalType()))
		fmt.Printf("Term: %s\n", parser.FormatTerm(term))
		state.proofMode = nil
		return true

	case line == ":abort":
		fmt.Println("Proof aborted.")
		state.proofMode = nil
		return true

	case strings.HasPrefix(line, ":type "):
		expr := strings.TrimPrefix(line, ":type ")
		handleTypeCommand(state, expr)
		return true

	case strings.HasPrefix(line, ":reduce "):
		expr := strings.TrimPrefix(line, ":reduce ")
		handleReduceCommand(state, expr)
		return true

	case strings.HasPrefix(line, ":focus "):
		expr := strings.TrimPrefix(line, ":focus ")
		handleFocusCommand(state, expr)
		return true

	case line == ":history" || line == ":hist":
		handleHistoryCommand(state)
		return true

	case strings.HasPrefix(line, ":checkpoint ") || strings.HasPrefix(line, ":cp "):
		var name string
		if strings.HasPrefix(line, ":checkpoint ") {
			name = strings.TrimPrefix(line, ":checkpoint ")
		} else {
			name = strings.TrimPrefix(line, ":cp ")
		}
		handleCheckpointCommand(state, name)
		return true

	case strings.HasPrefix(line, ":restore "):
		name := strings.TrimPrefix(line, ":restore ")
		handleRestoreCommand(state, name)
		return true

	case line == ":checkpoints" || line == ":cps":
		handleListCheckpointsCommand(state)
		return true

	case strings.HasPrefix(line, ":tactic ") || strings.HasPrefix(line, ":t "):
		var rest string
		if strings.HasPrefix(line, ":tactic ") {
			rest = strings.TrimPrefix(line, ":tactic ")
		} else {
			rest = strings.TrimPrefix(line, ":t ")
		}
		parts := strings.Fields(rest)
		if len(parts) == 0 {
			fmt.Fprintf(os.Stderr, "usage: :tactic NAME [ARGS]\n")
			return true
		}
		tacticName := parts[0]
		tacticArgs := parts[1:]
		applyTacticWithOutput(state, tacticName, tacticArgs)
		return true

	default:
		// In proof mode, bare words are treated as tactics
		parts := strings.Fields(line)
		if len(parts) > 0 && !strings.HasPrefix(line, ":") {
			tacticName := parts[0]
			tacticArgs := parts[1:]
			applyTacticWithOutput(state, tacticName, tacticArgs)
			return true
		}
	}

	return false
}

// applyTacticWithOutput applies a tactic and shows the result with optional verbose info.
func applyTacticWithOutput(state *replState, name string, args []string) {
	pm := state.proofMode

	msg, err := pm.ApplyTactic(name, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tactic error: %v\n", err)
		return
	}

	fmt.Println(msg)

	// In verbose mode, show additional info
	if state.settings.verbose {
		fmt.Printf("  [%d goal(s) remaining]\n", pm.GoalCount())
	}

	if pm.IsComplete() {
		fmt.Println("No more goals. Type :qed to complete the proof.")
	} else {
		fmt.Println(pm.FormatCurrentGoal())
	}
}

// handleTypeCommand shows the type of a term in proof mode context.
func handleTypeCommand(state *replState, expr string) {
	if state.proofMode == nil {
		fmt.Fprintf(os.Stderr, "not in proof mode\n")
		return
	}

	term, err := parser.ParseTerm(expr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		return
	}

	// Try to synthesize the type
	ty, checkErr := state.checker.Synth(nil, check.Span{}, term)
	if checkErr != nil {
		fmt.Fprintf(os.Stderr, "type error: %v\n", checkErr)
		return
	}

	fmt.Printf("%s : %s\n", parser.FormatTerm(term), parser.FormatTerm(ty))
}

// handleReduceCommand normalizes and displays a term.
func handleReduceCommand(state *replState, expr string) {
	term, err := parser.ParseTerm(expr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		return
	}

	reduced := eval.EvalNBE(term)
	fmt.Printf("%s\n", parser.FormatTerm(reduced))
}

// handleFocusCommand focuses on a specific goal by ID.
func handleFocusCommand(state *replState, expr string) {
	if state.proofMode == nil {
		fmt.Fprintf(os.Stderr, "not in proof mode\n")
		return
	}

	var goalID int
	if _, err := fmt.Sscanf(expr, "%d", &goalID); err != nil {
		fmt.Fprintf(os.Stderr, "invalid goal ID: %s\n", expr)
		return
	}

	// Access the underlying proof state to focus
	ps := state.proofMode.state
	if err := ps.Focus(proofstate.GoalID(goalID)); err != nil {
		fmt.Fprintf(os.Stderr, "focus error: %v\n", err)
		return
	}

	fmt.Printf("Focused on goal %d\n", goalID)
	fmt.Println(state.proofMode.FormatCurrentGoal())
}

// handleHistoryCommand displays the tactic history.
func handleHistoryCommand(state *replState) {
	if state.proofMode == nil {
		fmt.Fprintf(os.Stderr, "not in proof mode\n")
		return
	}

	history := state.proofMode.History()
	if len(history) == 0 {
		fmt.Println("No tactics applied yet.")
		return
	}

	fmt.Println("Tactic History:")
	for i, entry := range history {
		argsStr := ""
		if len(entry.Args) > 0 {
			argsStr = " " + strings.Join(entry.Args, " ")
		}
		fmt.Printf("  %d. %s%s\n", i+1, entry.Name, argsStr)
	}
}

// handleCheckpointCommand saves a checkpoint with the given name.
func handleCheckpointCommand(state *replState, name string) {
	if state.proofMode == nil {
		fmt.Fprintf(os.Stderr, "not in proof mode\n")
		return
	}

	name = strings.TrimSpace(name)
	if name == "" {
		fmt.Fprintf(os.Stderr, "usage: :checkpoint NAME\n")
		return
	}

	state.proofMode.SaveCheckpoint(name)
	fmt.Printf("Saved checkpoint '%s'\n", name)
}

// handleRestoreCommand restores a previously saved checkpoint.
func handleRestoreCommand(state *replState, name string) {
	if state.proofMode == nil {
		fmt.Fprintf(os.Stderr, "not in proof mode\n")
		return
	}

	name = strings.TrimSpace(name)
	if name == "" {
		fmt.Fprintf(os.Stderr, "usage: :restore NAME\n")
		return
	}

	if err := state.proofMode.RestoreCheckpoint(name); err != nil {
		fmt.Fprintf(os.Stderr, "restore error: %v\n", err)
		return
	}

	fmt.Printf("Restored checkpoint '%s'\n", name)
	fmt.Println(state.proofMode.FormatCurrentGoal())
}

// handleListCheckpointsCommand lists all saved checkpoints.
func handleListCheckpointsCommand(state *replState) {
	if state.proofMode == nil {
		fmt.Fprintf(os.Stderr, "not in proof mode\n")
		return
	}

	checkpoints := state.proofMode.ListCheckpoints()
	if len(checkpoints) == 0 {
		fmt.Println("No checkpoints saved.")
		return
	}

	fmt.Println("Saved Checkpoints:")
	for _, name := range checkpoints {
		fmt.Printf("  - %s\n", name)
	}
}

// printHelp displays help information.
func printHelp(inProofMode bool) {
	fmt.Println("HoTTGo REPL Commands:")
	fmt.Println()
	fmt.Println("  :eval EXPR            Evaluate an expression")
	fmt.Println("  :synth EXPR           Synthesize the type of an expression")
	fmt.Println("  :define NAME TYPE TERM  Define a new constant")
	fmt.Println("  :axiom NAME TYPE      Postulate an axiom")
	fmt.Println("  :prove TYPE           Enter proof mode with goal TYPE")
	fmt.Println("  :prove NAME : TYPE    Enter proof mode with named theorem")
	fmt.Println("  :env [FILTER]         List environment (axioms/defs/inductives)")
	fmt.Println("  :print NAME           Print the body of a definition")
	fmt.Println("  :search PATTERN       Search for entries by type pattern")
	fmt.Println("  :set OPTION VALUE     Set display option (named, verbose)")
	fmt.Println("  :settings             Show current settings")
	fmt.Println("  :clear, :cls          Clear the terminal")
	fmt.Println("  :help, :h             Show this help")
	fmt.Println("  :quit, :q             Exit the REPL")
	fmt.Println()
	if inProofMode {
		fmt.Println("Proof Mode Commands:")
		fmt.Println()
		fmt.Println("  :goal, :g         Show current goal")
		fmt.Println("  :goals            Show all goals")
		fmt.Println("  :tactic NAME      Apply a tactic (or just type tactic name)")
		fmt.Println("  :undo, :u         Undo last tactic")
		fmt.Println("  :history, :hist   Show tactic history")
		fmt.Println("  :type TERM        Show type of a term")
		fmt.Println("  :reduce TERM      Normalize and display a term")
		fmt.Println("  :focus N          Focus on goal N")
		fmt.Println("  :checkpoint NAME  Save current proof state")
		fmt.Println("  :restore NAME     Restore saved checkpoint")
		fmt.Println("  :checkpoints      List saved checkpoints")
		fmt.Println("  :qed              Complete proof and add to environment")
		fmt.Println("  :abort            Exit proof mode")
		fmt.Println()
		fmt.Println("Available Tactics:")
		fmt.Println()
		fmt.Println("  Introduction:")
		fmt.Println("    intro [NAME]      Introduce a hypothesis from Pi type")
		fmt.Println("    intros            Introduce all hypotheses")
		fmt.Println()
		fmt.Println("  Proof Completion:")
		fmt.Println("    exact TERM        Provide exact proof term")
		fmt.Println("    assumption        Use hypothesis matching the goal")
		fmt.Println("    reflexivity       Prove by reflexivity (refl, Id, Path)")
		fmt.Println("    trivial           Try reflexivity and assumption")
		fmt.Println("    auto              Automatic proof search")
		fmt.Println()
		fmt.Println("  Equality Reasoning:")
		fmt.Println("    rewrite H         Rewrite using equality H : Id A x y")
		fmt.Println("    symmetry H        Reverse H : Id A x y to Id A y x")
		fmt.Println("    transitivity H1 H2  Chain H1 : Id A x y, H2 : Id A y z")
		fmt.Println("    ap FUNC H         Apply function to both sides of H")
		fmt.Println("    transport P X     Transport x : A along path P : Path Type A B")
		fmt.Println()
		fmt.Println("  Product/Sum Types:")
		fmt.Println("    split             Split Sigma goal into two subgoals")
		fmt.Println("    exists TERM       Provide witness for Sigma goal")
		fmt.Println("    left              Prove Sum goal with left injection")
		fmt.Println("    right             Prove Sum goal with right injection")
		fmt.Println()
		fmt.Println("  Case Analysis:")
		fmt.Println("    destruct H        Case analysis on H (Sum, Bool)")
		fmt.Println("    induction H       Induction on H (Nat, List)")
		fmt.Println("    cases H           Non-recursive case analysis")
		fmt.Println("    constructor       Apply first applicable constructor")
		fmt.Println("    contradiction     Prove from Empty hypothesis")
		fmt.Println()
		fmt.Println("  Simplification:")
		fmt.Println("    simpl             Simplify (normalize) the goal")
		fmt.Println("    unfold NAME       Unfold a definition in the goal")
		fmt.Println("    apply TERM        Apply function/theorem to the goal")
	}
}

// printExamples displays example proofs for the REPL.
func printExamples() {
	fmt.Println("HoTTGo Examples")
	fmt.Println("===============")
	fmt.Println()
	fmt.Println("Example 1: Reflexivity")
	fmt.Println("-----------------------")
	fmt.Println("  :prove (Id Nat zero zero)")
	fmt.Println("  reflexivity")
	fmt.Println("  :qed")
	fmt.Println()
	fmt.Println("Example 2: Function composition")
	fmt.Println("-------------------------------")
	fmt.Println("  :prove (Pi (A Type) (Pi (B Type) (Pi (C Type)")
	fmt.Println("    (Pi (f (Pi (_ A) B)) (Pi (g (Pi (_ B) C)) (Pi (_ A) C))))))")
	fmt.Println("  intros")
	fmt.Println("  exact (g (f x))")
	fmt.Println("  :qed")
	fmt.Println()
	fmt.Println("Example 3: Dependent pair")
	fmt.Println("--------------------------")
	fmt.Println("  :prove (Sigma (n Nat) (Id Nat n n))")
	fmt.Println("  split")
	fmt.Println("  exact zero")
	fmt.Println("  reflexivity")
	fmt.Println("  :qed")
	fmt.Println()
	fmt.Println("Example 4: Symmetry of equality")
	fmt.Println("-------------------------------")
	fmt.Println("  :prove (Pi (A Type) (Pi (x A) (Pi (y A)")
	fmt.Println("    (Pi (p (Id A x y)) (Id A y x)))))")
	fmt.Println("  intros")
	fmt.Println("  symmetry p")
	fmt.Println("  :qed")
	fmt.Println()
}

// printTutorial displays an interactive tutorial.
func printTutorial() {
	fmt.Println("HoTTGo Interactive Tutorial")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Println("Welcome to HoTTGo! This tutorial will guide you through basic proof")
	fmt.Println("construction in Homotopy Type Theory.")
	fmt.Println()
	fmt.Println("STEP 1: Starting a Proof")
	fmt.Println("------------------------")
	fmt.Println("Use ':prove TYPE' to start proving a theorem. For example:")
	fmt.Println()
	fmt.Println("  :prove (Pi (A Type) (Pi (x A) A))")
	fmt.Println()
	fmt.Println("This starts a proof that, given any type A and element x : A, we can")
	fmt.Println("produce an element of A (the identity function).")
	fmt.Println()
	fmt.Println("STEP 2: Understanding Goals")
	fmt.Println("---------------------------")
	fmt.Println("After entering proof mode, you'll see your current goal:")
	fmt.Println()
	fmt.Println("  Goal 0 (focused):")
	fmt.Println("    ========================")
	fmt.Println("    (Pi (A Type) (Pi (x A) A))")
	fmt.Println()
	fmt.Println("The line above '====' shows your hypotheses (assumptions),")
	fmt.Println("and below shows what you need to prove.")
	fmt.Println()
	fmt.Println("STEP 3: Introduction")
	fmt.Println("--------------------")
	fmt.Println("For Pi types, use 'intro' to move the argument into hypotheses:")
	fmt.Println()
	fmt.Println("  intro A")
	fmt.Println("  intro x")
	fmt.Println()
	fmt.Println("Now your goal becomes:")
	fmt.Println()
	fmt.Println("    A : Type")
	fmt.Println("    x : A")
	fmt.Println("    ========================")
	fmt.Println("    A")
	fmt.Println()
	fmt.Println("STEP 4: Completing the Proof")
	fmt.Println("----------------------------")
	fmt.Println("Use 'assumption' when the goal matches a hypothesis, or")
	fmt.Println("use 'exact TERM' to provide an explicit proof term:")
	fmt.Println()
	fmt.Println("  assumption  -- uses x since x : A and goal is A")
	fmt.Println()
	fmt.Println("STEP 5: Finishing Up")
	fmt.Println("--------------------")
	fmt.Println("When all goals are solved, use ':qed' to complete the proof:")
	fmt.Println()
	fmt.Println("  :qed")
	fmt.Println()
	fmt.Println("The theorem will be type-checked and added to your environment.")
	fmt.Println()
	fmt.Println("OTHER USEFUL COMMANDS:")
	fmt.Println("  :undo          -- undo the last tactic")
	fmt.Println("  :goal          -- show current goal")
	fmt.Println("  :goals         -- show all goals")
	fmt.Println("  :help          -- show available tactics")
	fmt.Println("  :examples      -- show more examples")
	fmt.Println()
}

// handleSetCommand processes :set option value commands.
func handleSetCommand(state *replState, args string) {
	parts := strings.Fields(args)
	if len(parts) < 2 {
		fmt.Fprintf(os.Stderr, "usage: :set OPTION VALUE\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  named on|off    Show named variables (default: on)\n")
		fmt.Fprintf(os.Stderr, "  verbose on|off  Verbose output (default: off)\n")
		return
	}

	option := parts[0]
	value := parts[1]

	switch option {
	case "named":
		switch value {
		case "on", "true", "1":
			state.settings.namedVariables = true
			fmt.Println("Named variables: on")
		case "off", "false", "0":
			state.settings.namedVariables = false
			fmt.Println("Named variables: off")
		default:
			fmt.Fprintf(os.Stderr, "invalid value for 'named': use on or off\n")
		}
	case "verbose":
		switch value {
		case "on", "true", "1":
			state.settings.verbose = true
			fmt.Println("Verbose: on")
		case "off", "false", "0":
			state.settings.verbose = false
			fmt.Println("Verbose: off")
		default:
			fmt.Fprintf(os.Stderr, "invalid value for 'verbose': use on or off\n")
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown option: %s\n", option)
		fmt.Fprintf(os.Stderr, "Available options: named, verbose\n")
	}
}

// printSettings displays current REPL settings.
func printSettings(state *replState) {
	fmt.Println("Current Settings:")
	fmt.Printf("  named variables: %v\n", boolToOnOff(state.settings.namedVariables))
	fmt.Printf("  verbose: %v\n", boolToOnOff(state.settings.verbose))
}

func boolToOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

// handleEnvCommand displays the contents of the global environment.
func handleEnvCommand(state *replState, filter string) {
	globals := state.checker.Globals()
	filter = strings.TrimSpace(strings.ToLower(filter))

	// Filter options: "axioms", "defs", "inductives", "all" (default)
	showAxioms := filter == "" || filter == "all" || filter == "axioms"
	showDefs := filter == "" || filter == "all" || filter == "defs" || filter == "definitions"
	showInductives := filter == "" || filter == "all" || filter == "inductives" || filter == "types"

	anyShown := false

	if showAxioms {
		axioms := globals.Axioms()
		if len(axioms) > 0 {
			anyShown = true
			fmt.Println("Axioms:")
			for _, name := range axioms {
				ty := globals.LookupType(name)
				fmt.Printf("  %s : %s\n", name, parser.FormatTerm(ty))
			}
			fmt.Println()
		}
	}

	if showDefs {
		defs := globals.Definitions()
		if len(defs) > 0 {
			anyShown = true
			fmt.Println("Definitions:")
			for _, name := range defs {
				ty := globals.LookupType(name)
				fmt.Printf("  %s : %s\n", name, parser.FormatTerm(ty))
			}
			fmt.Println()
		}
	}

	if showInductives {
		inds := globals.Inductives()
		if len(inds) > 0 {
			anyShown = true
			fmt.Println("Inductive Types:")
			for _, name := range inds {
				ty := globals.LookupType(name)
				fmt.Printf("  %s : %s\n", name, parser.FormatTerm(ty))
			}
			fmt.Println()
		}
	}

	if !anyShown {
		if filter != "" {
			fmt.Printf("No entries matching '%s'.\n", filter)
		} else {
			fmt.Println("Environment is empty (no user-defined entries).")
		}
	}
}

// handlePrintCommand prints the body of a definition.
func handlePrintCommand(state *replState, name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		fmt.Fprintln(os.Stderr, "usage: :print NAME")
		return
	}

	globals := state.checker.Globals()

	// First check what kind of entry it is
	kind := globals.GetKind(name)

	switch kind {
	case check.KindAxiom:
		ty := globals.LookupType(name)
		fmt.Printf("axiom %s : %s\n", name, parser.FormatTerm(ty))

	case check.KindDefinition:
		ty := globals.LookupType(name)
		body, _ := globals.LookupDefinitionBodyForced(name)
		fmt.Printf("def %s : %s\n", name, parser.FormatTerm(ty))
		fmt.Printf("  := %s\n", parser.FormatTerm(body))

	case check.KindInductive:
		ty := globals.LookupType(name)
		fmt.Printf("inductive %s : %s\n", name, parser.FormatTerm(ty))
		// Print constructors
		ind := globals.GetInductive(name)
		if ind != nil && len(ind.Constructors) > 0 {
			fmt.Println("  constructors:")
			for _, c := range ind.Constructors {
				fmt.Printf("    %s : %s\n", c.Name, parser.FormatTerm(c.Type))
			}
		}

	case check.KindConstructor:
		ty := globals.LookupType(name)
		fmt.Printf("constructor %s : %s\n", name, parser.FormatTerm(ty))

	case check.KindPrimitive:
		ty := globals.LookupType(name)
		fmt.Printf("primitive %s : %s\n", name, parser.FormatTerm(ty))

	default:
		fmt.Fprintf(os.Stderr, "unknown name: %s\n", name)
	}
}

// handleSearchCommand searches for definitions matching a type pattern.
func handleSearchCommand(state *replState, query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		fmt.Fprintln(os.Stderr, "usage: :search TYPE_PATTERN")
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  :search Nat         - find entries with Nat in their type")
		fmt.Fprintln(os.Stderr, "  :search -> Nat      - find entries returning Nat")
		fmt.Fprintln(os.Stderr, "  :search Id          - find identity type entries")
		return
	}

	globals := state.checker.Globals()
	queryLower := strings.ToLower(query)

	// Search through all entries
	found := false
	for _, name := range globals.Order() {
		ty := globals.LookupType(name)
		if ty == nil {
			continue
		}

		// Format the type and search for the pattern
		tyStr := parser.FormatTerm(ty)
		tyStrLower := strings.ToLower(tyStr)

		if strings.Contains(tyStrLower, queryLower) {
			if !found {
				fmt.Println("Matching entries:")
				found = true
			}
			kind := globals.GetKind(name)
			fmt.Printf("  %s %s : %s\n", kind, name, tyStr)
		}
	}

	if !found {
		fmt.Printf("No entries matching '%s'.\n", query)
	}
}
