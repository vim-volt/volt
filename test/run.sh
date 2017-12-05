#!/bin/bash

cd $(dirname $0)
if [ $? -ne 0 ]; then
  echo 'cannot change to parent directory' >&2
  exit 1
fi

VOLT=../bin/volt
if ! [ -x "$VOLT" ]; then
  echo "volt command is not executable: $VOLT" >&2
  exit 2
fi

new_voltpath() {
  mktemp -d /tmp/volt-test-XXXXXXXXXX
}

run_volt() {
  OUT=$($VOLT $@ 2>&1)
  EXITCODE=$?
}

inc_indent() {
  INDENT="  $INDENT"
}

dec_indent() {
  INDENT="${INDENT:2}"
}

msg() {
  echo "$INDENT$@" >&2
}

assert() {
  local name=$1; shift
  $@ >/dev/null 2>&1
  if [ $? -eq 0 ]; then
    echo "$INDENT$name ... ok"
    return 0
  else
    echo "$INDENT$name ... NOT ok ($@)"
    return 1
  fi
}

test_repos() {
  local repos=$VOLTPATH/repos/$1
  assert "check git repository (1)" [ -d "$repos" ]
  assert "check git repository (2)" git "--git-dir=$repos/.git" rev-parse HEAD
}

test_plugconf() {
  local plugconf=$VOLTPATH/plugconf/$1.vim
  assert "check plugconf" [ -f "$plugconf" ]
  # TODO: has s:config(), s:loaded_on(), depends()
}

test_vim_repos() {
  local repos=$HOME/.vim/pack/volt/opt/$(convert_vim_repos_name "$1")
  assert "check vim repository" [ -d "$repos" ]
  # TODO: has same content as $VOLTPATH/repos/<repos>
}

convert_vim_repos_name() {
  echo "$1" | sed -E 's/_/__/g' | sed -E 's@/@_@g'
}

test_success_exit() {
  if echo "$OUT" | grep -q -E '\[WARN|ERROR\]'; then
    assert "no error msg" false
  else
    assert "no error msg" true
  fi
  assert "exit code = 0" [ $EXITCODE -eq 0 ]
}

test_failure_exit() {
  if echo "$OUT" | grep -q -E '\[WARN|ERROR\]'; then
    assert "showed error msg" true
  else
    assert "showed error msg" false
  fi
  assert "exit code != 0" [ $EXITCODE -ne 0 ]
}

test_volt_get_one_plugin() {
  msg "* test_volt_get_one_plugin"
  inc_indent

  export VOLTPATH=$(new_voltpath)
  run_volt get tyru/caw.vim

  test_repos github.com/tyru/caw.vim
  test_plugconf github.com/tyru/caw.vim
  test_vim_repos github.com/tyru/caw.vim
  test_success_exit

  dec_indent
}

test_volt_get_two_or_more_plugins() {
  msg "* test_volt_get_two_or_more_plugins"
  inc_indent

  export VOLTPATH=$(new_voltpath)
  run_volt get tyru/caw.vim tyru/skk.vim

  test_repos github.com/tyru/caw.vim
  test_repos github.com/tyru/skk.vim
  test_plugconf github.com/tyru/caw.vim
  test_plugconf github.com/tyru/skk.vim
  test_vim_repos github.com/tyru/caw.vim
  test_vim_repos github.com/tyru/skk.vim
  test_success_exit

  dec_indent
}

test_volt_get_invalid_args() {
  msg "* test_volt_get_invalid_args"
  inc_indent

  export VOLTPATH=$(new_voltpath)
  run_volt get caw.vim

  assert "repos is not cloned" [ ! -d $VOLTPATH/repos/caw.vim ]
  assert "repos is not cloned" [ ! -d $VOLTPATH/repos/github.com/caw.vim ]
  assert "plugconf is not created" [ ! -f "$VOLTPATH/plugconf/caw.vim.vim" ]
  assert "plugconf is not created" [ ! -f "$VOLTPATH/plugconf/github.com/caw.vim.vim" ]
  assert "check vim repository" [ ! -d "$HOME/.vim/pack/volt/opt/$(convert_vim_repos_name "caw.vim")" ]
  test_failure_exit

  dec_indent
}

test_volt_get_not_found() {
  msg "* test_volt_get_not_found"
  inc_indent

  export VOLTPATH=$(new_voltpath)
  run_volt get vim-volt/not_found

  assert "repos is not cloned" [ ! -d $VOLTPATH/repos/github.com/tyru/caw.vim ]
  assert "plugconf is not created" [ ! -f "$VOLTPATH/plugconf/github.com/tyru/caw.vim.vim" ]
  assert "check vim repository" [ -d "$HOME/.vim/pack/volt/opt/$(convert_vim_repos_name "github.com/tyru/caw.vim")" ]
  test_failure_exit

  dec_indent
}

test_volt_get() {
  msg "* test_volt_get"
  inc_indent

  test_volt_get_one_plugin
  test_volt_get_two_or_more_plugins

  test_volt_get_invalid_args
  test_volt_get_not_found

  dec_indent
}

test_volt_rm() {
  msg "* test_volt_rm"
  inc_indent

  # test_volt_rm_one_plugin
  # test_volt_rm_two_or_more_plugins
  #
  # test_volt_rm_invalid_args
  # test_volt_rm_not_found

  dec_indent
}

run_tests() {
  test_volt_get
}

run_tests
