package ast

// ExArg represents extra arguments of command.
type ExArg struct {
	Forceit    bool
	AddrCount  int
	Line1      int
	Line2      int
	Flags      int
	DoEcmdCmd  string
	DoEcmdLnum int
	Append     int
	Usefilter  bool
	Amount     int
	Regname    int
	ForceBin   int
	ReadEdit   int
	ForceFf    string // int
	ForceEnc   string // int
	BadChar    string // int
	Linepos    *Pos
	Cmdpos     *Pos
	Argpos     *Pos
	Cmd        *Cmd // Ex-command. It's not nil for most case?
	Modifiers  []interface{}
	Range      []interface{}
	Argopt     map[string]interface{}
	Argcmd     map[string]interface{}
}

// Cmd represents command.
type Cmd struct {
	Name   string
	Minlen int
	Flags  string
	Parser string
}
