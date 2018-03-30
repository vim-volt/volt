package plugconf

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/vim-volt/volt/httputil"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/pathutil"

	"github.com/haya14busa/go-vimlparser"
	"github.com/haya14busa/go-vimlparser/ast"
	"github.com/haya14busa/go-vimlparser/token"
)

type loadOnType string

// TODO: make this uint
const (
	loadOnStart    loadOnType = "(loadOnStart)"
	loadOnFileType            = "FileType"
	loadOnExcmd               = "(loadOnExcmd)"
)

const (
	// TODO: Check duplicate variable for excmdLoadPlugin
	excmdLoadPlugin   = "s:__volt_excmd_load_plugin"
	lazyLoadExcmdFunc = "s:__volt_lazy_load_excmd"
	completeFunc      = "s:__volt_complete"
)

func isProhibitedFuncName(name string) bool {
	return name == lazyLoadExcmdFunc ||
		name == completeFunc
}

type Plugconf struct {
	reposID        int
	reposPath      pathutil.ReposPath
	functions      []string
	onLoadPreFunc  string
	onLoadPostFunc string
	loadOnFunc     string
	loadOn         loadOnType
	loadOnArg      string
	dependsFunc    string
	depends        pathutil.ReposPathList
}

// This type does not provide Error() because I don't want let it pretend like
// error type. Receivers of a value of this type must decide how to handle.
type ParseError struct {
	filename string
	Errs     *multierror.Error
	Warns    *multierror.Error
}

func newParseError(filename string) *ParseError {
	var e ParseError
	e.Errs = newParseErrorMultiError("parse errors in ", filename)
	e.Warns = newParseErrorMultiError("parse warnings in ", filename)
	return &e
}

func newParseErrorMultiError(prefix, filename string) *multierror.Error {
	return &multierror.Error{
		Errors: make([]error, 0, 8),
		ErrorFormat: func(errs []error) string {
			var buf bytes.Buffer
			buf.WriteString(prefix)
			buf.WriteString(filename)
			buf.WriteString(":")
			for _, e := range errs {
				buf.WriteString("\n* ")
				buf.WriteString(e.Error())
			}
			return buf.String()
		},
	}
}

func (e *ParseError) HasErrsOrWarns() bool {
	return e != nil && (e.Errs.ErrorOrNil() != nil || e.Warns.ErrorOrNil() != nil)
}

func (e *ParseError) HasErrs() bool {
	return e != nil && e.Errs.ErrorOrNil() != nil
}

func (e *ParseError) HasWarns() bool {
	return e != nil && e.Warns.ErrorOrNil() != nil
}

func (e *ParseError) ErrorsAndWarns() *multierror.Error {
	if e == nil {
		return nil
	}
	var result *multierror.Error
	if e.Errs != nil {
		result = multierror.Append(result, e.Errs.Errors...)
	}
	if e.Warns != nil {
		result = multierror.Append(result, e.Warns.Errors...)
	}
	return result
}

func (e *ParseError) merge(e2 *ParseError) {
	if e == nil || e2 == nil {
		return
	}
	if e2.Errs != nil {
		e.Errs = multierror.Append(e.Errs, e2.Errs.Errors...)
	}
	if e2.Warns != nil {
		e.Warns = multierror.Append(e.Warns, e2.Warns.Errors...)
	}
}

type MultiParseError []ParseError

func (errs MultiParseError) HasErrs() bool {
	for _, e := range errs {
		if e.HasErrs() {
			return true
		}
	}
	return false
}

func (errs MultiParseError) HasWarns() bool {
	for _, e := range errs {
		if e.HasWarns() {
			return true
		}
	}
	return false
}

func (errs MultiParseError) Errors() *multierror.Error {
	return errs.concatErrors(false)
}

func (errs MultiParseError) Warns() *multierror.Error {
	return errs.concatErrors(true)
}

func (errs MultiParseError) concatErrors(showWarns bool) *multierror.Error {
	result := &multierror.Error{
		Errors: make([]error, 0, len(errs)),
		ErrorFormat: func(errs []error) string {
			var buf bytes.Buffer
			for _, e := range errs {
				buf.WriteString(e.Error())
				buf.WriteString("\n")
			}
			return buf.String()
		},
	}
	for _, e := range errs {
		merr := e.Errs
		if showWarns {
			merr = e.Warns
		}
		if merr != nil {
			// Call merr.Error() to apply error format func
			result = multierror.Append(result, errors.New(merr.Error()))
		}
	}
	return result
}

