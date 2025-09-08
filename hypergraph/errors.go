package hypergraph

import "errors"

var (
    // ErrDuplicateEdge is returned when adding an edge with an existing ID.
    ErrDuplicateEdge = errors.New("duplicate edge ID")
    // ErrUnknownEdge is returned when an operation references a missing edge.
    ErrUnknownEdge = errors.New("unknown edge ID")
    // ErrUnknownVertex is returned when an operation references a missing vertex.
    ErrUnknownVertex = errors.New("unknown vertex")
    // ErrCutoff indicates an algorithm terminated early due to a configured cutoff.
    ErrCutoff = errors.New("operation cutoff reached")
)
