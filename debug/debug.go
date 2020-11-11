package debug

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/go-dawn/pkg/deck"
)

var osExit = deck.OsExit

var defaultDebugger = &debugger{
	out: os.Stdout,
}

// DP dumps variables in prettier format
func DP(vars ...interface{}) {
	defaultDebugger.DP(vars...)
}

// DD dumps variables in prettier format and exit
func DD(vars ...interface{}) {
	defaultDebugger.DD(vars...)
}

type debugger struct {
	buf bytes.Buffer

	out io.Writer
}

// DP dumps variables in prettier format
func (d *debugger) DP(vars ...interface{}) {
	for i, v := range vars {
		// write index
		d.writeIndex(i + 1)

		d.dump(v)

		d.buf.WriteString("\n")
	}

	if _, err := d.buf.WriteTo(d.out); err != nil {
		panic(err)
	}
}

// DD dumps variables in prettier format and exit
func (d *debugger) DD(vars ...interface{}) {
	d.DP(vars...)
	osExit(0)
}

func (d *debugger) dump(v interface{}) {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	d.buf.WriteString(fmt.Sprintf("%s: %v\n", typ.String(), val.Interface()))
}

func (d *debugger) writeIndex(n int) {
	var b [4]byte
	bb := b[:]
	i := len(bb)
	var q int
	for n >= 10 {
		i--
		q = n / 10
		bb[i] = '0' + byte(n-q*10)
		n = q
	}
	i--
	bb[i] = '0' + byte(n)

	d.buf.Write(bb[i:])
	d.buf.WriteByte(' ')
}