func (errs MultiParseError) ErrorsAndWarns() *multierror.Error {
	var result *multierror.Error
	for _, e := range errs {
		merr := e.ErrorsAndWarns()
		if merr != nil {
			result = multierror.Append(result, merr.Errors...)
		}
	}
	return result
}

func ParsePlugconfFile(plugConf string, reposID int, reposPath pathutil.ReposPath) (result *Plugconf, parseErr *ParseError) {
	// this function always returns non-nil parseErr
	// (which may have empty errors / warns)
	parseErr = new(ParseError)

	content, err := ioutil.ReadFile(plugConf)
	if err != nil {
		err = multierror.Append(nil, err)
		return
	}
	src := string(content)
	file, err := vimlparser.ParseFile(strings.NewReader(src), plugConf, nil)
	if err != nil {
		err = multierror.Append(nil, err)
		return
	}
	result, parseErr = ParsePlugconf(file, src, plugConf)
	if result != nil {
		result.reposID = reposID
		result.reposPath = reposPath
	}
	return
}

// this function always returns non-nil parseErr
// (which may have empty errors / warns)
func ParsePlugconf(file *ast.File, src, filename string) (*Plugconf, *ParseError) {
	var loadOn loadOnType = loadOnStart
	var loadOnArg string
	var loadOnFunc string
	var onLoadPreFunc string
	var onLoadPostFunc string
	var functions []string
	var dependsFunc string
	var depends pathutil.ReposPathList

	parseErr := newParseError(filename)

	// Inspect nodes and get above values from plugconf script
	ast.Inspect(file, func(node ast.Node) bool {
		// Cast to function node (return if it's not a function node)
		var fn *ast.Function
		if f, ok := node.(*ast.Function); !ok {
			return true
		} else {
			fn = f
		}

		// Get function name
		var name string
		if ident, ok := fn.Name.(*ast.Ident); !ok {
			return true
		} else {
			name = ident.Name
		}

		switch {
		case name == "s:loaded_on":
			if loadOnFunc != "" {
				parseErr.Errs = multierror.Append(parseErr.Errs,
					errors.New("duplicate s:loaded_on()"))
				return true
			}
			if !isEmptyFunc(fn) {
				loadOnFunc = extractBody(fn, src)
				var err error
				loadOn, loadOnArg, err = inspectReturnValue(fn)
				if err != nil {
					parseErr.Errs = multierror.Append(parseErr.Errs, err)
				}
			}
		case name == "s:config":
			if onLoadPreFunc != "" {
				parseErr.Errs = multierror.Append(parseErr.Errs,
					errors.New("duplicate s:on_load_pre() and s:config()"))
				return true
			}
			parseErr.Warns = multierror.Append(parseErr.Warns,
				errors.New("s:config() is deprecated. please use s:on_load_pre() instead"))
			if !isEmptyFunc(fn) {
				onLoadPreFunc = extractBody(fn, src)
				onLoadPreFunc = rxFuncName.ReplaceAllString(
					onLoadPreFunc, "${1}on_load_pre",
				)
			}
		case name == "s:on_load_pre":
			if onLoadPreFunc != "" {
				parseErr.Errs = multierror.Append(parseErr.Errs,
					errors.New("duplicate s:on_load_pre() and s:config()"))
				return true
			}
			if !isEmptyFunc(fn) {
				onLoadPreFunc = extractBody(fn, src)
			}
		case name == "s:on_load_post":
			if onLoadPostFunc != "" {
				parseErr.Errs = multierror.Append(parseErr.Errs,
					errors.New("duplicate s:on_load_post()"))
				return true
			}
			if !isEmptyFunc(fn) {
				onLoadPostFunc = extractBody(fn, src)
			}
		case name == "s:depends":
			if dependsFunc != "" {
				parseErr.Errs = multierror.Append(parseErr.Errs,
					errors.New("duplicate s:depends()"))
				return true
			}
			if !isEmptyFunc(fn) {
				dependsFunc = extractBody(fn, src)
				var err error
				depends, err = getDependencies(fn, src)
				if err != nil {
					parseErr.Errs = multierror.Append(parseErr.Errs, err)
				}
			}
		case isProhibitedFuncName(name):
			parseErr.Errs = multierror.Append(parseErr.Errs,
				fmt.Errorf(
					"'%s' is prohibited function name. Please use other function name.", name))
		default:
			functions = append(functions, extractBody(fn, src))
		}

		return true
	})

	if parseErr.HasErrs() {
		return nil, parseErr
	}

	return &Plugconf{
		functions:      functions,
		onLoadPreFunc:  onLoadPreFunc,
		onLoadPostFunc: onLoadPostFunc,
		loadOnFunc:     loadOnFunc,
		loadOn:         loadOn,
		loadOnArg:      loadOnArg,
		dependsFunc:    dependsFunc,
		depends:        depends,
	}, parseErr
}

