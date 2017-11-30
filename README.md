Volt
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


## Want to join development of volt?

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
