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
//	:prove TYPE                Start proof mode with goal TYPE
//	:quit                      Exit the REPL
//
// Proof Mode Commands (when in proof mode):
//
//	:tactic NAME [ARGS]        Apply a tactic
//	:goal                      Show current goal
//	:goals                     Show all goals
//	:undo                      Undo last tactic
//	:qed                       Extract and verify proof
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

func repl() {
	checker := check.NewCheckerWithStdlib()
	scanner := bufio.NewScanner(os.Stdin)
	var proofMode *ProofMode

	for {
		// Show different prompt based on mode
		if proofMode != nil {
			fmt.Printf("proof[%d]> ", proofMode.GoalCount())
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
			if proofMode != nil {
				fmt.Println("Aborting proof mode.")
				proofMode = nil
			}
			break
		}

		if line == ":help" || line == ":h" {
			printHelp(proofMode != nil)
			continue
		}

		// Handle proof mode commands
		if proofMode != nil {
			handled := handleProofModeCommand(proofMode, line, &proofMode)
			if handled {
				continue
			}
		}

		// Handle :prove command to enter proof mode
		if strings.HasPrefix(line, ":prove ") {
			expr := strings.TrimPrefix(line, ":prove ")
			goalTy, err := parser.ParseTerm(expr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
				continue
			}
			// Verify it's a valid type
			_, checkErr := checker.Synth(nil, check.Span{}, goalTy)
			if checkErr != nil {
				fmt.Fprintf(os.Stderr, "type error: %v\n", checkErr)
				continue
			}
			proofMode = NewProofMode(goalTy, checker)
			fmt.Println("Entering proof mode.")
			fmt.Println(proofMode.FormatCurrentGoal())
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
			ty, checkErr := checker.Synth(nil, check.Span{}, term)
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
		ty, checkErr := checker.Synth(nil, check.Span{}, term)
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

// handleProofModeCommand processes proof mode specific commands.
// Returns true if the command was handled.
func handleProofModeCommand(pm *ProofMode, line string, proofModePtr **ProofMode) bool {
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
		fmt.Println("Proof complete!")
		fmt.Printf("Term: %s\n", parser.FormatTerm(term))
		*proofModePtr = nil
		return true

	case line == ":abort":
		fmt.Println("Proof aborted.")
		*proofModePtr = nil
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
		msg, err := pm.ApplyTactic(tacticName, tacticArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tactic error: %v\n", err)
		} else {
			fmt.Println(msg)
			if pm.IsComplete() {
				fmt.Println("No more goals. Type :qed to complete the proof.")
			} else {
				fmt.Println(pm.FormatCurrentGoal())
			}
		}
		return true

	default:
		// In proof mode, bare words are treated as tactics
		parts := strings.Fields(line)
		if len(parts) > 0 && !strings.HasPrefix(line, ":") {
			tacticName := parts[0]
			tacticArgs := parts[1:]
			msg, err := pm.ApplyTactic(tacticName, tacticArgs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "tactic error: %v\n", err)
			} else {
				fmt.Println(msg)
				if pm.IsComplete() {
					fmt.Println("No more goals. Type :qed to complete the proof.")
				} else {
					fmt.Println(pm.FormatCurrentGoal())
				}
			}
			return true
		}
	}

	return false
}

// printHelp displays help information.
func printHelp(inProofMode bool) {
	fmt.Println("HoTTGo REPL Commands:")
	fmt.Println()
	fmt.Println("  :eval EXPR        Evaluate an expression")
	fmt.Println("  :synth EXPR       Synthesize the type of an expression")
	fmt.Println("  :prove TYPE       Enter proof mode with goal TYPE")
	fmt.Println("  :help, :h         Show this help")
	fmt.Println("  :quit, :q         Exit the REPL")
	fmt.Println()
	if inProofMode {
		fmt.Println("Proof Mode Commands:")
		fmt.Println()
		fmt.Println("  :goal, :g         Show current goal")
		fmt.Println("  :goals            Show all goals")
		fmt.Println("  :tactic NAME      Apply a tactic (or just type tactic name)")
		fmt.Println("  :undo, :u         Undo last tactic")
		fmt.Println("  :qed              Complete and verify the proof")
		fmt.Println("  :abort            Exit proof mode")
		fmt.Println()
		fmt.Println("Available Tactics:")
		fmt.Println("  intro [NAME]      Introduce a hypothesis")
		fmt.Println("  intros            Introduce all hypotheses")
		fmt.Println("  exact TERM        Provide exact proof term")
		fmt.Println("  assumption        Use a hypothesis matching the goal")
		fmt.Println("  reflexivity       Prove equality by reflexivity")
		fmt.Println("  split             Split a Sigma goal into two subgoals")
		fmt.Println("  left              Prove Sum goal with left injection")
		fmt.Println("  right             Prove Sum goal with right injection")
		fmt.Println("  destruct H        Case analysis on H (Sum, Bool)")
		fmt.Println("  induction H       Induction on H (Nat, List)")
		fmt.Println("  cases H           Non-recursive case analysis")
		fmt.Println("  constructor       Apply first applicable constructor")
		fmt.Println("  exists TERM       Provide witness for existential")
		fmt.Println("  contradiction     Prove from Empty hypothesis")
		fmt.Println("  rewrite H         Rewrite using equality H")
		fmt.Println("  simpl             Simplify the goal")
		fmt.Println("  trivial           Try reflexivity and assumption")
		fmt.Println("  auto              Automatic proof search")
	}
}
