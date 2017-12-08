

let s:vimlparser = vimlparser#import()

let s:sdir = expand('<sfile>:p:h')

function! s:run()
  for vimfile in glob(s:sdir . '/test*.vim', 0, 1)
    let okfile = fnamemodify(vimfile, ':r') . '.ok'
    let outfile = fnamemodify(vimfile, ':r') . '.out'
    let src = readfile(vimfile)
    let r = s:vimlparser.StringReader.new(src)
    if vimfile =~# 'test_neo'
        let l:neovim = 1
    else
        let l:neovim = 0
    endif
    let p = s:vimlparser.VimLParser.new(l:neovim)
    let c = s:vimlparser.Compiler.new()
    try
      let out = c.compile(p.parse(r))
      call writefile(out, outfile)
    catch
      call writefile([v:exception], outfile)
    endtry
    if system(printf('diff %s %s', shellescape(okfile), shellescape(outfile))) == ""
      let line = printf('%s => ok', fnamemodify(vimfile, ':.'))
    else
      let line = printf('%s => ng', fnamemodify(vimfile, ':.'))
    endif
    call append(line('$'), line)
  endfor
  syntax enable
  match Error /^.* => ng$/
endfunction

call s:run()
