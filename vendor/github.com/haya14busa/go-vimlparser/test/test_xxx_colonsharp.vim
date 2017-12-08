" XXX: no parse error but invalid expression
echo :
echo #
echo ::
echo ##
echo :#:#
echo #:#:
echo :foo
echo #bar
echo x:y:z
echo x:y:1
echo x:1:y
echo 1:x:y
echo x[::]
echo x[::y]
echo x[y:]
echo x[y:z]
echo x[#:#]
echo x[y:#]
echo {"x"::}
" NOTE: vim stop parse at first colon because ":" is undefined variable
echo {: : :}
" NOTE: curly name
echo {:}
echo {::}
echo {x:y}
echo (0 ? 1 : :)
echo (0 ? : : 1)
