var x = 1
func Funcname() {
	var y = 1
	y = 2
	x = 1
}

func Funcname(z) {
	z = z
}

func (self *VimLParser) hoge(a) {
	a = a
}

var b, c = d
b, c = d
node.pattern, b = hoge
node.pattern, _ = hoge
var e = 1
if f {
	e = g
	var h = 0
	if i {
		h = 1
	}
}
var xs = 1
if x {
	xs[0] = 1
}
func F() {
	var cmd *Cmd = nil
	for _, x := range builtin_commands {
		if viml_stridx(x.name, name) == 0 && len(name) >= x.minlen {
			cmd = x
			break
		}
	}
	if self.neovim {
		for _, x := range neovim_additional_commands {
			if viml_stridx(x.name, name) == 0 && len(name) >= x.minlen {
				cmd = x
				break
			}
		}
		for _, x := range neovim_removed_commands {
			if viml_stridx(x.name, name) == 0 && len(name) >= x.minlen {
				cmd = nil
				break
			}
		}
	}
}