// Inspect return value of s:loaded_on() function in plugconf
func inspectReturnValue(fn *ast.Function) (loadOnType, string, error) {
	var loadOn loadOnType
	var loadOnArg string
	var err error
	ast.Inspect(fn, func(node ast.Node) bool {
		// Cast to return node (return if it's not a return node)
		var ret *ast.Return
		if r, ok := node.(*ast.Return); !ok {
			return true
		} else {
			ret = r
		}

		// Parse the argument of :return
		rhs, ok := ret.Result.(*ast.BasicLit)
		if ok && rhs.Kind == token.STRING {
			value := rhs.Value[1 : len(rhs.Value)-1]
			if value == "start" {
				loadOn = loadOnStart
			} else if strings.HasPrefix(value, "filetype=") {
				loadOn = loadOnFileType
				loadOnArg = strings.TrimPrefix(value, "filetype=")
			} else if strings.HasPrefix(value, "excmd=") {
				loadOn = loadOnExcmd
				loadOnArg = strings.TrimPrefix(value, "excmd=")
			} else {
				err = errors.New("Invalid rhs of ':return': " + rhs.Value)
			}
		}

		return true
	})
	if string(loadOn) == "" {
		return "", "", errors.New("can't detect return value of s:loaded_on()")
	}
	return loadOn, loadOnArg, err
}

// Returns true if fn.Body is empty or has only comment nodes
func isEmptyFunc(fn *ast.Function) bool {
	for i := range fn.Body {
		empty := true
		ast.Inspect(fn.Body[i], func(node ast.Node) bool {
			if _, ok := node.(*ast.Comment); ok {
				return true
			}
			empty = false
			return false
		})
		if !empty {
			return false
		}
	}
	return true
}

func extractBody(fn *ast.Function, src string) string {
	pos := fn.Pos()

	endpos := fn.EndFunction.Pos()
	endfunc := fn.EndFunction.ExArg
	cmdlen := endfunc.Argpos.Offset - endfunc.Cmdpos.Offset
	endpos.Offset += cmdlen

	return src[pos.Offset:endpos.Offset]
}

func getDependencies(fn *ast.Function, src string) (pathutil.ReposPathList, error) {
	var deps pathutil.ReposPathList
	var parseErr error

	ast.Inspect(fn, func(node ast.Node) bool {
		// Cast to return node (return if it's not a return node)
		var ret *ast.Return
		if r, ok := node.(*ast.Return); !ok {
			return true
		} else {
			ret = r
		}
		if list, ok := ret.Result.(*ast.List); ok {
			for i := range list.Values {
				if str, ok := list.Values[i].(*ast.BasicLit); ok {
					if deps == nil {
						deps = make(pathutil.ReposPathList, 0, len(list.Values))
					}
					if str.Kind == token.STRING {
						reposPath, err := pathutil.NormalizeRepos(str.Value[1 : len(str.Value)-1])
						if err != nil {
							parseErr = err
							return false
						}
						deps = append(deps, reposPath)
					}
				}
			}
		}
		return true
	})

	return deps, parseErr
}

