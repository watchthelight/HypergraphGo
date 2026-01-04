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

	case VId:
		b.WriteString("(Id ")
		writeValue(b, val.A)
		b.WriteString(" ")
		writeValue(b, val.X)
		b.WriteString(" ")
		writeValue(b, val.Y)
		b.WriteString(")")

	case VRefl:
		b.WriteString("(refl ")
		writeValue(b, val.A)
		b.WriteString(" ")
		writeValue(b, val.X)
		b.WriteString(")")

	// --- Cubical Values ---

	case VI0:
		b.WriteString("i0")

	case VI1:
		b.WriteString("i1")

	case VIVar:
		b.WriteString("i{")
		b.WriteString(strconv.Itoa(val.Level))
		b.WriteString("}")

	case VPath:
		b.WriteString("(Path ")
		writeValue(b, val.A)
		b.WriteString(" ")
		writeValue(b, val.X)
		b.WriteString(" ")
		writeValue(b, val.Y)
		b.WriteString(")")

	case VPathP:
		b.WriteString("(PathP <closure> ")
		writeValue(b, val.X)
		b.WriteString(" ")
		writeValue(b, val.Y)
		b.WriteString(")")

	case VPathLam:
		b.WriteString("(<_> <closure>)")

	case VTransport:
		b.WriteString("(transport <closure> ")
		writeValue(b, val.E)
		b.WriteString(")")

	case VFaceTop:
		b.WriteString("⊤")

	case VFaceBot:
		b.WriteString("⊥")

	case VFaceEq:
		b.WriteString("(i{")
		b.WriteString(strconv.Itoa(val.ILevel))
		b.WriteString("} = ")
		if val.IsOne {
			b.WriteString("1")
		} else {
			b.WriteString("0")
		}
		b.WriteString(")")

	case VFaceAnd:
		b.WriteString("(")
		writeFaceValue(b, val.Left)
		b.WriteString(" ∧ ")
		writeFaceValue(b, val.Right)
		b.WriteString(")")

	case VFaceOr:
		b.WriteString("(")
		writeFaceValue(b, val.Left)
		b.WriteString(" ∨ ")
		writeFaceValue(b, val.Right)
		b.WriteString(")")

	case VPartial:
		b.WriteString("(Partial ")
		writeFaceValue(b, val.Phi)
		b.WriteString(" ")
		writeValue(b, val.A)
		b.WriteString(")")

	case VSystem:
		b.WriteString("[")
		for i, br := range val.Branches {
			if i > 0 {
				b.WriteString(", ")
			}
			writeFaceValue(b, br.Phi)
			b.WriteString(" ↦ ")
			writeValue(b, br.Term)
		}
		b.WriteString("]")

	case VComp:
		b.WriteString("(comp <closure> [")
		writeFaceValue(b, val.Phi)
		b.WriteString(" ↦ <closure>] ")
		writeValue(b, val.Base)
		b.WriteString(")")

	case VHComp:
		b.WriteString("(hcomp ")
		writeValue(b, val.A)
		b.WriteString(" [")
		writeFaceValue(b, val.Phi)
		b.WriteString(" ↦ <closure>] ")
		writeValue(b, val.Base)
		b.WriteString(")")

	case VFill:
		b.WriteString("(fill <closure> [")
		writeFaceValue(b, val.Phi)
		b.WriteString(" ↦ <closure>] ")
		writeValue(b, val.Base)
		b.WriteString(")")

	case VGlue:
		b.WriteString("(Glue ")
		writeValue(b, val.A)
		b.WriteString(" [...])")

	case VGlueElem:
		b.WriteString("(glue [...] ")
		writeValue(b, val.Base)
		b.WriteString(")")

	case VUnglue:
		b.WriteString("(unglue ")
		writeValue(b, val.G)
		b.WriteString(")")

	case VUA:
		b.WriteString("(ua ")
		writeValue(b, val.A)
		b.WriteString(" ")
		writeValue(b, val.B)
		b.WriteString(" ")
		writeValue(b, val.Equiv)
		b.WriteString(")")

	case VUABeta:
		b.WriteString("(ua-β ")
		writeValue(b, val.Equiv)
		b.WriteString(" ")
		writeValue(b, val.Arg)
		b.WriteString(")")

	// --- Higher Inductive Types ---

	case VHITPathCtor:
		b.WriteString("(")
		b.WriteString(val.CtorName)
		for _, arg := range val.Args {
			b.WriteString(" ")
			writeValue(b, arg)
		}
		for _, iarg := range val.IArgs {
			b.WriteString(" @ ")
			writeValue(b, iarg)
		}
		b.WriteString(")")

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

