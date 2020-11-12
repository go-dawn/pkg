package debug

import (
	"testing"
)

type x struct {
	a  int
	ch chan<- struct{}
}

func Test_Debug_Dump(t *testing.T) {
	s := [][]chan struct{}{{nil, make(chan struct{}), nil}, {make(chan struct{})}}
	dbg.DP([2]int{1, 1}, s)
}
