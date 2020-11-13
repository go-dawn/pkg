package debug

import (
	"testing"
)

type x struct {
	a  int
	ch chan<- struct{}
}

func Test_Debug_Dump(t *testing.T) {
	dbg.DP([]func(){})
}
