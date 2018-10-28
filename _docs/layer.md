
## Layered architecture

The volt commands like `volt get` which may modify lock.json, config.toml,
filesystem, are executed in several steps:

1. (UI layer): Passes subcommand arguments, lock.json & config.toml structure to Gateway layer
2. (Gateway layer): Invokes usecase(s). This layer cannot touch filesystem, do network requests, because it makes unit testing difficult
3. (Usecase layer): Modify files, do network requests, and other business logic

Below is the dependency graph:

```
UI --> Gateway --> Usecase
```

* UI only depends Gateway
* Gateway doesn't know UI
* Gateway only depends Usecase
* Usecase doesn't know Gateway
