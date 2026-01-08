package elab

import (
	"fmt"
	"sync"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/kernel/ctx"
)

// MetaID uniquely identifies a metavariable.
type MetaID int

// MetaState represents the state of a metavariable.
type MetaState int

const (
	// MetaUnsolved means the metavariable has no solution yet.
	MetaUnsolved MetaState = iota
	// MetaSolved means the metavariable has been unified with a term.
	MetaSolved
	// MetaFrozen means the metavariable should not be solved (e.g., user-specified).
	MetaFrozen
)

func (s MetaState) String() string {
	switch s {
	case MetaUnsolved:
		return "unsolved"
	case MetaSolved:
		return "solved"
	case MetaFrozen:
		return "frozen"
	default:
		return "unknown"
	}
}

// MetaEntry represents a single metavariable and its current state.
type MetaEntry struct {
	ID MetaID // Unique identifier

	// Type is the expected type of this metavariable (in core syntax).
	// This is computed at creation time and should be zonked before use.
	Type ast.Term

	// Ctx is the context in which this metavariable was created.
	// Solutions must only use variables from this context (or weakenings of it).
	Ctx *ctx.Ctx

	// Solution is the term this metavariable was unified with.
	// nil if State != MetaSolved.
	Solution ast.Term

	// State tracks whether this meta is solved, unsolved, or frozen.
	State MetaState

	// Span is the source location where this metavariable was created.
	// Used for error messages.
	Span Span

	// Name is an optional user-provided name for ?foo holes.
	// Empty for anonymous _ holes.
	Name string

	// Dependencies tracks which other metavariables this one depends on.
	// Used for cycle detection during unification.
	Dependencies []MetaID
}

// IsSolved returns true if this metavariable has a solution.
func (e *MetaEntry) IsSolved() bool {
	return e.State == MetaSolved && e.Solution != nil
}


// MetaStore manages all metavariables created during elaboration.
// It is the central authority for metavariable state.
type MetaStore struct {
	mu      sync.RWMutex
	entries map[MetaID]*MetaEntry
	next    MetaID
}

// NewMetaStore creates a new empty metavariable store.
func NewMetaStore() *MetaStore {
	return &MetaStore{
		entries: make(map[MetaID]*MetaEntry),
		next:    0,
	}
}

// Fresh creates a new unsolved metavariable with the given type and context.
func (m *MetaStore) Fresh(ty ast.Term, c *ctx.Ctx, span Span) MetaID {
	return m.FreshNamed(ty, c, span, "")
}

// FreshNamed creates a new unsolved metavariable with a user-provided name.
func (m *MetaStore) FreshNamed(ty ast.Term, c *ctx.Ctx, span Span, name string) MetaID {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := m.next
	m.next++

	// Make a copy of the context to avoid aliasing issues
	var ctxCopy *ctx.Ctx
	if c != nil {
		ctxCopy = &ctx.Ctx{Tele: make([]ctx.Binding, len(c.Tele))}
		copy(ctxCopy.Tele, c.Tele)
	} else {
		ctxCopy = &ctx.Ctx{}
	}

	m.entries[id] = &MetaEntry{
		ID:       id,
		Type:     ty,
		Ctx:      ctxCopy,
		Solution: nil,
		State:    MetaUnsolved,
		Span:     span,
		Name:     name,
	}

	return id
}

// Lookup retrieves a metavariable entry by ID.
func (m *MetaStore) Lookup(id MetaID) (*MetaEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.entries[id]
	return entry, ok
}

// MustLookup retrieves a metavariable entry by ID, panicking if not found.
func (m *MetaStore) MustLookup(id MetaID) *MetaEntry {
	entry, ok := m.Lookup(id)
	if !ok {
		panic(fmt.Sprintf("metavariable ?%d not found", id))
	}
	return entry
}

// Solve records a solution for a metavariable.
// Returns an error if the metavariable is already solved or frozen.
func (m *MetaStore) Solve(id MetaID, solution ast.Term) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.entries[id]
	if !ok {
		return fmt.Errorf("metavariable ?%d not found", id)
	}

	switch entry.State {
	case MetaSolved:
		return fmt.Errorf("metavariable ?%d already solved", id)
	case MetaFrozen:
		return fmt.Errorf("metavariable ?%d is frozen", id)
	}

	entry.Solution = solution
	entry.State = MetaSolved
	return nil
}

