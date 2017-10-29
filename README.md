go-volt
-------

A meta-level vim package manager.

## Install

```
$ go get github.com/vim-volt/volt
```

## Build environment

Go 1.9.1, or Go 1.9.0 with [a patch for os.RemoveAll()](https://go-review.googlesource.com/c/go/+/62970) ([#1](https://github.com/vim-volt/go-volt/issues/1))

## Want to join development of volt?

If you want to join developing volt, to build release binaries:

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

To build develop binary:

```
$ make setup
$ make precompile # this speeds up 'go build'
```

and edit sources, and

```
$ make
$ bin/volt ... # run volt command
```
