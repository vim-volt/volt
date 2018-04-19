
[Original (Japanese)](https://gist.github.com/tyru/819e593b2d996321298f6338bbaa34e0)

# Volt refactoring note: JSON DSL and Transaction

## Example of JSON DSL

```json
["label",
  1,
  "installing plugins:",
  ["vimdir/with-install",
    ["parallel",
      ["label",
        2,
        "  github.com/tyru/open-browser.vim ... {{if .Done}}done!{{end}}",
        ["parallel",
          ["lockjson/add",
            ["repos/get", "github.com/tyru/open-browser.vim"],
            ["@", "default"]],
          ["plugconf/install", "github.com/tyru/open-browser.vim"]]],
      ["label",
        3,
        "  github.com/tyru/open-browser-github.vim ... {{if .Done}}done!{{end}}",
        ["parallel",
          ["lockjson/add",
            ["repos/get", "github.com/tyru/open-browser-github.vim"],
            ["@", "default"]],
          ["plugconf/install", "github.com/tyru/open-browser-github.vim"]]]]]]
```

## Wordings

* operator: "callable" object of DSL. this is generic name of function and macro
* function: the name of process
  * e.g. "label"
  * e.g. "parallel"
* macro: like function, but is expanded before execution
  * e.g. "@"
* expression: the form of operator application
  * e.g. `["label", ...]`
  * e.g. `["parallel", ...]`
* transaction log (file): a JSON file which is saved at
  `$VOLTPATH/trx/{id}/log.json`

## Goals

This refactoring allows us or makes it easy to implement the following issues:

1. JSON file of AST (abstract syntax tree) is saved under `$VOLTPATH/trx/{id}/`
2. The history feature (undo, redo, list, ...) like `yum history`
   [#147](https://github.com/vim-volt/volt/issues/147)
    * `volt history undo` executes `[$invert, expr]` for transaction log
    * `volt history redo` just executes saved expression in transaction log
3. Display progress bar [#118](https://github.com/vim-volt/volt/issues/188)
    * Updating progress bars according to `["label", ...]` expression
4. `volt watch` command can be easy to implement
   [#174](https://github.com/vim-volt/volt/issues/174)
    * Current `volt build` implementation installs all repositories of current
      profile, not specific repositories
5. Parallelism
    * Currently each command independently implements it using goroutine, but DSL
      provides higher level parallel processing
6. More detailed unit testing
    * Small component is easy to test
    * And especially "Subcmd layer" is easy because it does not access to
      filesystem
7. Gracefully rollback when an error occurs while processing a DSL [#200](https://github.com/vim-volt/volt/issues/200)

## Layered architecture

The volt commands like `volt get` which may modify lock.json, config.toml([#221](https://github.com/vim-volt/volt/issues/221)),
filesystem, are executed in several steps:

1. (Gateway layer): pass subcommand arguments, lock.json & config.toml structure
   to Subcmd layer
2. (Subcmd layer): Create an AST according to given information
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

This is an application form (called "expression" in this note).
An array literal value is written using `@` operator.

```json
["@", 1, 2, 3]
```

This expression is evaluated to `[1, 2, 3]`.

Each expression has 0 or more parameters.  And evaluation
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

One also might think that "why the it defines the existence, not content?" Because, if we
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
(`repos/delete`, `lockjson/remove`, `plugconf/delete`) does nothing if given
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
      ["repos/get", "github.com/tyru/caw.vim"],
      ["@", "default"]],
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
        ["repos/get", "github.com/tyru/caw.vim"],
        ["@", "default"]],
      ["plugconf/install", "github.com/tyru/caw.vim"]]]]
```

`["$invert", ["vimdir/with-install", expr]]` is expanded to
`["vimdir/with-install", ["$invert", expr]]`.  Internally, it is implemented as
calling `Invert()` method of `vimdir/with-install` operator struct.  See "Go
API" section.

```json
["vimdir/with-install",
  ["$invert",
    ["do",
      ["lockjson/add",
        ["repos/get", "github.com/tyru/caw.vim"],
        ["@", "default"]],
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
        ["repos/get", "github.com/tyru/caw.vim"],
        ["@", "default"]]]]]
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
      ["$invert", ["repos/get", "github.com/tyru/caw.vim"]],
      ["$invert", ["@", "default"]]]]]
