:zap: Volt
----

A meta-level vim package manager.

## Install

```
$ go get github.com/vim-volt/volt
```

## Build environment

* Go 1.9.1 or higher
* or Go 1.9.0 with [a patch for os.RemoveAll()](https://go-review.googlesource.com/c/go/+/62970) ([#1](https://github.com/vim-volt/go-volt/issues/1))

## Introduction

### VOLTPATH

You can change base directory of volt by `VOLTPATH` environment variable.
This is `$HOME/volt` by default.

### Install plugin(s)

```
$ volt get [repositories ...]
```

For example, installing [tyru/caw.vim](https://github.com/tyru/caw.vim) plugin:

```
$ volt get https://github.com/tyru/caw.vim   # most verbose way
$ volt get github.com/tyru/caw.vim           # you can omit https:// of repository URL
$ volt get tyru/caw.vim                      # you can omit github.com/ if the repository is on GitHub
```

And you can install multiple plugins (parallel download):

```
$ volt get tyru/open-browser.vim tyru/open-browser-github.vim
```

For example, what `volt get tyru/caw.vim` command does internally is:

* Clone and install the repository to `$VOLTPATH/repos/github.com/tyru/caw.vim`
    * Volt does not require `git` command because it's powered by [go-git](https://github.com/src-d/go-git)
* Update `$VOLTPATH/lock.json`
* Run `volt rebuild`
    * Copy repository files to `~/.vim/pack/volt/opt/github.com_tyru_caw.vim`
    * Install `~/.vim/pack/volt/start/system/plugin/bundled_plugconf.vim`
        * It loads plugins like `packadd github.com_tyru_caw.vim`

### Update plugins

To update all plugins:

```
$ volt get -l -u
```

`-l` works like all installed plugins are specified (the repositories list is read from `$VOLTPATH/lock.json`).
`-u` updates specified plugins.

To update specified plugin only:

```
$ volt get -u tyru/caw.vim
```

### Uninstall plugins

```
$ volt rm [repositories ...]
```

To uninstall `tyru/caw.vim` like:

```
$ volt rm tyru/caw.vim   # (sob)
```

### Easy setup on another PC

If you want to install set of plugins which you have installed by `volt get`, you can use `volt get -l`.

```
$ volt get -l   # install plugins listed in $VOLTPATH/lock.json
```

First, you have to manage the following files under `$VOLTPATH`.

```
$VOLTPATH/
├ ─ ─  lock.json
├ ─ ─  plugconf (optional)
└ ─ ─  rc (optional)
```

**NOTE: DO NOT RECOMMEND SHARING VOLT DIRECTORY ON DROPBOX** (see [related issues](https://github.com/vim-volt/volt/issues?utf8=%E2%9C%93&q=is%3Aissue+dropbox)).

For example, my actual setup is:

```
$ tree -L 1 ~/volt/
/home/tyru/volt/
├ ─ ─  lock.json -> /home/tyru/git/dotfiles/dotfiles/volt/lock.json
├ ─ ─  plugconf -> /home/tyru/git/dotfiles/dotfiles/volt/plugconf
├ ─ ─  rc -> /home/tyru/git/dotfiles/dotfiles/volt/rc
└ ─ ─  repos
```

See [volt directory](https://github.com/tyru/dotfiles/tree/75a37b4a640a5cffecf34d2a52406d0f53ee6f09/dotfiles/volt) in [tyru/dotfiles](https://github.com/tyru/dotfiles/) repository for example.

### Configuration per plugin ("Plugconf" feature)

You can write plugin configuration in "plugconf" file.
The files are placed at:

* `$VOLTPATH/plugconf/user/<repository>.vim`

For example, [tyru/open-browser-github.vim](https://github.com/tyru/open-browser-github.vim) configuration is `$VOLTPATH/plugconf/user/github.com/tyru/open-browser.vim.vim` because "github.com/tyru/open-browser-github.vim" is the repository URL.

Some special functions can be defined in plugconf file:

* `s:config()`
    * Plugin configuration
* `s:load_on()` (optional)
    * Return value: String (when to load a plugin by `:packadd`)
    * This function specifies when to load a plugin by `:packadd`
    * e.g.: `return "start"` (default, load on `VimEnter` autocommand)
    * e.g.: `return "filetype=<filetype>"` (load on `FileType` autocommand)
    * e.g.: `return "excmd=<excmd>"` (load on `CmdUndefined` autocommand)
* `s:depends()` (optional)
    * Return value: List (repository name)
    * The specified plugins by this function are loaded before the plugin of plugconf
    * e.g.: `["github.com/tyru/open-browser.vim"]`

However, you can also define global functions in plugconf (see [tyru/nextfile.vim example](https://github.com/tyru/dotfiles/blob/master/dotfiles/volt/plugconf/user/github.com/tyru/nextfile.vim.vim)).

An example config of [tyru/open-browser-github.vim](https://github.com/tyru/open-browser-github.vim):

```vim
function! s:config()
  let g:openbrowser_github_always_use_commit_hash = 1
endfunction

function! s:depends()
  return ['github.com/tyru/open-browser.vim']
endfunction
```

NOTE:

* Plugconf file is parsed by [go-vimlparser](https://github.com/haya14busa/go-vimlparser)
* The rhs of `:return` must be literal
* Breaking newline by backslash (`\`) in `s:load_on()` and `s:depends()` is safe, but the following code can not be recognized (currently not supported at least)

```vim
" Wrong
function! s:load_on()
  let when = 'filetype=vim'
  return when
endfunction

" Wrong
function! s:depends()
  let list =  ['github.com/tyru/open-browser.vim']
  return list
endfunction

" OK
function! s:depends()
  return [
  \  'github.com/tyru/open-browser.vim'
  \]
endfunction
```

See [plugconf directory](https://github.com/tyru/dotfiles/tree/75a37b4a640a5cffecf34d2a52406d0f53ee6f09/dotfiles/volt/plugconf) in [tyru/dotfiles](https://github.com/tyru/dotfiles/) repository for example.

### Switch set of plugins ("Profile" feature)

You can think this is similar feature of **branch** of `git`.
The default profile name is "default".

You can see profile list by `volt profile list`.

```
$ volt profile list
* default
```

You can create a new profile by `volt profile new`.

```
$ volt profile new foo   # will create profile "foo"
$ volt profile list
* default
  foo
```

You can switch current profile by `volt profile set`.

```
$ volt profile set foo   # will switch profile to "foo"
$ volt profile list
  default
* foo
```

You can delete profile by `volt profile destroy` (but you cannot delete current active profile which you are switching on).

```
$ volt profile destroy foo   # will delete profile "foo"
```

You can enable/disable plugin by `volt enable` (`volt profile add`), `volt disable` (`volt profile rm`).

```
$ volt enable tyru/caw.vim    # enable loading tyru/caw.vim on current profile
$ volt profile add foo tyru/caw.vim    # enable loading tyru/caw.vim on "foo" profile
```

```
$ volt disable tyru/caw.vim   # disable loading tyru/caw.vim on current profile
$ volt profile rm foo tyru/caw.vim    # disable loading tyru/caw.vim on "foo" profile
```

You can create a vimrc & gvimrc file for each profile:
* vimrc: `$VOLTPATH/rc/<profile name>/vimrc.vim`
* gvimrc: `$VOLTPATH/rc/<profile name>/gvimrc.vim`

This file is copied to `~/.vim/vimrc` and `~/.vim/gvimrc` with magic comment (shows error if existing vimrc/gvimrc files exist with no magic comment).

And you can enable/disable vimrc by `volt profile use` (or you can simply remove `$VOLTPATH/rc/<profile name>/vimrc.vim` file if you don't want vimrc for the profile).

```
$ volt profile use -current vimrc false   # Disable installing vimrc on current active profile
$ volt profile use default gvimrc true   # Enable installing gvimrc on profile default
```

See `volt help profile` for more detailed information.


## :tada: Join development

If you want to join developing volt, you can setup like:

```
$ make setup
$ make precompile   # this speeds up 'go build'
$ vim ...           # edit sources
$ make
$ bin/volt ...      # run volt command
```

## How to build release binaries

```
$ make setup
$ make release
$ ls -1 dist/
volt-v0.0.1-alpha-darwin-amd64
volt-v0.0.1-alpha-darwin-386
volt-v0.0.1-alpha-linux-amd64
volt-v0.0.1-alpha-linux-386
volt-v0.0.1-alpha-windows-amd64.exe
volt-v0.0.1-alpha-windows-386.exe
```
