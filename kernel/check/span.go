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

import "strconv"

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
			return strconv.Itoa(s.Start.Line) + ":" + strconv.Itoa(s.Start.Column)
		}
		return strconv.Itoa(s.Start.Line) + ":" + strconv.Itoa(s.Start.Column) + "-" + strconv.Itoa(s.End.Column)
	}
	return strconv.Itoa(s.Start.Line) + ":" + strconv.Itoa(s.Start.Column) + "-" + strconv.Itoa(s.End.Line) + ":" + strconv.Itoa(s.End.Column)
}
