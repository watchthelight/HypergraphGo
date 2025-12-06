//go:build cubical

// Package eval provides NbE for the HoTT kernel.
// This file contains cubical type theory extensions (gated by build tag).
package eval

import "github.com/watchthelight/HypergraphGo/internal/ast"

// --- Interval Values ---

// VI0 represents the left endpoint of the interval (i0).
type VI0 struct{}

func (VI0) isValue() {}

// VI1 represents the right endpoint of the interval (i1).
type VI1 struct{}

func (VI1) isValue() {}

// VIVar represents a neutral interval variable.
// Level is used for level-indexed reification (similar to term variables).
type VIVar struct{ Level int }

func (VIVar) isValue() {}

// --- Path Type Values ---

// VPath represents non-dependent path type values: Path A x y.
type VPath struct {
	A Value // Type
	X Value // Left endpoint
	Y Value // Right endpoint
}

func (VPath) isValue() {}

// VPathP represents dependent path type values: PathP A x y.
type VPathP struct {
	A *IClosure // Type family: I → Type
	X Value     // Left endpoint
	Y Value     // Right endpoint
}

func (VPathP) isValue() {}

// VPathLam represents path abstraction values: <i> t.
type VPathLam struct {
	Body *IClosure // Body with interval variable bound
}

func (VPathLam) isValue() {}

// VTransport represents stuck transport values.
// Used when A is not constant and reduction cannot proceed.
type VTransport struct {
	A *IClosure // Type family: I → Type
	E Value     // Element at i0
}

func (VTransport) isValue() {}

// --- Interval Closure and Environment ---

// IClosure captures both term and interval environments with a term.
// Used for constructs that bind an interval variable.
type IClosure struct {
	Env  *Env     // Term environment
	IEnv *IEnv    // Interval environment
	Term ast.Term // Body term
}

// IEnv represents an interval evaluation environment.
// Maps interval de Bruijn indices to interval values (VI0, VI1, VIVar).
type IEnv struct {
	Bindings []Value // Must be VI0, VI1, or VIVar
}

// EmptyIEnv returns an empty interval environment.
func EmptyIEnv() *IEnv {
	return &IEnv{Bindings: nil}
}

// Extend adds a new binding to the front of the interval environment.
func (ie *IEnv) Extend(v Value) *IEnv {
	if ie == nil {
		return &IEnv{Bindings: []Value{v}}
	}
	nb := make([]Value, 0, len(ie.Bindings)+1)
	nb = append(nb, v)
	nb = append(nb, ie.Bindings...)
	return &IEnv{Bindings: nb}
}

// Lookup retrieves an interval value by de Bruijn index.
func (ie *IEnv) Lookup(ix int) Value {
	if ie == nil || ix < 0 || ix >= len(ie.Bindings) {
		// Return neutral interval variable if out of bounds
		return VIVar{Level: ix}
	}
	return ie.Bindings[ix]
}

// ILen returns the number of bindings in the interval environment.
func (ie *IEnv) ILen() int {
	if ie == nil {
		return 0
	}
	return len(ie.Bindings)
}

// --- Cubical Evaluation ---

