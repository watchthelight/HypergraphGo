// Package eval implements Normalization by Evaluation (NbE) for the HoTT kernel.
//
// NbE is a technique for normalizing lambda calculus terms by:
//  1. Evaluating syntax into a semantic domain (Values)
//  2. Reifying Values back to syntax in normal form
//
// The semantic domain uses closures for lazy evaluation under binders,
// and neutral terms to represent stuck computations (e.g., variable applications).
//
// Key concepts:
//   - Value: semantic representation (VLam, VPi, VSigma, VPair, VSort, VNeutral)
//   - Closure: captures environment + term for delayed evaluation
//   - Neutral: stuck computation with head variable and argument spine
//   - Env: de Bruijn environment mapping indices to Values
//
// References:
//   - Abel, A. "Normalization by Evaluation: Dependent Types and Impredicativity"
//   - Coquand, T. "An Algorithm for Type-Checking Dependent Types"
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

// VId represents identity type values.
type VId struct {
	A Value
	X Value
	Y Value
}

func (VId) isValue() {}

// VRefl represents reflexivity proof values.
type VRefl struct {
	A Value
	X Value
}

func (VRefl) isValue() {}

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

// Eval evaluates a term in an environment to weak head normal form (WHNF).
//
// Evaluation proceeds recursively, handling each term constructor:
//   - Var: lookup in environment, return neutral if unbound
//   - Lam: create closure capturing current environment
//   - App: evaluate function and argument, then apply
//   - Pi/Sigma: evaluate domain, create closure for codomain
//   - Pair: evaluate both components
//   - Fst/Snd: evaluate pair and project
//   - Let: evaluate definition, extend environment, evaluate body
//   - Sort/Global: convert directly to corresponding Value
//
// If env is nil, an empty environment is used.
// If t is nil, returns VGlobal{"nil"} as a fallback.
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

	case ast.Id:
		a := Eval(env, tm.A)
		x := Eval(env, tm.X)
		y := Eval(env, tm.Y)
		return VId{A: a, X: x, Y: y}

	case ast.Refl:
		a := Eval(env, tm.A)
		x := Eval(env, tm.X)
		return VRefl{A: a, X: x}

	case ast.J:
		a := Eval(env, tm.A)
		c := Eval(env, tm.C)
		d := Eval(env, tm.D)
		x := Eval(env, tm.X)
		y := Eval(env, tm.Y)
		p := Eval(env, tm.P)
		return evalJ(a, c, d, x, y, p)

	default:
		// Try extension evaluators (e.g., cubical terms when built with -tags cubical)
		if val, ok := tryEvalCubical(env, t); ok {
			return val
		}
		// Fallback for unknown terms
		return VGlobal{Name: "unknown"}
	}
}

// Apply performs function application in the semantic domain.
//
// This is the key operation for beta reduction in NbE:
//   - VLam: performs beta reduction by extending the closure's environment
//     with the argument and evaluating the body
//   - VNeutral: extends the spine with the new argument (stuck application)
//   - Other: creates a "bad_app" neutral term (type error in well-typed terms)
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

// evalJ handles J elimination (path induction).
// The computation rule is: J A C d x x (refl A x) --> d
//
// Note: We only check if p is VRefl, not that x == y. This is correct because:
// - The type checker ensures p : Id A x y
// - If p is refl A z, then p : Id A z z
// - For well-typed terms, this means x == z == y
// - NbE assumes well-typed input (standard practice in type theory implementations)
func evalJ(a, c, d, x, y, p Value) Value {
	// Check if p is refl - this triggers the computation rule
	// For well-typed input, VRefl implies x == y by typing invariant
	if _, ok := p.(VRefl); ok {
		// J A C d x x (refl A x) --> d
		return d
	}
	// Stuck: return neutral J application
	head := Head{Glob: "J"}
	return VNeutral{N: Neutral{Head: head, Sp: []Value{a, c, d, x, y, p}}}
}

// Reify converts a Value back to an ast.Term in normal form.
//
// This is the "read back" phase of NbE that extracts syntax from semantics:
//   - VLam: applies to a fresh variable, reifies the result under a binder
//   - VPi/VSigma: reifies domain, applies closure to fresh var for codomain
//   - VPair: reifies both components
//   - VNeutral: reconstructs term from head and spine
//   - VSort/VGlobal: direct conversion
//
// The result is in beta-normal form (no reducible beta redexes).
// Uses level-indexed fresh variables for correct de Bruijn handling under binders.
func Reify(v Value) ast.Term {
	return reifyAt(0, v)
}

