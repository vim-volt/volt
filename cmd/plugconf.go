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

  plugconf bundle
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
func (cmd *plugconfCmd) doBundle(_ []string) error {
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
	output, merr := cmd.generateBundlePlugconf(reposList)
	for _, err := range merr.Errors {
		// Show vim script parse errors
		logger.Warn(err.Error())
	}
	os.Stdout.Write(output)
	return nil
}

func (cmd *plugconfCmd) generateBundlePlugconf(reposList []lockjson.Repos) ([]byte, *multierror.Error) {
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
			parsed, err = cmd.parsePlugConf(user, parsedList)
		} else if pathutil.Exists(system) {
			parsed, err = cmd.parsePlugConf(system, parsedList)
		}
		if err != nil {
			merr = multierror.Append(merr, err)
		} else {
			parsedList = append(parsedList, *parsed)
			funcCap += len(parsed.functions) + 1 /* +1 for s:config() */
		}
	}
	return cmd.makeBundledPlugConf(parsedList, funcCap), merr
}

type loadOnType string

const (
	loadOnStart    loadOnType = "VimEnter"
	loadOnFileType            = "FileType"
	loadOnExcmd               = "CmdUndefined"
)

type parsedPlugconf struct {
	number     int
	configFunc string
	functions  []string
	loadOn     loadOnType
}

func (cmd *plugconfCmd) parsePlugConf(plugConf string, parsedList []parsedPlugconf) (*parsedPlugconf, error) {
	bytes, err := ioutil.ReadFile(plugConf)
	if err != nil {
		return nil, err
	}
	src := string(bytes)

	file, err := vimlparser.ParseFile(strings.NewReader(src), plugConf, nil)
	if err != nil {
		return nil, err
	}

	var loadOn loadOnType
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
			var err error
			loadOn, err = cmd.inspectReturnValue(fn)
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
		configFunc: configFunc,
		functions:  functions,
		loadOn:     loadOn,
	}, nil
}

// Inspect return value of s:load_on() function in plugconf
func (cmd *plugconfCmd) inspectReturnValue(fn *ast.Function) (loadOnType, error) {
	var loadOn loadOnType
	ast.Inspect(fn, func(node ast.Node) bool {
		// Cast to return node (return if it's not a return node)
		var ret *ast.Return
		if r, ok := node.(*ast.Return); !ok {
			return true
		} else {
			ret = r
		}

		// TODO: Parse the argument of :return

		return true
	})
	if string(loadOn) == "" {
		return "", errors.New("can't detect return value of s:")
	}
	return loadOn, nil
}

func (cmd *plugconfCmd) extractBody(fn *ast.Function, src string) string {
	pos := fn.Pos()

	// TODO: Handle line break
	endpos := fn.EndFunction.Pos()
	endfunc := fn.EndFunction.ExArg
	cmdlen := endfunc.Argpos.Offset - endfunc.Cmdpos.Offset
	endpos.Offset += cmdlen
	endpos.Column += cmdlen

	return src[pos.Offset:endpos.Offset]
}

func (cmd *plugconfCmd) makeBundledPlugConf(parsedList []parsedPlugconf, funcCap int) []byte {
	functions := make([]string, 0, funcCap)
	autocommands := make([]string, 0, len(parsedList))
	for _, p := range parsedList {
		// TODO: replace only function name node
		functions = append(functions, strings.Replace(p.configFunc, "s:config", fmt.Sprintf("s:config_%d", p.number), -1))
		functions = append(functions, p.functions...)
		autocommands = append(autocommands, fmt.Sprintf("  autocmd %s * call s:config_%d()", string(p.loadOn), p.number))
	}
	return []byte(fmt.Sprintf(`%s

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