```

`["$invert", ["repos/get", path]]` becomes
`["repos/delete", ["$invert", path]]`.

```json
["vimdir/with-install",
  ["do",
    ["plugconf/delete", ["$invert", "github.com/tyru/caw.vim"]],
    ["lockjson/remove",
      ["repos/delete", ["$invert", "github.com/tyru/caw.vim"]],
      ["$invert", ["@", "default"]]]]]
```

And if `$invert` is applied to literals like string, JSON array, it just remains
as-is.

```json
["vimdir/with-install",
  ["do",
    ["plugconf/delete", "github.com/tyru/caw.vim"],
    ["lockjson/remove",
      ["repos/delete", "github.com/tyru/caw.vim"],
      ["@", "default"]]]]
```

We can successfully evaluate the inverse expression of the first expression :)

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
* `"repos/git/update"`

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

## JSON DSL API

TODO: Move to Godoc.

### Macro

All macros has `$` prefixed name for readability.
Macros are not saved in transaction log (expanded before saving).

* `["$invert", expr Expr[* => *]] Expr[* => *]`
  * Returns inverse expression of given expression.
  * Internally, this macro calls `InvertExpr()` method of each operator struct.
  * What value is returned depends on each operator's `InvertExpr()`
    implementation.

* `["$eval", expr Expr[* => *]] Expr[* => *]`
  * Evaluate `expr` at parsing time.
    This is useful to save evaluated value to transaction log,
    instead of its expression.
  * See `repos/git/fetch`, `repos/git/update` for concrete example.

### Basic operators

* `["label", linenum: number, tmpl string, expr Expr[* => R]] R`
  * Render `tmpl` by text/template to `linenum` line (1-origin).
    Returns the evaluated value of `expr`.
  * e.g.
    * `["$invert", ["label", linenum, "msg", expr]]` = `["label", ["$invert", linenum], "revert: \"msg\"", ["$invert", expr]]`
    * See `Label examples` section for more details

* `["do", expr1 Expr[* => R1], ..., expr_last Expr[* => R2]] R2`
  * Executes multiple expressions in series.
  * Returns the evaluated value of the last expression.
  * e.g.
    * `["$invert", ["do", expr1, expr2]]` = `["do", ["$invert", expr1], ["$invert", expr2]]`
      * Note that the arguments are reversed.

* `["parallel", msg string, expr1 Expr[* => R1], ..., expr_last Expr[* => R2]] R2`
  * Executes multiple expressions in parallel.
  * Returns the evaluated value of the last expression.
  * e.g.
    * `["$invert", ["parallel", expr1, expr2]]` = `["parallel", ["$invert", expr2], ["$invert", expr1]]`
      * The arguments are **not** reversed because parallel does not care of
        execution order of given expressions.

### Repository operators

* `["repos/get", path ReposPath] Repos`
  * If the repository does not exist, executes `git clone` on given `path`
    repository and saves to `$VOLTPATH/repos/{path}`.
  * If the repository already exists, returns the repository information.
    The information is the repository information on filesystem, not in lock.json.
    * `volt get A` emits `["lockjson/add", ["repos/get", path], profiles]`
    * If A is git repository, it updates lock.json information with repository
      information.
    * If A is static repository, it does nothing.
  * e.g.
    * `["repos/get", "github.com/tyru/caw.vim"]`
    * `["$invert", ["repos/get", path]]` = `["repos/delete", ["$invert", path]]`

* `["repos/delete", path ReposPath] Repos`
  * If the repository does not exist, it does nothing.
  * If the repository already exists, returns the repository information.
    The information is the repository information on filesystem, not in lock.json.
    * `volt rm -r A` emits `["lockjson/add", ["repos/delete", path], profiles]`
    * `volt rm A` emits `["lockjson/add", ["repos/info", path], profiles]`
    * If A is git repository, it deletes `path` repository's directory.
    * If A is static repository, shows warning `static repository cannot be
      deleted by 'volt rm' command. delete '{path}' manually`
      * To avoid removing local repository accidentally.
  * e.g.
    * `["repos/delete", "github.com/tyru/caw.vim"]`
    * `["$invert", ["repos/delete", path]]` = `["repos/get", ["$invert", path]]`

* `["repos/info", path ReposPath] Repos`
  * Returns `path` repository information.
  * e.g.
    * `["lockjson/add", ["repos/get", path], profiles]`
      * `volt rm A` emits this expression.
    * `["$invert", ["repos/info", path]]` = `["repos/info", ["$invert", path]]`

* `["repos/git/fetch", path ReposPath] head_hash string`
  * Executes `git fetch` on `path` repository.
    Returns the hash string of HEAD.
  * For bare repository, the result HEAD hash string is the hash string of
    default branch's HEAD.
  * e.g.
    * `["$invert", ["repos/git/fetch", path]]` = `["repos/git/fetch", ["$invert", path]]`

* `["repos/git/update", path ReposPath, target_hash string, prev_hash string] void`
  * This fails if the working tree is dirty.
  * If `target_hash` is not merged yet in current branch, try `git merge
    --ff-only` (it raises an error if cannot merge with fast-forward)
  * If `target_hash` is already merged in current branch, try `git reset --hard
    {target_hash}`
  * It does nothing for bare git repository.
  * e.g.
    * `["repos/git/update", "github.com/tyru/caw.vim", ["$eval", ["repos/git/rev-parse", "HEAD", "github.com/tyru/caw.vim"]], ["$eval", ["repos/git/fetch", "github.com/tyru/caw.vim"]]]`
      * To save evaluated hash string in transaction log instead of its
        expression, apply `$eval` to `repos/git/fetch` expression.
    * `["$invert", ["repos/git/update", path, target_hash, prev_hash]]` = `["repos/git/update", ["$invert", path], ["$invert", prev_hash], ["$invert", target_hash]]`

* `["repos/git/rev-parse", str string, path ReposPath] hash string`
  * Returns hash string from `str` argument.  This executes `git rev-parse
    {str}` on `path` repository.
  * e.g.
    * `["repos/git/rev-parse", "HEAD", "github.com/tyru/caw.vim"]`
    * `["$invert", ["repos/git/rev-parse", str, path]]` = `["repos/git/rev-parse", ["$invert", str], ["$invert", path]]`

### lock.json operators

* `["lockjson/add", repos Repos, profiles []string]`
  * Add `repos` information to `repos[]` array in lock.json.
    If `profiles` is not empty, the repository name is added to
    specified profile (`profiles[]` array in lock.json).
  * It fails if specified profile name does not exist.
    * Need to create profile before using `lockjson/profile/add`.
  * e.g.
    * `["lockjson/add", ["repos/get", "github.com/tyru/caw.vim"], ["@", "default"]]`
    * `["$invert", ["lockjson/add", repos, profiles]]` = `["lockjson/remove", ["$invert", repos], ["$invert", profiles]]`

* `["lockjson/profile/add", name string] Profile`
  * Add empty profile named `name` if it does not exist.
    If it exists, do nothing.
    Returns created/existed profile.
  * e.g.
    * `["lockjson/profile/add", "default"]`
    * `["$invert", ["lockjson/profile/add", name]]` = `["lockjson/profile/remove", ["$invert", name]]`

* `["lockjson/profile/remove", name string] Profile`
  * Remove specified profile named `name` if it exists.
    If it does not exist, do nothing.
    Returns removed profile.
  * e.g.
    * `["lockjson/profile/remove", "default"]`
    * `["$invert", ["lockjson/profile/remove", name]]` = `["lockjson/profile/add", ["$invert", name]]`

### Plugconf operators

* `["plugconf/install", path ReposPath] void`
  * Created plugconf of specified repository, or fetch a plugconf file from
    [vim-volt/plugconf-templates](https://github.com/vim-volt/plugconf-templates)
  * e.g.
    * `["plugconf/install", "github.com/tyru/caw.vim"]`
    * `["$invert", ["plugconf/install", path]]` = `["plugconf/delete", ["$invert", path]]`

* `["plugconf/delete", path ReposPath] void`
  * Delete a plugconf file of `path`.
    If it does not exist, do nothing.
  * e.g.
    * `["plugconf/delete", "github.com/tyru/caw.vim"]`
    * `["$invert", ["plugconf/delete", path]]` = `["plugconf/install", ["$invert", path]]`

### Vim directory operators

* `["vimdir/with-install", paths "all" | []ReposPath, expr Expr[* => R]] R`
  * `paths` is the list of repositories to build after `expr` is executed.
    * `"all"` means all repositories of current profile.
  * e.g.
    * `["$invert", ["vimdir/with-install", paths, expr]]` = `["vimdir/with-install", ["$invert", paths], ["$invert", expr]]`
    * See "Why `vimdir/install` and `vimdir/uninstall` operators do not exist?"
      section

### Why `vimdir/install` and `vimdir/uninstall` operators do not exist?

We'll describe why `vimdir/install` and `vimdir/uninstall` operators do not
exist, and `vimdir/with-install` exists instead.

For example, now we have the following expression with `vimdir/uninstall`.  It
removes lock.json, deletes repository, plugconf, and also the repository in vim
directory:

```json
[
  "do",
  ["lockjson/remove",
    {
      "type": "git",
      "path": "github.com/tyru/caw.vim",
      "version": "deadbeefcafebabe"
    },
    ["@", "default"]
  ],
  ["repos/delete", "github.com/tyru/caw.vim"],
  ["plugconf/delete", "github.com/tyru/caw.vim"],
  ["vimdir/uninstall", "github.com/tyru/caw.vim"]
]
```

And below is the inverse expression of above.

```json
[
  "do",
  ["vimdir/install", "github.com/tyru/caw.vim"],
  ["plugconf/install", "github.com/tyru/caw.vim"],
  ["repos/get", "github.com/tyru/caw.vim"],
  ["lockjson/add",
    {
      "type": "git",
      "path": "github.com/tyru/caw.vim",
      "version": "deadbeefcafebabe"
    },
    ["@", "default"]
  ]
]
```

1. Installs the repository to vim directory
   **EVEN THE REPOSITORY DOES NOT EXIST YET!**
2. Installs plugconf
3. Clones repository
4. Add repository information to lock.json

1 must raise an error!
The problem is that `["$invert", ["do", exprs...]]` simply reverses the `exprs`.
We have to install **always** the repository to vim directory after all
expressions.

This is what we expected.

```json
["vimdir/with-install",
  ["github.com/tyru/caw.vim"],
  ["do",
    ["lockjson/remove",
      {
        "type": "git",
        "path": "github.com/tyru/caw.vim",
        "version": "deadbeefcafebabe"
      },
      ["@", "default"]
    ],
    ["repos/delete", "github.com/tyru/caw.vim"],
    ["plugconf/delete", "github.com/tyru/caw.vim"]]]
