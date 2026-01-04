// Package eval provides NbE for the HoTT kernel.
// This file contains cubical type theory extensions.
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

// --- Face Formula Values ---

// VFaceTop represents the always-true face formula (⊤).
type VFaceTop struct{}

func (VFaceTop) isValue()     {}
func (VFaceTop) isFaceValue() {}

// VFaceBot represents the always-false face formula (⊥).
type VFaceBot struct{}

func (VFaceBot) isValue()     {}
func (VFaceBot) isFaceValue() {}

// VFaceEq represents an endpoint constraint: (i = 0) or (i = 1).
type VFaceEq struct {
	ILevel int  // Interval variable level (for reification)
	IsOne  bool // true for (i = 1), false for (i = 0)
}

func (VFaceEq) isValue()     {}
func (VFaceEq) isFaceValue() {}

// VFaceAnd represents conjunction of faces: φ ∧ ψ.
type VFaceAnd struct {
	Left  FaceValue
	Right FaceValue
}

func (VFaceAnd) isValue()     {}
func (VFaceAnd) isFaceValue() {}

// VFaceOr represents disjunction of faces: φ ∨ ψ.
type VFaceOr struct {
	Left  FaceValue
	Right FaceValue
}

func (VFaceOr) isValue()     {}
func (VFaceOr) isFaceValue() {}

// FaceValue is the interface for face formula values.
type FaceValue interface {
	Value
	isFaceValue()
}

// --- Partial Type Values ---

// VPartial represents a partial type value: Partial φ A.
type VPartial struct {
	Phi FaceValue // The face constraint
	A   Value     // The type
}

func (VPartial) isValue() {}

// VSystem represents a system of partial elements.
type VSystem struct {
	Branches []VSystemBranch
}

func (VSystem) isValue() {}

// VSystemBranch represents a single branch in a system value.
type VSystemBranch struct {
	Phi  FaceValue
	Term Value
}

// --- Composition Value Types ---

// VComp represents a stuck heterogeneous composition.
// Used when the face is not satisfied and reduction cannot proceed.
type VComp struct {
	A    *IClosure // Type line: I → Type
	Phi  FaceValue // Face constraint
	Tube *IClosure // Partial tube
	Base Value     // Base element
}

func (VComp) isValue() {}

// VHComp represents a stuck homogeneous composition.
type VHComp struct {
	A    Value     // Type (constant)
	Phi  FaceValue // Face constraint
	Tube *IClosure // Partial tube
	Base Value     // Base element
}

func (VHComp) isValue() {}

// VFill represents a stuck fill operation.
type VFill struct {
	A    *IClosure // Type line: I → Type
	Phi  FaceValue // Face constraint
	Tube *IClosure // Partial tube
	Base Value     // Base element
}

func (VFill) isValue() {}

// --- Glue Type Values ---

// VGlue represents a Glue type value: Glue A [φ ↦ (T, e)].
type VGlue struct {
	A      Value         // Base type
	System []VGlueBranch // System of equivalences
}

func (VGlue) isValue() {}

// VGlueBranch represents a branch in a VGlue value.
type VGlueBranch struct {
	Phi   FaceValue // Face constraint
	T     Value     // Fiber type
	Equiv Value     // Equivalence: Equiv T A
}

// VGlueElem represents a Glue element value: glue [φ ↦ t] a.
type VGlueElem struct {
	System []VGlueElemBranch // Partial element in fiber
	Base   Value             // Base element
}

func (VGlueElem) isValue() {}

// VGlueElemBranch represents a branch in a VGlueElem value.
type VGlueElemBranch struct {
	Phi  FaceValue
	Term Value
}

// VUnglue represents a stuck unglue operation.
type VUnglue struct {
	Ty Value // The Glue type
	G  Value // The Glue element
}

func (VUnglue) isValue() {}

// --- Univalence Values ---

// VUA represents ua e : Path Type A B.
// ua is defined as: ua e = <i> Glue B [(i=0) ↦ (A, e)]
type VUA struct {
	A     Value // Source type
	B     Value // Target type
	Equiv Value // Equivalence: Equiv A B
}

func (VUA) isValue() {}

// VUABeta represents the computation result: transport (ua e) a = e.fst a.
type VUABeta struct {
	Equiv Value // The equivalence e
	Arg   Value // The argument a : A
}