// s:loaded_on() function is not included
func makeBundledPlugconf(reposList []lockjson.Repos, plugconf map[pathutil.ReposPath]*Plugconf, vimrcPath, gvimrcPath string) ([]byte, error) {
	functions := make([]string, 0, 64)
	loadCmds := make([]string, 0, len(reposList))
	lazyExcmd := make(map[string]string, len(reposList))

	for _, repos := range reposList {
		p, hasPlugconf := plugconf[repos.Path]
		// :packadd <repos>
		optName := filepath.Base(pathutil.EncodeReposPath(repos.Path))
		packadd := fmt.Sprintf("packadd %s", optName)

		// s:on_load_pre(), invoked command, s:on_load_post()
		var invokedCmd string
		if hasPlugconf {
			cmds := make([]string, 0, 3)
			if p.onLoadPreFunc != "" {
				functions = append(functions, convertToDecodableFunc(p.onLoadPreFunc, p.reposPath, p.reposID))
				cmds = append(cmds, fmt.Sprintf("call s:on_load_pre_%d()", p.reposID))
			}
			cmds = append(cmds, packadd)
			if p.onLoadPostFunc != "" {
				functions = append(functions, convertToDecodableFunc(p.onLoadPostFunc, p.reposPath, p.reposID))
				cmds = append(cmds, fmt.Sprintf("call s:on_load_post_%d()", p.reposID))
			}
			invokedCmd = strings.Join(cmds, " | ")
		} else {
			invokedCmd = packadd
		}

		// Bootstrap statements
		switch {
		case !hasPlugconf || p.loadOn == loadOnStart:
			loadCmds = append(loadCmds, "  "+invokedCmd)
		case p.loadOn == loadOnFileType:
			loadCmds = append(loadCmds,
				fmt.Sprintf("  autocmd %s %s %s", string(p.loadOn), p.loadOnArg, invokedCmd))
		case p.loadOn == loadOnExcmd:
			// Define dummy Ex commands
			for _, excmd := range strings.Split(p.loadOnArg, ",") {
				lazyExcmd[excmd] = invokedCmd
				loadCmds = append(loadCmds,
					fmt.Sprintf("  command -complete=customlist,%[1]s -bang -bar -range -nargs=* %[3]s call %[2]s('%[3]s', <q-args>, expand('<bang>'), expand('<line1>'), expand('<line2>'))", completeFunc, lazyLoadExcmdFunc, excmd))
			}
		}

		// User defined functions in plugconf
		if hasPlugconf {
			functions = append(functions, p.functions...)
		}
	}

	var buf bytes.Buffer
	buf.WriteString(`if exists('g:loaded_volt_system_bundled_plugconf')
  finish
endif
let g:loaded_volt_system_bundled_plugconf = 1`)
	if len(functions) > 0 {
		buf.WriteString("\n\n")
		buf.WriteString(strings.Join(functions, "\n\n"))
	}
	if len(lazyExcmd) > 0 {
		lazyExcmdJSON, err := json.Marshal(lazyExcmd)
		if err != nil {
			return nil, err
		}
		// * dein#autoload#_on_cmd()
		//   https://github.com/Shougo/dein.vim/blob/2adba7655b23f2fc1ddcd35e15d380c5069a3712/autoload/dein/autoload.vim#L157-L175
		// * dein#autoload#_dummy_complete()
		//   https://github.com/Shougo/dein.vim/blob/2adba7655b23f2fc1ddcd35e15d380c5069a3712/autoload/dein/autoload.vim#L216-L232
		buf.WriteString(`

let ` + excmdLoadPlugin + ` = ` + string(lazyExcmdJSON) + `

function ` + lazyLoadExcmdFunc + `(command, args, bang, line1, line2) abort
  if exists(':' . a:command) is# 2
    execute 'delcommand' a:command
  endif
  execute get(` + excmdLoadPlugin + `, a:command, '')
  if exists(':' . a:command) isnot# 2
    echohl ErrorMsg
    echomsg printf('[volt] Lazy loading of Ex command ''%s'' failed: ''%s'' is not found', a:command, a:command)
    echohl None
    return
  endif
  let range = (a:line1 is# a:line2) ? '' :
        \ (a:line1 is# line("'<") && a:line2 is# line("'>")) ?
        \ "'<,'>" : a:line1 . ',' . a:line2
  try
    execute range . a:command . a:bang a:args
  catch /^Vim\%((\a\+)\)\=:E481/
    " E481: No range allowed
    execute a:command . a:bang a:args
  endtry
endfunction

function ` + completeFunc + `(arglead, cmdline, cursorpos) abort
  let command = matchstr(a:cmdline, '\h\w*')
  if exists(':' . command) is# 2
    execute 'delcommand' command
  endif
  execute get(` + excmdLoadPlugin + `, command, '')
  if exists(':' . command) is# 2
    call feedkeys("\<C-d>", 'n')
  endif
  return [a:arglead]
endfunction
`)
	}
	if len(loadCmds) > 0 {
		buf.WriteString("\n\n")
		buf.WriteString(`augroup volt-bundled-plugconf
  autocmd!
`)
		buf.WriteString(strings.Join(loadCmds, "\n"))
		buf.WriteString("\naugroup END")
	}

	if vimrcPath != "" || gvimrcPath != "" {
		buf.WriteString("\n\n")
		if vimrcPath != "" {
			vimrcPath = strings.Replace(vimrcPath, "'", "''", -1)
			buf.WriteString("let $MYVIMRC = '" + vimrcPath + "'")
		}
		if gvimrcPath != "" {
			gvimrcPath = strings.Replace(gvimrcPath, "'", "''", -1)
			buf.WriteString("let $MYGVIMRC = '" + gvimrcPath + "'")
		}
	}

	return buf.Bytes(), nil
}

