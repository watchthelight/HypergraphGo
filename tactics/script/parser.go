package script

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/watchthelight/HypergraphGo/internal/parser"
)

// ParseError represents a parsing error with location information.
type ParseError struct {
	Line    int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}

// Parse parses a tactic script from a reader.
func Parse(r io.Reader) (*Script, error) {
	scanner := bufio.NewScanner(r)
	script := &Script{}
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		// Parse theorem declaration
		if strings.HasPrefix(line, "Theorem ") {
			thm, endLine, err := parseTheorem(line, lineNum, scanner)
			if err != nil {
				return nil, err
			}
			script.Theorems = append(script.Theorems, *thm)
			lineNum = endLine
			continue
		}

		return nil, &ParseError{Line: lineNum, Message: fmt.Sprintf("unexpected: %s", line)}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading script: %w", err)
	}

	return script, nil
}

// ParseString parses a tactic script from a string.
func ParseString(s string) (*Script, error) {
	return Parse(strings.NewReader(s))
}

// parseTheorem parses a theorem declaration and its proof.
// Returns the theorem and the last line number consumed.
func parseTheorem(firstLine string, startLine int, scanner *bufio.Scanner) (*Theorem, int, error) {
	// Parse "Theorem NAME : TYPE"
	rest := strings.TrimPrefix(firstLine, "Theorem ")
	colonIdx := strings.Index(rest, ":")
	if colonIdx == -1 {
		return nil, startLine, &ParseError{Line: startLine, Message: "expected ':' in theorem declaration"}
	}

	name := strings.TrimSpace(rest[:colonIdx])
	typeStr := strings.TrimSpace(rest[colonIdx+1:])

	if name == "" {
		return nil, startLine, &ParseError{Line: startLine, Message: "theorem name cannot be empty"}
	}

	// Parse the type expression
	goalType, err := parser.ParseTerm(typeStr)
	if err != nil {
		return nil, startLine, &ParseError{Line: startLine, Message: fmt.Sprintf("parsing type: %v", err)}
	}

	thm := &Theorem{
		Name: name,
		Type: goalType,
	}

	lineNum := startLine
	inProof := false

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		if line == "Proof" {
			if inProof {
				return nil, lineNum, &ParseError{Line: lineNum, Message: "unexpected 'Proof' inside proof block"}
			}
			inProof = true
			continue
		}

		if line == "Qed" {
			if !inProof {
				return nil, lineNum, &ParseError{Line: lineNum, Message: "'Qed' without 'Proof'"}
			}
			return thm, lineNum, nil
		}

		if !inProof {
			return nil, lineNum, &ParseError{Line: lineNum, Message: fmt.Sprintf("expected 'Proof', got: %s", line)}
		}

		// Parse tactic command
		cmd, err := parseTacticCmd(line, lineNum)
		if err != nil {
			return nil, lineNum, err
		}
		thm.Proof = append(thm.Proof, *cmd)
	}

	if inProof {
		return nil, lineNum, &ParseError{Line: lineNum, Message: "unexpected end of file in proof block"}
	}
	return nil, lineNum, &ParseError{Line: lineNum, Message: "unexpected end of file after theorem declaration"}
}

// parseTacticCmd parses a single tactic command.
func parseTacticCmd(line string, lineNum int) (*TacticCmd, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, &ParseError{Line: lineNum, Message: "empty tactic command"}
	}

	return &TacticCmd{
		Name: parts[0],
		Args: parts[1:],
		Line: lineNum,
	}, nil
}
