// Command hottgo is the HoTT kernel CLI.
//
// Usage:
//
//	hottgo --version           Print version info
//	hottgo --check FILE        Type-check a file of S-expression terms
//	hottgo --eval EXPR         Evaluate an S-expression term
//	hottgo --synth EXPR        Synthesize the type of an S-expression term
//	hottgo                     Start interactive REPL
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
)

func main() {
	ver := flag.Bool("version", false, "print version and exit")
	checkFile := flag.String("check", "", "file to type-check")
	evalExpr := flag.String("eval", "", "S-expression term to evaluate")
	synthExpr := flag.String("synth", "", "S-expression term to synthesize type")
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

	// REPL mode
	fmt.Println("hottgo - HoTT Kernel REPL")
	fmt.Println("Commands: :eval EXPR, :synth EXPR, :quit")
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

func repl() {
	checker := check.NewCheckerWithPrimitives()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if line == ":quit" || line == ":q" {
			break
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
}
