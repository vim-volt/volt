```
 .----------------.  .----------------.  .----------------.  .----------------.
| .--------------. || .--------------. || .--------------. || .--------------. |
| | ____   ____  | || |     ____     | || |   _____      | || |  _________   | |
| ||_  _| |_  _| | || |   .'    `.   | || |  |_   _|     | || | |  _   _  |  | |
| |  \ \   / /   | || |  /  .--.  \  | || |    | |       | || | |_/ | | \_|  | |
| |   \ \ / /    | || |  | |    | |  | || |    | |   _   | || |     | |      | |
| |    \ ' /     | || |  \  `--'  /  | || |   _| |__/ |  | || |    _| |_     | |
| |     \_/      | || |   `.____.'   | || |  |________|  | || |   |_____|    | |
| |              | || |              | || |              | || |              | |
| '--------------' || '--------------' || '--------------' || '--------------' |
 '----------------'  '----------------'  '----------------'  '----------------'

Usage
  volt COMMAND ARGS

Command
  get [-l] [-u] [{repository} ...]
    Install or upgrade given {repository} list, or add local {repository} list as plugins

  rm [-r] [-p] {repository} [{repository2} ...]
    Remove vim plugin from ~/.vim/pack/volt/opt/ directory

  list [-f {text/template string}]
    Vim plugin information extractor.
    Unless -f flag was given, this command shows vim plugins of **current profile** (not all installed plugins) by default.

  enable {repository} [{repository2} ...]
    This is shortcut of:
    volt profile add -current {repository} [{repository2} ...]

  disable {repository} [{repository2} ...]
    This is shortcut of:
    volt profile rm -current {repository} [{repository2} ...]

  edit [-e|--editor {editor}] {repository} [{repository2} ...]
    Open the plugconf file(s) of one or more {repository} for editing.

  profile set {name}
    Set profile name

  profile show {name}
    Show profile info

  profile list
    List all profiles

  profile new {name}
    Create new profile

  profile destroy {name}
    Delete profile

  profile rename {old} {new}
    Rename profile {old} to {new}

  profile add {name} {repository} [{repository2} ...]
    Add one or more repositories to profile

  profile rm {name} {repository} [{repository2} ...]
    Remove one or more repositories to profile

  build [-full]
    Build ~/.vim/pack/volt/ directory

  migrate {migration operation}
    Perform miscellaneous migration operations.
    See 'volt migrate -help' for all available operations

  self-upgrade [-check]
    Upgrade to the latest volt command, or if -check was given, it only checks the newer version is available

  version
    Show volt command version
```

# volt build

```
Usage
  volt build [-help] [-full]

Quick example
  $ volt build        # builds directories under ~/.vim/pack/volt
  $ volt build -full  # full build (remove ~/.vim/pack/volt, and re-create all)

Description
  Build ~/.vim/pack/volt/opt/ directory:
    1. Copy repositories' files into ~/.vim/pack/volt/opt/
      * If the repository is git repository, extract files from locked revision of tree object and copy them into above vim directories
      * If the repository is static repository (imported non-git directory by "volt add" command), copy files into above vim directories
    2. Remove directories from above vim directories, which exist in ~/.vim/pack/volt/build-info.json but not in $VOLTPATH/lock.json

  ~/.vim/pack/volt/build-info.json is a file which holds the information that what vim plugins are installed in ~/.vim/pack/volt/ and its type (git repository, static repository, or system repository), its version. A user normally doesn't need to know the contents of build-info.json .

  If -full option was given, remove all directories in ~/.vim/pack/volt/opt/ , and copy repositories' files into above vim directories.
  Otherwise, it will perform smart build: copy / remove only changed repositories' files.

Options
  -full
        full build
```

# volt disable

```
Usage
  volt disable [-help] {repository} [{repository2} ...]

Quick example
  $ volt disable tyru/caw.vim # will disable tyru/caw.vim plugin in current profile

Description
  This is shortcut of:
  volt profile rm {current profile} {repository} [{repository2} ...]
```

# volt edit