// EvalCubical evaluates a term with both term and interval environments.
// This is the main entry point for cubical evaluation.
func EvalCubical(env *Env, ienv *IEnv, t ast.Term) Value {
	if t == nil {
		return VGlobal{Name: "nil"}
	}
	if env == nil {
		env = &Env{Bindings: nil}
	}
	if ienv == nil {
		ienv = EmptyIEnv()
	}

	switch tm := t.(type) {
	// Standard terms - delegate to regular Eval logic but track ienv
	case ast.Var:
		return env.Lookup(tm.Ix)

	case ast.Global:
		return vGlobal(tm.Name)

	case ast.Sort:
		return VSort{Level: int(tm.U)}

	case ast.Lam:
		// Term lambda - still uses term environment
		return VLam{Body: &Closure{Env: env, Term: tm.Body}}

	case ast.App:
		fun := EvalCubical(env, ienv, tm.T)
		arg := EvalCubical(env, ienv, tm.U)
		return Apply(fun, arg)

	case ast.Pair:
		fst := EvalCubical(env, ienv, tm.Fst)
		snd := EvalCubical(env, ienv, tm.Snd)
		return VPair{Fst: fst, Snd: snd}

	case ast.Fst:
		p := EvalCubical(env, ienv, tm.P)
		return Fst(p)

	case ast.Snd:
		p := EvalCubical(env, ienv, tm.P)
		return Snd(p)

	case ast.Pi:
		a := EvalCubical(env, ienv, tm.A)
		return VPi{A: a, B: &Closure{Env: env, Term: tm.B}}

	case ast.Sigma:
		a := EvalCubical(env, ienv, tm.A)
		return VSigma{A: a, B: &Closure{Env: env, Term: tm.B}}

	case ast.Let:
		val := EvalCubical(env, ienv, tm.Val)
		newEnv := env.Extend(val)
		return EvalCubical(newEnv, ienv, tm.Body)

	case ast.Id:
		a := EvalCubical(env, ienv, tm.A)
		x := EvalCubical(env, ienv, tm.X)
		y := EvalCubical(env, ienv, tm.Y)
		return VId{A: a, X: x, Y: y}

	case ast.Refl:
		a := EvalCubical(env, ienv, tm.A)
		x := EvalCubical(env, ienv, tm.X)
		return VRefl{A: a, X: x}

	case ast.J:
		a := EvalCubical(env, ienv, tm.A)
		c := EvalCubical(env, ienv, tm.C)
		d := EvalCubical(env, ienv, tm.D)
		x := EvalCubical(env, ienv, tm.X)
		y := EvalCubical(env, ienv, tm.Y)
		p := EvalCubical(env, ienv, tm.P)
		return evalJ(a, c, d, x, y, p)

	// --- Cubical-specific terms ---

	case ast.Interval:
		// The interval type itself
		return VGlobal{Name: "I"}

	case ast.I0:
		return VI0{}

	case ast.I1:
		return VI1{}

	case ast.IVar:
		return ienv.Lookup(tm.Ix)

	case ast.Path:
		a := EvalCubical(env, ienv, tm.A)
		x := EvalCubical(env, ienv, tm.X)
		y := EvalCubical(env, ienv, tm.Y)
		return VPath{A: a, X: x, Y: y}

	case ast.PathP:
		// A binds an interval variable
		aClosure := &IClosure{Env: env, IEnv: ienv, Term: tm.A}
		x := EvalCubical(env, ienv, tm.X)
		y := EvalCubical(env, ienv, tm.Y)
		return VPathP{A: aClosure, X: x, Y: y}

	case ast.PathLam:
		// Creates an interval closure
		return VPathLam{Body: &IClosure{Env: env, IEnv: ienv, Term: tm.Body}}

	case ast.PathApp:
		p := EvalCubical(env, ienv, tm.P)
		r := EvalCubical(env, ienv, tm.R)
		return PathApply(p, r)

	case ast.Transport:
		// A binds an interval variable
		aClosure := &IClosure{Env: env, IEnv: ienv, Term: tm.A}
		e := EvalCubical(env, ienv, tm.E)
		return EvalTransport(aClosure, e)

	default:
		return VGlobal{Name: "unknown_cubical"}
	}
}

// PathApply applies a path to an interval argument.
// Implements the computation rules:
//   (<i> t) @ i0  -->  t[i0/i]
//   (<i> t) @ i1  -->  t[i1/i]
//   (<i> t) @ j   -->  t[j/i]
func PathApply(p Value, r Value) Value {
	switch pv := p.(type) {
	case VPathLam:
		// Beta reduction: evaluate body with interval substituted
		newIEnv := pv.Body.IEnv.Extend(r)
		return EvalCubical(pv.Body.Env, newIEnv, pv.Body.Term)

	case VNeutral:
		// Stuck path application
		head := Head{Glob: "@"}
		return VNeutral{N: Neutral{Head: head, Sp: []Value{p, r}}}

	default:
		// Not a path - stuck
		head := Head{Glob: "@"}
		return VNeutral{N: Neutral{Head: head, Sp: []Value{p, r}}}
	}
}

// EvalTransport evaluates transport (λi. A) e.
// The computation rule is: transport (λi. A) e --> e when A is constant in i.
func EvalTransport(aClosure *IClosure, e Value) Value {
	// Check if A is constant by evaluating at i0 and i1
	if isConstantFamily(aClosure) {
		// A doesn't depend on i, so transport is identity
		return e
	}
	// Cannot reduce - return stuck transport
	return VTransport{A: aClosure, E: e}
}

// isConstantFamily checks if an interval closure produces the same value at i0 and i1.
// This is used for the transport computation rule.
func isConstantFamily(c *IClosure) bool {
	v0 := EvalCubical(c.Env, c.IEnv.Extend(VI0{}), c.Term)
	v1 := EvalCubical(c.Env, c.IEnv.Extend(VI1{}), c.Term)
	// Use alpha equality on reified terms
	t0 := ReifyCubicalAt(0, 0, v0)
	t1 := ReifyCubicalAt(0, 0, v1)
	return alphaEqCubical(t0, t1)
}

