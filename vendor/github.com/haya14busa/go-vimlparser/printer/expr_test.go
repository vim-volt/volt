package printer

import (
	"bytes"
	"strings"
	"testing"

	vimlparser "github.com/haya14busa/go-vimlparser"
	"github.com/haya14busa/go-vimlparser/ast"
	"github.com/haya14busa/go-vimlparser/token"
)

func TestFprint_expr(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: `xyz`, want: `xyz`},                        // Ident
		{in: `"double quote"`, want: `"double quote"`},  // BasicLit
		{in: `14`, want: `14`},                          // BasicLit
		{in: `+1`, want: `+1`},                          // UnaryExpr
		{in: `-  1`, want: `-1`},                        // UnaryExpr
		{in: `! + - 1`, want: `!+-1`},                   // UnaryExpr
		{in: `x+1`, want: `x + 1`},                      // BinaryExpr
		{in: `1+2*3`, want: `1 + 2 * 3`},                // BinaryExpr
		{in: `1*2+3`, want: `1 * 2 + 3`},                // BinaryExpr
		{in: `(1+2)*(3-4)`, want: `(1 + 2) * (3 - 4)`},  // ParenExpr
		{in: `1+(2*3)`, want: `1 + (2 * 3)`},            // ParenExpr
		{in: `(((x+(1))))`, want: `(x + (1))`},          // ParenExpr
		{in: `x+1==14 ||-1`, want: `x + 1 == 14 || -1`}, // BinaryExpr
		{in: `x[ y ]`, want: `x[y]`},                    // SubscriptExpr
	}

	for _, tt := range tests {
		r := strings.NewReader(tt.in)
		node, err := vimlparser.ParseExpr(r)
		if err != nil {
			t.Fatal(err)
		}
		buf := new(bytes.Buffer)
		if err := Fprint(buf, node, nil); err != nil {
			if !tt.wantErr {
				t.Errorf("got unexpected error: %v", err)
			}
			continue
		}
		if got := buf.String(); got != tt.want {
			t.Errorf("got: %v, want: %v", got, tt.want)
		}
	}
}

func TestFprint_expr_insert_paren_to_binary(t *testing.T) {
	want := `(x + y) * z`
	buf := new(bytes.Buffer)
	left := &ast.BinaryExpr{
		Left:  &ast.Ident{Name: "x"},
		Op:    token.PLUS,
		Right: &ast.Ident{Name: "y"},
	}
	node := &ast.BinaryExpr{
		Left:  left,
		Op:    token.STAR,
		Right: &ast.Ident{Name: "z"},
	}
	if err := Fprint(buf, node, nil); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFprint_expr_insert_paren_to_unary(t *testing.T) {
	want := `(-x)[y]`
	buf := new(bytes.Buffer)
	node := &ast.SubscriptExpr{
		Left:  &ast.UnaryExpr{Op: token.MINUS, X: &ast.Ident{Name: "x"}},
		Right: &ast.Ident{Name: "y"},
	}
	if err := Fprint(buf, node, nil); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
