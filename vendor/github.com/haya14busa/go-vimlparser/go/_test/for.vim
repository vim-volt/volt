for x in xs
  let y = x
  let x = 1
endfor

function! Func() abort
  for y in ys
  endfor
  " for creates scope
  let y = 1
endfunction

for [k,v] in kv
  let s = k
  let k = 1
  let v = 1
endfor
