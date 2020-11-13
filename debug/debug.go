package debug

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

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

	typStr, val, kind := normalize(v)

	_, _ = bb.WriteString(typStr)

	indent := strings.Repeat(d.indent, lvl)

	if kind != reflect.Slice && kind != reflect.Map {
		_ = bb.WriteByte('\n')
		_, _ = bb.WriteString(indent)
	}

	if !val.IsValid() {
		_, _ = bb.WriteString("<nil>\n")
		return
	}

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
	case reflect.Map:
		d.dumpMap(bb, val, lvl)
	case reflect.Struct:
		d.dumpStruct(bb, val, lvl)
	default:
		_, _ = bb.WriteString(fmt.Sprintf("%#v", val.Interface()))
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
		d.dumpElem(bb, val.Index(i).Interface(), lvl+1, last, isInterface)
	}

	_ = bb.WriteByte(']')
}

func (d *debugger) dumpMap(bb *bytebufferpool.ByteBuffer, val reflect.Value, lvl int) {
	indent := strings.Repeat(d.indent, lvl)
	isKeyInterface := strings.Contains(val.Type().String(), "interface {}]")
	isValueInterface := strings.Contains(val.Type().String(), "]interface {}")

	l := val.Len()

	_, _ = bb.WriteString(" (len=")
	bb.B = strconv.AppendInt(bb.B, int64(l), 10)
	_, _ = bb.WriteString(")\n")
	_, _ = bb.WriteString(indent)

	_, _ = bb.WriteString("{\n")

	var i int
	for iter := val.MapRange(); iter.Next(); {
		last := i == l-1
		d.dumpMapKey(bb, iter.Key().Interface(), lvl+1, isKeyInterface)
		_, _ = bb.WriteString(" : ")
		d.dumpValue(bb, iter.Value().Interface(), lvl+1, last, isValueInterface)
		i++
	}

	_, _ = bb.WriteString("\n")
	_, _ = bb.WriteString(indent)
	_ = bb.WriteByte('}')
}

func (d *debugger) dumpStruct(bb *bytebufferpool.ByteBuffer, val reflect.Value, lvl int) {
	indent := strings.Repeat(d.indent, lvl)
	_, _ = bb.WriteString("{\n")

	typ := val.Type()

	clone := reflect.New(typ).Elem()
	clone.Set(val)

	for i, l := 0, val.NumField(); i < l; i++ {
		_, _ = bb.WriteString(strings.Repeat(d.indent, lvl+1))
		_, _ = bb.WriteString(typ.Field(i).Name)
		_, _ = bb.WriteString(" : ")

		typStr := typ.Field(i).Type.String()
		_, _ = bb.WriteString("(")
		_, _ = bb.WriteString(typStr)
		_, _ = bb.WriteString(") ")

		isInterface := strings.Contains(typStr, "interface {}")

		f := clone.Field(i)
		f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()

		last := i == l-1
		d.dumpValue(bb, f.Interface(), lvl+1, last, isInterface)
	}

	//for iter := val.MapRange(); iter.Next(); {
	//	last := i == l-1
	//	d.dumpMapKey(bb, iter.Key().Interface(), lvl+1, isKeyInterface)
	//	_, _ = bb.WriteString(" : ")
	//	d.dumpValue(bb, iter.Value().Interface(), lvl+1, last, isValueInterface)
	//	i++
	//}

	_, _ = bb.WriteString("\n")
	_, _ = bb.WriteString(indent)
	_ = bb.WriteByte('}')
}

func (d *debugger) dumpElem(bb *bytebufferpool.ByteBuffer, v interface{}, lvl int, last, isInterface bool) {
	if v == nil {
		_, _ = bb.WriteString("<nil>")
		if !last {
			_, _ = bb.WriteString(", ")
		}
		return
	}

	typStr, val, kind := normalize(v)

	if isInterface && kind != reflect.Slice {
		_, _ = bb.WriteString("(")
		_, _ = bb.WriteString(typStr)
		_, _ = bb.WriteString(")")
	}

	if !val.IsValid() {
		_, _ = bb.WriteString("<nil>")
		if !last {
			_, _ = bb.WriteString(", ")
		}
		return
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
		_, _ = bb.WriteString(typStr)
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

func (d *debugger) dumpMapKey(bb *bytebufferpool.ByteBuffer, key interface{}, lvl int, isInterface bool) {
	indent := strings.Repeat(d.indent, lvl)
	_, _ = bb.WriteString(indent)

	if key == nil {
		_, _ = bb.WriteString("<nil>")
		return
	}

	typStr, val, kind := normalize(key)

	if isInterface {
		_, _ = bb.WriteString("(")
		_, _ = bb.WriteString(typStr)
		_, _ = bb.WriteString(") ")
	}

	if !val.IsValid() {
		_, _ = bb.WriteString("<nil>")
		return
	}

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
			_, _ = bb.WriteString("<nil>\n")
			return
		}
		_, _ = bb.WriteString("len=")
		bb.B = strconv.AppendInt(bb.B, int64(val.Len()), 10)
		_, _ = bb.WriteString(", cap=")
		bb.B = strconv.AppendInt(bb.B, int64(val.Cap()), 10)
	case reflect.Struct:
		_, _ = bb.WriteString(fmt.Sprintf("%#v", val.Interface()))
	default:
		_, _ = bb.WriteString(fmt.Sprintf("%#v", val.Interface()))
	}
}

func (d *debugger) dumpValue(bb *bytebufferpool.ByteBuffer, v interface{}, lvl int, last, isInterface bool) {
	if v == nil {
		_, _ = bb.WriteString("<nil>,")
		if !last {
			_, _ = bb.WriteString("\n")
		}
		return
	}

	typStr, val, kind := normalize(v)

	if isInterface {
		_, _ = bb.WriteString("(")
		_, _ = bb.WriteString(typStr)
		_, _ = bb.WriteString(")")
	}

	switch kind {
	case reflect.Bool:
		_, _ = bb.WriteString(fmt.Sprintf("%#v", val))
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
	case reflect.Struct:
		d.dumpStruct(bb, val, lvl+1)
	default:
		_, _ = bb.WriteString(fmt.Sprintf("%v", val.Interface()))
	}

	_, _ = bb.WriteString(",")
	if !last {
		_, _ = bb.WriteString("\n")
	}
}

func normalize(v interface{}) (string, reflect.Value, reflect.Kind) {
	val := reflect.ValueOf(v)
	typ := val.Type()

	typStr := typ.String()

	for typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	return typStr, val, typ.Kind()
}