var rxFuncName = regexp.MustCompile(`\A(fu\w+!?\s+s:)(\w+)`)

func convertToDecodableFunc(funcBody string, reposPath pathutil.ReposPath, reposID int) string {
	// Change function name (e.g. s:loaded_on() -> s:loaded_on_1())
	funcBody = rxFuncName.ReplaceAllString(funcBody, fmt.Sprintf("${1}${2}_%d", reposID))
	// Add repos path as comment
	funcBody = "\" " + reposPath.String() + "\n" + funcBody
	return funcBody
}

type reposDepTree struct {
	// The nodes' dependTo are nil. These repos's ranks are always 0.
	leaves []reposDepNode
}

type reposDepNode struct {
	repos      *lockjson.Repos
	dependTo   []reposDepNode
	dependedBy []reposDepNode
}

func ParseEachPlugconf(reposList []lockjson.Repos) (*MultiPlugconf, MultiParseError) {
	plugconfMap, parseErr := parsePlugconfAsMap(reposList)
	if parseErr.HasErrs() {
		return nil, parseErr
	}
	sortByDepends(reposList, plugconfMap)
	return &MultiPlugconf{
		plugconfMap: plugconfMap,
		reposList:   reposList,
	}, parseErr
}

type PlugconfMap map[pathutil.ReposPath]*Plugconf

type MultiPlugconf struct {
	plugconfMap PlugconfMap
	reposList   []lockjson.Repos
}

// vimrcPath and gvimrcPath become an empty string when each path does not
// exist.
func (mp *MultiPlugconf) GenerateBundlePlugconf(vimrcPath, gvimrcPath string) ([]byte, error) {
	return makeBundledPlugconf(mp.reposList, mp.plugconfMap, vimrcPath, gvimrcPath)
}

func RdepsOf(reposPath pathutil.ReposPath, reposList []lockjson.Repos) (pathutil.ReposPathList, error) {
	plugconfMap, parseErr := parsePlugconfAsMap(reposList)
	if parseErr.HasErrs() {
		return nil, parseErr.ErrorsAndWarns()
	}
	_, _, rdepsMap := getDepMaps(reposList, plugconfMap)
	rdeps := rdepsMap[reposPath]
	if rdeps == nil {
		rdeps = make(pathutil.ReposPathList, 0)
	}
	return rdeps, nil
}

// Parse plugconf of reposList and return parsed plugconf info as map
func parsePlugconfAsMap(reposList []lockjson.Repos) (map[pathutil.ReposPath]*Plugconf, MultiParseError) {
	parseErrAll := make(MultiParseError, 0, len(reposList))
	plugconfMap := make(PlugconfMap, len(reposList))
	reposID := 1
	for _, repos := range reposList {
		path := pathutil.Plugconf(repos.Path)
		if !pathutil.Exists(path) {
			continue
		}
		result, parseErr := ParsePlugconfFile(path, reposID, repos.Path)
		if parseErr.HasErrsOrWarns() {
			parseErrAll = append(parseErrAll, *parseErr)
		}
		if result != nil {
			plugconfMap[repos.Path] = result
			reposID += 1
		}
	}
	return plugconfMap, parseErrAll
}

