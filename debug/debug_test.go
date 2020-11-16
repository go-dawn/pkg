package debug

import (
	"bytes"
	"errors"
	"testing"
	"unsafe"

	"github.com/go-dawn/pkg/deck"

	"github.com/stretchr/testify/assert"
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

func Test_Debug_DP(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var b bytes.Buffer
		dbg.out = &b
		DP(1)
		assert.Contains(t, b.String(), "int\n  1")
	})

	t.Run("panic", func(t *testing.T) {
		dbg.out = errorWriter{}
		assert.Panics(t, func() {
			DP(1)
		})
	})
}

func Test_Debug_DD(t *testing.T) {
	deck.SetupOsExit()
	defer deck.TeardownOsExit()

	var b bytes.Buffer
	dbg.out = &b
	DD(1)
	assert.Contains(t, b.String(), "int\n  1")
}

func Test_Debug_Debugger_DumpInterface(t *testing.T) {
	t.Parallel()

	var nilCh chan int
	bufferedCh := make(chan int, 1)
	bufferedCh <- 1

	cases := []struct {
		name   string
		input  interface{}
		expect string
	}{
		//// integer
		//{"int", 1, "int\n 1"},
		//{"int8", int8(1), "int8\n 1"},
		//{"int16", int16(1), "int16\n 1"},
		//{"int32", int32(1), "int32\n 1"},
		//{"int64", int64(1), "int64\n 1"},
		//{"uint", uint(1), "int\n 1"},
		//{"uint8", uint8(1), "uint8\n 1"},
		//{"uint16", uint16(1), "uint16\n 1"},
		//{"uint32", uint32(1), "uint32\n 1"},
		//{"uint64", uint64(1), "uint64\n 1"},
		//// float
		//{"float32", float32(1.1), "float32\n 1.1"},
		//{"float64", 1.1, "float64\n 1.1"},
		//// bool
		//{"true", true, "bool\n true"},
		//{"false", false, "bool\n false"},
		////string
		//{"string", "debugger", "string\n \"debugger\""},
		//// nil
		//{"nil", nil, "<nil>"},
		// channel
		{"nil chan", nilCh, "chan int\n <nil>"},
		{"buffered chan", bufferedCh, "(len=1, cap=1)"},
		//// array
		//{"array", [2]int{1, 2}, "[2]int\n (len=2, cap=2)[1, 2]"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, b := getDebugger()

			d.DP(tc.input)

			assert.Contains(t, b.String(), tc.expect)
		})
	}
}

func getDebugger() (*debugger, *bytes.Buffer) {
	b := &bytes.Buffer{}
	d := &debugger{out: b, indent: " ", maxDepth: 5}
	return d, b
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("")
}
