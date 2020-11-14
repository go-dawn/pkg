package debug

import (
	"testing"
	"unsafe"
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
	p unsafe.Pointer
	q uintptr
	R []rune
	S string
	t bool
	U uint32
	u uint16
	x complex64
	X complex128
	Y []*fake
	z **fake
	Z []map[int]int
}

func Test_Debug_Dump(t *testing.T) {
	//dbg.DP(make(chan *fake, 2), fake{
	//	b: []byte("bbb"),
	//	B: []interface{}{
	//		[]byte("ccc"),
	//		struct {
	//			ss string
	//			vv **string
	//		}{ss: "sss"},
	//		map[uint8]int8{1: 2, 2: 9}},
	//	C: make(chan struct{}, 16),
	//	M: map[interface{}]interface{}{
	//		make(chan *fake, 2): make(chan *fake, 2),
	//		make(chan *fake, 2): 23,
	//	},
	//	p: unsafe.Pointer(&fake{}),
	//	Y: []*fake{{}, {}},
	//	Z: []map[int]int{{1: 1}, {2: 2, 3: 3}},
	//},
	//)
	dbg.DP(fake{})
}
