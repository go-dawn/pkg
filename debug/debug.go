package debug

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-dawn/pkg/deck"
	"github.com/valyala/bytebufferpool"
)

var osExit = deck.OsExit

// dbg default debugger
var dbg = &debugger{
	out:    os.Stdout,
	indent: "..",
}

// DP dumps variables in prettier format
func DP(vars ...interface{}) {
	dbg.DP(vars...)
}

// DD dumps variables in prettier format and exit
func DD(vars ...interface{}) {
	dbg.DD(vars...)
}

type debugger struct {
	indent string
	out    io.Writer
}

// DP dumps variables in prettier format
func (d *debugger) DP(vars ...interface{}) {
	bb := bytebufferpool.Get()
	defer bytebufferpool.Put(bb)

	for i, v := range vars {
		// write index
		d.writeIndex(bb, i+1)

		d.dump(bb, v, 1)
	}

	if _, err := bb.WriteTo(d.out); err != nil {
		panic(err)
	}
}

// DD dumps variables in prettier format and exit
func (d *debugger) DD(vars ...interface{}) {
	d.DP(vars...)
	osExit(0)
}

func (d *debugger) dump(bb *bytebufferpool.ByteBuffer, v interface{}, lvl int) {
	if v == nil {
		_, _ = bb.WriteString("<nil>\n")
		return
	}

	val := reflect.ValueOf(v)
	typ := val.Type()

	_, _ = bb.WriteString(typ.String())

	for typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	kind := typ.Kind()

	indent := strings.Repeat(d.indent, lvl)

	if kind != reflect.Slice && kind != reflect.Map {
		_ = bb.WriteByte('\n')
		_, _ = bb.WriteString(indent)
	}

	switch kind {
	case reflect.Bool:
		bb.B = strconv.AppendBool(bb.B, val.Bool())
		_ = bb.WriteByte(',')
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bb.B = strconv.AppendInt(bb.B, val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		bb.B = strconv.AppendUint(bb.B, val.Uint(), 10)
	case reflect.String:
		_ = bb.WriteByte('"')
		_, _ = bb.WriteString(val.String())
		_ = bb.WriteByte('"')
	case reflect.Chan:
		if val.IsNil() {
			_, _ = bb.WriteString("<nil>\n")
			return
		}
		_, _ = bb.WriteString("len=")
		bb.B = strconv.AppendInt(bb.B, int64(val.Len()), 10)
		_, _ = bb.WriteString(", cap=")
		bb.B = strconv.AppendInt(bb.B, int64(val.Cap()), 10)
	case reflect.Func:
		if val.IsNil() {
			_, _ = bb.WriteString("<nil>\n")
			return
		}
		_, _ = bb.WriteString(fmt.Sprintf("%v", val.Interface()))
	case reflect.Array:
		d.dumpArrayOrSlice(bb, val, lvl)
	case reflect.Slice:
		if val.IsNil() {
			_ = bb.WriteByte('\n')
			_, _ = bb.WriteString(indent)
			_, _ = bb.WriteString("<nil>\n")
			return
		}
		_, _ = bb.WriteString(" (len=")
		bb.B = strconv.AppendInt(bb.B, int64(val.Len()), 10)
		_, _ = bb.WriteString(", cap=")
		bb.B = strconv.AppendInt(bb.B, int64(val.Cap()), 10)
		_, _ = bb.WriteString(")\n")
		_, _ = bb.WriteString(indent)

		d.dumpArrayOrSlice(bb, val, lvl)
	case reflect.Map, reflect.Struct:
		_, _ = bb.WriteString(fmt.Sprintf("%#v", v))
	default:
		_, _ = bb.WriteString(fmt.Sprintf("%v", val.Interface()))
	}

	_ = bb.WriteByte('\n')
}

func (d *debugger) writeIndex(bb *bytebufferpool.ByteBuffer, n int) {
	bb.B = strconv.AppendInt(bb.B, int64(n), 10)
	_ = bb.WriteByte(' ')
}

func (d *debugger) dumpArrayOrSlice(bb *bytebufferpool.ByteBuffer, val reflect.Value, lvl int) {
	isInterface := strings.Contains(val.Type().String(), "interface {}")

	_, _ = bb.WriteString("[")

	for i, l := 0, val.Len(); i < l; i++ {
		last := i == l-1
		d.dumpElement(bb, val.Index(i).Interface(), lvl+1, last, isInterface)
	}

	_ = bb.WriteByte(']')
}

func (d *debugger) dumpElement(bb *bytebufferpool.ByteBuffer, v interface{}, lvl int, last, isInterface bool) {
	if v == nil {
		_, _ = bb.WriteString("<nil>")
		if !last {
			_, _ = bb.WriteString(", ")
		}
		return
	}

	val := reflect.ValueOf(v)
	typ := val.Type()

	originTypStr := typ.String()

	for typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	kind := typ.Kind()

	if isInterface && kind != reflect.Slice {
		_, _ = bb.WriteString("(")
		_, _ = bb.WriteString(originTypStr)
		_, _ = bb.WriteString(")")
	}

	indent := strings.Repeat(d.indent, lvl)

	switch kind {
	case reflect.Bool:
		bb.B = strconv.AppendBool(bb.B, val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bb.B = strconv.AppendInt(bb.B, val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		bb.B = strconv.AppendUint(bb.B, val.Uint(), 10)
	case reflect.String:
		_ = bb.WriteByte('"')
		_, _ = bb.WriteString(val.String())
		_ = bb.WriteByte('"')
	case reflect.Chan:
		if val.IsNil() {
			_, _ = bb.WriteString("<nil>")
			break
		}
		_, _ = bb.WriteString(fmt.Sprintf("%v", val.Interface()))

		_, _ = bb.WriteString("(len=")
		bb.B = strconv.AppendInt(bb.B, int64(val.Len()), 10)
		_, _ = bb.WriteString(", cap=")
		bb.B = strconv.AppendInt(bb.B, int64(val.Cap()), 10)
		_, _ = bb.WriteString(")")

	case reflect.Func:
		if val.IsNil() {
			_, _ = bb.WriteString("<nil>")
			break
		}
		_, _ = bb.WriteString(fmt.Sprintf("%v", val.Interface()))
	case reflect.Array:
		d.dumpArrayOrSlice(bb, val, lvl)
	case reflect.Slice:
		_, _ = bb.WriteString("\n")
		_, _ = bb.WriteString(indent)

		_, _ = bb.WriteString("(")
		_, _ = bb.WriteString(originTypStr)
		_, _ = bb.WriteString(")")

		if val.IsNil() {
			_, _ = bb.WriteString("<nil>")
		} else {
			d.dumpArrayOrSlice(bb, val, lvl)
		}

		if last {
			_, _ = bb.WriteString(",\n")
			_, _ = bb.WriteString(strings.Repeat(d.indent, lvl-1))
			return
		}
	case reflect.Map, reflect.Struct:
		_, _ = bb.WriteString(fmt.Sprintf("%#v", v))
	default:
		_, _ = bb.WriteString(fmt.Sprintf("%v", val.Interface()))
	}

	if !last {
		_, _ = bb.WriteString(", ")
	}
}
