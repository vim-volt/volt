package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"

	"github.com/haya14busa/go-vimlparser"
	"github.com/haya14busa/go-vimlparser/ast"
	"github.com/haya14busa/go-vimlparser/token"
)

type plugconfFlagsType struct {
	helped   bool
	lockJSON bool
	upgrade  bool
	verbose  bool
}

var plugconfFlags plugconfFlagsType

var plugconfSubCmd = make(map[string]func([]string) error)

func init() {
	cmd := plugconfCmd{}
	plugconfSubCmd["list"] = cmd.doList
	plugconfSubCmd["export"] = cmd.doExport
	plugconfSubCmd["import"] = cmd.doImport

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  plugconf list [-a]
    List all user plugconfs. If -a option was given, list also system plugconfs.

  plugconf export
    Outputs bundled plugconf to stdout.
    Note that the output differs a bit from the file written by "volt rebuild"
    (~/.vim/pack/volt/opt/system/plugin/bundled_plugconf.vim).
    Some functions are removed in the file because it is unnecessary for Vim.
    But this command shows them because this command must export all in plugconfs.

  plugconf import
    Input bundled plugconf (volt plugconf export) from stdin, import the plugconf, and put files to each plugin's plugconf.

Quick example
  $ volt plugconf list
  github.com/tyru/open-browser.vim.vim
  github.com/tpope/vim-markdown.vim

  $ volt plugconf export

  " github.com/tyru/open-browser.vim.vim
  function s:config_1()
    let g:netrw_nogx = 1
    nmap gx <Plug>(openbrowser-smart-search)
    xmap gx <Plug>(openbrowser-smart-search)

    command! OpenBrowserCurrent execute 'OpenBrowser' 'file://' . expand('%:p:gs?\\?/?')
  endfunction

  " github.com/tpope/vim-markdown.vim
  function s:config_2()
    " no config
  endfunction

  augroup volt-bundled-plugconf
    autocmd!
    autocmd VimEnter * call s:config_1()
    autocmd FileType markdown call s:config_2()
  augroup END

  $ volt plugconf export >exported.vim
  $ vim exported.vim  # edit config
  $ volt plugconf import <exported.vim` + "\n\n")
		fs.PrintDefaults()
		fmt.Println()
		plugconfFlags.helped = true
	}

	cmdFlagSet["plugconf"] = fs
}

type plugconfCmd struct{}

func Plugconf(args []string) int {
	cmd := plugconfCmd{}

	// Parse args
	args, err := cmd.parseArgs(args)
	if err == ErrShowedHelp {
		return 0
	}
	if err != nil {
		logger.Error(err.Error())
		return 10
	}

	if fn, exists := plugconfSubCmd[args[0]]; exists {
		err = fn(args[1:])
		if err != nil {
			logger.Error(err.Error())
			return 11
		}
	}

	return 0
}

func (cmd *plugconfCmd) parseArgs(args []string) ([]string, error) {
	fs := cmdFlagSet["plugconf"]
	fs.Parse(args)
	if plugconfFlags.helped {
		return nil, ErrShowedHelp
	}
	if len(fs.Args()) == 0 {
		return nil, errors.New("volt plugconf: must specify subcommand")
	}

	subCmd := fs.Args()[0]
	if _, exists := plugconfSubCmd[subCmd]; !exists {
		return nil, errors.New("unknown subcommand: " + subCmd)
	}
	return fs.Args(), nil
}

func (*plugconfCmd) doList(args []string) error {
	var showSystem bool
	if len(args) > 0 && args[0] == "-a" {
		showSystem = true
	}

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("Could not read lock.json: " + err.Error())
	}

	for i := range lockJSON.Repos {
		repos := &lockJSON.Repos[i]
		user := pathutil.UserPlugconfOf(repos.Path)
		system := pathutil.SystemPlugconfOf(repos.Path)
		if pathutil.Exists(user) {
			fmt.Println(user)
		}
		if showSystem && pathutil.Exists(system) {
			fmt.Println(system)
		}
	}

	return nil
}

// Output bundled plugconf content
func (cmd *plugconfCmd) doExport(args []string) error {
	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("could not read lock.json: " + err.Error())
	}

	// Find active profile
	profile, err := lockJSON.Profiles.FindByName(lockJSON.ActiveProfile)
	if err != nil {
		// this must not be occurred because lockjson.Read()
		// validates that the matching profile exists
		return err
	}

	// Get active profile's repos list
	reposList, err := lockJSON.GetReposListByProfile(profile)
	if err != nil {
		return err
	}

	// Output bundled plugconf content
	exportAll := true
	output, merr := cmd.generateBundlePlugconf(exportAll, reposList)
	if merr.ErrorOrNil() != nil {
		for _, err := range merr.Errors {
			// Show vim script parse errors
			logger.Error(err.Error())
		}
		return nil
	}
	os.Stdout.Write(output)
	return nil
}

func (cmd *plugconfCmd) generateBundlePlugconf(exportAll bool, reposList []lockjson.Repos) ([]byte, *multierror.Error) {
	// Parse plugconfs and make parsed plugconf info
	var merr *multierror.Error
	plugconf := make(map[string]*parsedPlugconf, len(reposList))
	funcCap := 0
	reposID := 1
	for _, repos := range reposList {
		var parsed *parsedPlugconf
		var err error
		user := pathutil.UserPlugconfOf(repos.Path)
		system := pathutil.SystemPlugconfOf(repos.Path)
		if pathutil.Exists(user) {
			parsed, err = cmd.parsePlugConf(user, reposID, repos.Path)
		} else if pathutil.Exists(system) {
			parsed, err = cmd.parsePlugConf(system, reposID, repos.Path)
		} else {
			continue
		}
		if err != nil {
			merr = multierror.Append(merr, err)
		} else {
			plugconf[repos.Path] = parsed
			funcCap += len(parsed.functions) + 1 /* +1 for s:config() */
			reposID += 1
		}
	}
	cmd.sortByDepends(reposList, plugconf)
	return cmd.makeBundledPlugConf(exportAll, reposList, plugconf, funcCap), merr
}

// Move the plugins which was depended to previous plugin which depends to them.
// reposList is sorted in-place.
func (cmd *plugconfCmd) sortByDepends(reposList []lockjson.Repos, plugconf map[string]*parsedPlugconf) {
	reposMap := make(map[string]*lockjson.Repos, len(reposList))
	depsMap := make(map[string][]string, len(reposList))
	rdepsMap := make(map[string][]string, len(reposList))
	rank := make(map[string]int, len(reposList))
	for i := range reposList {
		reposPath := reposList[i].Path
		reposMap[reposPath] = &reposList[i]
		if p, exists := plugconf[reposPath]; exists {
			depsMap[reposPath] = p.depends
			for _, dep := range p.depends {
				rdepsMap[dep] = append(rdepsMap[dep], reposPath)
			}
		}
	}
	for i := range reposList {
		if _, exists := rank[reposList[i].Path]; !exists {
			tree := cmd.makeDepTree(reposList[i].Path, reposMap, depsMap, rdepsMap)
			for i := range tree.leaves {
				cmd.makeRank(rank, &tree.leaves[i], 0)
			}
		}
	}
	sort.Slice(reposList, func(i, j int) bool {
		return rank[reposList[i].Path] < rank[reposList[j].Path]
	})
}

func (cmd *plugconfCmd) makeDepTree(reposPath string, reposMap map[string]*lockjson.Repos, depsMap map[string][]string, rdepsMap map[string][]string) *reposDepTree {
	visited := make(map[string]*reposDepNode, len(reposMap))
	node := cmd.makeNodes(visited, reposPath, reposMap, depsMap, rdepsMap)
	leaves := make([]reposDepNode, 0, 10)
	cmd.visitNode(node, func(n *reposDepNode) {
		if len(n.dependTo) == 0 {
			leaves = append(leaves, *n)
		}
	})
	return &reposDepTree{leaves: leaves}
}

func (cmd *plugconfCmd) makeNodes(visited map[string]*reposDepNode, reposPath string, reposMap map[string]*lockjson.Repos, depsMap map[string][]string, rdepsMap map[string][]string) *reposDepNode {
	if node, exists := visited[reposPath]; exists {
		return node
	}
	node := &reposDepNode{repos: reposMap[reposPath]}
	visited[reposPath] = node
	for i := range depsMap[reposPath] {
		dep := cmd.makeNodes(visited, depsMap[reposPath][i], reposMap, depsMap, rdepsMap)
		node.dependTo = append(node.dependTo, *dep)
	}
	for i := range rdepsMap[reposPath] {
		rdep := cmd.makeNodes(visited, rdepsMap[reposPath][i], reposMap, depsMap, rdepsMap)
		node.dependedBy = append(node.dependedBy, *rdep)
	}
	return node
}

func (cmd *plugconfCmd) visitNode(node *reposDepNode, callback func(*reposDepNode)) {
	visited := make(map[string]bool, 10)
	cmd.doVisitNode(visited, node, callback)
}

func (cmd *plugconfCmd) doVisitNode(visited map[string]bool, node *reposDepNode, callback func(*reposDepNode)) {
	if visited[node.repos.Path] {
		return
	}
	visited[node.repos.Path] = true
	callback(node)
	for i := range node.dependTo {
		cmd.doVisitNode(visited, &node.dependTo[i], callback)
	}
	for i := range node.dependedBy {
		cmd.doVisitNode(visited, &node.dependedBy[i], callback)
	}
}

func (cmd *plugconfCmd) makeRank(rank map[string]int, node *reposDepNode, value int) {
	rank[node.repos.Path] = value
	for i := range node.dependedBy {
		cmd.makeRank(rank, &node.dependedBy[i], value+1)
	}
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

type loadOnType string

const (
	loadOnStart    loadOnType = "VimEnter"
	loadOnFileType            = "FileType"
	loadOnExcmd               = "CmdUndefined"
)

type parsedPlugconf struct {
	reposID    int
	reposPath  string
	functions  []string
	configFunc string
	loadOnFunc string
	loadOn     loadOnType
	loadOnArg  string
	depends    []string
}

func (cmd *plugconfCmd) parsePlugConf(plugConf string, reposID int, reposPath string) (*parsedPlugconf, error) {
	bytes, err := ioutil.ReadFile(plugConf)
	if err != nil {
		return nil, err
	}
	src := string(bytes)

	file, err := vimlparser.ParseFile(strings.NewReader(src), plugConf, nil)
	if err != nil {
		return nil, err
	}

	var loadOn loadOnType = loadOnStart
	var loadOnArg string
	var loadOnFunc string
	var configFunc string
	var functions []string
	var depends []string
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
		case "s:load_on":
			loadOnFunc = cmd.extractBody(fn, src)
			var err error
			loadOn, loadOnArg, err = cmd.inspectReturnValue(fn)
			if err != nil {
				parseErr = err
			}
		case "s:config":
			configFunc = cmd.extractBody(fn, src)
		case "s:depends":
			depends = cmd.getDependencies(fn, src)
		default:
			functions = append(functions, cmd.extractBody(fn, src))
		}

		return true
	})

	if parseErr != nil {
		return nil, parseErr
	}

	return &parsedPlugconf{
		reposID:    reposID,
		reposPath:  reposPath,
		functions:  functions,
		configFunc: configFunc,
		loadOnFunc: loadOnFunc,
		loadOn:     loadOn,
		loadOnArg:  loadOnArg,
		depends:    depends,
	}, nil
}

// Inspect return value of s:load_on() function in plugconf
func (cmd *plugconfCmd) inspectReturnValue(fn *ast.Function) (loadOnType, string, error) {
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
		return "", "", errors.New("can't detect return value of s:load_on()")
	}
	return loadOn, loadOnArg, err
}

func (cmd *plugconfCmd) extractBody(fn *ast.Function, src string) string {
	pos := fn.Pos()

	endpos := fn.EndFunction.Pos()
	endfunc := fn.EndFunction.ExArg
	cmdlen := endfunc.Argpos.Offset - endfunc.Cmdpos.Offset
	endpos.Offset += cmdlen

	return src[pos.Offset:endpos.Offset]
}

func (cmd *plugconfCmd) getDependencies(fn *ast.Function, src string) []string {
	var deps []string

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
						deps = make([]string, 0, len(list.Values))
					}
					if str.Kind == token.STRING {
						deps = append(deps, str.Value[1:len(str.Value)-1])
					}
				}
			}
		}
		return true
	})

	return deps
}

func (cmd *plugconfCmd) makeBundledPlugConf(exportAll bool, reposList []lockjson.Repos, plugconf map[string]*parsedPlugconf, funcCap int) []byte {
	functions := make([]string, 0, funcCap)
	autocommands := make([]string, 0, len(reposList))
	packadds := make([]string, 0, len(reposList))
	for _, repos := range reposList {
		optName := filepath.Base(pathutil.PackReposPathOf(repos.Path))
		packadds = append(packadds, fmt.Sprintf("packadd %s", optName))
		if p, exists := plugconf[repos.Path]; exists {
			if exportAll && p.loadOnFunc != "" {
				functions = append(functions, cmd.convertToDecodableFunc(p.loadOnFunc, p.reposPath, p.reposID))
			}
			if p.configFunc != "" {
				functions = append(functions, cmd.convertToDecodableFunc(p.configFunc, p.reposPath, p.reposID))
				var pattern string
				if p.loadOn == loadOnStart || p.loadOnArg == "" {
					pattern = "*"
				} else {
					pattern = p.loadOnArg
				}
				autocommands = append(autocommands, fmt.Sprintf("  autocmd %s %s call s:config_%d()", string(p.loadOn), pattern, p.reposID))
			}
			functions = append(functions, p.functions...)
		}
	}
	return []byte(fmt.Sprintf(`if exists('g:loaded_volt_system_bundled_plugconf')
  finish
endif
let g:loaded_volt_system_bundled_plugconf = 1

%s

augroup volt-bundled-plugconf
  autocmd!
%s
augroup END

%s
`, strings.Join(functions, "\n\n"),
		strings.Join(autocommands, "\n"),
		strings.Join(packadds, "\n"),
	))
}

var rxFuncName = regexp.MustCompile(`^(fu\w+!?\s+s:\w+)`)

func (cmd *plugconfCmd) convertToDecodableFunc(funcBody string, reposPath string, reposID int) string {
	// Change function name (e.g. s:load_on() -> s:load_on_1())
	funcBody = rxFuncName.ReplaceAllString(funcBody, fmt.Sprintf("${1}_%d", reposID))
	// Add repos path as comment
	funcBody = "\" " + reposPath + "\n" + funcBody
	return funcBody
}

func (*plugconfCmd) doImport(_ []string) error {

	// TODO
	logger.Error("Sorry, currently this feature is not implemented")

	return nil
}