```

The inverse expression of the above is:

```json
["vimdir/with-install",
  ["github.com/tyru/caw.vim"],
  ["do",
    ["plugconf/install", "github.com/tyru/caw.vim"],
    ["repos/get", "github.com/tyru/caw.vim"],
    ["lockjson/add",
      {
        "type": "git",
        "path": "github.com/tyru/caw.vim",
        "version": "deadbeefcafebabe"
      },
      ["@", "default"]]]]
```

1. Installs plugconf
2. Clones repository
3. Add repository information to lock.json
4. Installs the repository to vim directory (yes!)

We successfully installs [tyru/caw.vim](https://github.com/tyru/caw.vim)
plugin :)

But, of course if we placed `vimdir/with-install` at before `repos/delete` or
`plugconf/delete` not at top-level.

```json
["do",
  ["vimdir/with-install",
    ["github.com/tyru/caw.vim"],
    "dummy"],
  ["lockjson/remove",
    {
      "type": "git",
      "path": "github.com/tyru/caw.vim",
      "version": "deadbeefcafebabe"
    },
    ["@", "default"]
  ],
  ["repos/delete", "github.com/tyru/caw.vim"],
  ["plugconf/delete", "github.com/tyru/caw.vim"]]
```

```json
["do",
  ["plugconf/install", "github.com/tyru/caw.vim"],
  ["repos/get", "github.com/tyru/caw.vim"],
  ["lockjson/add",
    {
      "type": "git",
      "path": "github.com/tyru/caw.vim",
      "version": "deadbeefcafebabe"
    },
    ["@", "default"]],
  ["vimdir/with-install",
    ["github.com/tyru/caw.vim"],
    "dummy"]]