```
Usage
  volt edit [-help] [-e|--editor {editor}] {repository} [{repository2} ...]

Quick example
  $ volt edit tyru/caw.vim # will open the plugconf file for tyru/caw.vim for editing

Description
  Open the plugconf file(s) of one or more {repository} for editing.

  If the -e option was given, use the given editor for editing those files (unless it cannot be found)

  It also calls "volt build" afterwards if modifications were made to the plugconf file(s).
```

# volt enable

```
Usage
  volt enable [-help] {repository} [{repository2} ...]

Quick example
  $ volt enable tyru/caw.vim # will enable tyru/caw.vim plugin in current profile

Description
  This is shortcut of:
  volt profile add {current profile} {repository} [{repository2} ...]
```

# volt get

```
Usage
  volt get [-help] [-l] [-u] [{repository} ...]

Quick example
  $ volt get tyru/caw.vim     # will install tyru/caw.vim plugin
  $ volt get -u tyru/caw.vim  # will upgrade tyru/caw.vim plugin
  $ volt get -l -u            # will upgrade all plugins in current profile
  $ VOLT_DEBUG=1 volt get tyru/caw.vim  # will output more verbosely

  $ mkdir -p ~/volt/repos/localhost/local/hello/plugin
  $ echo 'command! Hello echom "hello"' >~/volt/repos/localhost/local/hello/plugin/hello.vim
  $ volt get localhost/local/hello     # will add the local repository as a plugin
  $ vim -c Hello                       # will output "hello"

Description
  Install or upgrade given {repository} list, or add local {repository} list as plugins.

  And fetch skeleton plugconf from:
    https://github.com/vim-volt/plugconf-templates
  and install it to:
    $VOLTPATH/plugconf/{repository}.vim

Repository List
  {repository} list (=target to perform installing, upgrading, and so on) is determined as followings:
  * If -l option is specified, all plugins in current profile are used
  * If one or more {repository} arguments are specified, the arguments are used

Action
  The action (install, upgrade, or add only) is determined as follows:
    1. If -u option is specified (upgrade):
      * Upgrade git repositories in {repository} list (static repositories are ignored).
      * Add {repository} list to lock.json (if not found)
    2. Or (install):
      * Fetch {repository} list from remotes
      * Add {repository} list to lock.json (if not found)

Static repository
    Volt can manage a local directory as a repository. It's called "static repository".
    When you have unpublished plugins, or you want to manage ~/.vim/* files as one repository
    (this is useful when you use profile feature, see "volt help profile" for more details),
    static repository is useful.
    All you need is to create a directory in "$VOLTPATH/repos/<repos>".

    When -u was not specified (install) and given repositories exist, volt does not make a request to clone the repositories.
    Therefore, "volt get" tries to fetch repositories but skip it because the directory exists.
    then it adds repositories to lock.json if not found.

      $ mkdir -p ~/volt/repos/localhost/local/hello/plugin
      $ echo 'command! Hello echom "hello"' >~/volt/repos/localhost/local/hello/plugin/hello.vim
      $ volt get localhost/local/hello     # will add the local repository as a plugin
      $ vim -c Hello                       # will output "hello"

Repository path
  {repository}'s format is one of the followings:

  1. {user}/{name}
       This is same as "github.com/{user}/{name}"
  2. {site}/{user}/{name}
  3. https://{site}/{user}/{name}
  4. http://{site}/{user}/{name}

Options
  -l    use all plugins in current profile as targets
  -u    upgrade plugins
```

# volt list

