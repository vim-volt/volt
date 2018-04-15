
[Original (Japanese)](https://gist.github.com/tyru/819e593b2d996321298f6338bbaa34e0)

# Volt refactoring note: JSON DSL and Transaction

## Example of JSON DSL

```json
["label",
  0,
  "installing plugins:",
  ["vimdir/with-install",
    ["parallel",
      ["label",
        1,
        "github.com/tyru/open-browser.vim ... {{if .Done}}done!{{end}}",
        ["parallel",
          ["lockjson/add",
            ["repos/git/clone", "github.com/tyru/open-browser.vim"],
            ["default"]],
          ["plugconf/install", "github.com/tyru/open-browser.vim"]]],
      ["label",
        1,
        "github.com/tyru/open-browser-github.vim ... {{if .Done}}done!{{end}}",
        ["parallel",
          ["lockjson/add",
            ["repos/git/clone", "github.com/tyru/open-browser-github.vim"],
            ["default"]],
          ["plugconf/install", "github.com/tyru/open-browser-github.vim"]]]]]]
```

## Wordings

* operator: "function" of DSL
  * e.g. "label"
  * e.g. "parallel"
* macro: like "function", but is expanded before execution
  * see JSON DSL note (TODO)
* expression: the form of operator application
  * e.g. `["label", ...]`
  * e.g. `["parallel", ...]`
* transaction log (file): a JSON file which is saved at
  `$VOLTPATH/trx/{id}/log.json`

## Goals

This refactoring allows us or makes it easy to implement the following issues:

1. JSON file of AST is saved under `$VOLTPATH/trx/{id}/`
2. The history feature (undo, redo, list, ...) like `yum history`
   [#147](https://github.com/vim-volt/volt/issues/147)
  * `volt history undo` executes `[$invert, expr]` for transaction log
  * `volt history redo` just executes saved expression in transaction log
3. Display progress bar [#118](https://github.com/vim-volt/volt/issues/188)
  * Updating before/after each expression node is executed
4. `volt watch` command can be easy to implement
   [#174](https://github.com/vim-volt/volt/issues/174)
  * Current `volt build` implementation installs all repositories of current
    profile, not specific repositories
5. Parallelism
  * Currently each command independently implements it using goroutine, but DSL
    provides higher level parallel processing
6. More detailed unit testing
  * Small component which is easy to test
  * And especially "Subcmd layer" is easy because it does not access to
    filesystem
7. Gracefully rollback when an error occurs while processing a DSL

## Layered architecture

The volt commands like `volt get` which may modify lock.json, config.toml(#221),
filesystem, are executed in several steps:

1. (Gateway layer): pass subcommand arguments, lock.json & config.toml structure
   to Subcmd layer
2. (Subcmd layer): Create an AST (abstract syntax tree) according to given information
  * This layer cannot touch filesystem, because it makes unit testing difficult
3. (DSL layer): Execute the AST. This note mainly describes this layer's design

Below is the dependency graph:

```
Gateway --> Subcmd --> DSL
```

* Gateway only depends Subcmd
* Subcmd doesn't know Gateway
* Subcmd only depends DSL
* DSL doesn't know Subcmd

## Abstract

JSON DSL is a S-expression like DSL represented as JSON format.

```json
["op", "arg1", "arg2"]
```

This is an application form (called "expression" in this note) when `op` is a
known operator name.  But if `op` is not a known operator, it is just an array
literal value.  Each expression has 0 or more parameters.  And evaluation
strategy is a non-strict evaluation.

Parameter types are

* JSON types
  * boolean
  * string
  * number
  * array
  * object
* expression

But all values must be able to be serialized to JSON.  Because AST of whole
process is serialized and saved as a "transaction log file".  The process can be
rolled back, when an error occur while the process, or user send SIGINT signal,
or `volt history undo` command is executed.  The transaction log file does not
have ID but the saved directory `{id}` does:

```
$VOLTPATH/trx/{id}/{logfile}
```

`{id}` is called transaction ID, a simple serial number assigned `max + 1` like
DB's AUTOINCREMENT.

JSON DSL has the following characteristic:

* Idempotent
* Invertible

## Idempotent

All operators have an idempotency: "even if an expression is executed twice, it
guarantees the existence (not the content) of a requested resource."

One also might think that "why the definition is so ambiguos?" Because, if we
define operator's idempotency as "after an expression was executed twice at
different times, lock.json, filesystem must be the same." But `volt get A`
installs the latest plugin of remote repository.  At the first time and second
time, the repository's HEAD may be different.  But it guarantees that the
existence of specified property of lock.json, and the repository on filesystem.

Here is a more concrete example:

1. Install plugin A, B, C by `volt get A B C`.
2. Uninstall plugin B.
3. Re-run 1's operation by `volt history redo {id}`

At 3, one might think that "3 should raise an error because plugin B is already
uninstalled!" But volt does raise an error, because operators when uninstalling
(`repos/git/delete`, `lockjson/remove`, `plugconf/delete`) does nothing if given
plugin does not exist, like HTTP's DELETE method.  Those operator guarantees
that "the specified resource is deleted after the execution."

## Invertible

All operators have an inverse expression.  Here is the example of JSON DSL when
[tyru/caw.vim](https://github.com/tyru/caw.vim) plugin is installed (it may be
different with latest volt's JSON DSL when installing. this is a simplified
version for easiness).

```json
["vimdir/with-install",
  ["do",
    ["lockjson/add",
      ["repos/git/clone", "github.com/tyru/caw.vim"],
      ["default"]],
    ["plugconf/install", "github.com/tyru/caw.vim"]]]
```

I show you what happens in several steps when you "invert" the expression like
`volt history undo`.

At first, to invert the expression, `$invert` macro is used:

```json
["$invert",
  ["vimdir/with-install",
    ["do",
      ["lockjson/add",
        ["repos/git/clone", "github.com/tyru/caw.vim"],
        ["default"]],
      ["plugconf/install", "github.com/tyru/caw.vim"]]]]
```

`["$invert", ["vimdir/with-install", expr]]` is expanded to
`["vimdir/with-install", ["$invert", expr]]`.  Internally, it is implemented as
calling `Invert()` method of `vimdir/with-install` operator struct.  See "Go
API" section of JSONDSL note (TODO).

```json
["vimdir/with-install",
  ["$invert",
    ["do",
      ["lockjson/add",
        ["repos/git/clone", "github.com/tyru/caw.vim"],
        ["default"]],
      ["plugconf/install", "github.com/tyru/caw.vim"]]]]
```

And `["$invert", ["do", expr1, expr2]]` becomes
`["do", ["$invert", expr2], ["$invert", expr1]]`.
Note that `expr1` and `expr2` becomes reversed order.

```json
["vimdir/with-install",
  ["do",
    ["$invert",
      ["plugconf/install", "github.com/tyru/caw.vim"]],
    ["$invert",
      ["lockjson/add",
        ["repos/git/clone", "github.com/tyru/caw.vim"],
        ["default"]]]]]
```

And
* `["$invert", ["lockjson/add", repos, profiles]]` becomes
  `["lockjson/remove", ["$invert", repos], ["$invert", profiles]]`
* `["$invert", ["plugconf/install", repos]]` becomes
  `["plugconf/delete", ["$invert", repos]]`

```json
["vimdir/with-install",
  ["do",
    ["plugconf/delete", ["$invert", "github.com/tyru/caw.vim"]],
    ["lockjson/remove",
      ["$invert", ["repos/git/clone", "github.com/tyru/caw.vim"]],
      ["$invert", ["default"]]]]]
```

`["$invert", ["repos/git/clone", repos_path]]` becomes
`["repos/git/delete", ["$invert", repos_path]]`.

```json
["vimdir/with-install",
  ["do",
    ["plugconf/delete", ["$invert", "github.com/tyru/caw.vim"]],
    ["lockjson/remove",
      ["repos/git/delete", ["$invert", "github.com/tyru/caw.vim"]],
      ["$invert", ["default"]]]]]
```

And if `$invert` is applied to literals like string, JSON array, it just remains
as-is.

```json
["vimdir/with-install",
  ["do",
    ["plugconf/delete", "github.com/tyru/caw.vim"],
    ["lockjson/remove",
      ["repos/git/delete", "github.com/tyru/caw.vim"],
      ["default"]]]]
```

We can successfully evaluate the inverse expression of the first expression :)

## The implementation of an operator

To achieve goals as mentioned above,
the signature of Go function of an operator is:

```go
func (ctx Context, args []Value) (ret Value, rollback func(), err error)
```

* [Go context](https://golang.org/pkg/context/) makes graceful rollback easy.
* `Value` is the interface which is serializable to JSON.
  All types of JSON DSL must implement this interface.
* `rollback` function is to rollback this operator's process.
  Invoking rollback function after invoking this operator function
  must rollback lock.json, config.toml, filesystem to the previous state.
  * TODO: inverse expression may be enough for rollback?

## Operator responsibility

As the above signature shows, operators must take care the following points:

1. is cancellable with given context, because Go's context is not a magic to
   make it cancellable if receiving it as an argument :)
2. must rollback to the previous state invoking rollback function
  * TODO: inverse expression may be enough for rollback?
3. must guarantee idempotency: it must not destroy user environment if
   expression is executed twice
4. must have an inverse expression (not inverse operator)

### Uninstall operation should be "Recovable"?

If `volt history undo {id}` takes uninstall operation, it executes `git clone`
to install plugin(s) from remote.  But should it be recovable like uninstall
operation just "archive" to specific directory, and `volt history undo {id}`
"unarchive" the repository?

I don't think it is what a user wants to do.  I think, when a user undoes
uninstall operation, it should just clone new repositories from remote.  A user
just wants repositories to be back, of the latest one not the old version.

It is possible to design to map commands to archive/unarchive operations (I
thought it in the initial design).  But I found it is redundant.

## Update operation must be Recovable!

TODO

* `"repos/git/fetch"`
* `"repos/git/reset"`

## JSON DSL API

TODO

## Go API

TODO
