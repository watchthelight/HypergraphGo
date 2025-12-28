package hypergraph

import "errors"

var (
	// ErrDuplicateEdge is returned when adding an edge with an existing ID.
	ErrDuplicateEdge = errors.New("duplicate edge ID")
	// ErrCutoff indicates an algorithm terminated early due to a configured cutoff.
	ErrCutoff = errors.New("operation cutoff reached")
)
