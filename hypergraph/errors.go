package hypergraph

import "errors"

// Exported package errors.
var (
	ErrDuplicateEdge  = errors.New("duplicate edge ID")
	ErrUnknownEdge    = errors.New("unknown edge ID")
	ErrUnknownVertex  = errors.New("unknown vertex")
	ErrCutoff         = errors.New("operation cutoff reached")
)