```

But a user does not touch JSON DSL.  In other words, constructing "wrong" AST
must not occur without Volt's bug.

### Install examples

Here is the simple JSON to install
[tyru/caw.vim](https://github.com/tyru/caw.vim) using Git.

```json
["vimdir/with-install",
  ["do",
    ["lockjson/add",
      ["repos/get", "github.com/tyru/caw.vim"],
      ["@", "default"]],
    ["plugconf/install", "github.com/tyru/caw.vim"]]]
```

Here is the inverse expression of above.

```json
["vimdir/with-install",
  ["do",
    ["plugconf/delete", "github.com/tyru/caw.vim"],
    ["lockjson/remove",
      ["repos/delete", "github.com/tyru/caw.vim"],
      ["@", "default"]]]]
```

Here is the JSON to install plugins from local directory (static repository).

```json
["vimdir/with-install",
  ["lockjson/add",
    { ... (repository information of local directory) ... },
    ["@", "default"]],
  ["plugconf/install", "localhost/local/myplugin"]]
```

Here is the inverse expression of above.

```json
["vimdir/with-install",
  ["plugconf/delete", "localhost/local/myplugin"],
  ["lockjson/remove",
    { ... (repository information of local directory) ... },
    ["@", "default"]]]
```

### Label examples

Here is the simple example of installing
[tyru/caw.vim](https://github.com/tyru/caw.vim) plugin.

```json
["label",
  1,
  "installing github.com/caw.vim...",
  ["vimdir/with-install",
    ["do",
      ["lockjson/add",
        ["repos/get", "github.com/tyru/caw.vim"],
        ["@", "default"]],
      ["plugconf/install", "github.com/tyru/caw.vim"]]]]
