package ast

import (
	"bytes"
	"strconv"
)

// Sprint returns a compact S-expression-like string for debugging.
func Sprint(t Term) string {
	var b bytes.Buffer
	write(&b, t)
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

func write(b *bytes.Buffer, t Term) {
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
		write(b, t.A)
		b.WriteString(" . ")
		write(b, t.B)
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
			write(b, t.Ann)
		}
		b.WriteString(" => ")
		write(b, t.Body)
		b.WriteString(")")
	case App:
		fun, args := collectSpine(t)
		b.WriteString("(")
		write(b, fun)
		for _, arg := range args {
			b.WriteString(" ")
			write(b, arg)
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
		write(b, t.A)
		b.WriteString(" . ")
		write(b, t.B)
		b.WriteString(")")
	case Pair:
		b.WriteString("(")
		write(b, t.Fst)
		b.WriteString(" , ")
		write(b, t.Snd)
		b.WriteString(")")
	case Fst:
		b.WriteString("(fst ")
		write(b, t.P)
		b.WriteString(")")
	case Snd:
		b.WriteString("(snd ")
		write(b, t.P)
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
			write(b, t.Ann)
		}
		b.WriteString(" = ")
		write(b, t.Val)
		b.WriteString(" in ")
		write(b, t.Body)
		b.WriteString(")")
	default:
		b.WriteString("<?>")
	}
}
