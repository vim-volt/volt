package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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
		}
	}
	return cmd.makeBundledPlugConf(exportAll, reposList, plugconf, funcCap), merr
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
		default:
			functions = append(functions, cmd.extractBody(fn, src))
		}

		return true
	})

	if parseErr != nil {
		return nil, parseErr
	}

	if configFunc == "" {
		return nil, errors.New("no s:config() function in plugconf: " + plugConf)
	}

	return &parsedPlugconf{
		reposID:    reposID,
		reposPath:  reposPath,
		functions:  functions,
		configFunc: configFunc,
		loadOnFunc: loadOnFunc,
		loadOn:     loadOn,
		loadOnArg:  loadOnArg,
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
			functions = append(functions, cmd.convertToDecodableFunc(p.configFunc, p.reposPath, p.reposID))
			functions = append(functions, p.functions...)
			var pattern string
			if p.loadOn == loadOnStart || p.loadOnArg == "" {
				pattern = "*"
			} else {
				pattern = p.loadOnArg
			}
			autocommands = append(autocommands, fmt.Sprintf("  autocmd %s %s call s:config_%d()", string(p.loadOn), pattern, p.reposID))
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
