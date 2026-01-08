// Package elab provides elaboration from surface syntax to core terms.
//
// Surface syntax (STerm) extends raw terms with:
//   - Implicit arguments: {x : A} -> B, \{x}. t, f {arg}
//   - Holes: _ (anonymous) or ?name (named metavariables)
//   - Source spans for error messages
//
// The elaborator transforms STerm into ast.Term, solving for implicit
// arguments and metavariables through unification.
package elab

// Icity marks whether a binder or argument is implicit or explicit.
type Icity int

const (
	// Explicit arguments are written and inferred normally: (x : A) -> B
	Explicit Icity = iota
	// Implicit arguments are inferred by unification: {x : A} -> B
	Implicit
	// Instance arguments are resolved by type class search: [x : A] -> B
	// (Reserved for future use)
	Instance
)

func (i Icity) String() string {
	switch i {
	case Explicit:
		return "explicit"
	case Implicit:
		return "implicit"
	case Instance:
		return "instance"
	default:
		return "unknown"
	}
}

// Span represents a source location for error messages.
type Span struct {
	File   string // Source file name
	Line   int    // Line number (1-indexed)
	Col    int    // Column number (1-indexed)
	EndCol int    // End column (for highlighting)
}

// NoSpan is the zero value span for generated terms.
var NoSpan = Span{}

// STerm is the interface for all surface syntax terms.
// Surface terms carry:
//   - User-written names (not de Bruijn indices)
//   - Implicit/explicit markers
//   - Holes for inference
//   - Source spans for errors
type STerm interface {
	isSurface()
	Span() Span
}

// base implements common STerm functionality
type base struct{ span Span }

func (b base) Span() Span { return b.span }

// ---------- Variables ----------

// SVar is a variable reference (resolved to de Bruijn index during elaboration).
type SVar struct {
	base
	Name string // Variable name
}

func (SVar) isSurface() {}

// SGlobal is a reference to a global definition.
type SGlobal struct {
	base
	Name string // Global name
}

func (SGlobal) isSurface() {}

// ---------- Universes ----------

// SType is a universe Type^level.
type SType struct {
	base
	Level uint // Universe level
}

func (SType) isSurface() {}

// ---------- Pi Types ----------

// SPi is a Pi type: (x : A) -> B or {x : A} -> B for implicit.
type SPi struct {
	base
	Binder string // Binder name
	Icity  Icity  // Implicit or explicit
	Dom    STerm  // Domain type A
	Cod    STerm  // Codomain type B (may reference Binder)
}

func (SPi) isSurface() {}

// SArrow is non-dependent function type: A -> B (sugar for Pi with unused binder).
type SArrow struct {
	base
	Dom STerm // Domain
	Cod STerm // Codomain
}

func (SArrow) isSurface() {}

// ---------- Lambdas ----------

// SLam is a lambda abstraction: \x. t or \{x}. t for implicit.
type SLam struct {
	base
	Binder string // Binder name
	Icity  Icity  // Implicit or explicit parameter
	Ann    STerm  // Optional type annotation (may be nil)
	Body   STerm  // Lambda body
}

func (SLam) isSurface() {}

// ---------- Application ----------

// SApp is function application: f x or f {x} for explicit implicit.
type SApp struct {
	base
	Fn     STerm // Function
	Arg    STerm // Argument
	Icity  Icity // Whether argument is explicitly marked implicit
	Spread bool  // Whether this is f @args... (reserved for future)
}

func (SApp) isSurface() {}

// ---------- Sigma Types ----------

// SSigma is a dependent pair type: (x : A) * B.
type SSigma struct {
	base
	Binder string
	Fst    STerm // First component type
	Snd    STerm // Second component type (may reference Binder)
}

func (SSigma) isSurface() {}

// SProd is non-dependent product type: A * B (sugar for Sigma with unused binder).
type SProd struct {
	base
	Fst STerm
	Snd STerm
}

func (SProd) isSurface() {}

// ---------- Pairs and Projections ----------

// SPair is a pair constructor: (a, b).
type SPair struct {
	base
	Fst STerm
	Snd STerm
}

func (SPair) isSurface() {}

// SFst is the first projection: fst p or p.1.
type SFst struct {
	base
	Pair STerm
}

func (SFst) isSurface() {}

// SSnd is the second projection: snd p or p.2.
type SSnd struct {
	base
	Pair STerm
}

func (SSnd) isSurface() {}

// ---------- Let Bindings ----------

