go-volt
-------

A meta-level vim package manager.

## Build & Install

NOTE: Normal user does not require this step to use volt command.
Because this command was normally invoked by vim-volt,
and vim-volt automatically installs volt binary from [GitHub releases](https://github.com/vim-volt/go-volt/releases).

To build release binaries:

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

To develop:

```
$ make setup
$ make precompile # this speeds up 'go build'
```

and edit sources, and

```
$ make
$ bin/volt ... # run volt command
```

## Build environment

Go 1.9 with [a patch for os.RemoveAll()](https://go-review.googlesource.com/c/go/+/62970) ([#1](https://github.com/vim-volt/go-volt/issues/1))
