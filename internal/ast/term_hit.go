package ast

// Higher Inductive Types (HITs) extend inductive types with path constructors.
// Path constructors produce paths between elements, not just points.
//
// Example: Circle S¹
//   - Point constructor: base : S¹
//   - Path constructor: loop : Path S¹ base base
//
// Example: Suspension Susp A
//   - Point constructors: north, south : Susp A
//   - Path constructor: merid : A → Path (Susp A) north south

// Boundary specifies what a path constructor reduces to at interval endpoints.
// For a path constructor like `loop : Path S1 base base`:
//
//	AtZero = base (value at i=0)
//	AtOne  = base (value at i=1)
type Boundary struct {
	AtZero Term // Value when interval variable = i0
	AtOne  Term // Value when interval variable = i1
}

// PathConstructor represents a path-level constructor in a HIT.
//
// Level indicates the dimension of the path:
//   - Level 1: path between points (binds 1 interval variable)
//   - Level 2: path between paths (binds 2 interval variables)
//   - etc.
//
// Example for Circle:
//
//	PathConstructor{
//	    Name: "loop",
//	    Level: 1,
//	    Type: Path S1 base base,
//	    Boundaries: []Boundary{{AtZero: base, AtOne: base}},
//	}
type PathConstructor struct {
	Name       string     // Constructor name (e.g., "loop")
	Level      int        // Dimension: 1=path, 2=path-of-path, etc.
	Type       Term       // Full type (e.g., Path S1 base base)
	Boundaries []Boundary // One Boundary per interval dimension
}

// HITApp represents application of a HIT path constructor to interval arguments.
// When applied to interval endpoints (i0 or i1), it computes to boundary values:
//
//	loop @ i0 --> base
//	loop @ i1 --> base
//
// When applied to an interval variable, it remains stuck until eliminated.
type HITApp struct {
	HITName string // HIT type name (e.g., "S1")
	Ctor    string // Path constructor name (e.g., "loop")
	Args    []Term // Type/term parameters applied to constructor
	IArgs   []Term // Interval arguments (I0, I1, or IVar)
}

func (HITApp) isCoreTerm() {}

// HITSpec represents a Higher Inductive Type specification.
// Used for declaring and validating HITs.
type HITSpec struct {
	Name       string            // Type name (e.g., "S1", "Trunc")
	Type       Term              // Full type signature including parameters
	NumParams  int               // Number of type parameters
	ParamTypes []Term            // Types of parameters
	PointCtors []Constructor     // Point constructors (level 0)
	PathCtors  []PathConstructor // Path constructors (level >= 1)
	Eliminator string            // Eliminator name (e.g., "S1-elim")
}

// Constructor represents a point constructor in an inductive type.
// This mirrors the existing Constructor in env.go but is included here
// for completeness in HIT specifications.
type Constructor struct {
	Name string
	Type Term
}

// IsHIT returns true if the specification has path constructors.
func (s *HITSpec) IsHIT() bool {
	return len(s.PathCtors) > 0
}

// MaxLevel returns the maximum constructor level in the HIT.
// Returns 0 for ordinary inductive types, 1 for 1-HITs, 2 for 2-HITs, etc.
func (s *HITSpec) MaxLevel() int {
	max := 0
	for _, pc := range s.PathCtors {
		if pc.Level > max {
			max = pc.Level
		}
	}
	return max
}