```

Here is the inverse expression of above.
Note that:

* Message becomes `revert "%s"`
* `$invert` is applied to the third argument, thus the argument of `do` is
  inverted

```json
["label",
  1,
  "revert \"installing github.com/caw.vim...\"",
  ["vimdir/with-install",
    ["do",
      ["plugconf/install", "github.com/tyru/caw.vim"],
      ["lockjson/add",
        ["repos/get", "github.com/tyru/caw.vim"],
        ["@", "default"]]]]]
```

Here is more complex example to install two plugins "tyru/open-browser.vim",
"tyru/open-browser-github.vim".  Two levels of `label` expression exist.

```json
["label",
  1,
  "installing plugins:",
  ["vimdir/with-install",
    ["parallel",
      ["label",
        2,
        "  github.com/tyru/open-browser.vim ... {{if .Done}}done!{{end}}",
        ["parallel",
          ["lockjson/add",
            ["repos/get", "github.com/tyru/open-browser.vim"],
            ["@", "default"]],
          ["plugconf/install", "github.com/tyru/open-browser.vim"]]],
      ["label",
        3,
        "  github.com/tyru/open-browser-github.vim ... {{if .Done}}done!{{end}}",
        ["parallel",
          ["lockjson/add",
            ["repos/get", "github.com/tyru/open-browser-github.vim"],
            ["@", "default"]],
          ["plugconf/install", "github.com/tyru/open-browser-github.vim"]]]]]]
```

### Go API

* dsl package
  * `Execute(Context, Expr) (val Value, rollback func(), err error)`
    * Executes in new transacton.
    * In the given context, if the following keys are missing, returns an error.
      * lock.json information `*lockjson.LockJSON`
      * config.toml information `*config.Config`
    * And this function sets the below key before execution.
      * Transaction ID: assigns `max + 1`

* dsl/types package
  * Value interface
    * `Eval(Context) (val Value, rollback func(), err error)`
      * Evaluate this value. the value of JSON literal just returns itself as-is.
    * `Invert() (Value, error)`
      * Returns inverse expression.
  * Null struct
    * implements Value
  * NullType Type = 1
  * var NullValue Null
  * Bool struct
    * implements Value
  * BoolType Type = 2
  * var TrueValue Bool
  * var FalseValue Bool
  * Number struct
    * implements Value
  * NumberType Type = 3
  * String struct
    * implements Value
  * StringType Type = 4
  * Array struct
    * implements Value
  * ArrayType Type = 5
  * Object struct
    * implements Value
  * ObjectType Type = 6
  * Expr struct
    * implements Value
    * Op Op
    * Args []Value
    * Type Type

* dsl/op package
  * `signature(types Type...) *sigChecker`
  * `sigChecker.check(args []Value) error`
    * Returns nil if the types of args match.
    * Returns non-nil error otherwise.
  * Op interface
    * `Bind(args []Value...) (*Expr, error)`
      * Binds arguments to this operator, and returns an expression.
      * Checks semantics (type) of each argument.
      * Sets the result type of Expr.
    * `InvertExpr(args []Value) (*Expr, error)`
      * Returns inverse expression. This normally inverts operator and arguments
        and call Bind() (not all operators do it).
    * `Execute(ctx Context, args []Value) (val Value, rollback func(), err error)`
      * Executes expression (operator + args).