// alphaEqCubical checks alpha-equality of cubical terms.
// For now, uses simple structural equality.
func alphaEqCubical(a, b ast.Term) bool {
	return ast.Sprint(a) == ast.Sprint(b)
}

// --- Cubical Reification ---

// ReifyCubicalAt reifies a value at given term and interval levels.
func ReifyCubicalAt(level int, ilevel int, v Value) ast.Term {
	switch val := v.(type) {
	case VNeutral:
		return reifyNeutralCubicalAt(level, ilevel, val.N)

	case VLam:
		freshVar := vVar(level)
		bodyVal := Apply(val, freshVar)
		bodyTerm := ReifyCubicalAt(level+1, ilevel, bodyVal)
		return ast.Lam{Binder: "_", Body: bodyTerm}

	case VPair:
		fst := ReifyCubicalAt(level, ilevel, val.Fst)
		snd := ReifyCubicalAt(level, ilevel, val.Snd)
		return ast.Pair{Fst: fst, Snd: snd}

	case VSort:
		return ast.Sort{U: ast.Level(val.Level)}

	case VGlobal:
		return ast.Global{Name: val.Name}

	case VPi:
		a := ReifyCubicalAt(level, ilevel, val.A)
		freshVar := vVar(level)
		bVal := Apply(VLam{Body: val.B}, freshVar)
		b := ReifyCubicalAt(level+1, ilevel, bVal)
		return ast.Pi{Binder: "_", A: a, B: b}

	case VSigma:
		a := ReifyCubicalAt(level, ilevel, val.A)
		freshVar := vVar(level)
		bVal := Apply(VLam{Body: val.B}, freshVar)
		b := ReifyCubicalAt(level+1, ilevel, bVal)
		return ast.Sigma{Binder: "_", A: a, B: b}

	case VId:
		a := ReifyCubicalAt(level, ilevel, val.A)
		x := ReifyCubicalAt(level, ilevel, val.X)
		y := ReifyCubicalAt(level, ilevel, val.Y)
		return ast.Id{A: a, X: x, Y: y}

	case VRefl:
		a := ReifyCubicalAt(level, ilevel, val.A)
		x := ReifyCubicalAt(level, ilevel, val.X)
		return ast.Refl{A: a, X: x}

	// --- Cubical values ---

	case VI0:
		return ast.I0{}

	case VI1:
		return ast.I1{}

	case VIVar:
		// Convert from level to de Bruijn index
		ix := ilevel - val.Level - 1
		if ix < 0 {
			ix = val.Level
		}
		return ast.IVar{Ix: ix}

	case VPath:
		a := ReifyCubicalAt(level, ilevel, val.A)
		x := ReifyCubicalAt(level, ilevel, val.X)
		y := ReifyCubicalAt(level, ilevel, val.Y)
		return ast.Path{A: a, X: x, Y: y}

	case VPathP:
		// Reify type family under interval binder
		freshIVar := VIVar{Level: ilevel}
		aVal := EvalCubical(val.A.Env, val.A.IEnv.Extend(freshIVar), val.A.Term)
		a := ReifyCubicalAt(level, ilevel+1, aVal)
		x := ReifyCubicalAt(level, ilevel, val.X)
		y := ReifyCubicalAt(level, ilevel, val.Y)
		return ast.PathP{A: a, X: x, Y: y}

	case VPathLam:
		// Reify body under interval binder
		freshIVar := VIVar{Level: ilevel}
		bodyVal := EvalCubical(val.Body.Env, val.Body.IEnv.Extend(freshIVar), val.Body.Term)
		body := ReifyCubicalAt(level, ilevel+1, bodyVal)
		return ast.PathLam{Binder: "_", Body: body}

	case VTransport:
		// Reify stuck transport
		freshIVar := VIVar{Level: ilevel}
		aVal := EvalCubical(val.A.Env, val.A.IEnv.Extend(freshIVar), val.A.Term)
		a := ReifyCubicalAt(level, ilevel+1, aVal)
		e := ReifyCubicalAt(level, ilevel, val.E)
		return ast.Transport{A: a, E: e}

	default:
		return ast.Global{Name: "reify_cubical_error"}
	}
}

