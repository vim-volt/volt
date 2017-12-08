@echo off
vim -u NONE -N --cmd "let &rtp .= ',' . getcwd()" -S test/run.vim
