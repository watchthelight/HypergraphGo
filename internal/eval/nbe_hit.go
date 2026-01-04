package eval

// Higher Inductive Types (HITs) evaluation support.
//
// HITs extend inductive types with path constructors. When a path constructor
// is applied to interval arguments, it may compute to boundary values at
// endpoints (i0 or i1) or remain stuck at interval variables.

// VHITPathCtor represents a path constructor applied to interval argument(s).
// This value reduces to boundary values when applied at endpoints:
//
//	loop @ i0 --> base
//	loop @ i1 --> base
//
// When applied to an interval variable, it remains stuck until eliminated.
type VHITPathCtor struct {
	HITName    string        // HIT type name (e.g., "S1")
	CtorName   string        // Path constructor name (e.g., "loop")
	Args       []Value       // Type/term parameters applied to constructor
	IArgs      []Value       // Interval arguments (VI0, VI1, or VIVar)
	Boundaries []BoundaryVal // Boundary values at endpoints
}

func (VHITPathCtor) isValue() {}

// BoundaryVal stores evaluated boundary values for a path constructor.
type BoundaryVal struct {
	AtZero Value // Value when interval = i0
	AtOne  Value // Value when interval = i1
}

// evalHITApp evaluates a HIT path constructor application.
// When all interval arguments are endpoints, it computes to the boundary value.
func evalHITApp(hitName, ctorName string, args []Value, iargs []Value, boundaries []BoundaryVal) Value {
	// Check if any interval arg is an endpoint
	for dim, iarg := range iargs {
		if dim >= len(boundaries) {
			// No boundary info for this dimension, stay stuck
			break
		}
		switch iarg.(type) {
		case VI0:
			// Reduce to AtZero boundary at this dimension
			return boundaries[dim].AtZero
		case VI1:
			// Reduce to AtOne boundary at this dimension
			return boundaries[dim].AtOne
		}
		// VIVar or other stuck interval - continue checking
	}

	// Not reduced - return stuck value
	return VHITPathCtor{
		HITName:    hitName,
		CtorName:   ctorName,
		Args:       args,
		IArgs:      iargs,
		Boundaries: boundaries,
	}
}

// lookupHITBoundaries retrieves boundary values for a HIT path constructor.
// This looks up the HIT info from the recursor registry and evaluates boundaries.
func lookupHITBoundaries(hitName, ctorName string, _ []Value) []BoundaryVal {
	// Look up HIT recursor info
	info := LookupRecursor(hitName + "-elim")
	if info == nil || !info.IsHIT {
		return nil
	}

	// Find the path constructor
	for _, pc := range info.PathCtors {
		if pc.Name == ctorName {
			// Build boundary values from boundary specs
			boundaries := make([]BoundaryVal, len(pc.Boundaries))
			for i, bspec := range pc.Boundaries {
				boundaries[i] = BoundaryVal{
					AtZero: bspec.AtZeroVal,
					AtOne:  bspec.AtOneVal,
				}
			}
			return boundaries
		}
	}

	return nil
}
