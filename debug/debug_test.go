package debug

import (
	"io"
	"os"
	"testing"
)

type x struct {
	a  int
	ch chan<- struct{}
}

func Test_Debug_Dump(t *testing.T) {
	m := map[interface{}]interface{}{1.1: 2.2, true: true, false: 23, 1: 1, "2": 1.23, io.Writer(os.Stdout): nil, "xx": x{}}
	var a *int
	dbg.DP(m, []interface{}{a, a, a})
}
