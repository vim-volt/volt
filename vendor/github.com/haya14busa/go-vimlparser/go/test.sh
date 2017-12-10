#!/bin/sh
vim -u NONE -N --cmd "let &rtp .= ',' . getcwd()" -S go/test.vim
