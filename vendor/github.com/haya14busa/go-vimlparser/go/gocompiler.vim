function! ImportGoCompiler()
  return s:
endfunction

call extend(s:, vimlparser#import())

let s:opprec = {}
let s:opprec[s:NODE_TERNARY] = 1
let s:opprec[s:NODE_PARENEXPR] = 1
let s:opprec[s:NODE_OR] = 2
let s:opprec[s:NODE_AND] = 3
let s:opprec[s:NODE_EQUAL] = 4
let s:opprec[s:NODE_EQUALCI] = 4
let s:opprec[s:NODE_EQUALCS] = 4
let s:opprec[s:NODE_NEQUAL] = 4
let s:opprec[s:NODE_NEQUALCI] = 4
let s:opprec[s:NODE_NEQUALCS] = 4
let s:opprec[s:NODE_GREATER] = 4
let s:opprec[s:NODE_GREATERCI] = 4
let s:opprec[s:NODE_GREATERCS] = 4
let s:opprec[s:NODE_GEQUAL] = 4
let s:opprec[s:NODE_GEQUALCI] = 4
let s:opprec[s:NODE_GEQUALCS] = 4
let s:opprec[s:NODE_SMALLER] = 4
let s:opprec[s:NODE_SMALLERCI] = 4
let s:opprec[s:NODE_SMALLERCS] = 4
let s:opprec[s:NODE_SEQUAL] = 4
let s:opprec[s:NODE_SEQUALCI] = 4
let s:opprec[s:NODE_SEQUALCS] = 4
let s:opprec[s:NODE_MATCH] = 4
let s:opprec[s:NODE_MATCHCI] = 4
let s:opprec[s:NODE_MATCHCS] = 4
let s:opprec[s:NODE_NOMATCH] = 4
let s:opprec[s:NODE_NOMATCHCI] = 4
let s:opprec[s:NODE_NOMATCHCS] = 4
let s:opprec[s:NODE_IS] = 4
let s:opprec[s:NODE_ISCI] = 4
let s:opprec[s:NODE_ISCS] = 4
let s:opprec[s:NODE_ISNOT] = 4
let s:opprec[s:NODE_ISNOTCI] = 4
let s:opprec[s:NODE_ISNOTCS] = 4
let s:opprec[s:NODE_ADD] = 5
let s:opprec[s:NODE_SUBTRACT] = 5
let s:opprec[s:NODE_CONCAT] = 5
let s:opprec[s:NODE_MULTIPLY] = 6
let s:opprec[s:NODE_DIVIDE] = 6
let s:opprec[s:NODE_REMAINDER] = 6
let s:opprec[s:NODE_NOT] = 7
let s:opprec[s:NODE_MINUS] = 7
let s:opprec[s:NODE_PLUS] = 7
let s:opprec[s:NODE_SUBSCRIPT] = 8
let s:opprec[s:NODE_SLICE] = 8
let s:opprec[s:NODE_CALL] = 8
let s:opprec[s:NODE_DOT] = 8
let s:opprec[s:NODE_NUMBER] = 9
let s:opprec[s:NODE_STRING] = 9
let s:opprec[s:NODE_LIST] = 9
let s:opprec[s:NODE_DICT] = 9
let s:opprec[s:NODE_OPTION] = 9
let s:opprec[s:NODE_IDENTIFIER] = 9
let s:opprec[s:NODE_CURLYNAME] = 9
let s:opprec[s:NODE_ENV] = 9
let s:opprec[s:NODE_REG] = 9

let s:GoCompiler = {}

function s:GoCompiler.new(...)
  let obj = copy(self)
  call call(obj.__init__, a:000, obj)
  return obj
endfunction

function s:GoCompiler.__init__(typedefs)
  let self.indent = ['']
  let self.lines = []
  let self.scopes = [{}]
  let self.typedefs = a:typedefs
endfunction

function s:GoCompiler.out(...)
  if len(a:000) == 1
    if a:000[0] =~ '^)\+$'
      let self.lines[-1] .= a:000[0]
    else
      call add(self.lines, self.indent[0] . a:000[0])
    endif
  else
    call add(self.lines, self.indent[0] . call('printf', a:000))
  endif
endfunction

function s:GoCompiler.emptyline()
  call add(self.lines, '')
endfunction

function s:GoCompiler.incindent(s)
  call insert(self.indent, self.indent[0] . a:s)
endfunction

function s:GoCompiler.decindent()
  call remove(self.indent, 0)
endfunction

function s:GoCompiler.inscope()
  call insert(self.scopes, {})
endfunction

function s:GoCompiler.descope()
  call remove(self.scopes, 0)
endfunction

function s:GoCompiler.addscope(varid)
  let self.scopes[0][a:varid] = 1
endfunction

function s:GoCompiler.isinscope(varid)
  for scope in self.scopes
    if has_key(scope, a:varid)
      return 1
    endif
  endfor
  return 0
endfunction

function s:GoCompiler.compile(node)
  if a:node.type == s:NODE_TOPLEVEL
    return self.compile_toplevel(a:node)
  elseif a:node.type == s:NODE_COMMENT
    return self.compile_comment(a:node)
  elseif a:node.type == s:NODE_EXCMD
    return self.compile_excmd(a:node)
  elseif a:node.type == s:NODE_FUNCTION
    return self.compile_function(a:node)
  elseif a:node.type == s:NODE_DELFUNCTION
    return self.compile_delfunction(a:node)
  elseif a:node.type == s:NODE_RETURN
    return self.compile_return(a:node)
  elseif a:node.type == s:NODE_EXCALL
    return self.compile_excall(a:node)
  elseif a:node.type == s:NODE_LET
    return self.compile_let(a:node)
  elseif a:node.type == s:NODE_UNLET
    return self.compile_unlet(a:node)
  elseif a:node.type == s:NODE_LOCKVAR
    return self.compile_lockvar(a:node)
  elseif a:node.type == s:NODE_UNLOCKVAR
    return self.compile_unlockvar(a:node)
  elseif a:node.type == s:NODE_IF
    return self.compile_if(a:node)
  elseif a:node.type == s:NODE_WHILE
    return self.compile_while(a:node)
  elseif a:node.type == s:NODE_FOR
    return self.compile_for(a:node)
  elseif a:node.type == s:NODE_CONTINUE
    return self.compile_continue(a:node)
  elseif a:node.type == s:NODE_BREAK
    return self.compile_break(a:node)
  elseif a:node.type == s:NODE_TRY
    return self.compile_try(a:node)
  elseif a:node.type == s:NODE_THROW
    return self.compile_throw(a:node)
  elseif a:node.type == s:NODE_ECHO
    return self.compile_echo(a:node)
  elseif a:node.type == s:NODE_ECHON
    return self.compile_echon(a:node)
  elseif a:node.type == s:NODE_ECHOHL
    return self.compile_echohl(a:node)
  elseif a:node.type == s:NODE_ECHOMSG
    return self.compile_echomsg(a:node)
  elseif a:node.type == s:NODE_ECHOERR
    return self.compile_echoerr(a:node)
  elseif a:node.type == s:NODE_EXECUTE
    return self.compile_execute(a:node)
  elseif a:node.type == s:NODE_TERNARY
    return self.compile_ternary(a:node)
  elseif a:node.type == s:NODE_OR
    return self.compile_or(a:node)
  elseif a:node.type == s:NODE_AND
    return self.compile_and(a:node)
  elseif a:node.type == s:NODE_EQUAL
    return self.compile_equal(a:node)
  elseif a:node.type == s:NODE_EQUALCI
    return self.compile_equalci(a:node)
  elseif a:node.type == s:NODE_EQUALCS
    return self.compile_equalcs(a:node)
  elseif a:node.type == s:NODE_NEQUAL
    return self.compile_nequal(a:node)
  elseif a:node.type == s:NODE_NEQUALCI
    return self.compile_nequalci(a:node)
  elseif a:node.type == s:NODE_NEQUALCS
    return self.compile_nequalcs(a:node)
  elseif a:node.type == s:NODE_GREATER
    return self.compile_greater(a:node)
  elseif a:node.type == s:NODE_GREATERCI
    return self.compile_greaterci(a:node)
  elseif a:node.type == s:NODE_GREATERCS
    return self.compile_greatercs(a:node)
  elseif a:node.type == s:NODE_GEQUAL
    return self.compile_gequal(a:node)
  elseif a:node.type == s:NODE_GEQUALCI
    return self.compile_gequalci(a:node)
  elseif a:node.type == s:NODE_GEQUALCS
    return self.compile_gequalcs(a:node)
  elseif a:node.type == s:NODE_SMALLER
    return self.compile_smaller(a:node)
  elseif a:node.type == s:NODE_SMALLERCI
    return self.compile_smallerci(a:node)
  elseif a:node.type == s:NODE_SMALLERCS
    return self.compile_smallercs(a:node)
  elseif a:node.type == s:NODE_SEQUAL
    return self.compile_sequal(a:node)
  elseif a:node.type == s:NODE_SEQUALCI
    return self.compile_sequalci(a:node)
  elseif a:node.type == s:NODE_SEQUALCS
    return self.compile_sequalcs(a:node)
  elseif a:node.type == s:NODE_MATCH
    return self.compile_match(a:node)
  elseif a:node.type == s:NODE_MATCHCI
    return self.compile_matchci(a:node)
  elseif a:node.type == s:NODE_MATCHCS
    return self.compile_matchcs(a:node)
  elseif a:node.type == s:NODE_NOMATCH
    return self.compile_nomatch(a:node)
  elseif a:node.type == s:NODE_NOMATCHCI
    return self.compile_nomatchci(a:node)
  elseif a:node.type == s:NODE_NOMATCHCS
    return self.compile_nomatchcs(a:node)
  elseif a:node.type == s:NODE_IS
    return self.compile_is(a:node)
  elseif a:node.type == s:NODE_ISCI
    return self.compile_isci(a:node)
  elseif a:node.type == s:NODE_ISCS
    return self.compile_iscs(a:node)
  elseif a:node.type == s:NODE_ISNOT
    return self.compile_isnot(a:node)
  elseif a:node.type == s:NODE_ISNOTCI
    return self.compile_isnotci(a:node)
  elseif a:node.type == s:NODE_ISNOTCS
    return self.compile_isnotcs(a:node)
  elseif a:node.type == s:NODE_ADD
    return self.compile_add(a:node)
  elseif a:node.type == s:NODE_SUBTRACT
    return self.compile_subtract(a:node)
  elseif a:node.type == s:NODE_CONCAT
    return self.compile_concat(a:node)
  elseif a:node.type == s:NODE_MULTIPLY
    return self.compile_multiply(a:node)
  elseif a:node.type == s:NODE_DIVIDE
    return self.compile_divide(a:node)
  elseif a:node.type == s:NODE_REMAINDER
    return self.compile_remainder(a:node)
  elseif a:node.type == s:NODE_NOT
    return self.compile_not(a:node)
  elseif a:node.type == s:NODE_PLUS
    return self.compile_plus(a:node)
  elseif a:node.type == s:NODE_MINUS
    return self.compile_minus(a:node)
  elseif a:node.type == s:NODE_SUBSCRIPT
    return self.compile_subscript(a:node)
  elseif a:node.type == s:NODE_SLICE
    return self.compile_slice(a:node)
  elseif a:node.type == s:NODE_DOT
    return self.compile_dot(a:node)
  elseif a:node.type == s:NODE_CALL
    return self.compile_call(a:node)
  elseif a:node.type == s:NODE_NUMBER
    return self.compile_number(a:node)
  elseif a:node.type == s:NODE_STRING
    return self.compile_string(a:node)
  elseif a:node.type == s:NODE_LIST
    return self.compile_list(a:node)
  elseif a:node.type == s:NODE_DICT
    return self.compile_dict(a:node)
  elseif a:node.type == s:NODE_OPTION
    return self.compile_option(a:node)
  elseif a:node.type == s:NODE_IDENTIFIER
    return self.compile_identifier(a:node)
  elseif a:node.type == s:NODE_CURLYNAME
    return self.compile_curlyname(a:node)
  elseif a:node.type == s:NODE_ENV
    return self.compile_env(a:node)
  elseif a:node.type == s:NODE_REG
    return self.compile_reg(a:node)
  elseif a:node.type == s:NODE_PARENEXPR
    return self.compile_parenexpr(a:node)
  else
    throw self.err('Compiler: unknown node: %s', string(a:node))
  endif
endfunction

function s:GoCompiler.compile_body(body)
  let empty = 1
  for node in a:body
    call self.compile(node)
    if node.type != s:NODE_COMMENT
      let empty = 0
    endif
  endfor
endfunction

function s:GoCompiler.compile_toplevel(node)
  call self.compile_body(a:node.body)
  return self.lines
endfunction

function s:GoCompiler.compile_comment(node)
  call self.out('//%s', a:node.str)
endfunction

function s:GoCompiler.compile_excmd(node)
  throw 'NotImplemented: excmd'
endfunction

function s:GoCompiler.compile_function(node)
  let left = self.compile(a:node.left)
  let rlist = map(a:node.rlist, 'self.compile(v:val)')
  if !empty(rlist) && rlist[-1] == '...'
    let rlist[-1] = 'a000'
  endif
  " type annotation
  let typedef = get(self.typedefs.func, left, {})
  let args = rlist
  let out = ''
  if !empty(typedef)
    let args = []
    let types_in = get(typedef, 'in', [])
    try
      for i in range(len(rlist))
        let args = add(args, rlist[i] . ' ' . types_in[i])
      endfor
    catch
      echomsg v:exception
      echomsg left
      echomsg rlist
      echomsg types_in
      echomsg i
    endtry
    let types_out = get(typedef, 'out', [])
    let out = join(types_out, ', ')
    if len(types_out) > 1
      let out = '(' . out . ')'
    endif
    if out != ''
      let out .= ' '
    endif
  endif
  if left =~ '^\(ExArg\|Node\|Err\)$'
    return
  elseif left =~ '^\(VimLParser\|ExprTokenizer\|ExprParser\|LvalueParser\|StringReader\|Compiler\|RegexpParser\)\.'
    let [_0, struct, name; _] = matchlist(left, '^\(.*\)\.\(.*\)$')
    if name == 'new'
    \ || (struct == 'ExprTokenizer' && name == 'token')
    \ || (struct == 'StringReader' && (name == 'getpos' || name == '__init__'))
    \ || (struct == 'VimLParser' && (name =~ '\(push\|pop\)_context\|__init__'))
    \ || (struct == 'Compiler' && (
    \        name == '__init__'
    \     || name == 'out'
    \     || name == 'incindent'
    \     || name == 'decindent'
    \     || name == 'compile_curlynameexpr'
    \     || name == 'compile_list'
    \     || name == 'compile_curlyname'
    \     || name == 'compile_dict'
    \     || name == 'compile_parenexpr'
    \     ))
      return
    endif
    call self.out('func (self *%s) %s(%s) %s{', struct, name, join(args, ', '), out)
    call self.incindent("\t")
    call self.inscope()
    for r in rlist
      call self.addscope(r)
    endfor
    call self.compile_body(a:node.body)
    call self.descope()
    call self.decindent()
    call self.out('}')
  else
    call self.out('func %s(%s) %s{', left, join(args, ', '), out)
    call self.incindent("\t")
    call self.inscope()
    for r in rlist
      call self.addscope(r)
    endfor
    call self.compile_body(a:node.body)
    call self.descope()
    call self.decindent()
    call self.out('}')
  endif
  call self.emptyline()
endfunction

function s:GoCompiler.compile_delfunction(node)
  throw 'NotImplemented: delfunction'
endfunction

function s:GoCompiler.compile_return(node)
  if a:node.left is s:NIL
    call self.out('return')
  else
    let r = self.compile(a:node.left)
    if r == 'x[1]'
      call self.out('return %s.(*ExprToken)', r)
      return
    elseif r == 'node.value'
      call self.out('return %s.(string)', r)
      return
    endif
    let ms = matchlist(r, '\V\^[]interface{}{\(\.\*\)}\$')
    if len(ms) > 1
      let r = ms[1]
    endif
    call self.out('return %s', r)
  endif
endfunction

function s:GoCompiler.compile_excall(node)
  let left = self.compile(a:node.left)
  if left =~ '^append('
    let [_, list, item; __] = matchlist(left, '^append(\(.\{-}\),\(.*\))$')
    if list == 'node.value'
      call self.out('%s = append(%s,%s)', list, list . '.([]interface{})', item)
    else
      call self.out('%s = %s', list, left)
    endif
    return
  endif
  call self.out('%s', left)
endfunction

function s:GoCompiler.compile_let(node)
  let op = a:node.op
  if op == '.='
    let op = '+='
  endif
  let right = self.compile(a:node.right)
  if a:node.left isnot s:NIL
    let left = self.compile(a:node.left)
    if left =~ '^\(VimLParser\|ExprTokenizer\|ExprParser\|LvalueParser\|StringReader\|Compiler\|RegexpParser\)$'
      return
    elseif left =~ '^\(VimLParser\|ExprTokenizer\|ExprParser\|LvalueParser\|StringReader\|Compiler\|RegexpParser\)\.'
      let left = matchstr(left, '\.\zs.*')
      " throw 'CaonnotImplement: Class.var'
      " echom left
      " =>
      "   builtin_commands
      "   RE_VERY_NOMAGIC
      "   RE_NOMAGIC
      "   RE_MAGIC
      "   RE_VERY_MAGIC
      return
    elseif left =~ '^\v(self\.(find_command_cache|cache|buf|pos|context)|toplevel.body|lhs.list|(node\.(body|attr|else_|elseif|catch|finally|pattern|end(function|if|for|try))))$' && op == '='
      " skip initialization
      return
    elseif left =~ '^\v(node\.(list|depth))$' && op == '='
      if right == 'nil' || right == '[]interface{}{}'
        return
      endif
      call self.out('%s %s %s', left, op, right)
      return
    elseif left =~ 'node.rlist' && op == '='
      if right == '[]interface{}{}'
        return
      endif
      let m = matchstr(right, '\V[]interface{}{\zs\.\*\ze}\$')
      if m != ''
        call self.out('%s = []*VimNode{%s}', left, m)
      else
        call self.out('%s = %s', left, right)
      endif
      return
    elseif left =~ '^\v(list|curly_parts)$' && op == '=' && right == '[]interface{}{}'
      call self.out('var %s []*VimNode', left)
      return
    elseif left == 'cmd' && op == '=' && (right == 'nil' || right =~ '^\Vmap[string]interface{}{')
      if right == 'nil'
        if self.isinscope(left)
          call self.out('cmd = nil')
        else
          call self.out('var cmd *Cmd = nil')
          call self.addscope(left)
        endif
      else
        let m = matchstr(right, '^\Vmap[string]interface{}{\zs\(\.\*\)\ze}\$')
        let rs = []
        for kv in split(m, ', ')
          let [k, v] = split(kv, ':')
          call add(rs, k[1:-2] . ': ' . v)
        endfor
        call self.out('cmd = &Cmd{%s}', join(rs, ', '))
      endif
      return
    elseif left == 's' && right == 'left.value'
      call self.out('var %s %s %s.(string)', left, op, right)
      return
    elseif left =~ '\.'
      call self.out('%s %s %s', left, op, right)
      return
    elseif left == 'lhs' && right =~ '^\Vmap[string]interface{}{'
      call self.out('var lhs = &lhs{}')
      return
    endif
    if self.isinscope(left)
      call self.out('%s %s %s', left, op, right)
    elseif left =~ '\[[^]]*\]$'
      call self.out('%s %s %s', left, op, right)
    else
      call self.out('var %s %s %s', left, op, right)
      call self.addscope(left)
    endif
  else " let [x,y] = ...
    let list = map(a:node.list, 'self.compile(v:val)')
    if a:node.rest isnot s:NIL
      throw 'NotImplemented: let [x,y; z] ='
    endif
    let var = ''
    for l in list
      if l !~ '\.' && l != '_' && !self.isinscope(l)
        let var = 'var '
        call self.addscope(l)
      endif
    endfor
    let left = join(list, ', ')
    call self.out('%s%s %s %s', var, left, op, right)
  endif
endfunction

function s:GoCompiler.compile_unlet(node)
  echom 'NotImplemented: unlet'
endfunction

function s:GoCompiler.compile_lockvar(node)
  throw 'NotImplemented: lockvar'
endfunction

function s:GoCompiler.compile_unlockvar(node)
  throw 'NotImplemented: unlockvar'
endfunction

function s:GoCompiler.compile_if(node)
  call self.out('if %s {', self.compile(a:node.cond))
  call self.incindent("\t")
  call self.inscope()
  call self.compile_body(a:node.body)
  call self.descope()
  call self.decindent()
  for node in a:node.elseif
    call self.out('} else if %s {', self.compile(node.cond))
    call self.incindent("\t")
    call self.inscope()
    call self.compile_body(node.body)
    call self.descope()
    call self.decindent()
  endfor
  if a:node.else isnot s:NIL
    call self.out('} else {')
    call self.incindent("\t")
    call self.inscope()
    call self.compile_body(a:node.else.body)
    call self.descope()
    call self.decindent()
  endif
  call self.out('}')
endfunction

function s:GoCompiler.compile_while(node)
  let cond = self.compile(a:node.cond)
  if cond == '1'
    let cond = ''
  else
    let cond .= ' '
  endif
  call self.out('for %s{', cond)
  call self.incindent("\t")
  call self.compile_body(a:node.body)
  call self.decindent()
  call self.out('}')
endfunction

function s:GoCompiler.compile_for(node)
  if a:node.left isnot s:NIL
    let left = self.compile(a:node.left)
    let right = self.compile(a:node.right)
    call self.out('for _, %s := range %s {', left, right)
    call self.inscope()
    call self.addscope(left)
  else
    let list = map(a:node.list, 'self.compile(v:val)')
    let right = self.compile(a:node.right)
    if a:node.rest isnot s:NIL
      throw 'NotImplemented: for [x,y;z] in ss'
    endif
    let [k, v; _] = list
    call self.out('for %s, %s := range %s {', k, v, right)
    call self.inscope()
    call self.addscope(k)
    call self.addscope(v)
  endif
  call self.incindent("\t")
  call self.compile_body(a:node.body)
  call self.descope()
  call self.decindent()
  call self.out('}')
endfunction

function s:GoCompiler.compile_continue(node)
  call self.out('continue')
endfunction

function s:GoCompiler.compile_break(node)
  call self.out('break')
endfunction

function s:GoCompiler.compile_try(node)
  " throw 'NotImplemented: try'
  echom 'NotImplemented: try'
endfunction

function s:GoCompiler.compile_throw(node)
  call self.out('panic(%s)', self.compile(a:node.left))
endfunction

function s:GoCompiler.compile_echo(node)
  throw 'NotImplemented: echo'
endfunction

function s:GoCompiler.compile_echon(node)
  throw 'NotImplemented: echon'
endfunction

function s:GoCompiler.compile_echohl(node)
  throw 'NotImplemented: echohl'
endfunction

function s:GoCompiler.compile_echomsg(node)
  throw 'NotImplemented: echomsg'
endfunction

function s:GoCompiler.compile_echoerr(node)
  " throw 'NotImplemented: echoerr'
  echom 'NotImplemented: echoerr'
endfunction

function s:GoCompiler.compile_execute(node)
  throw 'NotImplemented: execute'
endfunction

function s:GoCompiler.compile_ternary(node)
  let cond = self.compile(a:node.cond)
  let left = self.compile(a:node.left)
  let right = self.compile(a:node.right)
  if cond =~ '^node\.rlist\[\d\]' && left == '"nil"'
    return printf('func() string { if %s {return %s} else {return %s.(string)} }()', cond, left, right)
  else
    return printf('viml_ternary(%s, %s, %s)', cond, left, right)
  endif
endfunction

function s:GoCompiler.compile_or(node)
  return self.compile_op2(a:node, '||')
endfunction

function s:GoCompiler.compile_and(node)
  return self.compile_op2(a:node, '&&')
endfunction

function s:GoCompiler.compile_equal(node)
  return self.compile_op2(a:node, '==')
endfunction

function s:GoCompiler.compile_equalci(node)
  return printf('viml_equalci(%s, %s)', self.compile(a:node.left), self.compile(a:node.right))
endfunction

function s:GoCompiler.compile_equalcs(node)
  return self.compile_op2(a:node, '==')
endfunction

function s:GoCompiler.compile_nequal(node)
  return self.compile_op2(a:node, '!=')
endfunction

function s:GoCompiler.compile_nequalci(node)
  return printf('!viml_equalci(%s, %s)', self.compile(a:node.left), self.compile(a:node.right))
endfunction

function s:GoCompiler.compile_nequalcs(node)
  return self.compile_op2(a:node, '!=')
endfunction

function s:GoCompiler.compile_greater(node)
  return self.compile_op2(a:node, '>')
endfunction

function s:GoCompiler.compile_greaterci(node)
  throw 'NotImplemented: >?'
endfunction

function s:GoCompiler.compile_greatercs(node)
  throw 'NotImplemented: >#'
endfunction

function s:GoCompiler.compile_gequal(node)
  return self.compile_op2(a:node, '>=')
endfunction

function s:GoCompiler.compile_gequalci(node)
  throw 'NotImplemented: >=?'
endfunction

function s:GoCompiler.compile_gequalcs(node)
  throw 'NotImplemented: >=#'
endfunction

function s:GoCompiler.compile_smaller(node)
  return self.compile_op2(a:node, '<')
endfunction

function s:GoCompiler.compile_smallerci(node)
  throw 'NotImplemented: <?'
endfunction

function s:GoCompiler.compile_smallercs(node)
  throw 'NotImplemented: <#'
endfunction

function s:GoCompiler.compile_sequal(node)
  return self.compile_op2(a:node, '<=')
endfunction

function s:GoCompiler.compile_sequalci(node)
  throw 'NotImplemented: <=?'
endfunction

function s:GoCompiler.compile_sequalcs(node)
  throw 'NotImplemented: <=#'
endfunction

function s:GoCompiler.compile_match(node)
  return printf('viml_eqreg(%s, %s)', self.compile(a:node.left), self.compile(a:node.right))
endfunction

function s:GoCompiler.compile_matchci(node)
  return printf('viml_eqregq(%s, %s)', self.compile(a:node.left), self.compile(a:node.right))
endfunction

function s:GoCompiler.compile_matchcs(node)
  return printf('viml_eqregh(%s, %s)', self.compile(a:node.left), self.compile(a:node.right))
endfunction

function s:GoCompiler.compile_nomatch(node)
  return printf('!viml_eqreg(%s, %s)', self.compile(a:node.left), self.compile(a:node.right))
endfunction

function s:GoCompiler.compile_nomatchci(node)
  return printf('!viml_eqregq(%s, %s, flags=re.IGNORECASE)', self.compile(a:node.left), self.compile(a:node.right))
endfunction

function s:GoCompiler.compile_nomatchcs(node)
  return printf('!viml_eqregh(%s, %s)', self.compile(a:node.left), self.compile(a:node.right))
endfunction

function s:GoCompiler.compile_is(node)
  return self.compile_op2(a:node, '==')
endfunction

function s:GoCompiler.compile_isci(node)
  throw 'NotImplemented: is?'
endfunction

function s:GoCompiler.compile_iscs(node)
  throw 'NotImplemented: is#'
endfunction

function s:GoCompiler.compile_isnot(node)
  return self.compile_op2(a:node, '!=')
endfunction

function s:GoCompiler.compile_isnotci(node)
  throw 'NotImplemented: isnot?'
endfunction

function s:GoCompiler.compile_isnotcs(node)
  throw 'NotImplemented: isnot#'
endfunction

function s:GoCompiler.compile_add(node)
  return self.compile_op2(a:node, '+')
endfunction

function s:GoCompiler.compile_subtract(node)
  return self.compile_op2(a:node, '-')
endfunction

function s:GoCompiler.compile_concat(node)
  return self.compile_op2(a:node, '+')
endfunction

function s:GoCompiler.compile_multiply(node)
  return self.compile_op2(a:node, '*')
endfunction

function s:GoCompiler.compile_divide(node)
  return self.compile_op2(a:node, '/')
endfunction

function s:GoCompiler.compile_remainder(node)
  return self.compile_op2(a:node, '%')
endfunction

function s:GoCompiler.compile_not(node)
  return self.compile_op1(a:node, '!')
endfunction

function s:GoCompiler.compile_plus(node)
  return self.compile_op1(a:node, '+')
endfunction

function s:GoCompiler.compile_minus(node)
  return self.compile_op1(a:node, '-')
endfunction

function s:GoCompiler.compile_subscript(node)
  let left = self.compile(a:node.left)
  let right = self.compile(a:node.right)
  if right =~ '^-\d\+'
    let right = printf('len(%s)%s', left, right)
  endif
  return printf('%s[%s]', left, right)
endfunction

function s:GoCompiler.compile_slice(node)
  throw 'NotImplemented: slice'
endfunction

function s:GoCompiler.compile_dot(node)
  let left = self.compile(a:node.left)
  let right = self.compile(a:node.right)
  let out = printf('%s.%s', left, right)
  let cmds = matchstr(out, 'self\.\zs\(builtin_commands\|neovim_additional_commands\|neovim_removed_commands\)')
  if cmds != ''
    return cmds
  endif
  return out
endfunction

function s:GoCompiler.compile_call(node)
  let rlist = map(a:node.rlist, 'self.compile(v:val)')
  let left = self.compile(a:node.left)
  if left == 'map' && len(rlist) == 2 && rlist[1] == '"self.compile(v:val)"'
    " throw 'NotImplemented: map()'
    return printf(join([
    \   'func() []string {',
    \   'var ss []string',
    \   'for _, vval := range %s {',
    \   'ss = append(ss, %s.(string))',
    \   '}',
    \   'return ss',
    \   '}()',
    \ ], ";"), rlist[0], substitute(rlist[1][1:-2], 'v:val', 'vval', 'g'))
  elseif left == 'call' && rlist[0][0] =~ '[''"]'
    return printf('viml_%s(*%s)', rlist[0][1:-2], rlist[1])
  elseif left =~ 'ExArg'
    return printf('&%s{}', left)
  elseif left == 'isvarname' && len(rlist) == 1 && rlist[0] == 'node.value'
    return printf('%s(%s.(string))', left, rlist[0])
  elseif left == 'self.reader.seek_set' && len(rlist) == 1 && rlist[0] == 'x[0]'
    return printf('%s(%s.(int))', left, rlist[0])
  elseif left == 'self.compile' && len(rlist) == 1 && rlist[0] =~ '\v^node\.(left|rest)$'
    return printf('%s(%s).(string)', left, rlist[0])
  endif
  if left =~ '\.new$'
    let left = 'New' . matchstr(left, '.*\ze\.new$')
  endif
  if index(s:viml_builtin_functions, left) != -1
    if left == 'add'
      let left = 'append'
    elseif left == 'len'
      let left = 'len'
    else
      let left = printf('viml_%s', left)
    endif
  endif
  if left == 'range_'
    let left = 'viml_range'
  endif
  return printf('%s(%s)', left, join(rlist, ', '))
endfunction

function s:GoCompiler.compile_number(node)
  return a:node.value
endfunction

function s:GoCompiler.compile_string(node)
  if a:node.value == '"\<C-V>"'
    " XXX: workaround
    return '`\<C-V>`'
  endif
  if a:node.value[0] == "'"
    let s = substitute(a:node.value[1:-2], "''", "'", 'g')
    return '"' . escape(s, '\"') . '"'
  else
    return a:node.value
  endif
endfunction

function s:GoCompiler.compile_list(node)
  let value = map(a:node.value, 'self.compile(v:val)')
  if empty(value)
    return '[]interface{}{}'
  else
    return printf('[]interface{}{%s}', join(value, ', '))
  endif
endfunction

function s:GoCompiler.compile_dict(node)
  let value = map(a:node.value, 'self.compile(v:val[0]) . ":" . self.compile(v:val[1])')
  if empty(value)
    return 'map[string]interface{}{}'
  else
    return printf('map[string]interface{}{%s}', join(value, ', '))
  endif
endfunction

function s:GoCompiler.compile_option(node)
  throw 'NotImplemented: option'
endfunction

function s:GoCompiler.compile_identifier(node)
  let name = a:node.value
  if name == 'a:000'
    let name = 'a000'
  elseif name == 'v:val'
    let name = 'vval'
  elseif name =~ '^[sa]:'
    let name = name[2:]
  endif
  if name =~ '^\(range\|type\|else\)$' " keywords
    let name .= '_'
  endif
  if name == 'NIL'
    let name = 'nil'
  elseif name == 'TRUE'
    let name = 'true'
  elseif name == 'FALSE'
    let name = 'false'
  endif
  return name
endfunction

function s:GoCompiler.compile_curlyname(node)
  throw 'NotImplemented: curlyname'
endfunction

function s:GoCompiler.compile_env(node)
  throw 'NotImplemented: env'
endfunction

function s:GoCompiler.compile_reg(node)
  throw 'NotImplemented: reg'
endfunction

function s:GoCompiler.compile_parenexpr(node)
  return self.compile(a:node.value)
endfunction

function s:GoCompiler.compile_op1(node, op)
  let left = self.compile(a:node.left)
  if s:opprec[a:node.type] > s:opprec[a:node.left.type]
    let left = '(' . left . ')'
  endif
  return printf('%s%s', a:op, left)
endfunction

function s:GoCompiler.compile_op2(node, op)
  let left = self.compile(a:node.left)
  if s:opprec[a:node.type] > s:opprec[a:node.left.type]
    let left = '(' . left . ')'
  endif
  let right = self.compile(a:node.right)
  if s:opprec[a:node.type] > s:opprec[a:node.right.type]
    let right = '(' . right . ')'
  endif

  if left == 'cnode.pattern' && right == 'nil'
    let right = '""'
  elseif left == 'node.depth' && right == 'nil'
    let right = '0'
  endif

  return printf('%s %s %s', left, a:op, right)
endfunction

let s:viml_builtin_functions = ['abs', 'acos', 'add', 'and', 'append', 'append', 'argc', 'argidx', 'argv', 'argv', 'asin', 'atan', 'atan2', 'browse', 'browsedir', 'bufexists', 'buflisted', 'bufloaded', 'bufname', 'bufnr', 'bufwinnr', 'byte2line', 'byteidx', 'call', 'ceil', 'changenr', 'char2nr', 'cindent', 'clearmatches', 'col', 'complete', 'complete_add', 'complete_check', 'confirm', 'copy', 'cos', 'cosh', 'count', 'cscope_connection', 'cursor', 'cursor', 'deepcopy', 'delete', 'did_filetype', 'diff_filler', 'diff_hlID', 'empty', 'escape', 'eval', 'eventhandler', 'executable', 'exists', 'extend', 'exp', 'expand', 'feedkeys', 'filereadable', 'filewritable', 'filter', 'finddir', 'findfile', 'float2nr', 'floor', 'fmod', 'fnameescape', 'fnamemodify', 'foldclosed', 'foldclosedend', 'foldlevel', 'foldtext', 'foldtextresult', 'foreground', 'function', 'garbagecollect', 'get', 'get', 'getbufline', 'getbufvar', 'getchar', 'getcharmod', 'getcmdline', 'getcmdpos', 'getcmdtype', 'getcwd', 'getfperm', 'getfsize', 'getfontname', 'getftime', 'getftype', 'getline', 'getline', 'getloclist', 'getmatches', 'getpid', 'getpos', 'getqflist', 'getreg', 'getregtype', 'gettabvar', 'gettabwinvar', 'getwinposx', 'getwinposy', 'getwinvar', 'glob', 'globpath', 'has', 'has_key', 'haslocaldir', 'hasmapto', 'histadd', 'histdel', 'histget', 'histnr', 'hlexists', 'hlID', 'hostname', 'iconv', 'indent', 'index', 'input', 'inputdialog', 'inputlist', 'inputrestore', 'inputsave', 'inputsecret', 'insert', 'invert', 'isdirectory', 'islocked', 'items', 'join', 'keys', 'len', 'libcall', 'libcallnr', 'line', 'line2byte', 'lispindent', 'localtime', 'log', 'log10', 'luaeval', 'map', 'maparg', 'mapcheck', 'match', 'matchadd', 'matcharg', 'matchdelete', 'matchend', 'matchlist', 'matchstr', 'max', 'min', 'mkdir', 'mode', 'mzeval', 'nextnonblank', 'nr2char', 'or', 'pathshorten', 'pow', 'prevnonblank', 'printf', 'pumvisible', 'pyeval', 'py3eval', 'range', 'readfile', 'reltime', 'reltimestr', 'remote_expr', 'remote_foreground', 'remote_peek', 'remote_read', 'remote_send', 'remove', 'remove', 'rename', 'repeat', 'resolve', 'reverse', 'round', 'screencol', 'screenrow', 'search', 'searchdecl', 'searchpair', 'searchpairpos', 'searchpos', 'server2client', 'serverlist', 'setbufvar', 'setcmdpos', 'setline', 'setloclist', 'setmatches', 'setpos', 'setqflist', 'setreg', 'settabvar', 'settabwinvar', 'setwinvar', 'sha256', 'shellescape', 'shiftwidth', 'simplify', 'sin', 'sinh', 'sort', 'soundfold', 'spellbadword', 'spellsuggest', 'split', 'sqrt', 'str2float', 'str2nr', 'strchars', 'strdisplaywidth', 'strftime', 'stridx', 'string', 'strlen', 'strpart', 'strridx', 'strtrans', 'strwidth', 'submatch', 'substitute', 'synID', 'synIDattr', 'synIDtrans', 'synconcealed', 'synstack', 'system', 'tabpagebuflist', 'tabpagenr', 'tabpagewinnr', 'taglist', 'tagfiles', 'tempname', 'tan', 'tanh', 'tolower', 'toupper', 'tr', 'trunc', 'type', 'undofile', 'undotree', 'values', 'virtcol', 'visualmode', 'wildmenumode', 'winbufnr', 'wincol', 'winheight', 'winline', 'winnr', 'winrestcmd', 'winrestview', 'winsaveview', 'winwidth', 'writefile', 'xor']

function! s:test()
  let vimfile = 'autoload/vimlparser.vim'
  let pyfile = 'py/vimlparser.py'
  let vimlfunc = 'py/vimlfunc.py'
  let head = readfile(vimlfunc)
  try
    let r = s:StringReader.new(readfile(vimfile))
    let p = s:VimLParser.new()
    let c = s:GoCompiler.new({})
    let lines = c.compile(p.parse(r))
    unlet lines[0 : index(lines, 'NIL = []') - 1]
    let tail = [
    \   'if __name__ == ''__main__'':',
    \   '    main()',
    \ ]
    call writefile(head + lines + tail, pyfile)
  catch
    echoerr substitute(v:throwpoint, '\.\.\zs\d\+', '\=s:numtoname(submatch(0))', 'g') . "\n" . v:exception
  endtry
endfunction

function! s:numtoname(num)
  let sig = printf("function('%s')", a:num)
  for k in keys(s:)
    if type(s:[k]) == type({})
      for name in keys(s:[k])
        if type(s:[k][name]) == type(function('tr')) && string(s:[k][name]) == sig
          return printf('%s.%s', k, name)
        endif
      endfor
    endif
  endfor
  return a:num
endfunction

" call s:test()
