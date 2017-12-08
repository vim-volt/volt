package vimlparser

import (
	"reflect"
	"testing"
)

func TestViml_eqregh(t *testing.T) {
	tests := []struct {
		in   string
		reg  string
		want bool
	}{
		{in: `deletel`, reg: "\\v^d%[elete][lp]$", want: true},
		{in: `deleteL`, reg: "\\v^d%[elete][lp]$", want: false},
		{in: `++bad=keep`, reg: "^++bad=keep", want: true},
		{in: `++bad=KEEP`, reg: "^++bad=keep", want: false},
	}
	for _, tt := range tests {
		if got := viml_eqregh(tt.in, tt.reg); got != tt.want {
			t.Errorf("viml_eqregh(%q, %q) = %v, want %v", tt.in, tt.reg, got, tt.want)
		}
	}
}

func TestViml_printf(t *testing.T) {
	tests := []struct {
		f    string
		args []interface{}
		want string
	}{
		{"hoge%s", []interface{}{"foo"}, "hogefoo"},
	}
	for _, tt := range tests {
		if got := viml_printf(tt.f, tt.args...); got != tt.want {
			t.Errorf("viml_printf(%q, %v) = %v, want %v", tt.f, tt.args, got, tt.want)
		}
	}
}

func TestViml_stridx(t *testing.T) {
	tests := []struct {
		heystack string
		needle   string
		want     int
	}{
		{"hoge", "", 0},
		{"hoge", "oge", 1},
		{"hoge", "xxx", -1},
		{"hoge", "x", -1},
		{"hoge", "hogehogehog", -1},
		{"", "hogehogehog", -1},
		{"An Example", "Example", 3},
	}

	for _, tt := range tests {
		if got := viml_stridx(tt.heystack, tt.needle); got != tt.want {
			t.Errorf("viml_stridx(%q, %q) = %v\nVim.stridx(%q, %q) = %v",
				tt.heystack, tt.needle, got, tt.heystack, tt.needle, tt.want)
		}
	}
}

func TestViml_has_key(t *testing.T) {
	tests := []struct {
		m    interface{}
		k    interface{}
		want bool
	}{
		{m: map[string]string{"a": "a"}, k: "a", want: true},
		{m: map[string]string{"a": "a"}, k: "", want: false},
		{m: map[string]string{"a": "a"}, k: "b", want: false},

		{m: map[string]int{"a": 1}, k: "a", want: true},
		{m: map[string]int{"a": 1}, k: "b", want: false},

		{m: map[int]int{1: 1}, k: 1, want: true},
		{m: map[int]int{1: 1}, k: 2, want: false},
	}

	for _, tt := range tests {
		got := viml_has_key(tt.m, tt.k)
		if got != tt.want {
			t.Errorf("viml_has_key(%v, %v) = %v, want %v", tt.m, tt.k, got, tt.want)
		}
	}
}

func TestViml_join(t *testing.T) {
	tests := []struct {
		lst  interface{}
		sep  string
		want string
	}{
		{[]interface{}{"foo", "bar"}, ":", "foo:bar"},
		{[]interface{}{}, ":", ""},
		{[]interface{}{}, "", ""},
		{[]interface{}{"foo", "bar"}, "", "foobar"},
	}
	for _, tt := range tests {
		if got := viml_join(tt.lst, tt.sep); got != tt.want {
			t.Errorf("viml_join(%q, %q) = %v, want %v", tt.lst, tt.sep, got, tt.want)
		}
	}
}

func TestViml_range(t *testing.T) {
	tests := []struct {
		start int
		end   int
		want  []int
	}{
		{0, 3, []int{0, 1, 2, 3}},
		{0, 0, []int{0}},
	}
	for _, tt := range tests {
		if got := viml_range(tt.start, tt.end); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("viml_range(%v, %v) = %v, want %v", tt.start, tt.end, got, tt.want)
		}
	}
}

func TestViml_escape(t *testing.T) {
	tests := []struct {
		str   string
		chars string
		want  string
	}{
		{"hoge", "og", `h\o\ge`},
	}
	for _, tt := range tests {
		if got := viml_escape(tt.str, tt.chars); got != tt.want {
			t.Errorf("viml_escape(%v, %v) = %v, want %v", tt.str, tt.chars, got, tt.want)
		}
	}
}