func (VUABeta) isValue() {}

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
		return evalError("nil term (cubical)")
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

	// --- Face Formulas ---

	case ast.FaceTop:
		return VFaceTop{}

	case ast.FaceBot:
		return VFaceBot{}

	case ast.FaceEq:
		// Look up the interval variable to see if it's resolved
		iVal := ienv.Lookup(tm.IVar)
		return evalFaceEq(iVal, tm.IVar, tm.IsOne, ienv.ILen())

	case ast.FaceAnd:
		left := evalFace(env, ienv, tm.Left)
		right := evalFace(env, ienv, tm.Right)
		return simplifyFaceAnd(left, right)

	case ast.FaceOr:
		left := evalFace(env, ienv, tm.Left)
		right := evalFace(env, ienv, tm.Right)
		return simplifyFaceOr(left, right)

	// --- Partial Types and Systems ---

	case ast.Partial:
		phi := evalFace(env, ienv, tm.Phi)
		a := EvalCubical(env, ienv, tm.A)
		return VPartial{Phi: phi, A: a}

	case ast.System:
		branches := make([]VSystemBranch, len(tm.Branches))
		for i, br := range tm.Branches {
			phi := evalFace(env, ienv, br.Phi)
			term := EvalCubical(env, ienv, br.Term)
			branches[i] = VSystemBranch{Phi: phi, Term: term}
		}
		return VSystem{Branches: branches}

	// --- Composition Operations ---

	case ast.Comp:
		// Evaluate the face constraint
		phi := evalFace(env, ienv, tm.Phi)
		// Create closures for A and Tube (they bind interval variables)
		aClosure := &IClosure{Env: env, IEnv: ienv, Term: tm.A}
		tubeClosure := &IClosure{Env: env, IEnv: ienv, Term: tm.Tube}
		base := EvalCubical(env, ienv, tm.Base)
		return EvalComp(aClosure, phi, tubeClosure, base)

	case ast.HComp:
		// Evaluate components
		a := EvalCubical(env, ienv, tm.A)
		phi := evalFace(env, ienv, tm.Phi)
		tubeClosure := &IClosure{Env: env, IEnv: ienv, Term: tm.Tube}
		base := EvalCubical(env, ienv, tm.Base)
		return EvalHComp(a, phi, tubeClosure, base)

	case ast.Fill:
		// Create closures for A and Tube
		phi := evalFace(env, ienv, tm.Phi)
		aClosure := &IClosure{Env: env, IEnv: ienv, Term: tm.A}
		tubeClosure := &IClosure{Env: env, IEnv: ienv, Term: tm.Tube}
		base := EvalCubical(env, ienv, tm.Base)
		return EvalFill(aClosure, phi, tubeClosure, base)

	// --- Glue Types ---

	case ast.Glue:
		a := EvalCubical(env, ienv, tm.A)
		branches := make([]VGlueBranch, len(tm.System))
		for i, br := range tm.System {
			phi := evalFace(env, ienv, br.Phi)
			t := EvalCubical(env, ienv, br.T)
			equiv := EvalCubical(env, ienv, br.Equiv)
			branches[i] = VGlueBranch{Phi: phi, T: t, Equiv: equiv}
		}
		return EvalGlue(a, branches)

	case ast.GlueElem:
		branches := make([]VGlueElemBranch, len(tm.System))
		for i, br := range tm.System {
			phi := evalFace(env, ienv, br.Phi)
			term := EvalCubical(env, ienv, br.Term)
			branches[i] = VGlueElemBranch{Phi: phi, Term: term}
		}
		base := EvalCubical(env, ienv, tm.Base)
		return EvalGlueElem(branches, base)

	case ast.Unglue:
		ty := EvalCubical(env, ienv, tm.Ty)
		g := EvalCubical(env, ienv, tm.G)
		return EvalUnglue(ty, g)

	// --- Univalence ---

	case ast.UA:
		a := EvalCubical(env, ienv, tm.A)
		b := EvalCubical(env, ienv, tm.B)
		equiv := EvalCubical(env, ienv, tm.Equiv)
		return EvalUA(a, b, equiv)

	case ast.UABeta:
		equiv := EvalCubical(env, ienv, tm.Equiv)
		arg := EvalCubical(env, ienv, tm.Arg)
		return EvalUABeta(equiv, arg)

	// --- Higher Inductive Types ---

	case ast.HITApp:
		// Evaluate all term arguments
		args := make([]Value, len(tm.Args))
		for i, arg := range tm.Args {
			args[i] = EvalCubical(env, ienv, arg)
		}
		// Evaluate all interval arguments
		iargs := make([]Value, len(tm.IArgs))
		for i, iarg := range tm.IArgs {
			iargs[i] = EvalCubical(env, ienv, iarg)
		}
		// Look up boundary info from recursor registry
		boundaries := lookupHITBoundaries(tm.HITName, tm.Ctor, args)
		return evalHITApp(tm.HITName, tm.Ctor, args, iargs, boundaries)

	default:
		return evalError("unknown cubical term type")
	}
}