// SLet is a let binding: let x : A = v in body.
type SLet struct {
	base
	Binder string
	Ann    STerm // Type annotation (may be nil)
	Val    STerm // Bound value
	Body   STerm // Body expression
}

func (SLet) isSurface() {}

// ---------- Holes (Metavariables) ----------

// SHole is a hole to be filled by elaboration: _ or ?name.
type SHole struct {
	base
	Name string // Empty for anonymous _, otherwise ?name
}

func (SHole) isSurface() {}

// ---------- Identity Types ----------

// SId is the identity type: Id A x y.
type SId struct {
	base
	A STerm // Type
	X STerm // Left endpoint
	Y STerm // Right endpoint
}

func (SId) isSurface() {}

// SRefl is the reflexivity constructor: refl or refl A x.
type SRefl struct {
	base
	A STerm // Type (may be nil, to be inferred)
	X STerm // Term (may be nil, to be inferred)
}

func (SRefl) isSurface() {}

// SJ is the J eliminator: J A C d x y p.
type SJ struct {
	base
	A STerm // Type
	C STerm // Motive
	D STerm // Base case
	X STerm // Left endpoint
	Y STerm // Right endpoint
	P STerm // Proof
}

func (SJ) isSurface() {}

// ---------- Cubical Types ----------

// SPath is the path type: Path A x y.
type SPath struct {
	base
	A STerm // Type (constant over interval)
	X STerm // Left endpoint
	Y STerm // Right endpoint
}

func (SPath) isSurface() {}

// SPathP is the dependent path type: PathP A x y.
type SPathP struct {
	base
	A STerm // Type family: I -> Type
	X STerm // Left endpoint
	Y STerm // Right endpoint
}

func (SPathP) isSurface() {}

// SPathLam is path abstraction: <i> t.
type SPathLam struct {
	base
	Binder string // Interval variable name
	Body   STerm  // Body with interval variable bound
}

func (SPathLam) isSurface() {}

// SPathApp is path application: p @ r.
type SPathApp struct {
	base
	Path STerm // Path
	Arg  STerm // Interval argument
}

func (SPathApp) isSurface() {}

// SI0 is the interval endpoint 0.
type SI0 struct{ base }

func (SI0) isSurface() {}

// SI1 is the interval endpoint 1.
type SI1 struct{ base }

func (SI1) isSurface() {}

// STransport is cubical transport: transport A e.
type STransport struct {
	base
	A STerm // Type family
	E STerm // Element
}

func (STransport) isSurface() {}

// ---------- Inductive Types ----------

// SIndApp is inductive type constructor application.
type SIndApp struct {
	base
	Name string  // Inductive type name
	Args []STerm // Arguments
}

func (SIndApp) isSurface() {}

// SCtorApp is constructor application.
type SCtorApp struct {
	base
	Ind  string  // Inductive type name
	Ctor string  // Constructor name
	Args []STerm // Arguments
}

func (SCtorApp) isSurface() {}

// SElim is eliminator/recursor application.
type SElim struct {
	base
	Name    string  // Eliminator name
	Motive  STerm   // Motive
	Methods []STerm // Methods for each constructor
	Target  STerm   // Term being eliminated
}

func (SElim) isSurface() {}

// ---------- Helpers ----------

// MkSVar creates a variable with no span.
func MkSVar(name string) *SVar {
	return &SVar{Name: name}
}

// MkSApp creates an explicit application with no span.
func MkSApp(fn, arg STerm) *SApp {
	return &SApp{Fn: fn, Arg: arg, Icity: Explicit}
}

// MkSApps creates a chain of explicit applications.
func MkSApps(fn STerm, args ...STerm) STerm {
	for _, arg := range args {
		fn = MkSApp(fn, arg)
	}
	return fn
}

// MkSPi creates an explicit Pi type with no span.
func MkSPi(binder string, dom, cod STerm) *SPi {
	return &SPi{Binder: binder, Icity: Explicit, Dom: dom, Cod: cod}
}

// MkSLam creates an explicit lambda with no span.
func MkSLam(binder string, body STerm) *SLam {
	return &SLam{Binder: binder, Icity: Explicit, Body: body}
}

// MkSHole creates an anonymous hole.
func MkSHole() *SHole {
	return &SHole{}
}

// MkSNamedHole creates a named hole.
func MkSNamedHole(name string) *SHole {
	return &SHole{Name: name}
}

// WithSpan returns a copy of the term with the given span.
// This is used by the parser to attach source locations.
func WithSpan[T STerm](t T, span Span) T {
	// Note: This requires reflection or type switches in practice.
	// For now, we rely on the parser setting spans directly.
	return t
}
