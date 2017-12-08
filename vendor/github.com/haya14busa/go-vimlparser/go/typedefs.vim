" typedfes for autoload/vimlparser.vim

let s:typedefs = {
\   'func': {
\   },
\ }

call extend(s:typedefs.func, {
\   'isalpha': { 'in': ['string'], 'out': ['bool'] },
\   'isalnum': { 'in': ['string'], 'out': ['bool'] },
\   'isdigit': { 'in': ['string'], 'out': ['bool'] },
\   'isodigit': { 'in': ['string'], 'out': ['bool'] },
\   'isxdigit': { 'in': ['string'], 'out': ['bool'] },
\   'iswordc': { 'in': ['string'], 'out': ['bool'] },
\   'iswordc1': { 'in': ['string'], 'out': ['bool'] },
\   'iswhite': { 'in': ['string'], 'out': ['bool'] },
\   'isnamec': { 'in': ['string'], 'out': ['bool'] },
\   'isnamec1': { 'in': ['string'], 'out': ['bool'] },
\   'isargname': { 'in': ['string'], 'out': ['bool'] },
\   'isvarname': { 'in': ['string'], 'out': ['bool'] },
\   'isidc': { 'in': ['string'], 'out': ['bool'] },
\   'isupper': { 'in': ['string'], 'out': ['bool'] },
\   'islower': { 'in': ['string'], 'out': ['bool'] },
\ })

call extend(s:typedefs.func, {
\   'VimLParser.push_context': {
\     'in': ['*VimNode'],
\     'out': [],
\   },
\   'VimLParser.find_context': {
\     'in': ['int'],
\     'out': ['int'],
\   },
\   'VimLParser.add_node': {
\     'in': ['*VimNode'],
\     'out': [],
\   },
\   'VimLParser.check_missing_endfunction': { 'in': ['string', '*pos'], 'out': [] },
\   'VimLParser.check_missing_endif': { 'in': ['string', '*pos'], 'out': [] },
\   'VimLParser.check_missing_endtry': { 'in': ['string', '*pos'], 'out': [] },
\   'VimLParser.check_missing_endwhile': { 'in': ['string', '*pos'], 'out': [] },
\   'VimLParser.check_missing_endfor': { 'in': ['string', '*pos'], 'out': [] },
\   'VimLParser.parse': {
\     'in': ['*StringReader'],
\     'out': ['*VimNode'],
\   },
\   'VimLParser.parse_pattern': {
\     'in': ['string'],
\     'out': ['string', 'string'],
\   },
\   'VimLParser.find_command': {
\     'in': [],
\     'out': ['*Cmd'],
\   },
\   'VimLParser.read_cmdarg': {
\     'in': [],
\     'out': ['string'],
\   },
\   'VimLParser.separate_nextcmd': {
\     'in': [],
\     'out': ['*pos'],
\   },
\   'VimLParser.parse_expr': {
\     'in': [],
\     'out': ['*VimNode'],
\   },
\   'VimLParser.parse_exprlist': {
\     'in': [],
\     'out': ['[]*VimNode'],
\   },
\   'VimLParser.parse_lvalue_func': {
\     'in': [],
\     'out': ['*VimNode'],
\   },
\   'VimLParser.parse_lvalue': {
\     'in': [],
\     'out': ['*VimNode'],
\   },
\   'VimLParser.parse_lvaluelist': {
\     'in': [],
\     'out': ['[]*VimNode'],
\   },
\   'VimLParser.parse_letlhs': {
\     'in': [],
\     'out': ['*lhs'],
\   },
\   'VimLParser.ends_excmds': {
\     'in': ['string'],
\     'out': ['bool'],
\   },
\
\   'VimLParser._parse_command': {
\     'in': ['string'],
\     'out': [],
\   },
\ })

call extend(s:typedefs.func, {
\   'ExprTokenizer.__init__': {
\     'in': ['*StringReader'],
\     'out': [],
\   },
\   'ExprTokenizer.token': {
\     'in': ['int', 'string', '*pos'],
\     'out': ['*ExprToken'],
\   },
\   'ExprTokenizer.peek': { 'in': [], 'out': ['*ExprToken'] },
\   'ExprTokenizer.get': { 'in': [], 'out': ['*ExprToken'] },
\   'ExprTokenizer.get2': { 'in': [], 'out': ['*ExprToken'] },
\   'ExprTokenizer.get_sstring': { 'in': [], 'out': ['string'] },
\   'ExprTokenizer.get_dstring': { 'in': [], 'out': ['string'] },
\ })

call extend(s:typedefs.func, {
\   'ExprParser.__init__': {
\     'in': ['*StringReader'],
\     'out': [],
\   },
\   'ExprParser.parse': { 'in': [], 'out': ['*VimNode'] },
\   'ExprParser.parse_expr1': { 'in': [], 'out': ['*VimNode'] },
\   'ExprParser.parse_expr2': { 'in': [], 'out': ['*VimNode'] },
\   'ExprParser.parse_expr3': { 'in': [], 'out': ['*VimNode'] },
\   'ExprParser.parse_expr4': { 'in': [], 'out': ['*VimNode'] },
\   'ExprParser.parse_expr5': { 'in': [], 'out': ['*VimNode'] },
\   'ExprParser.parse_expr6': { 'in': [], 'out': ['*VimNode'] },
\   'ExprParser.parse_expr7': { 'in': [], 'out': ['*VimNode'] },
\   'ExprParser.parse_expr8': { 'in': [], 'out': ['*VimNode'] },
\   'ExprParser.parse_expr9': { 'in': [], 'out': ['*VimNode'] },
\   'ExprParser.parse_dot': { 'in': ['*ExprToken', '*VimNode'], 'out': ['*VimNode'] },
\   'ExprParser.parse_identifier': { 'in': [], 'out': ['*VimNode'] },
\   'ExprParser.parse_curly_parts': { 'in': [], 'out': ['[]*VimNode'] },
\ })

