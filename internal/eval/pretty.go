package eval

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// SprintValue returns a string representation of a Value for debugging and testing.
func SprintValue(v Value) string {
	var b bytes.Buffer
	writeValue(&b, v)
	return b.String()
}

// SprintNeutral returns a string representation of a Neutral for debugging and testing.
func SprintNeutral(n Neutral) string {
	var b bytes.Buffer
	writeNeutral(&b, n)
	return b.String()
}

func writeValue(b *bytes.Buffer, v Value) {
	switch val := v.(type) {
	case VNeutral:
		writeNeutral(b, val.N)

	case VLam:
		b.WriteString("(\\_ => ")
		// For pretty printing, we could evaluate with a dummy variable
		// but for now, just indicate it's a closure
		b.WriteString("<closure>")
		b.WriteString(")")

	case VPair:
		b.WriteString("(")
		writeValue(b, val.Fst)
		b.WriteString(" , ")
		writeValue(b, val.Snd)
		b.WriteString(")")

	case VSort:
		b.WriteString("Type")
		b.WriteString(strconv.Itoa(val.Level))

	case VGlobal:
		b.WriteString(val.Name)

	case VPi:
		b.WriteString("(Pi _ : ")
		writeValue(b, val.A)
		b.WriteString(" . <closure>)")

	case VSigma:
		b.WriteString("(Sigma _ : ")
		writeValue(b, val.A)
		b.WriteString(" . <closure>)")

	default:
		b.WriteString("<?value?>")
	}
}

func writeNeutral(b *bytes.Buffer, n Neutral) {
	// Check upfront if we need parentheses (avoid string reallocation)
	needsParens := len(n.Sp) > 0
	if needsParens {
		b.WriteString("(")
	}

	// Write the head
	writeHead(b, n.Head)

	// Write the spine
	for _, arg := range n.Sp {
		b.WriteString(" ")
		writeValue(b, arg)
	}

	if needsParens {
		b.WriteString(")")
	}
}

func writeHead(b *bytes.Buffer, h Head) {
	if h.Var >= 0 && h.Glob == "" {
		b.WriteString("{")
		b.WriteString(strconv.Itoa(h.Var))
		b.WriteString("}")
	} else if h.Glob != "" {
		b.WriteString(h.Glob)
	} else {
		b.WriteString("<?head?>")
	}
}

// NormalizeNBE normalizes a term using NbE and returns its string representation.
// This function provides compatibility with existing test expectations.
func NormalizeNBE(t ast.Term) string {
	normalized := EvalNBE(t)
	return ast.Sprint(normalized)
}

// PrettyValue converts a Value to a stable string representation for testing.
// This ensures deterministic output for golden tests.
func PrettyValue(v Value) string {
	// For testing purposes, we reify the value and use ast.Sprint for consistency
	term := Reify(v)
	return ast.Sprint(term)
}

// PrettyNeutral converts a Neutral to a stable string representation for testing.
func PrettyNeutral(n Neutral) string {
	term := reifyNeutral(n)
	return ast.Sprint(term)
}

// ValueEqual compares two Values for structural equality (useful for testing).
func ValueEqual(v1, v2 Value) bool {
	switch val1 := v1.(type) {
	case VNeutral:
		if val2, ok := v2.(VNeutral); ok {
			return NeutralEqual(val1.N, val2.N)
		}
		return false

	case VLam:
		if val2, ok := v2.(VLam); ok {
			// For lambda equality, we'd need to compare under fresh variables
			// For now, just compare the closure terms structurally
			return termEqual(val1.Body.Term, val2.Body.Term) && envEqual(val1.Body.Env, val2.Body.Env)
		}
		return false

	case VPair:
		if val2, ok := v2.(VPair); ok {
			return ValueEqual(val1.Fst, val2.Fst) && ValueEqual(val1.Snd, val2.Snd)
		}
		return false

	case VSort:
		if val2, ok := v2.(VSort); ok {
			return val1.Level == val2.Level
		}
		return false

	case VGlobal:
		if val2, ok := v2.(VGlobal); ok {
			return val1.Name == val2.Name
		}
		return false

	case VPi:
		if val2, ok := v2.(VPi); ok {
			return ValueEqual(val1.A, val2.A) && closureEqual(val1.B, val2.B)
		}
		return false

	case VSigma:
		if val2, ok := v2.(VSigma); ok {
			return ValueEqual(val1.A, val2.A) && closureEqual(val1.B, val2.B)
		}
		return false

	default:
		return false
	}
}

// NeutralEqual compares two Neutrals for structural equality.
func NeutralEqual(n1, n2 Neutral) bool {
	if !headEqual(n1.Head, n2.Head) {
		return false
	}
	if len(n1.Sp) != len(n2.Sp) {
		return false
	}
	for i, arg1 := range n1.Sp {
		if !ValueEqual(arg1, n2.Sp[i]) {
			return false
		}
	}
	return true
}

func headEqual(h1, h2 Head) bool {
	return h1.Var == h2.Var && h1.Glob == h2.Glob
}

func termEqual(t1, t2 ast.Term) bool {
	// Simple structural equality check
	return ast.Sprint(t1) == ast.Sprint(t2)
}

func envEqual(e1, e2 *Env) bool {
	if e1 == nil && e2 == nil {
		return true
	}
	if e1 == nil || e2 == nil {
		return false
	}
	if len(e1.Bindings) != len(e2.Bindings) {
		return false
	}
	for i, v1 := range e1.Bindings {
		if !ValueEqual(v1, e2.Bindings[i]) {
			return false
		}
	}
	return true
}

func closureEqual(c1, c2 *Closure) bool {
	if c1 == nil && c2 == nil {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}
	return termEqual(c1.Term, c2.Term) && envEqual(c1.Env, c2.Env)
}

// DebugValue provides detailed debug information about a Value.
func DebugValue(v Value) string {
	var parts []string
	parts = append(parts, "Value: "+SprintValue(v))
	parts = append(parts, "Type: "+valueTypeName(v))
	if reified := Reify(v); reified != nil {
		parts = append(parts, "Reified: "+ast.Sprint(reified))
	}
	return strings.Join(parts, "\n")
}

func valueTypeName(v Value) string {
	switch v.(type) {
	case VNeutral:
		return "VNeutral"
	case VLam:
		return "VLam"
	case VPair:
		return "VPair"
	case VSort:
		return "VSort"
	case VGlobal:
		return "VGlobal"
	case VPi:
		return "VPi"
	case VSigma:
		return "VSigma"
	default:
		return "Unknown"
	}
}
