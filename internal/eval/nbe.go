package eval

import "github.com/watchthelight/HypergraphGo/internal/ast"

// Value is the semantic domain for NbE.
type Value interface {
	isValue()
}

// VNeutral represents neutral terms (stuck computations).
type VNeutral struct{ N Neutral }

func (VNeutral) isValue() {}

// VLam represents lambda closures.
type VLam struct{ Body *Closure }

func (VLam) isValue() {}

// VPi represents Pi type closures (optional for now).
type VPi struct {
	A Value
	B *Closure
}

func (VPi) isValue() {}

// VSigma represents Sigma type closures (optional for now).
type VSigma struct {
	A Value
	B *Closure
}

func (VSigma) isValue() {}

// VPair represents pairs in WHNF.
type VPair struct{ Fst, Snd Value }

func (VPair) isValue() {}

// VSort represents universe sorts.
type VSort struct{ Level int }

func (VSort) isValue() {}

// VGlobal represents global constants.
type VGlobal struct{ Name string }

func (VGlobal) isValue() {}

// Closure captures an environment and a term for lazy evaluation.
type Closure struct {
	Env  *Env
	Term ast.Term
}

// Neutral represents stuck terms with a head and spine of arguments.
type Neutral struct {
	Head Head
	Sp   []Value
}

// Head represents the head of a neutral term.
type Head struct {
	Var  int    // de Bruijn index
	Glob string // global name
}

// Env represents an evaluation environment mapping de Bruijn indices to Values.
type Env struct {
	Bindings []Value
}

// Extend adds a new binding to the front of the environment.
func (e *Env) Extend(v Value) *Env {
	nb := make([]Value, 0, len(e.Bindings)+1)
	nb = append(nb, v)
	nb = append(nb, e.Bindings...)
	return &Env{Bindings: nb}
}

// Lookup retrieves a value by de Bruijn index.
func (e *Env) Lookup(ix int) Value {
	if ix < 0 || ix >= len(e.Bindings) {
		// Return neutral variable if out of bounds
		return vVar(ix)
	}
	return e.Bindings[ix]
}

// Helper constructors
func vVar(ix int) Value {
	return VNeutral{N: Neutral{Head: Head{Var: ix}}}
}

func vGlobal(name string) Value {
	return VNeutral{N: Neutral{Head: Head{Glob: name}}}
}

// Eval evaluates a term in an environment to weak head normal form.
func Eval(env *Env, t ast.Term) Value {
	if t == nil {
		return VGlobal{Name: "nil"} // fallback for nil terms
	}
	if env == nil {
		env = &Env{Bindings: nil} // use empty environment if nil
	}

	switch tm := t.(type) {
	case ast.Var:
		return env.Lookup(tm.Ix)

	case ast.Global:
		return vGlobal(tm.Name)

	case ast.Sort:
		return VSort{Level: int(tm.U)}

	case ast.Lam:
		return VLam{Body: &Closure{Env: env, Term: tm.Body}}

	case ast.App:
		fun := Eval(env, tm.T)
		arg := Eval(env, tm.U)
		return Apply(fun, arg)

	case ast.Pair:
		fst := Eval(env, tm.Fst)
		snd := Eval(env, tm.Snd)
		return VPair{Fst: fst, Snd: snd}

	case ast.Fst:
		p := Eval(env, tm.P)
		return Fst(p)

	case ast.Snd:
		p := Eval(env, tm.P)
		return Snd(p)

	case ast.Pi:
		a := Eval(env, tm.A)
		return VPi{A: a, B: &Closure{Env: env, Term: tm.B}}

	case ast.Sigma:
		a := Eval(env, tm.A)
		return VSigma{A: a, B: &Closure{Env: env, Term: tm.B}}

	case ast.Let:
		val := Eval(env, tm.Val)
		newEnv := env.Extend(val)
		return Eval(newEnv, tm.Body)

	default:
		// Fallback for unknown terms
		return VGlobal{Name: "unknown"}
	}
}

