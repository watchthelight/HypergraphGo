package ast

import (
	"bytes"
	"strconv"
)

// tryWriteExtension is the cubical implementation that handles cubical terms.
// Returns true if the term was a cubical term and was handled.
func tryWriteExtension(b *bytes.Buffer, t Term) bool {
	return writeCubical(b, t)
}

// writeCubical writes a cubical term to the buffer.
// Returns true if the term was a cubical term and was handled.
func writeCubical(b *bytes.Buffer, t Term) bool {
	switch t := t.(type) {
	case Interval:
		b.WriteString("I")
		return true

	case I0:
		b.WriteString("i0")
		return true

	case I1:
		b.WriteString("i1")
		return true

	case IVar:
		b.WriteString("i{")
		b.WriteString(strconv.Itoa(t.Ix))
		b.WriteString("}")
		return true

	case Path:
		b.WriteString("(Path ")
		write(b, t.A)
		b.WriteString(" ")
		write(b, t.X)
		b.WriteString(" ")
		write(b, t.Y)
		b.WriteString(")")
		return true

	case PathP:
		b.WriteString("(PathP ")
		write(b, t.A)
		b.WriteString(" ")
		write(b, t.X)
		b.WriteString(" ")
		write(b, t.Y)
		b.WriteString(")")
		return true

	case PathLam:
		b.WriteString("(<")
		if t.Binder != "" {
			b.WriteString(t.Binder)
		} else {
			b.WriteString("_")
		}
		b.WriteString("> ")
		write(b, t.Body)
		b.WriteString(")")
		return true

	case PathApp:
		b.WriteString("(")
		write(b, t.P)
		b.WriteString(" @ ")
		write(b, t.R)
		b.WriteString(")")
		return true

	case Transport:
		b.WriteString("(transport ")
		write(b, t.A)
		b.WriteString(" ")
		write(b, t.E)
		b.WriteString(")")
		return true

	// --- Face Formulas ---

	case FaceTop:
		b.WriteString("⊤")
		return true

	case FaceBot:
		b.WriteString("⊥")
		return true

	case FaceEq:
		b.WriteString("(i{")
		b.WriteString(strconv.Itoa(t.IVar))
		b.WriteString("} = ")
		if t.IsOne {
			b.WriteString("1")
		} else {
			b.WriteString("0")
		}
		b.WriteString(")")
		return true

	case FaceAnd:
		b.WriteString("(")
		writeFace(b, t.Left)
		b.WriteString(" ∧ ")
		writeFace(b, t.Right)
		b.WriteString(")")
		return true

	case FaceOr:
		b.WriteString("(")
		writeFace(b, t.Left)
		b.WriteString(" ∨ ")
		writeFace(b, t.Right)
		b.WriteString(")")
		return true

	// --- Partial Types and Systems ---

	case Partial:
		b.WriteString("(Partial ")
		writeFace(b, t.Phi)
		b.WriteString(" ")
		write(b, t.A)
		b.WriteString(")")
		return true

	case System:
		b.WriteString("[")
		for i, branch := range t.Branches {
			if i > 0 {
				b.WriteString(", ")
			}
			writeFace(b, branch.Phi)
			b.WriteString(" ↦ ")
			write(b, branch.Term)
		}
		b.WriteString("]")
		return true

	// --- Composition Operations ---

	case Comp:
		b.WriteString("(comp")
		if t.IBinder != "" {
			b.WriteString("^")
			b.WriteString(t.IBinder)
		}
		b.WriteString(" ")
		write(b, t.A)
		b.WriteString(" [")
		writeFace(b, t.Phi)
		b.WriteString(" ↦ ")
		write(b, t.Tube)
		b.WriteString("] ")
		write(b, t.Base)
		b.WriteString(")")
		return true

	case HComp:
		b.WriteString("(hcomp ")
		write(b, t.A)
		b.WriteString(" [")
		writeFace(b, t.Phi)
		b.WriteString(" ↦ ")
		write(b, t.Tube)
		b.WriteString("] ")
		write(b, t.Base)
		b.WriteString(")")
		return true

	case Fill:
		b.WriteString("(fill")
		if t.IBinder != "" {
			b.WriteString("^")
			b.WriteString(t.IBinder)
		}
		b.WriteString(" ")
		write(b, t.A)
		b.WriteString(" [")
		writeFace(b, t.Phi)
		b.WriteString(" ↦ ")
		write(b, t.Tube)
		b.WriteString("] ")
		write(b, t.Base)
		b.WriteString(")")
		return true

	// --- Glue Types ---

	case Glue:
		b.WriteString("(Glue ")
		write(b, t.A)
		b.WriteString(" [")
		for i, branch := range t.System {
			if i > 0 {
				b.WriteString(", ")
			}
			writeFace(b, branch.Phi)
			b.WriteString(" ↦ (")
			write(b, branch.T)
			b.WriteString(", ")
			write(b, branch.Equiv)
			b.WriteString(")")
		}
		b.WriteString("])")
		return true

	case GlueElem:
		b.WriteString("(glue [")
		for i, branch := range t.System {
			if i > 0 {
				b.WriteString(", ")
			}
			writeFace(b, branch.Phi)
			b.WriteString(" ↦ ")
			write(b, branch.Term)
		}
		b.WriteString("] ")
		write(b, t.Base)
		b.WriteString(")")
		return true

	case Unglue:
		b.WriteString("(unglue ")
		write(b, t.G)
		b.WriteString(")")
		return true

	// --- Univalence ---

	case UA:
		b.WriteString("(ua ")
		write(b, t.A)
		b.WriteString(" ")
		write(b, t.B)
		b.WriteString(" ")
		write(b, t.Equiv)
		b.WriteString(")")
		return true

	case UABeta:
		b.WriteString("(ua-β ")
		write(b, t.Equiv)
		b.WriteString(" ")
		write(b, t.Arg)
		b.WriteString(")")
		return true

	default:
		return false
	}
}

// writeFace writes a face formula to the buffer.
func writeFace(b *bytes.Buffer, f Face) {
	if f == nil {
		b.WriteString("⊥")
		return
	}
	switch t := f.(type) {
	case FaceTop:
		b.WriteString("⊤")
	case FaceBot:
		b.WriteString("⊥")
	case FaceEq:
		b.WriteString("(i{")
		b.WriteString(strconv.Itoa(t.IVar))
		b.WriteString("} = ")
		if t.IsOne {
			b.WriteString("1")
		} else {
			b.WriteString("0")
		}
		b.WriteString(")")
	case FaceAnd:
		b.WriteString("(")
		writeFace(b, t.Left)
		b.WriteString(" ∧ ")
		writeFace(b, t.Right)
		b.WriteString(")")
	case FaceOr:
		b.WriteString("(")
		writeFace(b, t.Left)
		b.WriteString(" ∨ ")
		writeFace(b, t.Right)
		b.WriteString(")")
	default:
		b.WriteString("?face")
	}
}