// reifyAt reifies a value at a given binding level.
// The level tracks how many binders we've gone under during reification.
// Fresh variables use the level as their index, and neutral variables are
// converted from levels to de Bruijn indices: index = level - varLevel - 1.
func reifyAt(level int, v Value) ast.Term {
	switch val := v.(type) {
	case VNeutral:
		return reifyNeutralAt(level, val.N)

	case VLam:
		// Create a fresh variable at current level and reify body at level+1
		freshVar := vVar(level)
		bodyVal := Apply(val, freshVar)
		bodyTerm := reifyAt(level+1, bodyVal)
		return ast.Lam{Binder: "_", Body: bodyTerm}

	case VPair:
		fst := reifyAt(level, val.Fst)
		snd := reifyAt(level, val.Snd)
		return ast.Pair{Fst: fst, Snd: snd}

	case VSort:
		return ast.Sort{U: ast.Level(val.Level)}

	case VGlobal:
		return ast.Global{Name: val.Name}

	case VPi:
		a := reifyAt(level, val.A)
		// For Pi types, apply closure to fresh var and reify at level+1
		freshVar := vVar(level)
		bVal := Apply(VLam{Body: val.B}, freshVar)
		b := reifyAt(level+1, bVal)
		return ast.Pi{Binder: "_", A: a, B: b}

	case VSigma:
		a := reifyAt(level, val.A)
		// For Sigma types, apply closure to fresh var and reify at level+1
		freshVar := vVar(level)
		bVal := Apply(VLam{Body: val.B}, freshVar)
		b := reifyAt(level+1, bVal)
		return ast.Sigma{Binder: "_", A: a, B: b}

	case VId:
		a := reifyAt(level, val.A)
		x := reifyAt(level, val.X)
		y := reifyAt(level, val.Y)
		return ast.Id{A: a, X: x, Y: y}

	case VRefl:
		a := reifyAt(level, val.A)
		x := reifyAt(level, val.X)
		return ast.Refl{A: a, X: x}

	default:
		// Try extension reifiers (e.g., cubical values when built with -tags cubical)
		if term, ok := tryReifyCubical(level, v); ok {
			return term
		}
		return ast.Global{Name: "reify_error"}
	}
}

// reifyNeutralAt converts a Neutral back to an ast.Term at a given level.
// For variables, converts from level-indexed to de Bruijn: index = level - varLevel - 1.
func reifyNeutralAt(level int, n Neutral) ast.Term {
	var head ast.Term

	if n.Head.Glob == "" {
		// Variable: convert from level to de Bruijn index
		// A variable created at level L, when reified at level M, becomes index M-L-1
		ix := level - n.Head.Var - 1
		if ix < 0 {
			// Free variable (created before reification started)
			// Keep original index as fallback
			ix = n.Head.Var
		}
		head = ast.Var{Ix: ix}
	} else {
		switch n.Head.Glob {
		case "fst":
			if len(n.Sp) >= 1 {
				// First arg is the pair being projected
				arg := reifyAt(level, n.Sp[0])
				base := ast.Fst{P: arg}
				// Apply remaining spine arguments (if projection result is a function)
				var result ast.Term = base
				for _, spArg := range n.Sp[1:] {
					argTerm := reifyAt(level, spArg)
					result = ast.App{T: result, U: argTerm}
				}
				return result
			}
			head = ast.Global{Name: n.Head.Glob}
		case "snd":
			if len(n.Sp) >= 1 {
				// First arg is the pair being projected
				arg := reifyAt(level, n.Sp[0])
				base := ast.Snd{P: arg}
				// Apply remaining spine arguments (if projection result is a function)
				var result ast.Term = base
				for _, spArg := range n.Sp[1:] {
					argTerm := reifyAt(level, spArg)
					result = ast.App{T: result, U: argTerm}
				}
				return result
			}
			head = ast.Global{Name: n.Head.Glob}
		case "J":
			// Stuck J: spine is [a, c, d, x, y, p]
			if len(n.Sp) >= 6 {
				a := reifyAt(level, n.Sp[0])
				c := reifyAt(level, n.Sp[1])
				d := reifyAt(level, n.Sp[2])
				x := reifyAt(level, n.Sp[3])
				y := reifyAt(level, n.Sp[4])
				p := reifyAt(level, n.Sp[5])
				base := ast.J{A: a, C: c, D: d, X: x, Y: y, P: p}
				// Handle any additional spine arguments (if J result is applied)
				var result ast.Term = base
				for _, spArg := range n.Sp[6:] {
					argTerm := reifyAt(level, spArg)
					result = ast.App{T: result, U: argTerm}
				}
				return result
			}
			head = ast.Global{Name: n.Head.Glob}
		default:
			head = ast.Global{Name: n.Head.Glob}
		}
	}

	// Apply spine arguments
	result := head
	for _, arg := range n.Sp {
		argTerm := reifyAt(level, arg)
		result = ast.App{T: result, U: argTerm}
	}

	return result
}

// EvalNBE is a convenience function that evaluates and reifies a term using NbE.
func EvalNBE(t ast.Term) ast.Term {
	env := &Env{Bindings: nil}
	val := Eval(env, t)
	return Reify(val)
}
