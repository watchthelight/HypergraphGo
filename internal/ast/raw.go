package ast

// Raw terms carry user-chosen names before resolution.
type RTerm interface{ isRTerm() }

type RVar struct{ Name string } // free or bound
func (RVar) isRTerm()           {}

type RGlobal struct{ Name string } // global symbol
func (RGlobal) isRTerm()           {}

type RSort struct{ U Level } // Type^U
func (RSort) isRTerm()       {}

type RPi struct {
	Binder string
	A, B   RTerm
}

func (RPi) isRTerm() {}

type RLam struct {
	Binder string
	Ann    RTerm
	Body   RTerm
}                     // Ann may be nil
func (RLam) isRTerm() {}

type RApp struct{ T, U RTerm }

func (RApp) isRTerm() {}

type RSigma struct {
	Binder string
	A, B   RTerm
}

func (RSigma) isRTerm() {}

type RPair struct{ Fst, Snd RTerm }

func (RPair) isRTerm() {}

type RFst struct{ P RTerm }

func (RFst) isRTerm() {}

type RSnd struct{ P RTerm }

func (RSnd) isRTerm() {}

type RLet struct {
	Binder         string
	Ann, Val, Body RTerm
}                     // Ann may be nil
func (RLet) isRTerm() {}

// RId is the raw identity type: Id A x y
type RId struct {
	A RTerm // Type
	X RTerm // Left endpoint
	Y RTerm // Right endpoint
}

func (RId) isRTerm() {}

// RRefl is the raw reflexivity constructor: refl A x
type RRefl struct {
	A RTerm // Type
	X RTerm // The term being proven equal to itself
}

func (RRefl) isRTerm() {}

// RJ is the raw identity eliminator (path induction): J A C d x y p
type RJ struct {
	A RTerm // Type
	C RTerm // Motive
	D RTerm // Base case
	X RTerm // Left endpoint
	Y RTerm // Right endpoint
	P RTerm // Proof
}

func (RJ) isRTerm() {}
