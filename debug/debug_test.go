package debug

import (
	"testing"
)

type x struct {
	A int
	b float64
	C chan<- struct{}
	D interface{}
	E complex64
	f []byte
	G map[interface{}]interface{}
}

func Test_Debug_Dump(t *testing.T) {
	a := &x{A: 12, b: 12.4, C: make(chan struct{}), D: x{}, f: []byte("F"), G: map[interface{}]interface{}{"hello": "world"}}
	dbg.DP(&a)
}