// Apply performs function application in the semantic domain.
func Apply(fun Value, arg Value) Value {
	switch f := fun.(type) {
	case VLam:
		// Beta reduction: evaluate body in extended environment
		newEnv := f.Body.Env.Extend(arg)
		return Eval(newEnv, f.Body.Term)

	case VNeutral:
		// Extend the spine of the neutral term
		newSp := make([]Value, len(f.N.Sp)+1)
		copy(newSp, f.N.Sp)
		newSp[len(f.N.Sp)] = arg
		return VNeutral{N: Neutral{Head: f.N.Head, Sp: newSp}}

	default:
		// Non-function applied to argument becomes neutral
		// This shouldn't happen in well-typed terms, but we handle it gracefully
		head := Head{Glob: "bad_app"}
		return VNeutral{N: Neutral{Head: head, Sp: []Value{fun, arg}}}
	}
}

// Fst performs first projection in the semantic domain.
func Fst(v Value) Value {
	switch val := v.(type) {
	case VPair:
		return val.Fst

	case VNeutral:
		// Create a neutral fst projection
		head := Head{Glob: "fst"}
		return VNeutral{N: Neutral{Head: head, Sp: []Value{v}}}

	default:
		// Non-pair projected becomes neutral
		head := Head{Glob: "fst"}
		return VNeutral{N: Neutral{Head: head, Sp: []Value{v}}}
	}
}

// Snd performs second projection in the semantic domain.
func Snd(v Value) Value {
	switch val := v.(type) {
	case VPair:
		return val.Snd

	case VNeutral:
		// Create a neutral snd projection
		head := Head{Glob: "snd"}
		return VNeutral{N: Neutral{Head: head, Sp: []Value{v}}}

	default:
		// Non-pair projected becomes neutral
		head := Head{Glob: "snd"}
		return VNeutral{N: Neutral{Head: head, Sp: []Value{v}}}
	}
}

// Reify converts a Value back to an ast.Term.
func Reify(v Value) ast.Term {
	switch val := v.(type) {
	case VNeutral:
		return reifyNeutral(val.N)

	case VLam:
		// Create a fresh variable and reify the body
		freshVar := vVar(0)
		bodyVal := Apply(val, freshVar)
		bodyTerm := Reify(bodyVal)
		return ast.Lam{Binder: "_", Body: bodyTerm}

	case VPair:
		fst := Reify(val.Fst)
		snd := Reify(val.Snd)
		return ast.Pair{Fst: fst, Snd: snd}

	case VSort:
		return ast.Sort{U: ast.Level(val.Level)}

	case VGlobal:
		return ast.Global{Name: val.Name}

	case VPi:
		a := Reify(val.A)
		// For Pi types, we need to reify under a binder
		freshVar := vVar(0)
		bVal := Apply(VLam{Body: val.B}, freshVar)
		b := Reify(bVal)
		return ast.Pi{Binder: "_", A: a, B: b}

	case VSigma:
		a := Reify(val.A)
		// For Sigma types, we need to reify under a binder
		freshVar := vVar(0)
		bVal := Apply(VLam{Body: val.B}, freshVar)
		b := Reify(bVal)
		return ast.Sigma{Binder: "_", A: a, B: b}

	default:
		return ast.Global{Name: "reify_error"}
	}
}

// reifyNeutral converts a Neutral back to an ast.Term.
func reifyNeutral(n Neutral) ast.Term {
	var head ast.Term

	if n.Head.Var >= 0 && n.Head.Glob == "" {
		head = ast.Var{Ix: n.Head.Var}
	} else if n.Head.Glob != "" {
		switch n.Head.Glob {
		case "fst":
			if len(n.Sp) == 1 {
				arg := Reify(n.Sp[0])
				return ast.Fst{P: arg}
			}
			head = ast.Global{Name: n.Head.Glob}
		case "snd":
			if len(n.Sp) == 1 {
				arg := Reify(n.Sp[0])
				return ast.Snd{P: arg}
			}
			head = ast.Global{Name: n.Head.Glob}
		default:
			head = ast.Global{Name: n.Head.Glob}
		}
	} else {
		head = ast.Global{Name: "neutral_error"}
	}

	// Apply spine arguments
	result := head
	for _, arg := range n.Sp {
		argTerm := Reify(arg)
		result = ast.App{T: result, U: argTerm}
	}

	return result
}

// Reflect converts a Neutral to a Value (identity function, but useful for API completeness).
func Reflect(neu Neutral) Value {
	return VNeutral{N: neu}
}

// EvalNBE is a convenience function that evaluates and reifies a term using NbE.
func EvalNBE(t ast.Term) ast.Term {
	env := &Env{Bindings: nil}
	val := Eval(env, t)
	return Reify(val)
}
