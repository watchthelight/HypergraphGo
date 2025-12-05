// Package check implements bidirectional type checking for the HoTT kernel.
//
// Bidirectional type checking uses two modes:
//   - Synthesis (Synth): infer the type of a term
//   - Checking (Check): verify a term has an expected type
//
// This approach provides better error messages by tracking source positions
// and reduces the annotation burden by propagating type information.
//
// References:
//   - Dunfield, J. and Krishnaswami, N. "Bidirectional Typing"
//   - Löh, A. et al. "A Tutorial Implementation of a Dependently Typed Lambda Calculus"
package check

// Pos represents a position in source code.
type Pos struct {
	Line   int // 1-indexed line number
	Column int // 1-indexed column number
	Offset int // 0-indexed byte offset
}

// Span represents a range in source code.
type Span struct {
	File  string // Source file name (empty for REPL/tests)
	Start Pos
	End   Pos
}

// NoSpan returns an empty span for generated terms or tests.
func NoSpan() Span {
	return Span{}
}

// NewSpan creates a span from start and end positions.
func NewSpan(file string, startLine, startCol, endLine, endCol int) Span {
	return Span{
		File:  file,
		Start: Pos{Line: startLine, Column: startCol},
		End:   Pos{Line: endLine, Column: endCol},
	}
}

// IsEmpty returns true if this is a zero-value span.
func (s Span) IsEmpty() bool {
	return s.Start.Line == 0 && s.Start.Column == 0 && s.End.Line == 0 && s.End.Column == 0
}

// String returns a human-readable representation of the span.
func (s Span) String() string {
	if s.IsEmpty() {
		return "<no location>"
	}
	if s.File == "" {
		return formatSpanRange(s)
	}
	return s.File + ":" + formatSpanRange(s)
}

func formatSpanRange(s Span) string {
	if s.Start.Line == s.End.Line {
		if s.Start.Column == s.End.Column {
			return itoa(s.Start.Line) + ":" + itoa(s.Start.Column)
		}
		return itoa(s.Start.Line) + ":" + itoa(s.Start.Column) + "-" + itoa(s.End.Column)
	}
	return itoa(s.Start.Line) + ":" + itoa(s.Start.Column) + "-" + itoa(s.End.Line) + ":" + itoa(s.End.Column)
}

// itoa converts an int to string without importing strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	// reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}
