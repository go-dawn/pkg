package debug

import (
	"testing"
)

type fake struct {
	A [2]complex64
	b []byte
	B []interface{}
	C chan<- struct{}
	d *int
	f float32
	G float64
	I interface{}
	j []interface{}
	M map[interface{}]interface{}
	n map[string]int64
	R []rune
	S string
	U uint32
	u uint16
	x complex64
	X complex128
	Y *fake
	z **fake
}

func Test_Debug_Dump(t *testing.T) {
	dbg.DP(fake{b: []byte("bbb"), B: []interface{}{[]byte("ccc"), []byte("ddd"), map[uint8]int8{1: 2, 2: 9}}})
}
