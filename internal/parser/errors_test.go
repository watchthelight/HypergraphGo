package parser

import (
	"strings"
	"testing"
)

// --- ParseError Tests ---

func TestParseError_Error(t *testing.T) {
	err := &ParseError{Pos: 10, Message: "unexpected token"}
	result := err.Error()
	if !strings.Contains(result, "10") {
		t.Errorf("Error() should contain position: %s", result)
	}
	if !strings.Contains(result, "unexpected token") {
		t.Errorf("Error() should contain message: %s", result)
	}
}

func TestParseError_Position_Start(t *testing.T) {
	_, err := ParseTerm(")")
	if err == nil {
		t.Fatal("Expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("Expected *ParseError, got %T", err)
	}
	if pe.Pos != 0 {
		t.Errorf("Error position = %d, want 0", pe.Pos)
	}
}

func TestParseError_Position_AfterWhitespace(t *testing.T) {
	_, err := ParseTerm("   )")
	if err == nil {
		t.Fatal("Expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("Expected *ParseError, got %T", err)
	}
	// Position should be at the ')', after whitespace is skipped
	if pe.Pos != 3 {
		t.Errorf("Error position = %d, want 3", pe.Pos)
	}
}

func TestParseError_Position_InNestedExpr(t *testing.T) {
	// Error in nested expression
	_, err := ParseTerm("(App (Unknown) x)")
	if err == nil {
		t.Fatal("Expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("Expected *ParseError, got %T", err)
	}
	// Should point to or near "Unknown"
	if pe.Pos < 5 {
		t.Errorf("Error position = %d, should be >= 5 (inside nested expr)", pe.Pos)
	}
}

// --- Error Message Quality Tests ---

func TestErrorMessage_UnexpectedEOF(t *testing.T) {
	_, err := ParseTerm("(Pi x Nat")
	if err == nil {
		t.Fatal("Expected error")
	}
	msg := err.Error()
	// Should mention EOF or similar
	if !strings.Contains(strings.ToLower(msg), "eof") && !strings.Contains(msg, "')'") {
		t.Errorf("Error should mention EOF or expected ')': %s", msg)
	}
}

func TestErrorMessage_UnknownForm(t *testing.T) {
	_, err := ParseTerm("(Blah x y)")
	if err == nil {
		t.Fatal("Expected error")
	}
	msg := err.Error()
	// Should mention unknown form
	if !strings.Contains(strings.ToLower(msg), "unknown") {
		t.Errorf("Error should mention 'unknown': %s", msg)
	}
}

func TestErrorMessage_ExpectedIndex(t *testing.T) {
	_, err := ParseTerm("(Var abc)")
	if err == nil {
		t.Fatal("Expected error")
	}
	msg := err.Error()
	// Should mention index or number
	if !strings.Contains(strings.ToLower(msg), "index") && !strings.Contains(strings.ToLower(msg), "number") {
		t.Errorf("Error should mention index/number: %s", msg)
	}
}

func TestErrorMessage_ExpectedAtom(t *testing.T) {
	_, err := ParseTerm("()")
	if err == nil {
		t.Fatal("Expected error")
	}
	msg := err.Error()
	// Should mention form or atom
	if !strings.Contains(strings.ToLower(msg), "form") && !strings.Contains(strings.ToLower(msg), "atom") {
		t.Errorf("Error should mention form/atom: %s", msg)
	}
}

// --- Error Position Accuracy Tests ---

func TestErrorPosition_EmptyInput(t *testing.T) {
	_, err := ParseTerm("")
	if err == nil {
		t.Fatal("Expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("Expected *ParseError, got %T", err)
	}
	if pe.Pos != 0 {
		t.Errorf("Error position = %d, want 0 for empty input", pe.Pos)
	}
}

func TestErrorPosition_AfterComment(t *testing.T) {
	_, err := ParseTerm("; comment\n)")
	if err == nil {
		t.Fatal("Expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("Expected *ParseError, got %T", err)
	}
	// Position should be at the ')' after comment
	if pe.Pos < 10 {
		t.Errorf("Error position = %d, should be >= 10 (after comment)", pe.Pos)
	}
}

func TestErrorPosition_UnmatchedParen(t *testing.T) {
	_, err := ParseTerm("(Pi x Nat Nat))")
	if err == nil {
		t.Fatal("Expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("Expected *ParseError, got %T", err)
	}
	// Error should be at the extra ')'
	if pe.Pos != 14 {
		t.Errorf("Error position = %d, want 14 (at extra paren)", pe.Pos)
	}
}

// --- Multiple Error Scenarios ---

func TestParseErrors_Various(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errContains string
	}{
		{"Unexpected close", ")", "atom"},
		{"Missing form", "()", "form"},
		{"Bad sort level", "(Sort abc)", "level"},
		{"Incomplete Pi", "(Pi x)", "atom"},
		{"Unknown keyword", "(xyz)", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTerm(tt.input)
			if err == nil {
				t.Fatalf("Expected error for %q", tt.input)
			}
			msg := strings.ToLower(err.Error())
			if !strings.Contains(msg, tt.errContains) {
				t.Errorf("Error %q should contain %q", err.Error(), tt.errContains)
			}
		})
	}
}
