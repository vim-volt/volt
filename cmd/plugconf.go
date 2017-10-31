package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
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
	plugconfSubCmd["bundle"] = cmd.doBundle
	plugconfSubCmd["unbundle"] = cmd.doUnbundle

	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Print(`
Usage
  plugconf list [-a]
    List all user plugconfs. If -a option was given, list also system plugconfs.

  plugconf bundle [-system]
    Outputs bundled plugconf to stdout.

  plugconf unbundle
    Input bundled plugconf (volt plugconf bundle) from stdin, unbundle the plugconf, and put files to each plugin's plugconf.

Quick example
  $ volt plugconf list
  github.com/tyru/open-browser.vim.vim
  github.com/tpope/vim-markdown.vim

  $ volt plugconf bundle

  " github.com/tyru/open-browser.vim.vim
  function s:load_on_1()
    return 'load'
  endfunction

  " github.com/tyru/open-browser.vim.vim
  function s:config_1()
    let g:netrw_nogx = 1
    nmap gx <Plug>(openbrowser-smart-search)
    xmap gx <Plug>(openbrowser-smart-search)

    command! OpenBrowserCurrent execute 'OpenBrowser' 'file://' . expand('%:p:gs?\\?/?')
  endfunction

  " github.com/tpope/vim-markdown.vim
  function s:load_on_2()
    return 'filetype=markdown'
  endfunction

  " github.com/tpope/vim-markdown.vim
  function s:config_2()
    " no config
  endfunction

  $ volt plugconf bundle -system

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

  $ volt plugconf bundle >bundle-plugconf.vim
  $ vim bundle-plugconf.vim  # edit config
  $ volt plugconf unbundle <bundle-plugconf.vim` + "\n\n")
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

// Output bundle plugconf content
func (cmd *plugconfCmd) doBundle(args []string) error {
	isSystem := len(args) != 0 && args[0] == "-system"

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

	// Output bundle plugconf content
	output, merr := cmd.generateBundlePlugconf(isSystem, reposList)
	if merr != nil {
		for _, err := range merr.Errors {
			// Show vim script parse errors
			logger.Warn(err.Error())
		}
	}
	os.Stdout.Write(output)
	return nil
}

func (cmd *plugconfCmd) generateBundlePlugconf(isSystem bool, reposList []lockjson.Repos) ([]byte, *multierror.Error) {
	// Parse plugconfs and make parsed plugconf info
	var merr *multierror.Error
	var parsedList []parsedPlugconf
	var funcCap int
	for _, repos := range reposList {
		var parsed *parsedPlugconf
		var err error
		user := pathutil.UserPlugConfOf(repos.Path)
		system := pathutil.SystemPlugConfOf(repos.Path)
		if pathutil.Exists(user) {
			parsed, err = cmd.parsePlugConf(user, parsedList, repos.Path)
		} else if pathutil.Exists(system) {
			parsed, err = cmd.parsePlugConf(system, parsedList, repos.Path)
		} else {
			continue
		}
		if err != nil {
			merr = multierror.Append(merr, err)
		} else {
			parsedList = append(parsedList, *parsed)
			funcCap += len(parsed.functions) + 1 /* +1 for s:config() */
		}
	}
	return cmd.makeBundledPlugConf(isSystem, parsedList, funcCap), merr
}

type loadOnType string

const (
	loadOnStart    loadOnType = "VimEnter"
	loadOnFileType            = "FileType"
	loadOnExcmd               = "CmdUndefined"
)

type parsedPlugconf struct {
	number     int
	reposPath  string
	loadOnFunc string
	configFunc string
	functions  []string
	loadOn     loadOnType
	loadOnArg  string
}

func (cmd *plugconfCmd) parsePlugConf(plugConf string, parsedList []parsedPlugconf, reposPath string) (*parsedPlugconf, error) {
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

	if configFunc == "" {
		return nil, errors.New("no s:config() function in plugconf: " + plugConf)
	}

	return &parsedPlugconf{
		number:     len(parsedList) + 1,
		reposPath:  reposPath,
		configFunc: configFunc,
		functions:  functions,
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

	// TODO: Handle line break
	endpos := fn.EndFunction.Pos()
	endfunc := fn.EndFunction.ExArg
	cmdlen := endfunc.Argpos.Offset - endfunc.Cmdpos.Offset
	endpos.Offset += cmdlen

	return src[pos.Offset:endpos.Offset]
}

func (cmd *plugconfCmd) makeBundledPlugConf(isSystem bool, parsedList []parsedPlugconf, funcCap int) []byte {
	functions := make([]string, 0, funcCap)
	autocommands := make([]string, 0, len(parsedList))
	for _, p := range parsedList {
		if isSystem && p.loadOnFunc != "" {
			functions = append(functions, p.loadOnFunc)
		}
		// TODO: replace only function name node
		configFunc := p.configFunc
		configFunc = strings.Replace(configFunc, "s:config", fmt.Sprintf("s:config_%d", p.number), -1)
		configFunc = fmt.Sprintf("\" %s\n", p.reposPath) + configFunc
		functions = append(functions, configFunc)
		functions = append(functions, p.functions...)
		autocommands = append(autocommands, fmt.Sprintf("  autocmd %s * call s:config_%d()", string(p.loadOn), p.number))
	}
	return []byte(fmt.Sprintf(`
if exists('g:loaded_volt_system_bundled_plugconf')
  finish
endif
let g:loaded_volt_system_bundled_plugconf = 1

%s

augroup volt-bundled-plugconf
  autocmd!
%s
augroup END
`, strings.Join(functions, "\n\n"), strings.Join(autocommands, "\n")))
}

func (*plugconfCmd) doUnbundle(_ []string) error {

	// TODO

	return nil
}
