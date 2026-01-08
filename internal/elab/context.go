package elab

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/kernel/ctx"
)

// ElabBinding extends ctx.Binding with icity and definition information.
type ElabBinding struct {
	Name  string   // Variable name
	Type  ast.Term // Type of the binding
	Icity Icity    // Implicit or explicit
	Def   ast.Term // Definition if this is a let-binding (nil for lambda bindings)
}

// ElabCtx is the elaboration context, extending the kernel context
// with information needed for implicit argument inference.
type ElabCtx struct {
	// Bindings is the local context with icity information.
	Bindings []ElabBinding

	// Metas is the shared metavariable store.
	Metas *MetaStore

	// IBindings tracks interval variable bindings for cubical elaboration.
	// Stored separately because interval variables have a different namespace.
	IBindings []string

	// Globals provides access to global definitions.
	// This is set by the elaborator and used for lookups.
	Globals GlobalEnv
}

// GlobalEnv provides access to global definitions.
type GlobalEnv interface {
	// LookupGlobal looks up a global definition by name.
	LookupGlobal(name string) (ty ast.Term, def ast.Term, ok bool)

	// LookupInductive looks up an inductive type.
	LookupInductive(name string) (IndInfo, bool)

	// LookupConstructor looks up a constructor.
	LookupConstructor(name string) (CtorInfo, bool)
}

// IndInfo holds information about an inductive type.
type IndInfo struct {
	Name       string
	Type       ast.Term
	NumParams  int
	NumIndices int
	Ctors      []string // Constructor names
}

// CtorInfo holds information about a constructor.
type CtorInfo struct {
	Name    string
	IndName string
	Type    ast.Term
	Index   int // Constructor index (for eliminators)
}

// NewElabCtx creates a new elaboration context with a fresh metastore.
func NewElabCtx() *ElabCtx {
	return &ElabCtx{
		Bindings:  nil,
		Metas:     NewMetaStore(),
		IBindings: nil,
	}
}

// WithMetas creates an elaboration context with a shared metastore.
func WithMetas(metas *MetaStore) *ElabCtx {
	return &ElabCtx{
		Bindings:  nil,
		Metas:     metas,
		IBindings: nil,
	}
}

// WithGlobals sets the global environment.
func (e *ElabCtx) WithGlobals(g GlobalEnv) *ElabCtx {
	e.Globals = g
	return e
}

// Len returns the number of bindings in the context.
func (e *ElabCtx) Len() int {
	return len(e.Bindings)
}

// ILen returns the number of interval variable bindings.
func (e *ElabCtx) ILen() int {
	return len(e.IBindings)
}

// Extend adds a new binding to the context.
// Returns a new context (immutable pattern for safety).
func (e *ElabCtx) Extend(name string, ty ast.Term, icity Icity) *ElabCtx {
	newBindings := make([]ElabBinding, len(e.Bindings)+1)
	copy(newBindings, e.Bindings)
	newBindings[len(e.Bindings)] = ElabBinding{
		Name:  name,
		Type:  ty,
		Icity: icity,
		Def:   nil,
	}

	return &ElabCtx{
		Bindings:  newBindings,
		Metas:     e.Metas,
		IBindings: e.IBindings,
		Globals:   e.Globals,
	}
}

// ExtendDef adds a let-binding to the context.
func (e *ElabCtx) ExtendDef(name string, ty, def ast.Term) *ElabCtx {
	newBindings := make([]ElabBinding, len(e.Bindings)+1)
	copy(newBindings, e.Bindings)
	newBindings[len(e.Bindings)] = ElabBinding{
		Name:  name,
		Type:  ty,
		Icity: Explicit,
		Def:   def,
	}

	return &ElabCtx{
		Bindings:  newBindings,
		Metas:     e.Metas,
		IBindings: e.IBindings,
		Globals:   e.Globals,
	}
}

// ExtendI adds an interval variable binding.
func (e *ElabCtx) ExtendI(name string) *ElabCtx {
	newIBindings := make([]string, len(e.IBindings)+1)
	copy(newIBindings, e.IBindings)
	newIBindings[len(e.IBindings)] = name

	return &ElabCtx{
		Bindings:  e.Bindings,
		Metas:     e.Metas,
		IBindings: newIBindings,
		Globals:   e.Globals,
	}
}

