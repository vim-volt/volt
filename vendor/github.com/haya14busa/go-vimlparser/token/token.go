// Package token defines constants representing the lexical tokens of Vim script.
//
// ref: "go/token"
package token

import "strconv"

// Token is the set of lexical tokens of Vim script.
type Token int

// The list of tokens.
const (
	ILLEGAL Token = iota
	EOF
	EOL
	SPACE
	OROR
	ANDAND
	EQEQ
	EQEQCI
	EQEQCS
	NEQ
	NEQCI
	NEQCS
	GT
	GTCI
	GTCS
	GTEQ
	GTEQCI
	GTEQCS
	LT
	LTCI
	LTCS
	LTEQ
	LTEQCI
	LTEQCS
	MATCH
	MATCHCI
	MATCHCS
	NOMATCH
	NOMATCHCI
	NOMATCHCS
	IS
	ISCI
	ISCS
	ISNOT
	ISNOTCI
	ISNOTCS
	PLUS
	MINUS
	DOT
	STAR
	SLASH
	PERCENT
	NOT
	QUESTION
	COLON
	POPEN
	PCLOSE
	SQOPEN
	SQCLOSE
	COPEN
	CCLOSE
	COMMA
	NUMBER
	SQUOTE
	DQUOTE
	OPTION
	IDENTIFIER
	ENV
	REG
	EQ
	OR
	SEMICOLON
	BACKTICK
	DOTDOTDOT
	SHARP
	ARROW

	STRING // "abc", 'abc'
)

var tokens = [...]string{
	ILLEGAL:    "ILLEGAL",
	EOF:        "<EOF>",
	EOL:        "<EOL>",
	SPACE:      "<SPACE>",
	OROR:       "||",
	ANDAND:     "&&",
	EQEQ:       "==",
	EQEQCI:     "==?",
	EQEQCS:     "==#",
	NEQ:        "!=",
	NEQCI:      "!=?",
	NEQCS:      "!=#",
	GT:         ">",
	GTCI:       ">?",
	GTCS:       ">#",
	GTEQ:       ">=",
	GTEQCI:     ">=?",
	GTEQCS:     ">=#",
	LT:         "<",
	LTCI:       "<?",
	LTCS:       "<#",
	LTEQ:       "<=",
	LTEQCI:     "<=?",
	LTEQCS:     "<=#",
	MATCH:      "=~",
	MATCHCI:    "=~?",
	MATCHCS:    "=~#",
	NOMATCH:    "!~",
	NOMATCHCI:  "!~?",
	NOMATCHCS:  "!~#",
	IS:         "is",
	ISCI:       "is?",
	ISCS:       "is#",
	ISNOT:      "isnot",
	ISNOTCI:    "isnot?",
	ISNOTCS:    "isnot#",
	PLUS:       "+",
	MINUS:      "-",
	DOT:        ".",
	STAR:       "*",
	SLASH:      "/",
	PERCENT:    "%",
	NOT:        "!",
	QUESTION:   "?",
	COLON:      ":",
	POPEN:      "(",
	PCLOSE:     ")",
	SQOPEN:     "[",
	SQCLOSE:    "]",
	COPEN:      "{",
	CCLOSE:     "}",
	COMMA:      ",",
	NUMBER:     "<NUMBER>",
	SQUOTE:     "'",
	DQUOTE:     `"`,
	OPTION:     "<&OPTION>",
	IDENTIFIER: "<IDENTIFIER>",
	ENV:        "<$ENV>",
	REG:        "<@REG>",
	EQ:         "=",
	OR:         "|",
	SEMICOLON:  ";",
	BACKTICK:   "`",
	DOTDOTDOT:  "...",
	SHARP:      "#",
	ARROW:      "->",

	STRING: "<STRING>",
}

// String returns the string corresponding to the token tok.
func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}