// reifyNeutralCubicalAt converts a Neutral to an ast.Term with cubical support.
func reifyNeutralCubicalAt(level int, ilevel int, n Neutral) ast.Term {
	var head ast.Term

	if n.Head.Glob == "" {
		// Variable: convert from level to de Bruijn index
		ix := level - n.Head.Var - 1
		if ix < 0 {
			ix = n.Head.Var
		}
		head = ast.Var{Ix: ix}
	} else {
		switch n.Head.Glob {
		case "fst":
			if len(n.Sp) >= 1 {
				arg := ReifyCubicalAt(level, ilevel, n.Sp[0])
				base := ast.Fst{P: arg}
				var result ast.Term = base
				for _, spArg := range n.Sp[1:] {
					argTerm := ReifyCubicalAt(level, ilevel, spArg)
					result = ast.App{T: result, U: argTerm}
				}
				return result
			}
			head = ast.Global{Name: n.Head.Glob}
		case "snd":
			if len(n.Sp) >= 1 {
				arg := ReifyCubicalAt(level, ilevel, n.Sp[0])
				base := ast.Snd{P: arg}
				var result ast.Term = base
				for _, spArg := range n.Sp[1:] {
					argTerm := ReifyCubicalAt(level, ilevel, spArg)
					result = ast.App{T: result, U: argTerm}
				}
				return result
			}
			head = ast.Global{Name: n.Head.Glob}
		case "@":
			// Path application
			if len(n.Sp) >= 2 {
				p := ReifyCubicalAt(level, ilevel, n.Sp[0])
				r := ReifyCubicalAt(level, ilevel, n.Sp[1])
				base := ast.PathApp{P: p, R: r}
				var result ast.Term = base
				for _, spArg := range n.Sp[2:] {
					argTerm := ReifyCubicalAt(level, ilevel, spArg)
					result = ast.App{T: result, U: argTerm}
				}
				return result
			}
			head = ast.Global{Name: n.Head.Glob}
		default:
			head = ast.Global{Name: n.Head.Glob}
		}
	}

	result := head
	for _, arg := range n.Sp {
		argTerm := ReifyCubicalAt(level, ilevel, arg)
		result = ast.App{T: result, U: argTerm}
	}

	return result
}

// tryEvalCubical is the extension hook for Eval in cubical builds.
// Returns (value, true) if the term was handled, (nil, false) otherwise.
func tryEvalCubical(env *Env, t ast.Term) (Value, bool) {
	switch tm := t.(type) {
	case ast.Interval:
		return VGlobal{Name: "I"}, true
	case ast.I0:
		return VI0{}, true
	case ast.I1:
		return VI1{}, true
	case ast.IVar:
		// In non-cubical Eval context, treat as neutral
		return VIVar{Level: tm.Ix}, true
	case ast.Path:
		a := Eval(env, tm.A)
		x := Eval(env, tm.X)
		y := Eval(env, tm.Y)
		return VPath{A: a, X: x, Y: y}, true
	case ast.PathP:
		// Without ienv, create closure with empty interval env
		aClosure := &IClosure{Env: env, IEnv: EmptyIEnv(), Term: tm.A}
		x := Eval(env, tm.X)
		y := Eval(env, tm.Y)
		return VPathP{A: aClosure, X: x, Y: y}, true
	case ast.PathLam:
		return VPathLam{Body: &IClosure{Env: env, IEnv: EmptyIEnv(), Term: tm.Body}}, true
	case ast.PathApp:
		p := Eval(env, tm.P)
		r := Eval(env, tm.R)
		return PathApply(p, r), true
	case ast.Transport:
		aClosure := &IClosure{Env: env, IEnv: EmptyIEnv(), Term: tm.A}
		e := Eval(env, tm.E)
		return EvalTransport(aClosure, e), true
	default:
		return nil, false
	}
}

// tryReifyCubical is the extension hook for reifyAt in cubical builds.
// Returns (term, true) if the value was handled, (nil, false) otherwise.
func tryReifyCubical(level int, v Value) (ast.Term, bool) {
	switch val := v.(type) {
	case VI0:
		return ast.I0{}, true
	case VI1:
		return ast.I1{}, true
	case VIVar:
		// In non-cubical reify context, convert with ilevel=0
		ix := -val.Level - 1
		if ix < 0 {
			ix = val.Level
		}
		return ast.IVar{Ix: ix}, true
	case VPath:
		a := reifyAt(level, val.A)
		x := reifyAt(level, val.X)
		y := reifyAt(level, val.Y)
		return ast.Path{A: a, X: x, Y: y}, true
	case VPathP:
		// Reify with ilevel=0
		return ReifyCubicalAt(level, 0, val), true
	case VPathLam:
		return ReifyCubicalAt(level, 0, val), true
	case VTransport:
		return ReifyCubicalAt(level, 0, val), true
	default:
		return nil, false
	}
}