// writeFaceValue writes a FaceValue to the buffer.
func writeFaceValue(b *bytes.Buffer, f FaceValue) {
	if f == nil {
		b.WriteString("⊥")
		return
	}
	switch fv := f.(type) {
	case VFaceTop:
		b.WriteString("⊤")
	case VFaceBot:
		b.WriteString("⊥")
	case VFaceEq:
		b.WriteString("(i{")
		b.WriteString(strconv.Itoa(fv.ILevel))
		b.WriteString("} = ")
		if fv.IsOne {
			b.WriteString("1")
		} else {
			b.WriteString("0")
		}
		b.WriteString(")")
	case VFaceAnd:
		b.WriteString("(")
		writeFaceValue(b, fv.Left)
		b.WriteString(" ∧ ")
		writeFaceValue(b, fv.Right)
		b.WriteString(")")
	case VFaceOr:
		b.WriteString("(")
		writeFaceValue(b, fv.Left)
		b.WriteString(" ∨ ")
		writeFaceValue(b, fv.Right)
		b.WriteString(")")
	default:
		b.WriteString("?face")
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
	term := reifyNeutralAt(0, n)
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

	case VId:
		if val2, ok := v2.(VId); ok {
			return ValueEqual(val1.A, val2.A) && ValueEqual(val1.X, val2.X) && ValueEqual(val1.Y, val2.Y)
		}
		return false

	case VRefl:
		if val2, ok := v2.(VRefl); ok {
			return ValueEqual(val1.A, val2.A) && ValueEqual(val1.X, val2.X)
		}
		return false

	// --- Cubical Values ---

	case VI0:
		_, ok := v2.(VI0)
		return ok

	case VI1:
		_, ok := v2.(VI1)
		return ok

	case VIVar:
		if val2, ok := v2.(VIVar); ok {
			return val1.Level == val2.Level
		}
		return false

	case VPath:
		if val2, ok := v2.(VPath); ok {
			return ValueEqual(val1.A, val2.A) && ValueEqual(val1.X, val2.X) && ValueEqual(val1.Y, val2.Y)
		}
		return false

	case VPathP:
		if val2, ok := v2.(VPathP); ok {
			return iClosureEqual(val1.A, val2.A) && ValueEqual(val1.X, val2.X) && ValueEqual(val1.Y, val2.Y)
		}
		return false

	case VPathLam:
		if val2, ok := v2.(VPathLam); ok {
			return iClosureEqual(val1.Body, val2.Body)
		}
		return false

	case VTransport:
		if val2, ok := v2.(VTransport); ok {
			return iClosureEqual(val1.A, val2.A) && ValueEqual(val1.E, val2.E)
		}
		return false

	case VFaceTop:
		_, ok := v2.(VFaceTop)
		return ok

	case VFaceBot:
		_, ok := v2.(VFaceBot)
		return ok

	case VFaceEq:
		if val2, ok := v2.(VFaceEq); ok {
			return val1.ILevel == val2.ILevel && val1.IsOne == val2.IsOne
		}
		return false

	case VFaceAnd:
		if val2, ok := v2.(VFaceAnd); ok {
			return faceValueEqual(val1.Left, val2.Left) && faceValueEqual(val1.Right, val2.Right)
		}
		return false

	case VFaceOr:
		if val2, ok := v2.(VFaceOr); ok {
			return faceValueEqual(val1.Left, val2.Left) && faceValueEqual(val1.Right, val2.Right)
		}
		return false

	case VPartial:
		if val2, ok := v2.(VPartial); ok {
			return faceValueEqual(val1.Phi, val2.Phi) && ValueEqual(val1.A, val2.A)
		}
		return false

	case VSystem:
		if val2, ok := v2.(VSystem); ok {
			if len(val1.Branches) != len(val2.Branches) {
				return false
			}
			for i := range val1.Branches {
				if !faceValueEqual(val1.Branches[i].Phi, val2.Branches[i].Phi) ||
					!ValueEqual(val1.Branches[i].Term, val2.Branches[i].Term) {
					return false
				}
			}
			return true
		}
		return false

	case VComp:
		if val2, ok := v2.(VComp); ok {
			return iClosureEqual(val1.A, val2.A) && faceValueEqual(val1.Phi, val2.Phi) &&
				iClosureEqual(val1.Tube, val2.Tube) && ValueEqual(val1.Base, val2.Base)
		}
		return false

	case VHComp:
		if val2, ok := v2.(VHComp); ok {
			return ValueEqual(val1.A, val2.A) && faceValueEqual(val1.Phi, val2.Phi) &&
				iClosureEqual(val1.Tube, val2.Tube) && ValueEqual(val1.Base, val2.Base)
		}
		return false

	case VFill:
		if val2, ok := v2.(VFill); ok {
			return iClosureEqual(val1.A, val2.A) && faceValueEqual(val1.Phi, val2.Phi) &&
				iClosureEqual(val1.Tube, val2.Tube) && ValueEqual(val1.Base, val2.Base)
		}
		return false

	case VGlue:
		if val2, ok := v2.(VGlue); ok {
			if !ValueEqual(val1.A, val2.A) || len(val1.System) != len(val2.System) {
				return false
			}
			for i := range val1.System {
				if !faceValueEqual(val1.System[i].Phi, val2.System[i].Phi) ||
					!ValueEqual(val1.System[i].T, val2.System[i].T) ||
					!ValueEqual(val1.System[i].Equiv, val2.System[i].Equiv) {
					return false
				}
			}
			return true
		}
		return false

	case VGlueElem:
		if val2, ok := v2.(VGlueElem); ok {
			if len(val1.System) != len(val2.System) {
				return false
			}
			for i := range val1.System {
				if !faceValueEqual(val1.System[i].Phi, val2.System[i].Phi) ||
					!ValueEqual(val1.System[i].Term, val2.System[i].Term) {
					return false
				}
			}
			return ValueEqual(val1.Base, val2.Base)
		}
		return false

	case VUnglue:
		if val2, ok := v2.(VUnglue); ok {
			return ValueEqual(val1.Ty, val2.Ty) && ValueEqual(val1.G, val2.G)
		}
		return false

	case VUA:
		if val2, ok := v2.(VUA); ok {
			return ValueEqual(val1.A, val2.A) && ValueEqual(val1.B, val2.B) && ValueEqual(val1.Equiv, val2.Equiv)
		}
		return false

	case VUABeta:
		if val2, ok := v2.(VUABeta); ok {
			return ValueEqual(val1.Equiv, val2.Equiv) && ValueEqual(val1.Arg, val2.Arg)
		}
		return false

	// --- Higher Inductive Types ---

	case VHITPathCtor:
		if val2, ok := v2.(VHITPathCtor); ok {
			if val1.HITName != val2.HITName || val1.CtorName != val2.CtorName {
				return false
			}
			if len(val1.Args) != len(val2.Args) || len(val1.IArgs) != len(val2.IArgs) {
				return false
			}
			for i := range val1.Args {
				if !ValueEqual(val1.Args[i], val2.Args[i]) {
					return false
				}
			}
			for i := range val1.IArgs {
				if !ValueEqual(val1.IArgs[i], val2.IArgs[i]) {
					return false
				}
			}
			return true
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

// iClosureEqual compares two interval closures for structural equality.
func iClosureEqual(c1, c2 *IClosure) bool {
	if c1 == nil && c2 == nil {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}
	return termEqual(c1.Term, c2.Term) && envEqual(c1.Env, c2.Env) && ienvEqual(c1.IEnv, c2.IEnv)
}

// ienvEqual compares two interval environments for equality.
func ienvEqual(e1, e2 *IEnv) bool {
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

// faceValueEqual compares two FaceValues for structural equality.
func faceValueEqual(f1, f2 FaceValue) bool {
	if f1 == nil && f2 == nil {
		return true
	}
	if f1 == nil || f2 == nil {
		return false
	}
	switch fv1 := f1.(type) {
	case VFaceTop:
		_, ok := f2.(VFaceTop)
		return ok
	case VFaceBot:
		_, ok := f2.(VFaceBot)
		return ok
	case VFaceEq:
		if fv2, ok := f2.(VFaceEq); ok {
			return fv1.ILevel == fv2.ILevel && fv1.IsOne == fv2.IsOne
		}
		return false
	case VFaceAnd:
		if fv2, ok := f2.(VFaceAnd); ok {
			return faceValueEqual(fv1.Left, fv2.Left) && faceValueEqual(fv1.Right, fv2.Right)
		}
		return false
	case VFaceOr:
		if fv2, ok := f2.(VFaceOr); ok {
			return faceValueEqual(fv1.Left, fv2.Left) && faceValueEqual(fv1.Right, fv2.Right)
		}
		return false
	default:
		return false
	}
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
	case VId:
		return "VId"
	case VRefl:
		return "VRefl"
	// --- Cubical Values ---
	case VI0:
		return "VI0"
	case VI1:
		return "VI1"
	case VIVar:
		return "VIVar"
	case VPath:
		return "VPath"
	case VPathP:
		return "VPathP"
	case VPathLam:
		return "VPathLam"
	case VTransport:
		return "VTransport"
	case VFaceTop:
		return "VFaceTop"
	case VFaceBot:
		return "VFaceBot"
	case VFaceEq:
		return "VFaceEq"
	case VFaceAnd:
		return "VFaceAnd"
	case VFaceOr:
		return "VFaceOr"
	case VPartial:
		return "VPartial"
	case VSystem:
		return "VSystem"
	case VComp:
		return "VComp"
	case VHComp:
		return "VHComp"
	case VFill:
		return "VFill"
	case VGlue:
		return "VGlue"
	case VGlueElem:
		return "VGlueElem"
	case VUnglue:
		return "VUnglue"
	case VUA:
		return "VUA"
	case VUABeta:
		return "VUABeta"
	case VHITPathCtor:
		return "VHITPathCtor"
	default:
		return "Unknown"
	}
}