call extend(s:typedefs.func, {
\   'LvalueParser.parse': { 'in': [], 'out': ['*VimNode'] },
\   'LvalueParser.parse_lv8': { 'in': [], 'out': ['*VimNode'] },
\   'LvalueParser.parse_lv9': { 'in': [], 'out': ['*VimNode'] },
\ })

call extend(s:typedefs.func, {
\   'StringReader.__init__': {
\     'in': ['[]string'],
\     'out': [],
\   },
\   'StringReader.eof': {
\     'in': [],
\     'out': ['bool'],
\   },
\   'StringReader.tell': {
\     'in': [],
\     'out': ['int'],
\   },
\   'StringReader.seek_set': {
\     'in': ['int'],
\     'out': [],
\   },
\   'StringReader.seek_cur': {
\     'in': ['int'],
\     'out': [],
\   },
\   'StringReader.seek_end': {
\     'in': ['int'],
\     'out': [],
\   },
\   'StringReader.p': {
\     'in': ['int'],
\     'out': ['string'],
\   },
\   'StringReader.peek': {
\     'in': [],
\     'out': ['string'],
\   },
\   'StringReader.get': {
\     'in': [],
\     'out': ['string'],
\   },
\   'StringReader.peekn': {
\     'in': ['int'],
\     'out': ['string'],
\   },
\   'StringReader.getn': {
\     'in': ['int'],
\     'out': ['string'],
\   },
\   'StringReader.peekline': {
\     'in': [],
\     'out': ['string'],
\   },
\   'StringReader.readline': {
\     'in': [],
\     'out': ['string'],
\   },
\   'StringReader.getstr': {
\     'in': ['*pos', '*pos'],
\     'out': ['string'],
\   },
\   'StringReader.getpos': {
\     'in': [],
\     'out': ['*pos'],
\   },
\   'StringReader.setpos': {
\     'in': ['*pos'],
\     'out': [],
\   },
\   'StringReader.read_alpha': { 'in': [], 'out': ['string'] },
\   'StringReader.read_alnum': { 'in': [], 'out': ['string'] },
\   'StringReader.read_digit': { 'in': [], 'out': ['string'] },
\   'StringReader.read_odigit': { 'in': [], 'out': ['string'] },
\   'StringReader.read_xdigit': { 'in': [], 'out': ['string'] },
\   'StringReader.read_integer': { 'in': [], 'out': ['string'] },
\   'StringReader.read_word': { 'in': [], 'out': ['string'] },
\   'StringReader.read_white': { 'in': [], 'out': ['string'] },
\   'StringReader.read_nonwhite': { 'in': [], 'out': ['string'] },
\   'StringReader.read_name': { 'in': [], 'out': ['string'] },
\ })

call extend(s:typedefs.func, {
\   'Compiler.out': {
\     'in': ['...interface{}'],
\     'out': [],
\   },
\   'Compiler.incindent': {
\     'in': ['string'],
\     'out': [],
\   },
\   'Compiler.compile': {
\     'in': ['*VimNode'],
\     'out': ['interface{}'],
\   },
\   'Compiler.compile_body': {
\     'in': ['[]*VimNode'],
\     'out': [],
\   },
\   'Compiler.compile_toplevel': {
\     'in': ['*VimNode'],
\     'out': ['[]string'],
\   },
\   'Compiler.compile_comment': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_excmd': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_function': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_delfunction': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_return': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_excall': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_let': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_unlet': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_lockvar': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_unlockvar': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_if': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_while': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_for': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_continue': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_break': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_try': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_throw': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_echo': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_echon': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_echohl': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_echomsg': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_echoerr': { 'in': ['*VimNode'], 'out': [] },
\   'Compiler.compile_execute': { 'in': ['*VimNode'], 'out': [] },
\
\   'Compiler.compile_ternary': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_or': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_and': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_equal': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_equalci': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_equalcs': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_nequal': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_nequalci': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_nequalcs': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_greater': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_greaterci': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_greatercs': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_gequal': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_gequalci': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_gequalcs': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_smaller': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_smallerci': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_smallercs': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_sequal': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_sequalci': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_sequalcs': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_match': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_matchci': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_matchcs': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_nomatch': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_nomatchci': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_nomatchcs': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_is': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_isci': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_iscs': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_isnot': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_isnotci': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_isnotcs': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_add': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_subtract': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_concat': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_multiply': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_divide': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_remainder': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_not': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_plus': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_minus': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_subscript': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_slice': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_dot': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_call': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_number': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_string': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_list': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_dict': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_option': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_identifier': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_curlyname': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_env': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_reg': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_curlynamepart': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_curlynameexpr': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_lambda': { 'in': ['*VimNode'], 'out': ['string'] },
\   'Compiler.compile_parenexpr': { 'in': ['*VimNode'], 'out': ['string'] },
\ })

function! ImportTypedefs() abort
  return s:typedefs
endfunction