// PathApply applies a path to an interval argument.
// Implements the computation rules:
//
//	(<i> t) @ i0  -->  t[i0/i]
//	(<i> t) @ i1  -->  t[i1/i]
//	(<i> t) @ j   -->  t[j/i]
//
// For PathP types (dependent path types), endpoint values are returned:
//
//	(PathP A x y) @ i0  -->  x
//	(PathP A x y) @ i1  -->  y
func PathApply(p Value, r Value) Value {
	switch pv := p.(type) {
	case VPathLam:
		// Beta reduction: evaluate body with interval substituted
		newIEnv := pv.Body.IEnv.Extend(r)
		return EvalCubical(pv.Body.Env, newIEnv, pv.Body.Term)

	case VPathP:
		// PathP applied to endpoint returns the corresponding endpoint value
		switch r.(type) {
		case VI0:
			return pv.X
		case VI1:
			return pv.Y
		default:
			// Neutral interval variable - stuck
			head := Head{Glob: "@"}
			return VNeutral{N: Neutral{Head: head, Sp: []Value{p, r}}}
		}

	case VPath:
		// Non-dependent Path applied to endpoint returns the corresponding endpoint
		switch r.(type) {
		case VI0:
			return pv.X
		case VI1:
			return pv.Y
		default:
			// Neutral interval variable - stuck
			head := Head{Glob: "@"}
			return VNeutral{N: Neutral{Head: head, Sp: []Value{p, r}}}
		}

	case VUA:
		// UA applied to interval: (ua e) @ r
		return UAPathApply(pv, r)

	case VFill:
		// Fill applied to interval endpoint:
		//   fill^i A [φ ↦ u] a₀ @ i0 = a₀
		//   fill^i A [φ ↦ u] a₀ @ i1 = comp^i A [φ ↦ u] a₀
		switch r.(type) {
		case VI0:
			// At i0: return base
			return pv.Base
		case VI1:
			// At i1: return comp result
			return EvalComp(pv.A, pv.Phi, pv.Tube, pv.Base)
		default:
			// Neutral interval - stuck
			head := Head{Glob: "@"}
			return VNeutral{N: Neutral{Head: head, Sp: []Value{p, r}}}
		}

	case VNeutral:
		// Stuck path application
		head := Head{Glob: "@"}
		return VNeutral{N: Neutral{Head: head, Sp: []Value{p, r}}}

	case VHITPathCtor:
		// HIT path constructor applied to interval
		// Apply the interval argument and check for endpoint reduction
		newIArgs := make([]Value, len(pv.IArgs)+1)
		copy(newIArgs, pv.IArgs)
		newIArgs[len(pv.IArgs)] = r
		return evalHITApp(pv.HITName, pv.CtorName, pv.Args, newIArgs, pv.Boundaries)

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

// --- Face Formula Evaluation ---

// evalFace evaluates a face formula to a face value.
func evalFace(env *Env, ienv *IEnv, f ast.Face) FaceValue {
	if f == nil {
		return VFaceBot{}
	}
	switch face := f.(type) {
	case ast.FaceTop:
		return VFaceTop{}
	case ast.FaceBot:
		return VFaceBot{}
	case ast.FaceEq:
		iVal := ienv.Lookup(face.IVar)
		return evalFaceEq(iVal, face.IVar, face.IsOne, ienv.ILen())
	case ast.FaceAnd:
		left := evalFace(env, ienv, face.Left)
		right := evalFace(env, ienv, face.Right)
		return simplifyFaceAnd(left, right)
	case ast.FaceOr:
		left := evalFace(env, ienv, face.Left)
		right := evalFace(env, ienv, face.Right)
		return simplifyFaceOr(left, right)
	default:
		return VFaceBot{}
	}
}

// evalFaceEq evaluates a face equality constraint (i = 0) or (i = 1).
// If the interval variable is resolved to an endpoint, simplify immediately.
func evalFaceEq(iVal Value, ivar int, isOne bool, ienvLen int) FaceValue {
	switch iv := iVal.(type) {
	case VI0:
		// i is known to be i0
		if isOne {
			return VFaceBot{} // (i0 = 1) is false
		}
		return VFaceTop{} // (i0 = 0) is true
	case VI1:
		// i is known to be i1
		if isOne {
			return VFaceTop{} // (i1 = 1) is true
		}
		return VFaceBot{} // (i1 = 0) is false
	case VIVar:
		// Interval variable is unresolved - keep as constraint
		return VFaceEq{ILevel: iv.Level, IsOne: isOne}
	default:
		// Unknown interval value - use original index converted to level
		return VFaceEq{ILevel: ienvLen - ivar - 1, IsOne: isOne}
	}
}

// simplifyFaceAnd simplifies φ ∧ ψ using boolean identities.
func simplifyFaceAnd(left, right FaceValue) FaceValue {
	// ⊥ ∧ ψ = ⊥
	if _, ok := left.(VFaceBot); ok {
		return VFaceBot{}
	}
	// φ ∧ ⊥ = ⊥
	if _, ok := right.(VFaceBot); ok {
		return VFaceBot{}
	}
	// ⊤ ∧ ψ = ψ
	if _, ok := left.(VFaceTop); ok {
		return right
	}
	// φ ∧ ⊤ = φ
	if _, ok := right.(VFaceTop); ok {
		return left
	}
	// Check for (i=0) ∧ (i=1) = ⊥
	if leq, lok := left.(VFaceEq); lok {
		if req, rok := right.(VFaceEq); rok {
			if leq.ILevel == req.ILevel && leq.IsOne != req.IsOne {
				return VFaceBot{}
			}
		}
	}
	return VFaceAnd{Left: left, Right: right}
}

// simplifyFaceOr simplifies φ ∨ ψ using boolean identities.
func simplifyFaceOr(left, right FaceValue) FaceValue {
	// ⊤ ∨ ψ = ⊤
	if _, ok := left.(VFaceTop); ok {
		return VFaceTop{}
	}
	// φ ∨ ⊤ = ⊤
	if _, ok := right.(VFaceTop); ok {
		return VFaceTop{}
	}
	// ⊥ ∨ ψ = ψ
	if _, ok := left.(VFaceBot); ok {
		return right
	}
	// φ ∨ ⊥ = φ
	if _, ok := right.(VFaceBot); ok {
		return left
	}
	// Check for (i=0) ∨ (i=1) = ⊤
	if leq, lok := left.(VFaceEq); lok {
		if req, rok := right.(VFaceEq); rok {
			if leq.ILevel == req.ILevel && leq.IsOne != req.IsOne {
				return VFaceTop{}
			}
		}
	}
	return VFaceOr{Left: left, Right: right}
}

// IsFaceTrue checks if a face value is definitely true.
func IsFaceTrue(f FaceValue) bool {
	_, ok := f.(VFaceTop)
	return ok
}

// IsFaceFalse checks if a face value is definitely false.
func IsFaceFalse(f FaceValue) bool {
	_, ok := f.(VFaceBot)
	return ok
}

// --- Composition Evaluation ---

// EvalComp evaluates heterogeneous composition: comp^i A [φ ↦ u] a₀.
// Computation rules:
//
//	comp^i A [1 ↦ u] a₀  ⟶  u[i1/i]         (face satisfied)
//	comp^i A [0 ↦ _] a₀  ⟶  transport A a₀  (face empty)
func EvalComp(aClosure *IClosure, phi FaceValue, tubeClosure *IClosure, base Value) Value {
	// If face is satisfied (φ = ⊤), return tube at i1
	if IsFaceTrue(phi) {
		return EvalCubical(tubeClosure.Env, tubeClosure.IEnv.Extend(VI1{}), tubeClosure.Term)
	}

	// If face is empty (φ = ⊥), reduce to transport
	if IsFaceFalse(phi) {
		return EvalTransport(aClosure, base)
	}

	// Cannot reduce - return stuck composition
	return VComp{A: aClosure, Phi: phi, Tube: tubeClosure, Base: base}
}

// EvalHComp evaluates homogeneous composition: hcomp A [φ ↦ u] a₀.
// Computation rules:
//
//	hcomp A [1 ↦ u] a₀  ⟶  u[i1/i]   (face satisfied)
//	hcomp A [0 ↦ _] a₀  ⟶  a₀        (face empty, identity)
func EvalHComp(a Value, phi FaceValue, tubeClosure *IClosure, base Value) Value {
	// If face is satisfied (φ = ⊤), return tube at i1
	if IsFaceTrue(phi) {
		return EvalCubical(tubeClosure.Env, tubeClosure.IEnv.Extend(VI1{}), tubeClosure.Term)
	}

	// If face is empty (φ = ⊥), return base (identity)
	if IsFaceFalse(phi) {
		return base
	}

	// Cannot reduce - return stuck hcomp
	return VHComp{A: a, Phi: phi, Tube: tubeClosure, Base: base}
}

// EvalFill evaluates the filler: fill^i A [φ ↦ u] a₀.
// Fill produces a path-like value that can be applied to an interval.
// The computation rules are handled in PathApply:
//
//	fill^i A [φ ↦ u] a₀ @ i0 = a₀
//	fill^i A [φ ↦ u] a₀ @ i1 = comp^i A [φ ↦ u] a₀
//
// If the face is satisfied (φ = ⊤), we can reduce immediately:
//
//	fill^i A [1 ↦ u] a₀ = <j> u[j/i]
func EvalFill(aClosure *IClosure, phi FaceValue, tubeClosure *IClosure, base Value) Value {
	// If face is satisfied, fill reduces to a path lambda over the tube
	if IsFaceTrue(phi) {
		// fill^i A [1 ↦ u] a₀ = <j> u[j/i]
		// Return a path lambda that applies the tube at the given interval
		return VPathLam{Body: tubeClosure}
	}

	// Otherwise, return stuck fill (will be reduced in PathApply when applied to i0/i1)
	return VFill{A: aClosure, Phi: phi, Tube: tubeClosure, Base: base}
}

// --- Glue Type Evaluation ---

// EvalGlue evaluates a Glue type: Glue A [φ ↦ (T, e)].
// Computation rules:
//
//	Glue A [⊤ ↦ (T, e)] = T    (face satisfied)
//	Glue A []           = A    (no branches)
func EvalGlue(a Value, branches []VGlueBranch) Value {
	// Check if any branch has face ⊤
	for _, br := range branches {
		if IsFaceTrue(br.Phi) {
			return br.T
		}
	}

	// Filter out branches with ⊥ face
	var nonTrivialBranches []VGlueBranch
	for _, br := range branches {
		if !IsFaceFalse(br.Phi) {
			nonTrivialBranches = append(nonTrivialBranches, br)
		}
	}

	// If no branches remain, return base type
	if len(nonTrivialBranches) == 0 {
		return a
	}

	return VGlue{A: a, System: nonTrivialBranches}
}

// EvalGlueElem evaluates a Glue element: glue [φ ↦ t] a.
// Computation rules:
//
//	glue [⊤ ↦ t] a = t    (face satisfied)
func EvalGlueElem(branches []VGlueElemBranch, base Value) Value {
	// Check if any branch has face ⊤
	for _, br := range branches {
		if IsFaceTrue(br.Phi) {
			return br.Term
		}
	}

	// Filter out branches with ⊥ face
	var nonTrivialBranches []VGlueElemBranch
	for _, br := range branches {
		if !IsFaceFalse(br.Phi) {
			nonTrivialBranches = append(nonTrivialBranches, br)
		}
	}

	return VGlueElem{System: nonTrivialBranches, Base: base}
}

// EvalUnglue evaluates unglue: unglue g.
// Computation rules:
//
//	unglue (glue [φ ↦ t] a) = a    (definitional)
func EvalUnglue(ty Value, g Value) Value {
	// If g is a glue element, return its base
	if ge, ok := g.(VGlueElem); ok {
		return ge.Base
	}

	// Otherwise stuck
	return VUnglue{Ty: ty, G: g}
}

// --- Univalence Evaluation ---

// EvalUA evaluates ua e : Path Type A B.
// Definition via Glue:
//
//	ua e = <i> Glue B [(i=0) ↦ (A, e)]
//
// At i=0: Glue B [⊤ ↦ (A, e)] = A
// At i=1: Glue B [⊥ ↦ (A, e)] = B
func EvalUA(a, b, equiv Value) Value {
	// ua produces a path, which we represent as VUA
	// When applied to an interval, we compute the Glue type
	return VUA{A: a, B: b, Equiv: equiv}
}

// EvalUABeta evaluates the transport computation: transport (ua e) a = e.fst a.
// The equivalence e : Equiv A B has the form (fwd, proof) where fwd : A -> B.
// This computation rule extracts the forward function and applies it to the argument.
// When the equivalence is neutral (stuck), we return VUABeta to preserve the structure.
func EvalUABeta(equiv, arg Value) Value {
	// Only reduce when the equivalence is a concrete pair
	if pair, ok := equiv.(VPair); ok {
		// Extract the forward function and apply it to the argument
		return Apply(pair.Fst, arg)
	}

	// For neutral equivalences, return VUABeta (stuck value)
	return VUABeta{Equiv: equiv, Arg: arg}
}

// UAPathApply applies ua to an interval argument.
// This is called when we have (ua e) @ r.
//
//	(ua e) @ i0 = A
//	(ua e) @ i1 = B
//	(ua e) @ i  = Glue B [(i=0) ↦ (A, e)]
func UAPathApply(ua VUA, r Value) Value {
	switch r.(type) {
	case VI0:
		// At i=0: Glue B [⊤ ↦ (A, e)] = A
		return ua.A
	case VI1:
		// At i=1: Glue B [⊥ ↦ (A, e)] = B
		return ua.B
	default:
		// At i: Glue B [(i=0) ↦ (A, e)]
		// We need to construct the Glue type with the face constraint
		branch := VGlueBranch{
			Phi:   evalFaceEqForUA(r),
			T:     ua.A,
			Equiv: ua.Equiv,
		}
		return VGlue{A: ua.B, System: []VGlueBranch{branch}}
	}
}

// evalFaceEqForUA creates a face constraint (i = 0) for the given interval value.
func evalFaceEqForUA(r Value) FaceValue {
	if iv, ok := r.(VIVar); ok {
		return VFaceEq{ILevel: iv.Level, IsOne: false} // (i = 0)
	}
	// If r is not a variable, this shouldn't happen in well-typed code
	return VFaceBot{}
}

// isConstantFamily checks if an interval closure produces the same value at i0 and i1.
// This is used for the transport computation rule.
func isConstantFamily(c *IClosure) bool {
	v0 := EvalCubical(c.Env, c.IEnv.Extend(VI0{}), c.Term)
	v1 := EvalCubical(c.Env, c.IEnv.Extend(VI1{}), c.Term)
	// Use alpha equality on reified terms.
	// Use proper ilevel to correctly distinguish interval variables.
	// After extending IEnv by 1, the ilevel is c.IEnv.ILen() + 1.
	level := 0 // term level (no term binders in this context)
	ilevel := c.IEnv.ILen() + 1
	t0 := ReifyCubicalAt(level, ilevel, v0)
	t1 := ReifyCubicalAt(level, ilevel, v1)
	return alphaEqCubical(t0, t1)
}

// alphaEqCubical checks alpha-equality of cubical terms.
// Uses proper structural comparison via de Bruijn indices.
func alphaEqCubical(a, b ast.Term) bool {
	return AlphaEq(a, b)
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

	// --- Face Formula Values ---

	case VFaceTop:
		return ast.FaceTop{}

	case VFaceBot:
		return ast.FaceBot{}

	case VFaceEq:
		// Convert from level to de Bruijn index
		ix := ilevel - val.ILevel - 1
		if ix < 0 {
			ix = val.ILevel
		}
		return ast.FaceEq{IVar: ix, IsOne: val.IsOne}

	case VFaceAnd:
		left := reifyFaceAt(level, ilevel, val.Left)
		right := reifyFaceAt(level, ilevel, val.Right)
		return ast.FaceAnd{Left: left, Right: right}

	case VFaceOr:
		left := reifyFaceAt(level, ilevel, val.Left)
		right := reifyFaceAt(level, ilevel, val.Right)
		return ast.FaceOr{Left: left, Right: right}

	// --- Partial Type Values ---

	case VPartial:
		phi := reifyFaceAt(level, ilevel, val.Phi)
		a := ReifyCubicalAt(level, ilevel, val.A)
		return ast.Partial{Phi: phi, A: a}

	case VSystem:
		branches := make([]ast.SystemBranch, len(val.Branches))
		for i, br := range val.Branches {
			phi := reifyFaceAt(level, ilevel, br.Phi)
			term := ReifyCubicalAt(level, ilevel, br.Term)
			branches[i] = ast.SystemBranch{Phi: phi, Term: term}
		}
		return ast.System{Branches: branches}

	// --- Composition Values ---

	case VComp:
		// Reify stuck composition
		freshIVar := VIVar{Level: ilevel}
		aVal := EvalCubical(val.A.Env, val.A.IEnv.Extend(freshIVar), val.A.Term)
		a := ReifyCubicalAt(level, ilevel+1, aVal)
		phi := reifyFaceAt(level, ilevel+1, val.Phi)
		tubeVal := EvalCubical(val.Tube.Env, val.Tube.IEnv.Extend(freshIVar), val.Tube.Term)
		tube := ReifyCubicalAt(level, ilevel+1, tubeVal)
		base := ReifyCubicalAt(level, ilevel, val.Base)
		return ast.Comp{IBinder: "_", A: a, Phi: phi, Tube: tube, Base: base}

	case VHComp:
		// Reify stuck hcomp
		freshIVar := VIVar{Level: ilevel}
		a := ReifyCubicalAt(level, ilevel, val.A)
		phi := reifyFaceAt(level, ilevel+1, val.Phi)
		tubeVal := EvalCubical(val.Tube.Env, val.Tube.IEnv.Extend(freshIVar), val.Tube.Term)
		tube := ReifyCubicalAt(level, ilevel+1, tubeVal)
		base := ReifyCubicalAt(level, ilevel, val.Base)
		return ast.HComp{A: a, Phi: phi, Tube: tube, Base: base}

	case VFill:
		// Reify stuck fill
		freshIVar := VIVar{Level: ilevel}
		aVal := EvalCubical(val.A.Env, val.A.IEnv.Extend(freshIVar), val.A.Term)
		a := ReifyCubicalAt(level, ilevel+1, aVal)
		phi := reifyFaceAt(level, ilevel+1, val.Phi)
		tubeVal := EvalCubical(val.Tube.Env, val.Tube.IEnv.Extend(freshIVar), val.Tube.Term)
		tube := ReifyCubicalAt(level, ilevel+1, tubeVal)
		base := ReifyCubicalAt(level, ilevel, val.Base)
		return ast.Fill{IBinder: "_", A: a, Phi: phi, Tube: tube, Base: base}

	// --- Glue Type Values ---

	case VGlue:
		a := ReifyCubicalAt(level, ilevel, val.A)
		branches := make([]ast.GlueBranch, len(val.System))
		for i, br := range val.System {
			phi := reifyFaceAt(level, ilevel, br.Phi)
			t := ReifyCubicalAt(level, ilevel, br.T)
			equiv := ReifyCubicalAt(level, ilevel, br.Equiv)
			branches[i] = ast.GlueBranch{Phi: phi, T: t, Equiv: equiv}
		}
		return ast.Glue{A: a, System: branches}

	case VGlueElem:
		branches := make([]ast.GlueElemBranch, len(val.System))
		for i, br := range val.System {
			phi := reifyFaceAt(level, ilevel, br.Phi)
			term := ReifyCubicalAt(level, ilevel, br.Term)
			branches[i] = ast.GlueElemBranch{Phi: phi, Term: term}
		}
		base := ReifyCubicalAt(level, ilevel, val.Base)
		return ast.GlueElem{System: branches, Base: base}

	case VUnglue:
		g := ReifyCubicalAt(level, ilevel, val.G)
		return ast.Unglue{Ty: ReifyCubicalAt(level, ilevel, val.Ty), G: g}

	// --- Univalence Values ---

	case VUA:
		a := ReifyCubicalAt(level, ilevel, val.A)
		b := ReifyCubicalAt(level, ilevel, val.B)
		equiv := ReifyCubicalAt(level, ilevel, val.Equiv)
		return ast.UA{A: a, B: b, Equiv: equiv}

	case VUABeta:
		equiv := ReifyCubicalAt(level, ilevel, val.Equiv)
		arg := ReifyCubicalAt(level, ilevel, val.Arg)
		return ast.UABeta{Equiv: equiv, Arg: arg}

	// --- Higher Inductive Types ---

	case VHITPathCtor:
		// Reify all term arguments
		args := make([]ast.Term, len(val.Args))
		for i, arg := range val.Args {
			args[i] = ReifyCubicalAt(level, ilevel, arg)
		}
		// Reify all interval arguments
		iargs := make([]ast.Term, len(val.IArgs))
		for i, iarg := range val.IArgs {
			iargs[i] = ReifyCubicalAt(level, ilevel, iarg)
		}
		return ast.HITApp{
			HITName: val.HITName,
			Ctor:    val.CtorName,
			Args:    args,
			IArgs:   iargs,
		}

	default:
		return reifyError("unknown cubical value type")
	}
}

// reifyFaceAt converts a face value to an AST face formula.
func reifyFaceAt(level int, ilevel int, f FaceValue) ast.Face {
	switch fv := f.(type) {
	case VFaceTop:
		return ast.FaceTop{}
	case VFaceBot:
		return ast.FaceBot{}
	case VFaceEq:
		// Convert from level to de Bruijn index
		ix := ilevel - fv.ILevel - 1
		if ix < 0 {
			ix = fv.ILevel
		}
		return ast.FaceEq{IVar: ix, IsOne: fv.IsOne}
	case VFaceAnd:
		left := reifyFaceAt(level, ilevel, fv.Left)
		right := reifyFaceAt(level, ilevel, fv.Right)
		return ast.FaceAnd{Left: left, Right: right}
	case VFaceOr:
		left := reifyFaceAt(level, ilevel, fv.Left)
		right := reifyFaceAt(level, ilevel, fv.Right)
		return ast.FaceOr{Left: left, Right: right}
	default:
		return ast.FaceBot{}
	}
}

// reifyNeutralCubicalAt converts a Neutral to an ast.Term with cubical support.
// Uses the shared reifyNeutralWithReifier helper with cubical-specific extra cases.
func reifyNeutralCubicalAt(level int, ilevel int, n Neutral) ast.Term {
	reifier := func(v Value) ast.Term { return ReifyCubicalAt(level, ilevel, v) }

	// Handle cubical-specific "@" (path application)
	extraCases := func(glob string, sp []Value) (ast.Term, bool) {
		if glob == "@" && len(sp) >= 2 {
			p := ReifyCubicalAt(level, ilevel, sp[0])
			r := ReifyCubicalAt(level, ilevel, sp[1])
			base := ast.PathApp{P: p, R: r}
			return reifySpine(base, sp[2:], reifier), true
		}
		return nil, false
	}

	return reifyNeutralWithReifier(level, n, reifier, extraCases)
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
	// Face formulas
	case ast.FaceTop:
		return VFaceTop{}, true
	case ast.FaceBot:
		return VFaceBot{}, true
	case ast.FaceEq:
		// Without ienv, use index as level
		return VFaceEq{ILevel: tm.IVar, IsOne: tm.IsOne}, true
	case ast.FaceAnd:
		left := evalFace(env, EmptyIEnv(), tm.Left)
		right := evalFace(env, EmptyIEnv(), tm.Right)
		return simplifyFaceAnd(left, right), true
	case ast.FaceOr:
		left := evalFace(env, EmptyIEnv(), tm.Left)
		right := evalFace(env, EmptyIEnv(), tm.Right)
		return simplifyFaceOr(left, right), true
	// Partial types
	case ast.Partial:
		phi := evalFace(env, EmptyIEnv(), tm.Phi)
		a := Eval(env, tm.A)
		return VPartial{Phi: phi, A: a}, true
	case ast.System:
		branches := make([]VSystemBranch, len(tm.Branches))
		for i, br := range tm.Branches {
			phi := evalFace(env, EmptyIEnv(), br.Phi)
			term := Eval(env, br.Term)
			branches[i] = VSystemBranch{Phi: phi, Term: term}
		}
		return VSystem{Branches: branches}, true
	// Composition operations
	case ast.Comp:
		phi := evalFace(env, EmptyIEnv(), tm.Phi)
		aClosure := &IClosure{Env: env, IEnv: EmptyIEnv(), Term: tm.A}
		tubeClosure := &IClosure{Env: env, IEnv: EmptyIEnv(), Term: tm.Tube}
		base := Eval(env, tm.Base)
		return EvalComp(aClosure, phi, tubeClosure, base), true
	case ast.HComp:
		a := Eval(env, tm.A)
		phi := evalFace(env, EmptyIEnv(), tm.Phi)
		tubeClosure := &IClosure{Env: env, IEnv: EmptyIEnv(), Term: tm.Tube}
		base := Eval(env, tm.Base)
		return EvalHComp(a, phi, tubeClosure, base), true
	case ast.Fill:
		phi := evalFace(env, EmptyIEnv(), tm.Phi)
		aClosure := &IClosure{Env: env, IEnv: EmptyIEnv(), Term: tm.A}
		tubeClosure := &IClosure{Env: env, IEnv: EmptyIEnv(), Term: tm.Tube}
		base := Eval(env, tm.Base)
		return EvalFill(aClosure, phi, tubeClosure, base), true
	// Glue types
	case ast.Glue:
		a := Eval(env, tm.A)
		branches := make([]VGlueBranch, len(tm.System))
		for i, br := range tm.System {
			phi := evalFace(env, EmptyIEnv(), br.Phi)
			t := Eval(env, br.T)
			equiv := Eval(env, br.Equiv)
			branches[i] = VGlueBranch{Phi: phi, T: t, Equiv: equiv}
		}
		return EvalGlue(a, branches), true
	case ast.GlueElem:
		branches := make([]VGlueElemBranch, len(tm.System))
		for i, br := range tm.System {
			phi := evalFace(env, EmptyIEnv(), br.Phi)
			term := Eval(env, br.Term)
			branches[i] = VGlueElemBranch{Phi: phi, Term: term}
		}
		base := Eval(env, tm.Base)
		return EvalGlueElem(branches, base), true
	case ast.Unglue:
		ty := Eval(env, tm.Ty)
		g := Eval(env, tm.G)
		return EvalUnglue(ty, g), true
	// Univalence
	case ast.UA:
		a := Eval(env, tm.A)
		b := Eval(env, tm.B)
		equiv := Eval(env, tm.Equiv)
		return EvalUA(a, b, equiv), true
	case ast.UABeta:
		equiv := Eval(env, tm.Equiv)
		arg := Eval(env, tm.Arg)
		return EvalUABeta(equiv, arg), true
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
		// In non-cubical reify context, we don't have proper ilevel tracking.
		// Use the level directly as the index (best effort for simple cases).
		// For correct handling, callers should use ReifyCubicalAt directly.
		return ast.IVar{Ix: val.Level}, true
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
	// Face formulas
	case VFaceTop:
		return ast.FaceTop{}, true
	case VFaceBot:
		return ast.FaceBot{}, true
	case VFaceEq:
		return ast.FaceEq{IVar: val.ILevel, IsOne: val.IsOne}, true
	case VFaceAnd:
		return ReifyCubicalAt(level, 0, val), true
	case VFaceOr:
		return ReifyCubicalAt(level, 0, val), true
	// Partial types
	case VPartial:
		return ReifyCubicalAt(level, 0, val), true
	case VSystem:
		return ReifyCubicalAt(level, 0, val), true
	// Composition values
	case VComp:
		return ReifyCubicalAt(level, 0, val), true
	case VHComp:
		return ReifyCubicalAt(level, 0, val), true
	case VFill:
		return ReifyCubicalAt(level, 0, val), true
	// Glue types
	case VGlue:
		return ReifyCubicalAt(level, 0, val), true
	case VGlueElem:
		return ReifyCubicalAt(level, 0, val), true
	case VUnglue:
		return ReifyCubicalAt(level, 0, val), true
	// Univalence
	case VUA:
		return ReifyCubicalAt(level, 0, val), true
	case VUABeta:
		return ReifyCubicalAt(level, 0, val), true
	// Higher Inductive Types
	case VHITPathCtor:
		return ReifyCubicalAt(level, 0, val), true
	default:
		return nil, false
	}
}
