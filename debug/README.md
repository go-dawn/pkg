# Debug
Provide handy functions for debug.

## Usages
### DP
Dump any type of variables in prettier format. The max nested level for `slice`, `map` and `struct` is 5 for now.

```go
package main

import (
	"unsafe"

	"github.com/go-dawn/pkg/debug"
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

func main() {
	debug.DP(
		fake{
			b: []byte("bbb"),
			B: []interface{}{
				[]byte("ccc"),
				struct {
					ss string
					vv **string
				}{ss: "sss"},
				map[uint8]int8{1: 2, 2: 9}},
			C: make(chan struct{}, 16),
			M: map[interface{}]interface{}{
				make(chan *fake, 2): make(chan *fake, 2),
				make(chan *fake, 2): 23,
			},
			p: unsafe.Pointer(&fake{}),
			Y: []*fake{{}, {}},
			Z: []map[int]int{{1: 1}, {2: 2, 3: 3}},
		},
	)
}
```

And the output is 
```
/path/to/caller.go:35
1 main.fake
  {
    A : ([2]complex64) (len=2, cap=2)[(0+0i), (0+0i)],
    b : ([]byte) (len=3, cap=3)[98, 98, 98],
    B : ([]interface {}) (len=3, cap=3)[
      ([]byte) (len=3, cap=3)[99, 99, 99],
      (struct { ss string; vv **string }) {
        ss : (string) "sss",
        vv : (**string) <nil>,
      },
      (map[uint8]int8) (len=2) {
        1 : 2,
        2 : 9,
      },
    ],
    C : (chan<- struct {}) 0xc00010e0c0(len=0, cap=16),
    d : (*int) <nil>,
    f : (float32) 0,
    G : (float64) 0,
    I : (interface {}) <nil>,
    j : ([]interface {}) <nil>,
    M : (map[interface {}]interface {}) (len=2) {
      (chan *main.fake) 0xc00007e0c0(len=0, cap=2) : (chan *main.fake) 0xc00007e120(len=0, cap=2),
      (chan *main.fake) 0xc00007e180(len=0, cap=2) : (int) 23,
    },
    n : (map[string]int64) <nil>,
    p : (unsafe.Pointer) 0xc00044c140,
    q : (uintptr) 0,
    R : ([]rune) <nil>,
    S : (string) "",
    t : (bool) false,
    U : (uint32) 0,
    u : (uint16) 0,
    x : (complex64) (0+0i),
    X : (complex128) (0+0i),
    Y : ([]*main.fake) (len=2, cap=2)[
      {
        A : ([2]complex64) (len=2, cap=2)[(0+0i), (0+0i)],
        b : ([]byte) <nil>,
        B : ([]interface {}) <nil>,
        C : (chan<- struct {}) <nil>,
        d : (*int) <nil>,
        f : (float32) 0,
        G : (float64) 0,
        I : (interface {}) <nil>,
        j : ([]interface {}) <nil>,
        M : (map[interface {}]interface {}) <nil>,
        n : (map[string]int64) <nil>,
        p : (unsafe.Pointer) <nil>,
        q : (uintptr) 0,
        R : ([]rune) <nil>,
        S : (string) "",
        t : (bool) false,
        U : (uint32) 0,
        u : (uint16) 0,
        x : (complex64) (0+0i),
        X : (complex128) (0+0i),
        Y : ([]*main.fake) <nil>,
        z : (**main.fake) <nil>,
        Z : ([]map[int]int) <nil>,
      },
      {
        A : ([2]complex64) (len=2, cap=2)[(0+0i), (0+0i)],
        b : ([]byte) <nil>,
        B : ([]interface {}) <nil>,
        C : (chan<- struct {}) <nil>,
        d : (*int) <nil>,
        f : (float32) 0,
        G : (float64) 0,
        I : (interface {}) <nil>,
        j : ([]interface {}) <nil>,
        M : (map[interface {}]interface {}) <nil>,
        n : (map[string]int64) <nil>,
        p : (unsafe.Pointer) <nil>,
        q : (uintptr) 0,
        R : ([]rune) <nil>,
        S : (string) "",
        t : (bool) false,
        U : (uint32) 0,
        u : (uint16) 0,
        x : (complex64) (0+0i),
        X : (complex128) (0+0i),
        Y : ([]*main.fake) <nil>,
        z : (**main.fake) <nil>,
        Z : ([]map[int]int) <nil>,
      },
    ],
    z : (**main.fake) <nil>,
    Z : ([]map[int]int) (len=2, cap=2)[
      (len=1) {
        1 : 1,
      },
      (len=2) {
        2 : 2,
        3 : 3,
      },
    ],
  }
```

### DD
Dump variables in prettier format and exit.