```
Usage
  volt list [-help] [-f {text/template string}]

Quick example
  $ volt list # will list installed plugins

  Show all installed repositories:

  $ volt list -f '{{ range .Repos }}{{ println .Path }}{{ end }}'

  Show repositories used by current profile:

  $ volt list -f '{{ range .Profiles }}{{ if eq $.CurrentProfileName .Name }}{{ range .ReposPath }}{{ println . }}{{ end }}{{ end }}{{ end }}'

  Or (see "Additional property"):

  $ volt list -f '{{ range currentProfile.ReposPath }}{{ println . }}{{ end }}'

Template functions

  json value [prefix [indent]] (string)
    Returns JSON representation of value.
    The argument is same as json.MarshalIndent().

  currentProfile (Profile (see "Structures"))
    Returns current profile

  currentProfile (Profile (see "Structures"))
    Returns given name's profile

  version (string)
    Returns volt version string. format is "v{major}.{minor}.{patch}" (e.g. "v0.3.0")

  versionMajor (number)
    Returns volt major version

  versionMinor (number)
    Returns volt minor version

  versionPatch (number)
    Returns volt patch version

Structures
  This describes the structure of lock.json .
  {
    // lock.json structure compatibility version
    "version": <int64>,

    // Current profile name (e.g. "default")
    "current_profile_name": <string>,

    // All Installed repositories
    // ("volt list" shows current profile's repositories, which is not the same as this)
    "repos": [
      {
        // "git" (git repository) or "static" (static repository)
        "type": <string>,

        // Repository path like "github.com/vim-volt/vim-volt"
        "path": <string>,

        // Git commit hash. if "type" is "static" this property does not exist
        "version": <string>,
      },
    ],

    // Profiles
    "profiles": [
      // Profile name (.e.g. "default")
      "name": <string>,

      // Repositories ("volt list" shows these repositories)
      "repos_path": [ <string> ],
    ]
  }

Description
  Vim plugin information extractor.
  If -f flag is not given, this command shows vim plugins of **current profile** (not all installed plugins) by default.
  If -f flag is given, it renders by given template which can access the information of lock.json .
```

# volt migrate

```
Usage
  volt migrate [-help] {migration operation}

Description
  Perform miscellaneous migration operations.
  See detailed help for 'volt migrate -help {migration operation}'.

Available operations
  lockjson
    converts old lock.json format to the latest format
  plugconf/config-func
    converts s:config() function name to s:on_load_pre() in all plugconf files
```

# volt profile

```
Usage
  profile [-help] {command}

Command
  profile set [-n] {name}
    Set profile name to {name}.

  profile show [-current | {name}]
    Show profile info of {name}.

  profile list
    List all profiles.

  profile new {name}
    Create new profile of {name}. This command does not switch to profile {name}.

  profile destroy {name}
    Delete profile of {name}.
    NOTE: Cannot delete current profile.

  profile rename {old} {new}
    Rename profile {old} to {new}.

  profile add [-current | {name}] {repository} [{repository2} ...]
    Add one or more repositories to profile {name}.

  profile rm [-current | {name}] {repository} [{repository2} ...]
    Remove one or more repositories from profile {name}.

Quick example
  $ volt profile list   # default profile is "default"
  * default
  $ volt profile new foo   # will create profile "foo"
  $ volt profile list
  * default
    foo
  $ volt profile set foo   # will switch profile to "foo"
  $ volt profile list
    default
  * foo

  $ volt profile set default   # on profile "default"

  $ volt enable tyru/caw.vim    # enable loading tyru/caw.vim on current profile
  $ volt profile add foo tyru/caw.vim    # enable loading tyru/caw.vim on "foo" profile

  $ volt disable tyru/caw.vim   # disable loading tyru/caw.vim on current profile
  $ volt profile rm foo tyru/caw.vim    # disable loading tyru/caw.vim on "foo" profile

  $ volt profile destroy foo   # will delete profile "foo"
```

# volt rm

```
Usage
  volt rm [-help] [-r] [-p] {repository} [{repository2} ...]

Quick example
  $ volt rm tyru/caw.vim    # Remove tyru/caw.vim plugin from lock.json
  $ volt rm -r tyru/caw.vim # Remove tyru/caw.vim plugin from lock.json, and remove repository directory
  $ volt rm -p tyru/caw.vim # Remove tyru/caw.vim plugin from lock.json, and remove plugconf
  $ volt rm -r -p tyru/caw.vim # Remove tyru/caw.vim plugin from lock.json, and remove repository directory, plugconf

Description
  Uninstall one or more {repository} from every profile.
  This results in removing vim plugins from ~/.vim/pack/volt/opt/ directory.
  If {repository} is depended by other repositories, this command exits with an error.

  If -r option was given, remove also repository directories of specified repositories.
  If -p option was given, remove also plugconf files of specified repositories.

  {repository} is treated as same format as "volt get" (see "volt get -help").
```

# volt self-upgrade

```
Usage
  volt self-upgrade [-help] [-check]

Description
    Upgrade to the latest volt command, or if -check was given, it only checks the newer version is available.
```

# volt version

```
Usage
  volt version [-help]

Description
  Show current version of volt.
```