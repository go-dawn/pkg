package debug

import (
	"bytes"
	"errors"
	"testing"
	"unsafe"

	"github.com/go-dawn/pkg/deck"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/bytebufferpool"
)

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

func Test_Debug_Buffer_Dump(t *testing.T) {
	t.Parallel()

	var nilCh chan int
	bufferedCh := make(chan int, 1)
	bufferedCh <- 1

	type test struct {
		P string
		p string
	}

	var (
		nilSlice   []interface{}
		nilMap     map[int]int
		nilPointer **int
		nilStruct  *test
	)

	cases := []struct {
		name   string
		input  interface{}
		expect string
		reg    bool
	}{
		// integer
		{"int", 1, "int\n 1", false},
		{"int8", int8(1), "int8\n 1", false},
		{"int16", int16(1), "int16\n 1", false},
		{"int32", int32(1), "int32\n 1", false},
		{"int64", int64(1), "int64\n 1", false},
		{"uint", uint(1), "int\n 1", false},
		{"uint8", uint8(1), "uint8\n 1", false},
		{"uint16", uint16(1), "uint16\n 1", false},
		{"uint32", uint32(1), "uint32\n 1", false},
		{"uint64", uint64(1), "uint64\n 1", false},
		// float
		{"float32", float32(1.1), "float32\n 1.1", false},
		{"float64", 1.1, "float64\n 1.1", false},
		// bool
		{"true", true, "bool\n true", false},
		{"false", false, "bool\n false", false},
		//string
		{"string", "debugger", "string\n \"debugger\"", false},
		// nil
		{"nil", nil, "<nil>", false},
		// channel
		{"nil chan", nilCh, "chan int\n <nil>", false},
		{"buffered chan", bufferedCh, `chan int\n 0x\w{4,}\(len=1, cap=1\)`, true},
		// array
		{"array", [2]int{1, 2}, "[2]int\n (len=2, cap=2)[1, 2]", false},
		// slice
		{"nil slice", nilSlice, "[]interface {}\n <nil>", false},
		{"slice", []interface{}{1, 2.2, []byte("byte")}, "[]interface {}\n (len=3, cap=3)[(int) 1, (float64) 2.2, \n  ([]byte) (len=4, cap=4)[98, 121, 116, 101],\n ]", false},
		// map
		{"nil map", nilMap, "map[int]int\n <nil>", false},
		{"map", map[interface{}]interface{}{
			1: 1.1,
			2: [2]int{2, 2},
		}, "map[interface {}]interface {}\n (len=2) {\n  (int) 1 : (float64) 1.1,\n  (int) 2 : ([2]int) (len=2, cap=2)[2, 2],\n }", false},
		// pointer
		{"nil pointer", nilPointer, "**int\n <nil>", false},
		/* #nosec G103 */
		{"unsafe.Pointer", unsafe.Pointer(&bufferedCh), `unsafe.Pointer\n 0x\w{4,}`, true},
		// complex
		{"complex64", complex(float32(1), float32(1)), "complex64\n (1+1i)", false},
		{"complex128", complex(1, 1), "complex128\n (1+1i)", false},
		// struct
		{"nil struct", nilStruct, "*debug.test\n <nil>", false},
		{"struct", test{P: "P", p: "p"}, "debug.test\n {\n  P : (string) \"P\",\n  p : (string) \"p\",\n }", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := getBuffer()

			b.dump(tc.input, 1)

			if tc.reg {
				assert.Regexp(t, tc.expect, b.String())
			} else {
				assert.Contains(t, b.String(), tc.expect)
			}
		})
	}
}

func Test_Debug_Buffer_Ellipsis(t *testing.T) {
	cases := []struct {
		name   string
		input  interface{}
		expect string
	}{
		{"slice", []int{1}, "[...]"},
		{"map", map[int]int{1: 1}, "{...}"},
		{"struct", struct{ i int }{i: 1}, "{...}"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := getBuffer()
			b.maxDepth = 0

			b.dump(tc.input, 1)

			assert.Contains(t, b.String(), tc.expect)
		})
	}
}

func getBuffer() *buffer {
	return &buffer{
		ByteBuffer: bytebufferpool.Get(),
		indent:     " ",
		maxDepth:   5,
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("")
}
