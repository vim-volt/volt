
## Layered architecture

The volt commands like `volt get` which may modify lock.json, config.toml([#221](https://github.com/vim-volt/volt/issues/221)),
filesystem, are executed in several steps:

1. (Subcmd layer): pass subcommand arguments, lock.json & config.toml structure
   to Gateway layer
2. (Gateway layer): Create an AST according to given information
    * This layer cannot touch filesystem, because it makes unit testing difficult
3. (Usecase layer): Execute the AST. This note mainly describes this layer's design

Below is the dependency graph:

```
Subcmd --> Gateway --> Usecase
```

* Subcmd only depends Gateway
* Gateway doesn't know Subcmd
* Gateway only depends Usecase
* Usecase doesn't know Gateway
