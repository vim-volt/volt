#!/bin/bash
vim -u NONE -N --cmd "let &rtp .= ',' . getcwd()" -S go/generate.vim -c ":q"
vim -u NONE -N --cmd "let &rtp .= ',' . getcwd()" -S go/gen_builtin_commands.vim -c ":q"
gofmt -s -w go/*.go

