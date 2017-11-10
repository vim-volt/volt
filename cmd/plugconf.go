package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/volt/lockjson"
	"github.com/vim-volt/volt/logger"
	"github.com/vim-volt/volt/pathutil"
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
		user := pathutil.UserPlugConfOf(repos.Path)
		system := pathutil.SystemPlugConfOf(repos.Path)
		if pathutil.Exists(user) {
			fmt.Println(user)
		}
		if showSystem && pathutil.Exists(system) {
			fmt.Println(system)
		}
	}

	return nil
}

func (*plugconfCmd) doBundle(_ []string) error {

	// TODO

	return nil
}

func (*plugconfCmd) doUnbundle(_ []string) error {

	// TODO

	return nil
}
