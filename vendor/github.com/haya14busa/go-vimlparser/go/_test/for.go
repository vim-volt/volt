for _, x := range xs {
	var y = x
	x = 1
}
func Func() {
	for _, y := range ys {
	}
	// for creates scope
	var y = 1
}

for k, v := range kv {
	var s = k
	k = 1
	v = 1
}
