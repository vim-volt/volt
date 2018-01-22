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

const (
	loadOnStart    loadOnType = "(loadOnStart)"
	loadOnFileType            = "FileType"
	loadOnExcmd               = "(loadOnExcmd)"
)

type Plugconf struct {
	reposID     int
	reposPath   pathutil.ReposPath
	functions   []string
	configFunc  string
	loadOnFunc  string
	loadOn      loadOnType
	loadOnArg   string
	dependsFunc string
	depends     pathutil.ReposPathList
}

func ParsePlugconfFile(plugConf string, reposID int, reposPath pathutil.ReposPath) (*Plugconf, error) {
	content, err := ioutil.ReadFile(plugConf)
	if err != nil {
		return nil, err
	}
	src := string(content)
	file, err := vimlparser.ParseFile(strings.NewReader(src), plugConf, nil)
	if err != nil {
		return nil, err
	}
	parsed, err := ParsePlugconf(file, src)
	if err != nil {
		return nil, err
	}
	parsed.reposID = reposID
	parsed.reposPath = reposPath
	return parsed, nil
}

func ParsePlugconf(file *ast.File, src string) (*Plugconf, error) {
	var loadOn loadOnType = loadOnStart
	var loadOnArg string
	var loadOnFunc string
	var configFunc string
	var functions []string
	var dependsFunc string
	var depends pathutil.ReposPathList
	var parseErr error

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

		switch name {
		case "s:loaded_on":
			loadOnFunc = extractBody(fn, src)
			var err error
			loadOn, loadOnArg, err = inspectReturnValue(fn)
			if err != nil {
				parseErr = err
			}
		case "s:config":
			configFunc = extractBody(fn, src)
		case "s:depends":
			dependsFunc = extractBody(fn, src)
			var err error
			depends, err = getDependencies(fn, src)
			if err != nil {
				parseErr = err
			}
		default:
			functions = append(functions, extractBody(fn, src))
		}

		return true
	})

	if parseErr != nil {
		return nil, parseErr
	}

	return &Plugconf{
		functions:   functions,
		configFunc:  configFunc,
		loadOnFunc:  loadOnFunc,
		loadOn:      loadOn,
		loadOnArg:   loadOnArg,
		dependsFunc: dependsFunc,
		depends:     depends,
	}, nil
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
func makeBundledPlugconf(reposList []lockjson.Repos, plugconf map[pathutil.ReposPath]*Plugconf) ([]byte, error) {
	functions := make([]string, 0, 64)
	loadCmds := make([]string, 0, len(reposList))
	lazyExcmd := make(map[string]string, len(reposList))

	for _, repos := range reposList {
		p, hasPlugconf := plugconf[repos.Path]
		// :packadd <repos>
		optName := filepath.Base(pathutil.EncodeReposPath(repos.Path))
		packadd := fmt.Sprintf("packadd %s", optName)

		// s:config() and invoked command
		var invokedCmd string
		if hasPlugconf && p.configFunc != "" {
			functions = append(functions, convertToDecodableFunc(p.configFunc, p.reposPath, p.reposID))
			invokedCmd = fmt.Sprintf("call s:config_%d() | %s", p.reposID, packadd)
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
					fmt.Sprintf("  command -complete=customlist,s:_complete -bang -bar -range -nargs=* %[1]s call s:_lazy_load_excmd('%[1]s', <q-args>, expand('<bang>'), expand('<line1>'), expand('<line2>'))", excmd))
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

let s:_excmd_load_plugin = ` + string(lazyExcmdJSON) + `

function! s:_lazy_load_excmd(command, args, bang, line1, line2) abort
  if exists(':' . a:command) is# 2
    execute 'delcommand' a:command
  endif
  execute get(s:_excmd_load_plugin, a:command, '')
  if exists(':' . a:command) isnot# 2
    echohl ErrorMsg
    echomsg printf('[volt] Lazy loading of Ex command '%s' failed: '%s' is not found', a:command, a:command)
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

function! s:_complete(arglead, cmdline, cursorpos) abort
  let command = matchstr(a:cmdline, '\h\w*')
  if exists(':' . command) is# 2
    execute 'delcommand' command
  endif
  execute get(s:_excmd_load_plugin, command, '')
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

	return buf.Bytes(), nil
}

var rxFuncName = regexp.MustCompile(`^(fu\w+!?\s+s:\w+)`)

func convertToDecodableFunc(funcBody string, reposPath pathutil.ReposPath, reposID int) string {
	// Change function name (e.g. s:loaded_on() -> s:loaded_on_1())
	funcBody = rxFuncName.ReplaceAllString(funcBody, fmt.Sprintf("${1}_%d", reposID))
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

func GenerateBundlePlugconf(reposList []lockjson.Repos) ([]byte, *multierror.Error) {
	plugconfMap, merr := parsePlugconfAsMap(reposList)
	if merr.ErrorOrNil() != nil {
		return nil, merr
	}
	sortByDepends(reposList, plugconfMap)
	content, err := makeBundledPlugconf(reposList, plugconfMap)
	return content, multierror.Append(nil, err)
}

func RdepsOf(reposPath pathutil.ReposPath, reposList []lockjson.Repos) (pathutil.ReposPathList, error) {
	plugconfMap, merr := parsePlugconfAsMap(reposList)
	if merr.ErrorOrNil() != nil {
		return nil, merr
	}
	_, _, rdepsMap := getDepMaps(reposList, plugconfMap)
	rdeps := rdepsMap[reposPath]
	if rdeps == nil {
		rdeps = make(pathutil.ReposPathList, 0)
	}
	return rdeps, nil
}

// Parse plugconf of reposList and return parsed plugconf info as map
func parsePlugconfAsMap(reposList []lockjson.Repos) (map[pathutil.ReposPath]*Plugconf, *multierror.Error) {
	var merr *multierror.Error
	plugconfMap := make(map[pathutil.ReposPath]*Plugconf, len(reposList))
	reposID := 1
	for _, repos := range reposList {
		var parsed *Plugconf
		var err error
		path := pathutil.Plugconf(repos.Path)
		if pathutil.Exists(path) {
			parsed, err = ParsePlugconfFile(path, reposID, repos.Path)
		} else {
			continue
		}
		if err != nil {
			merr = multierror.Append(merr, err)
		} else {
			plugconfMap[repos.Path] = parsed
			reposID += 1
		}
	}
	return plugconfMap, merr
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

const skeletonPlugconfConfig = `function! s:config()
  " Plugin configuration like the code written in vimrc.
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

func GenPlugconfByTemplate(tmplPlugconf string, filename string) ([]byte, error) {
	// Parse fetched plugconf
	tmpl, err := vimlparser.ParseFile(strings.NewReader(tmplPlugconf), filename, nil)
	if err != nil {
		return nil, err
	}
	parsed, err := ParsePlugconf(tmpl, tmplPlugconf)
	if err != nil {
		return nil, err
	}

	// Merge result and return it
	var buf bytes.Buffer
	// s:config()
	if parsed.configFunc != "" {
		_, err = buf.WriteString(parsed.configFunc)
	} else {
		_, err = buf.WriteString(skeletonPlugconfConfig)
	}
	if err != nil {
		return nil, err
	}
	_, err = buf.WriteString("\n\n")
	if err != nil {
		return nil, err
	}
	// s:loaded_on()
	if parsed.loadOnFunc != "" {
		_, err = buf.WriteString(parsed.loadOnFunc)
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
	if parsed.dependsFunc != "" {
		_, err = buf.WriteString(parsed.dependsFunc)
	} else {
		_, err = buf.WriteString(skeletonPlugconfDepends)
	}
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
