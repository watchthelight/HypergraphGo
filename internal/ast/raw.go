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
