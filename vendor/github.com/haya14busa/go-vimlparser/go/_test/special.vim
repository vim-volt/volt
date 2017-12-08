let s:VimLParser = {}

function! s:VimLParser.new(...)
  " skip .new()
  let obj = copy(self)
  call call(obj.__init__, a:000, obj)
  return obj
endfunction

function! s:ExprTokenizer.token()
  " skip ExprTokenizer.token
endfunction

function! s:StringReader.__init__()
  " skip StringReader.__init__
endfunction

function! s:StringReader.getpos()
  " skip StringReader.getpos
endfunction

function! s:VimLParser.push_context()
  " skip VimLParser.push_context
endfunction

function! s:VimLParser.pop_context()
  " skip VimLParser.push_context
endfunction

function! s:VimLParser.__init__()
  " skip VimLParser.__init__
endfunction

function! s:Compiler.__init__()
  " skip Compiler.__init__
endfunction

function! s:Compiler.out()
  " skip Compiler.out
endfunction

function! s:Compiler.incindent()
  " skip
endfunction

function! s:Compiler.decindent()
  " skip
endfunction

function! s:Compiler.compile_curlynameexpr()
  " skip
endfunction

function! s:Compiler.compile_list()
  " skip
endfunction

function! s:Compiler.compile_curlyname()
  " skip
endfunction

function! s:Compiler.compile_dict()
  " skip
endfunction

let y = s:ExArg()

function! s:ExArg()
  " skip ExArg definition
endfunction

function! s:Err()
  " skip Err
endfunction

let self.hoge = 1
let self.ea.range = 1
let xxx.x = 1
let z = self.ea.range
let xs = range(10)

function! s:Node()
  " skip Node definition
endfunction

call s:Node()

let type = 1
let t = type
let at = a:type

let lhs = {}
let lhs = hoge()

for x in self.builtin_commands
endfor
for x in self.neovim_removed_commands
endfor
for x in self.neovim_additional_commands
endfor
function! s:LvalueParser.pos1() abort
  let pos = self.reader.tell()
endfunction
function! s:LvalueParser.pos2() abort
  let pos = self.reader.tell()
endfunction

let self.ea.forceit = s:TRUE
let self.ea.forceit = s:FALSE
let self.ea.usefilter = s:TRUE
let self.ea.usefilter = s:FALSE
let node.attr.range = s:TRUE
let node.attr.abort = s:TRUE
let node.attr.dict = s:TRUE

" skip
let self.find_command_cache = {}
let self.cache = {}
let self.buf = []
let self.pos = []
let self.context = {}
let toplevel.body = {}

let node.body = []
let node.rlist = []
let node.attr = {'range': 0, 'abort': 0, 'dict': 0}
let node.endfunction = s:NIL
let node.endif = s:NIL
let node.endfor = s:NIL
let node.endtry = s:NIL
let node.else = s:NIL
let node.elseif = s:NIL
let node.catch = []
let node.finally = []

let node.list = []
let node.depth = s:NIL
let node.pattern = s:NIL

let lhs.list = []
" end skip

" do not skip
let node.list = self.parse_lvaluelist()
let node.depth = hoge
let node.pattern = node
let node.rlist = [s:NIL, s:NIL]
let node.rlist = [right, s:NIL]
let node.rlist = F()
" end do not skip

let p = s:VimLParser.new()
let et = s:ExprTokenizer.new(r)
let ep = s:ExprParser.new(r)
let lp = s:LvalueParser.new(r)
let r = s:StringReader.new(lines)

let nl = s:NIL

let list = []
let curly_parts = []
let cmd = s:NIL
let cmd = {'name': name, 'flags': 'USERCMD', 'parser': 'parse_cmd_usercmd'}

" type assertion
let s = left.value
let vn = s:isvarname(node.value)
function! s:cache() abort
  call self.reader.seek_set(x[0])
  return x[1]
endfunction
call F(self.compile(node.left))
call F(self.compile(node.rest))

function! F()
  return node.value
endfunction
" end type assertion
call add(xs, x)
call add(node.value, [key, val])
if cnode.pattern != s:NIL
endif
if node.depth != s:NIL
endif
let rlist = map(a:node.rlist, 'self.compile(v:val)')
call F(a:node.rlist[0] is s:NIL ? 'nil' : self.compile(a:node.rlist[0]))
call F(a:node.rlist[1] is s:NIL ? 'nil' : self.compile(a:node.rlist[1]))
