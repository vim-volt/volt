package ops

type funcBase string

func (f *funcBase) String() string {
	return string(*f)
}

func (*funcBase) IsMacro() bool {
	return false
}
