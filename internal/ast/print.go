package ast

import (
	"bytes"
	"strconv"
)

// Print a compact S-expression-like string for debugging.
func Sprint(t Term) string {
	var b bytes.Buffer
	write(&b, t, 0)
	return b.String()
}

func collectSpine(t Term) (fun Term, args []Term) {
	for {
		if app, ok := t.(App); ok {
			args = append([]Term{app.U}, args...)
			t = app.T
		} else {
			fun = t
			break
		}
	}
	return
}

func write(b *bytes.Buffer, t Term, _ int) {
	switch t := t.(type) {
	case Sort:
		b.WriteString("Type")
		b.WriteString(strconv.FormatUint(uint64(t.U), 10))
	case Var:
		b.WriteString("{")
		b.WriteString(strconv.Itoa(t.Ix))
		b.WriteString("}")
	case Global:
		b.WriteString(t.Name)
	case Pi:
		// (Pi x:A . B)
		b.WriteString("(Pi ")
		if t.Binder != "" {
			b.WriteString(t.Binder)
		} else {
			b.WriteString("_")
		}
		b.WriteString(": ")
		write(b, t.A, 0)
		b.WriteString(" . ")
		write(b, t.B, 0)
		b.WriteString(")")
	case Lam:
		// (\x [:Ann] => Body)
		b.WriteString("(\\")
		if t.Binder != "" {
			b.WriteString(t.Binder)
		} else {
			b.WriteString("_")
		}
		if t.Ann != nil {
			b.WriteString(" : ")
			write(b, t.Ann, 0)
		}
		b.WriteString(" => ")
		write(b, t.Body, 0)
		b.WriteString(")")
	case App:
		fun, args := collectSpine(t)
		b.WriteString("(")
		write(b, fun, 0)
		for _, arg := range args {
			b.WriteString(" ")
			write(b, arg, 0)
		}
		b.WriteString(")")
	case Sigma:
		b.WriteString("(Sigma ")
		if t.Binder != "" {
			b.WriteString(t.Binder)
		} else {
			b.WriteString("_")
		}
		b.WriteString(": ")
		write(b, t.A, 0)
		b.WriteString(" . ")
		write(b, t.B, 0)
		b.WriteString(")")
	case Pair:
		b.WriteString("(")
		write(b, t.Fst, 0)
		b.WriteString(" , ")
		write(b, t.Snd, 0)
		b.WriteString(")")
	case Fst:
		b.WriteString("(fst ")
		write(b, t.P, 0)
		b.WriteString(")")
	case Snd:
		b.WriteString("(snd ")
		write(b, t.P, 0)
		b.WriteString(")")
	case Let:
		b.WriteString("(let ")
		if t.Binder != "" {
			b.WriteString(t.Binder)
		} else {
			b.WriteString("_")
		}
		if t.Ann != nil {
			b.WriteString(" : ")
			write(b, t.Ann, 0)
		}
		b.WriteString(" = ")
		write(b, t.Val, 0)
		b.WriteString(" in ")
		write(b, t.Body, 0)
		b.WriteString(")")
	default:
		b.WriteString("<?>")
	}
}
