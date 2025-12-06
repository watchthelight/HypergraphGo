//go:build cubical

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

	default:
		return false
	}
}