// Move the plugins which was depended to previous plugin which depends to them.
// reposList is sorted in-place.
func sortByDepends(reposList []lockjson.Repos, plugconfMap map[pathutil.ReposPath]*Plugconf) {
	reposMap, depsMap, rdepsMap := getDepMaps(reposList, plugconfMap)
	rank := make(map[pathutil.ReposPath]int, len(reposList))
	for i := range reposList {
		if _, exists := rank[reposList[i].Path]; !exists {
			tree := makeDepTree(reposList[i].Path, reposMap, depsMap, rdepsMap)
			for i := range tree.leaves {
				makeRank(rank, &tree.leaves[i], 0)
			}
		}
	}
	sort.Slice(reposList, func(i, j int) bool {
		return rank[reposList[i].Path] < rank[reposList[j].Path]
	})
}

func getDepMaps(reposList []lockjson.Repos, plugconfMap map[pathutil.ReposPath]*Plugconf) (map[pathutil.ReposPath]*lockjson.Repos, map[pathutil.ReposPath]pathutil.ReposPathList, map[pathutil.ReposPath]pathutil.ReposPathList) {
	reposMap := make(map[pathutil.ReposPath]*lockjson.Repos, len(reposList))
	depsMap := make(map[pathutil.ReposPath]pathutil.ReposPathList, len(reposList))
	rdepsMap := make(map[pathutil.ReposPath]pathutil.ReposPathList, len(reposList))
	for i := range reposList {
		reposPath := reposList[i].Path
		reposMap[reposPath] = &reposList[i]
		if p, exists := plugconfMap[reposPath]; exists {
			depsMap[reposPath] = p.depends
			for _, dep := range p.depends {
				rdepsMap[dep] = append(rdepsMap[dep], reposPath)
			}
		}
	}
	return reposMap, depsMap, rdepsMap
}

func makeDepTree(reposPath pathutil.ReposPath, reposMap map[pathutil.ReposPath]*lockjson.Repos, depsMap map[pathutil.ReposPath]pathutil.ReposPathList, rdepsMap map[pathutil.ReposPath]pathutil.ReposPathList) *reposDepTree {
	visited := make(map[pathutil.ReposPath]*reposDepNode, len(reposMap))
	node := makeNodes(visited, reposPath, reposMap, depsMap, rdepsMap)
	leaves := make([]reposDepNode, 0, 10)
	visitNode(node, func(n *reposDepNode) {
		if len(n.dependTo) == 0 {
			leaves = append(leaves, *n)
		}
	})
	return &reposDepTree{leaves: leaves}
}

func makeNodes(visited map[pathutil.ReposPath]*reposDepNode, reposPath pathutil.ReposPath, reposMap map[pathutil.ReposPath]*lockjson.Repos, depsMap map[pathutil.ReposPath]pathutil.ReposPathList, rdepsMap map[pathutil.ReposPath]pathutil.ReposPathList) *reposDepNode {
	if node, exists := visited[reposPath]; exists {
		return node
	}
	node := &reposDepNode{repos: reposMap[reposPath]}
	visited[reposPath] = node
	for i := range depsMap[reposPath] {
		dep := makeNodes(visited, depsMap[reposPath][i], reposMap, depsMap, rdepsMap)
		node.dependTo = append(node.dependTo, *dep)
	}
	for i := range rdepsMap[reposPath] {
		rdep := makeNodes(visited, rdepsMap[reposPath][i], reposMap, depsMap, rdepsMap)
		node.dependedBy = append(node.dependedBy, *rdep)
	}
	return node
}

func visitNode(node *reposDepNode, callback func(*reposDepNode)) {
	visited := make(map[pathutil.ReposPath]bool, 10)
	doVisitNode(visited, node, callback)
}

