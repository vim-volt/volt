if 1 | syntax on | endif
syntax
syntax enable
syntax list GroupName
syn match pythonError "[&|]\{2,}" display
syntax match qfFileName /^\zs\S[^|]\+\/\ze[^|\/]\+\/[^|\/]\+|/ conceal cchar=+
syntax region jsString start=+"+ skip=+\\\("\|$\)+ end=+"\|$+ contains=jsSpecial,@Spell extend
