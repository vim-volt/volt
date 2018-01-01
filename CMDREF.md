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
  volt get [-help] [-l] [-u] [-v] [{repository} ...]

Quick example
  $ volt get tyru/caw.vim     # will install tyru/caw.vim plugin
  $ volt get -u tyru/caw.vim  # will upgrade tyru/caw.vim plugin
  $ volt get -l -u            # will upgrade all installed plugins
  $ volt get -v tyru/caw.vim  # will output more verbosely

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

  If -v option was specified, output more verbosely.

Repository List
  {repository} list (=target to perform installing, upgrading, and so on) is determined as followings:
  * If -l option is specified, all installed vim plugins (regardless current profile) are used
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
  -l    use all installed repositories as targets
  -u    upgrade repositories
  -v    output more verbosely
```

# volt list

```
Usage
  volt list [-help]

Quick example
  $ volt list # will list installed plugins

Description
  This is shortcut of:
  volt profile show {current profile}
```

# volt migrate

```
Usage
  volt migrate [-help]

Description
    Perform migration of $VOLTPATH/lock.json, which means volt converts old version lock.json structure into the latest version. This is always done automatically when reading lock.json content. For example, 'volt get <repos>' will install plugin, and migrate lock.json structure, and write it to lock.json after all. so the migrated content is written to lock.json automatically.
    But, for example, 'volt list' does not write to lock.json but does read, so every time when running 'volt list' shows warning about lock.json is old.
    To suppress this, running this command simply reads and writes migrated structure to lock.json.
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

  profile use [-current | {name}] vimrc [true | false]
  profile use [-current | {name}] gvimrc [true | false]
    Set vimrc / gvimrc flag to true or false.

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

  $ volt profile use -current vimrc false   # Disable installing vimrc on current profile on "volt build"
  $ volt profile use default gvimrc true   # Enable installing gvimrc on profile default on "volt build"
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
  Uninstall {repository} on every profile.
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