func doVisitNode(visited map[pathutil.ReposPath]bool, node *reposDepNode, callback func(*reposDepNode)) {
	if node == nil || node.repos == nil || visited[node.repos.Path] {
		return
	}
	visited[node.repos.Path] = true
	callback(node)
	for i := range node.dependTo {
		doVisitNode(visited, &node.dependTo[i], callback)
	}
	for i := range node.dependedBy {
		doVisitNode(visited, &node.dependedBy[i], callback)
	}
}

func makeRank(rank map[pathutil.ReposPath]int, node *reposDepNode, value int) {
	rank[node.repos.Path] = value
	for i := range node.dependedBy {
		makeRank(rank, &node.dependedBy[i], value+1)
	}
}

func FetchPlugconf(reposPath pathutil.ReposPath) (string, error) {
	url := path.Join("https://raw.githubusercontent.com/vim-volt/plugconf-templates/master/templates", reposPath.String()+".vim")
	return httputil.GetContentString(url)
}

const skeletonPlugconfOnLoadPre = `function! s:on_load_pre()
  " Plugin configuration like the code written in vimrc.
  " This configuration is executed before a plugin is loaded.
endfunction`

const skeletonPlugconfOnLoadPost = `function! s:on_load_post()
  " Plugin configuration like the code written in vimrc.
  " This configuration is executed after a plugin is loaded.
endfunction`

const skeletonPlugconfLoadOn = `function! s:loaded_on()
  " This function determines when a plugin is loaded.
  "
  " Possible values are:
  " * 'start' (a plugin will be loaded at VimEnter event)
  " * 'filetype=<filetypes>' (a plugin will be loaded at FileType event)
  " * 'excmd=<excmds>' (a plugin will be loaded at CmdUndefined event)
  " <filetypes> and <excmds> can be multiple values separated by comma.
  "
  " This function must contain 'return "<str>"' code.
  " (the argument of :return must be string literal)

  return 'start'
endfunction`

const skeletonPlugconfDepends = `function! s:depends()
  " Dependencies of this plugin.
  " The specified dependencies are loaded after this plugin is loaded.
  "
  " This function must contain 'return [<repos>, ...]' code.
  " (the argument of :return must be list literal, and the elements are string)
  " e.g. return ['github.com/tyru/open-browser.vim']

  return []
endfunction`

func GenPlugconfByTemplate(tmplPlugconf string, filename string) ([]byte, *multierror.Error) {
	// Parse fetched plugconf
	tmpl, err := vimlparser.ParseFile(strings.NewReader(tmplPlugconf), filename, nil)
	if err != nil {
		return nil, multierror.Append(nil, err)
	}
	result, parseErr := ParsePlugconf(tmpl, tmplPlugconf, filename)
	if parseErr.HasErrs() {
		return nil, parseErr.ErrorsAndWarns()
	}
	content, err := generatePlugconf(result)
	if err != nil {
		return nil, multierror.Append(nil, err)
	}
	return content, nil
}

func generatePlugconf(result *Plugconf) ([]byte, error) {
	// Merge result and return it
	var buf bytes.Buffer
	var err error
	// s:on_load_pre()
	if result.onLoadPreFunc != "" {
		_, err = buf.WriteString(result.onLoadPreFunc)
	} else {
		_, err = buf.WriteString(skeletonPlugconfOnLoadPre)
	}
	if err != nil {
		return nil, err
	}
	_, err = buf.WriteString("\n\n")
	if err != nil {
		return nil, err
	}
	// s:on_load_post()
	if result.onLoadPostFunc != "" {
		_, err = buf.WriteString(result.onLoadPostFunc)
	} else {
		_, err = buf.WriteString(skeletonPlugconfOnLoadPost)
	}
	if err != nil {
		return nil, err
	}
	_, err = buf.WriteString("\n\n")
	if err != nil {
		return nil, err
	}
	// s:loaded_on()
	if result.loadOnFunc != "" {
		_, err = buf.WriteString(result.loadOnFunc)
	} else {
		_, err = buf.WriteString(skeletonPlugconfLoadOn)
	}
	if err != nil {
		return nil, err
	}
	_, err = buf.WriteString("\n\n")
	if err != nil {
		return nil, err
	}
	// s:depends()
	if result.dependsFunc != "" {
		_, err = buf.WriteString(result.dependsFunc)
	} else {
		_, err = buf.WriteString(skeletonPlugconfDepends)
	}
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