// TrySolve attempts to solve a metavariable, returning false if already solved.
// Does not return an error for already-solved metas (useful for unification).
func (m *MetaStore) TrySolve(id MetaID, solution ast.Term) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.entries[id]
	if !ok {
		return false
	}

	if entry.State != MetaUnsolved {
		return false
	}

	entry.Solution = solution
	entry.State = MetaSolved
	return true
}

// GetSolution returns the solution for a metavariable if it's solved.
func (m *MetaStore) GetSolution(id MetaID) (ast.Term, bool) {
	entry, ok := m.Lookup(id)
	if !ok || !entry.IsSolved() {
		return nil, false
	}
	return entry.Solution, true
}

// IsSolved returns true if the metavariable is solved.
func (m *MetaStore) IsSolved(id MetaID) bool {
	entry, ok := m.Lookup(id)
	return ok && entry.IsSolved()
}

// Freeze marks a metavariable as frozen (should not be solved).
func (m *MetaStore) Freeze(id MetaID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.entries[id]
	if !ok {
		return fmt.Errorf("metavariable ?%d not found", id)
	}

	if entry.State == MetaSolved {
		return fmt.Errorf("cannot freeze solved metavariable ?%d", id)
	}

	entry.State = MetaFrozen
	return nil
}

// Unsolved returns all unsolved metavariable IDs.
func (m *MetaStore) Unsolved() []MetaID {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []MetaID
	for id, entry := range m.entries {
		if entry.State == MetaUnsolved {
			result = append(result, id)
		}
	}
	return result
}

// AllSolved returns true if all metavariables have been solved.
func (m *MetaStore) AllSolved() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, entry := range m.entries {
		if entry.State == MetaUnsolved {
			return false
		}
	}
	return true
}

// Size returns the number of metavariables in the store.
func (m *MetaStore) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}

// Clone creates a copy of the metastore (useful for speculative solving).
func (m *MetaStore) Clone() *MetaStore {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clone := &MetaStore{
		entries: make(map[MetaID]*MetaEntry, len(m.entries)),
		next:    m.next,
	}

	for id, entry := range m.entries {
		entryCopy := *entry
		if entry.Ctx != nil {
			ctxCopy := &ctx.Ctx{Tele: make([]ctx.Binding, len(entry.Ctx.Tele))}
			copy(ctxCopy.Tele, entry.Ctx.Tele)
			entryCopy.Ctx = ctxCopy
		}
		clone.entries[id] = &entryCopy
	}

	return clone
}

// FormatMeta returns a string representation of a metavariable for debugging.
func (m *MetaStore) FormatMeta(id MetaID) string {
	entry, ok := m.Lookup(id)
	if !ok {
		return fmt.Sprintf("?%d[unknown]", id)
	}

	if entry.Name != "" {
		return fmt.Sprintf("?%s", entry.Name)
	}
	return fmt.Sprintf("?%d", id)
}

// Debug returns a string representation of all metavariables for debugging.
func (m *MetaStore) Debug() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result string
	for id, entry := range m.entries {
		name := fmt.Sprintf("?%d", id)
		if entry.Name != "" {
			name = fmt.Sprintf("?%s(%d)", entry.Name, id)
		}

		status := entry.State.String()
		if entry.IsSolved() {
			status = fmt.Sprintf("solved = %v", entry.Solution)
		}

		result += fmt.Sprintf("%s : %v [%s]\n", name, entry.Type, status)
	}
	return result
}

// MkMeta creates a metavariable term with no arguments.
func MkMeta(id MetaID) ast.Meta {
	return ast.Meta{ID: int(id), Args: nil}
}

// MkMetaApp creates a metavariable term applied to arguments.
func MkMetaApp(id MetaID, args ...ast.Term) ast.Meta {
	return ast.Meta{ID: int(id), Args: args}
}
