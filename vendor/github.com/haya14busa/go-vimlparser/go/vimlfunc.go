package vimlparser

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// copied and little modified(^++) from ./py/vimlfunc.py
var patVim2Go = map[string]string{
	"[0-9a-zA-Z]":                       "[0-9a-zA-Z]",
	"[@*!=><&~#]":                       "[@*!=><&~#]",
	"\\<ARGOPT\\>":                      "\\bARGOPT\\b",
	"\\<BANG\\>":                        "\\bBANG\\b",
	"\\<EDITCMD\\>":                     "\\bEDITCMD\\b",
	"\\<NOTRLCOM\\>":                    "\\bNOTRLCOM\\b",
	"\\<TRLBAR\\>":                      "\\bTRLBAR\\b",
	"\\<USECTRLV\\>":                    "\\bUSECTRLV\\b",
	"\\<USERCMD\\>":                     "\\bUSERCMD\\b",
	"\\<\\(XFILE\\|FILES\\|FILE1\\)\\>": "\\b(XFILE|FILES|FILE1)\\b",
	"\\S":                                      "\\S",
	"\\a":                                      "[A-Za-z]",
	"\\d":                                      "\\d",
	"\\h":                                      "[A-Za-z_]",
	"\\s":                                      "\\s",
	"\\v^d%[elete][lp]$":                       "^d(elete|elet|ele|el|e)[lp]$",
	"\\v^s%(c[^sr][^i][^p]|g|i[^mlg]|I|r[^e])": "^s(c[^sr][^i][^p]|g|i[^mlg]|I|r[^e])",
	"\\w":        "[0-9A-Za-z_]",
	"\\w\\|[:#]": "[0-9A-Za-z_]|[:#]",
	"\\x":        "[0-9A-Fa-f]",
	"^++":        "^\\+\\+",
	"^++bad=\\(keep\\|drop\\|.\\)\\>":                       "^\\+\\+bad=(keep|drop|.)\\b",
	"^++bad=drop":                                           "^\\+\\+bad=drop",
	"^++bad=keep":                                           "^\\+\\+bad=keep",
	"^++bin\\>":                                             "^\\+\\+bin\\b",
	"^++edit\\>":                                            "^\\+\\+edit\\b",
	"^++enc=\\S":                                            "^\\+\\+enc=\\S",
	"^++encoding=\\S":                                       "^\\+\\+encoding=\\S",
	"^++ff=\\(dos\\|unix\\|mac\\)\\>":                       "^\\+\\+ff=(dos|unix|mac)\\b",
	"^++fileformat=\\(dos\\|unix\\|mac\\)\\>":               "^\\+\\+fileformat=(dos|unix|mac)\\b",
	"^++nobin\\>":                                           "^\\+\\+nobin\\b",
	"^[A-Z]":                                                "^[A-Z]",
	"^\\$\\w\\+":                                            "^\\$[0-9A-Za-z_]+",
	"^\\(!\\|global\\|vglobal\\)$":                          "^(!|global|vglobal)$",
	"^\\(WHILE\\|FOR\\)$":                                   "^(WHILE|FOR)$",
	"^\\(vimgrep\\|vimgrepadd\\|lvimgrep\\|lvimgrepadd\\)$": "^(vimgrep|vimgrepadd|lvimgrep|lvimgrepadd)$",
	"^\\d":                     "^\\d",
	"^\\h":                     "^[A-Za-z_]",
	"^\\s":                     "^\\s",
	"^\\s*\\\\":                "^\\s*\\\\",
	"^[ \\t]$":                 "^[ \\t]$",
	"^[A-Za-z]$":               "^[A-Za-z]$",
	"^[0-9A-Za-z]$":            "^[0-9A-Za-z]$",
	"^[0-9]$":                  "^[0-9]$",
	"^[0-9A-Fa-f]$":            "^[0-9A-Fa-f]$",
	"^[0-9A-Za-z_]$":           "^[0-9A-Za-z_]$",
	"^[A-Za-z_]$":              "^[A-Za-z_]$",
	"^[0-9A-Za-z_:#]$":         "^[0-9A-Za-z_:#]$",
	"^[A-Za-z_][0-9A-Za-z_]*$": "^[A-Za-z_][0-9A-Za-z_]*$",
	"^[A-Z]$":                  "^[A-Z]$",
	"^[a-z]$":                  "^[a-z]$",
	"^[vgslabwt]:$\\|^\\([vgslabwt]:\\)\\?[A-Za-z_][0-9A-Za-z_#]*$": "^[vgslabwt]:$|^([vgslabwt]:)?[A-Za-z_][0-9A-Za-z_#]*$",
	"^[0-7]$": "^[0-7]$",
}

var patVim2GoRegh = make(map[string]*regexp.Regexp)

func init() {
	for k, v := range patVim2Go {
		patVim2GoRegh[k] = regexp.MustCompile(v)
	}
}

type vimlList interface{}

func viml_empty(obj interface{}) bool {
	return reflect.ValueOf(obj).Len() == 0
}

func viml_equalci(a, b string) bool {
	return strings.ToLower(a) == strings.ToLower(b)
}

func viml_eqregh(s, reg string) bool {
	if r, ok := patVim2GoRegh[reg]; ok {
		return r.MatchString(s)
	}
	panic("NotImplemented viml_eqregh")
}

func viml_escape(s string, chars string) string {
	r := ""
	for _, c := range s {
		if strings.IndexRune(chars, c) != -1 {
			r += `\` + string(c)
		} else {
			r += string(c)
		}
	}
	return r
}

func viml_join(lst vimlList, sep string) string {
	var ss []string
	s := reflect.ValueOf(lst)
	for i := 0; i < s.Len(); i++ {
		ss = append(ss, fmt.Sprintf("%v", s.Index(i)))
	}
	return strings.Join(ss, sep)
}

func viml_printf(f string, args ...interface{}) string {
	return fmt.Sprintf(f, args...)
}

func viml_range(start, end int) []int {
	var rs []int
	for i := start; i <= end; i++ {
		rs = append(rs, i)
	}
	return rs
}

func viml_split(s string, sep string) []string {
	if sep == `\zs` {
		ss := make([]string, 0, len(s))
		for _, r := range s {
			ss = append(ss, string(r))
		}
		return ss
	}
	panic("NotImplemented viml_split")
}

func viml_str2nr(s string, base int) int {
	r, err := strconv.ParseInt(s, base, 32)
	if err != nil {
		panic(fmt.Errorf("viml_str2nr: %v", err))
	}
	return int(r)
}

func viml_string(obj interface{}) string {
	panic("NotImplemented viml_string")
}

func viml_has_key(obj interface{}, key interface{}) bool {
	// Avoid using reflect as much as possible by listing type used as obj and
	// use type switch.
	switch o := obj.(type) {
	case map[string]*Cmd:
		_, ok := o[key.(string)]
		return ok
	case map[string]interface{}:
		_, ok := o[key.(string)]
		return ok
	case map[int][]interface{}:
		_, ok := o[key.(int)]
		return ok
	}
	// fallback to reflect. Shoul be unreachable.
	m := reflect.ValueOf(obj)
	v := m.MapIndex(reflect.ValueOf(key))
	return v.Kind() != reflect.Invalid
}

func viml_stridx(a, b string) int {
	return strings.Index(a, b)
}
