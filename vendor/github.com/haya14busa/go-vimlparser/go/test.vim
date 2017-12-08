silent edit `='==GoCompilerTest=='`
setlocal buftype=nowrite
setlocal noswapfile
setlocal bufhidden=wipe
setlocal buftype=nofile
setlocal filetype=diff

let s:vimlparser = vimlparser#import()

let s:sdir = expand('<sfile>:p:h')

function! s:run() abort
  source ./go/gocompiler.vim
  source ./go/typedefs.vim
  :1,$delete
  for vimfile in glob(s:sdir . '/_test/*.vim', 0, 1)
    let okfile = fnamemodify(vimfile, ':r') . '.go'
    let outfile = fnamemodify(vimfile, ':r') . '.out'
    let src = readfile(vimfile)
    let r = s:vimlparser.StringReader.new(src)
    let p = s:vimlparser.VimLParser.new()
    let c = g:ImportGoCompiler().GoCompiler.new(g:ImportTypedefs())
    try
      let out = c.compile(p.parse(r))
      call writefile(out, outfile)
    catch
      call writefile([v:exception], outfile)
    endtry
    let diff = system(printf('diff -u %s %s', shellescape(okfile), shellescape(outfile)))
    if diff == ""
      call s:append(printf('%s => ok', fnamemodify(vimfile, ':.')))
    else
      call s:append(printf("%s => ng", fnamemodify(vimfile, ':.')))
      for l in split(diff, "\n")
        call s:append(l)
      endfor
    endif
  endfor
  syntax enable
  match Error /^.* => ng$/
endfunction

function! s:append(line) abort
  call append(line('$'), a:line)
endfunction

call s:run()

command! Run call s:run()
nnoremap <buffer> <Space><Space> :<C-u>Run<CR>