// LookupName looks up a variable by name, returning its de Bruijn index,
// type, and icity if found.
func (e *ElabCtx) LookupName(name string) (ix int, ty ast.Term, icity Icity, ok bool) {
	// Search from most recent to oldest (reverse order)
	for i := len(e.Bindings) - 1; i >= 0; i-- {
		if e.Bindings[i].Name == name {
			ix = len(e.Bindings) - 1 - i
			return ix, e.Bindings[i].Type, e.Bindings[i].Icity, true
		}
	}
	return 0, nil, Explicit, false
}

// LookupIName looks up an interval variable by name.
func (e *ElabCtx) LookupIName(name string) (ix int, ok bool) {
	for i := len(e.IBindings) - 1; i >= 0; i-- {
		if e.IBindings[i] == name {
			return len(e.IBindings) - 1 - i, true
		}
	}
	return 0, false
}

// LookupVar looks up a type by de Bruijn index.
func (e *ElabCtx) LookupVar(ix int) (ast.Term, bool) {
	if ix < 0 || ix >= len(e.Bindings) {
		return nil, false
	}
	return e.Bindings[len(e.Bindings)-1-ix].Type, true
}

// GetIcity returns the icity of a binding by de Bruijn index.
func (e *ElabCtx) GetIcity(ix int) Icity {
	if ix < 0 || ix >= len(e.Bindings) {
		return Explicit
	}
	return e.Bindings[len(e.Bindings)-1-ix].Icity
}

// GetDef returns the definition of a let-binding by de Bruijn index.
func (e *ElabCtx) GetDef(ix int) (ast.Term, bool) {
	if ix < 0 || ix >= len(e.Bindings) {
		return nil, false
	}
	def := e.Bindings[len(e.Bindings)-1-ix].Def
	return def, def != nil
}

// ToKernelCtx converts the elaboration context to a kernel context.
// This is used when calling into the type checker.
func (e *ElabCtx) ToKernelCtx() *ctx.Ctx {
	result := &ctx.Ctx{Tele: make([]ctx.Binding, len(e.Bindings))}
	for i, b := range e.Bindings {
		result.Tele[i] = ctx.Binding{Name: b.Name, Ty: b.Type}
	}
	return result
}

// Clone creates a copy of the context (useful for speculative elaboration).
func (e *ElabCtx) Clone() *ElabCtx {
	newBindings := make([]ElabBinding, len(e.Bindings))
	copy(newBindings, e.Bindings)

	newIBindings := make([]string, len(e.IBindings))
	copy(newIBindings, e.IBindings)

	return &ElabCtx{
		Bindings:  newBindings,
		Metas:     e.Metas, // Shared metastore
		IBindings: newIBindings,
		Globals:   e.Globals,
	}
}

// Names returns all bound variable names (for error messages).
func (e *ElabCtx) Names() []string {
	names := make([]string, len(e.Bindings))
	for i, b := range e.Bindings {
		names[i] = b.Name
	}
	return names
}

// ImplicitBindings returns the de Bruijn indices of all implicit bindings.
// Useful for auto-inserting implicit lambdas.
func (e *ElabCtx) ImplicitBindings() []int {
	var result []int
	for i, b := range e.Bindings {
		if b.Icity == Implicit {
			result = append(result, len(e.Bindings)-1-i)
		}
	}
	return result
}

// Fresh creates a fresh metavariable in the current context.
func (e *ElabCtx) Fresh(ty ast.Term, span Span) MetaID {
	return e.Metas.Fresh(ty, e.ToKernelCtx(), span)
}

// FreshNamed creates a named metavariable.
func (e *ElabCtx) FreshNamed(ty ast.Term, span Span, name string) MetaID {
	return e.Metas.FreshNamed(ty, e.ToKernelCtx(), span, name)
}

// FreshMeta creates a metavariable term applied to all local variables.
// This is the standard way to create a metavariable that depends on the context.
func (e *ElabCtx) FreshMeta(ty ast.Term, span Span) ast.Term {
	id := e.Fresh(ty, span)

	// Apply the metavariable to all local variables
	args := make([]ast.Term, len(e.Bindings))
	for i := range e.Bindings {
		args[i] = ast.Var{Ix: len(e.Bindings) - 1 - i}
	}

	return MkMetaApp(id, args...)
}
