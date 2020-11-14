package debug

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/go-dawn/pkg/deck"
	"github.com/valyala/bytebufferpool"
)

type buffer struct {
	*bytebufferpool.ByteBuffer
	indent   string
	maxDepth int
}

var osExit = deck.OsExit

// dbg default debugger
var dbg = &debugger{
	out:      os.Stdout,
	indent:   "..",
	maxDepth: 5,
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
	out      io.Writer
	indent   string
	maxDepth int
}

// DP dumps variables in prettier format
func (d *debugger) DP(vars ...interface{}) {
	b := get(d.indent, d.maxDepth)
	defer put(b)

	for i, v := range vars {
		b.writeIndex(i + 1)
		b.dump(v, 1)
	}

	if _, err := b.WriteTo(d.out); err != nil {
		panic(err)
	}
}

// DD dumps variables in prettier format and exit
func (d *debugger) DD(vars ...interface{}) {
	d.DP(vars...)
	osExit(0)
}

func (b *buffer) dump(v interface{}, lvl int) {
	if v == nil {
		_, _ = b.WriteString("<nil>\n")
		return
	}

	typStr, val, kind := normalize(v)

	_, _ = b.WriteString(typStr)

	indent := strings.Repeat(b.indent, lvl)

	if kind != reflect.Slice && kind != reflect.Map {
		_ = b.WriteByte('\n')
		_, _ = b.WriteString(indent)
	}

	if !val.IsValid() {
		_, _ = b.WriteString("<nil>\n")
		return
	}

	switch kind {
	case reflect.Bool:
		b.B = strconv.AppendBool(b.B, val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.B = strconv.AppendInt(b.B, val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		b.B = strconv.AppendUint(b.B, val.Uint(), 10)
	case reflect.String:
		_ = b.WriteByte('"')
		_, _ = b.WriteString(val.String())
		_ = b.WriteByte('"')
	case reflect.Chan:
		if val.IsNil() {
			_, _ = b.WriteString("<nil>\n")
			return
		}
		_, _ = b.WriteString(fmt.Sprintf("%v", val.Interface()))
		_, _ = b.WriteString("(len=")
		b.B = strconv.AppendInt(b.B, int64(val.Len()), 10)
		_, _ = b.WriteString(", cap=")
		b.B = strconv.AppendInt(b.B, int64(val.Cap()), 10)
		_, _ = b.WriteString(")")
	case reflect.Func:
		if val.IsNil() {
			_, _ = b.WriteString("<nil>\n")
			return
		}
		_, _ = b.WriteString(fmt.Sprintf("%v", val.Interface()))
	case reflect.Array:
		b.dumpArrayOrSlice(val, lvl)
	case reflect.Slice:
		if val.IsNil() {
			_ = b.WriteByte('\n')
			_, _ = b.WriteString(indent)
			_, _ = b.WriteString("<nil>\n")
			return
		}
		_, _ = b.WriteString("\n")
		_, _ = b.WriteString(indent)

		b.dumpArrayOrSlice(val, lvl)
	case reflect.Map:
		b.dumpMap(val, lvl)
	case reflect.Struct:
		b.dumpStruct(val, lvl)
	default:
		_, _ = b.WriteString(fmt.Sprintf("%#v", val.Interface()))
	}

	_ = b.WriteByte('\n')
}

func (b *buffer) dumpArrayOrSlice(val reflect.Value, lvl int) {
	_, _ = b.WriteString("(len=")
	b.B = strconv.AppendInt(b.B, int64(val.Len()), 10)
	_, _ = b.WriteString(", cap=")
	b.B = strconv.AppendInt(b.B, int64(val.Cap()), 10)
	_, _ = b.WriteString(") ")

	isInterface := strings.Contains(val.Type().String(), "interface {}")

	_, _ = b.WriteString("[")

	if lvl <= b.maxDepth {
		for i, l := 0, val.Len(); i < l; i++ {
			last := i == l-1
			b.dumpElem(val.Index(i).Interface(), lvl+1, last, isInterface)
		}
	} else {
		_, _ = b.WriteString("...")
	}

	_ = b.WriteByte(']')
}

func (b *buffer) dumpMap(val reflect.Value, lvl int) {
	indent := strings.Repeat(b.indent, lvl)
	isKeyInterface := strings.Contains(val.Type().String(), "interface {}]")
	isValueInterface := strings.Contains(val.Type().String(), "]interface {}")

	l := val.Len()

	_, _ = b.WriteString("(len=")
	b.B = strconv.AppendInt(b.B, int64(l), 10)
	_, _ = b.WriteString(")\n")
	_, _ = b.WriteString(indent)

	if lvl > b.maxDepth {
		_, _ = b.WriteString("{...}")
		return
	}

	_, _ = b.WriteString("{\n")

	var i int
	for iter := val.MapRange(); iter.Next(); {
		last := i == l-1
		b.dumpMapKey(iter.Key().Interface(), lvl+1, isKeyInterface)
		_, _ = b.WriteString(" : ")
		b.dumpValue(iter.Value().Interface(), lvl+1, last, isValueInterface)
		i++
	}

	_, _ = b.WriteString("\n")
	_, _ = b.WriteString(indent)
	_ = b.WriteByte('}')
}

func (b *buffer) dumpStruct(val reflect.Value, lvl int) {
	indent := strings.Repeat(b.indent, lvl)
	if lvl > b.maxDepth {
		_, _ = b.WriteString("{...}")
		return
	}

	_, _ = b.WriteString("{\n")

	typ := val.Type()

	clone := reflect.New(typ).Elem()
	clone.Set(val)

	for i, l := 0, val.NumField(); i < l; i++ {
		_, _ = b.WriteString(strings.Repeat(b.indent, lvl+1))
		_, _ = b.WriteString(typ.Field(i).Name)
		_, _ = b.WriteString(" : ")

		typStr := typ.Field(i).Type.String()
		_, _ = b.WriteString("(")
		_, _ = b.WriteString(typStr)
		_, _ = b.WriteString(") ")

		f := clone.Field(i)
		/* #nosec G103 */
		f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()

		last := i == l-1
		b.dumpValue(f.Interface(), lvl+1, last, false)
	}

	_, _ = b.WriteString("\n")
	_, _ = b.WriteString(indent)
	_ = b.WriteByte('}')
}

func (b *buffer) dumpElem(v interface{}, lvl int, last, isInterface bool) {
	if v == nil {
		_, _ = b.WriteString("<nil>")
		if !last {
			_, _ = b.WriteString(", ")
		}
		return
	}

	typStr, val, kind := normalize(v)

	if isInterface && kind != reflect.Slice && kind != reflect.Map && kind != reflect.Struct {
		_, _ = b.WriteString("(")
		_, _ = b.WriteString(typStr)
		_, _ = b.WriteString(") ")
	}

	if !val.IsValid() {
		_, _ = b.WriteString("<nil>")
		if !last {
			_, _ = b.WriteString(", ")
		}
		return
	}

	indent := strings.Repeat(b.indent, lvl)

	switch kind {
	case reflect.Bool:
		b.B = strconv.AppendBool(b.B, val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.B = strconv.AppendInt(b.B, val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		b.B = strconv.AppendUint(b.B, val.Uint(), 10)
	case reflect.String:
		_ = b.WriteByte('"')
		_, _ = b.WriteString(val.String())
		_ = b.WriteByte('"')
	case reflect.Chan:
		if val.IsNil() {
			_, _ = b.WriteString("<nil>")
			break
		}
		_, _ = b.WriteString(fmt.Sprintf("%v", val.Interface()))
		_, _ = b.WriteString("(len=")
		b.B = strconv.AppendInt(b.B, int64(val.Len()), 10)
		_, _ = b.WriteString(", cap=")
		b.B = strconv.AppendInt(b.B, int64(val.Cap()), 10)
		_, _ = b.WriteString(")")

	case reflect.Func:
		if val.IsNil() {
			_, _ = b.WriteString("<nil>")
			break
		}
		_, _ = b.WriteString(fmt.Sprintf("%v", val.Interface()))
	case reflect.Array:
		b.dumpArrayOrSlice(val, lvl)
	case reflect.Slice:
		_, _ = b.WriteString("\n")
		_, _ = b.WriteString(indent)

		_, _ = b.WriteString("(")
		_, _ = b.WriteString(typStr)
		_, _ = b.WriteString(") ")

		if val.IsNil() {
			_, _ = b.WriteString("<nil>")
		} else {
			b.dumpArrayOrSlice(val, lvl)
		}

		if last {
			_, _ = b.WriteString(",\n")
			_, _ = b.WriteString(strings.Repeat(b.indent, lvl-1))
			return
		}
	case reflect.Map:
		_, _ = b.WriteString("\n")
		_, _ = b.WriteString(indent)

		_, _ = b.WriteString("(")
		_, _ = b.WriteString(typStr)
		_, _ = b.WriteString(") ")

		if val.IsNil() {
			_, _ = b.WriteString("<nil>")
		} else {
			b.dumpMap(val, lvl)
		}

		if last {
			_, _ = b.WriteString(",\n")
			_, _ = b.WriteString(strings.Repeat(b.indent, lvl-1))
			return
		}
	case reflect.Struct:
		_, _ = b.WriteString("\n")
		_, _ = b.WriteString(indent)

		_, _ = b.WriteString("(")
		_, _ = b.WriteString(typStr)
		_, _ = b.WriteString(") ")

		b.dumpStruct(val, lvl)

		if last {
			_, _ = b.WriteString(",\n")
			_, _ = b.WriteString(strings.Repeat(b.indent, lvl-1))
			return
		}
	default:
		_, _ = b.WriteString(fmt.Sprintf("%v", val.Interface()))
	}

	if !last {
		_, _ = b.WriteString(", ")
	}
}

func (b *buffer) dumpMapKey(key interface{}, lvl int, isInterface bool) {
	indent := strings.Repeat(b.indent, lvl)
	_, _ = b.WriteString(indent)

	if key == nil {
		_, _ = b.WriteString("<nil>")
		return
	}

	typStr, val, kind := normalize(key)

	if isInterface {
		_, _ = b.WriteString("(")
		_, _ = b.WriteString(typStr)
		_, _ = b.WriteString(") ")
	}

	if !val.IsValid() {
		_, _ = b.WriteString("<nil>")
		return
	}

	switch kind {
	case reflect.Bool:
		b.B = strconv.AppendBool(b.B, val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.B = strconv.AppendInt(b.B, val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		b.B = strconv.AppendUint(b.B, val.Uint(), 10)
	case reflect.String:
		_ = b.WriteByte('"')
		_, _ = b.WriteString(val.String())
		_ = b.WriteByte('"')
	case reflect.Chan:
		if val.IsNil() {
			_, _ = b.WriteString("<nil>\n")
			return
		}
		_, _ = b.WriteString(fmt.Sprintf("%v", val.Interface()))
		_, _ = b.WriteString("(len=")
		b.B = strconv.AppendInt(b.B, int64(val.Len()), 10)
		_, _ = b.WriteString(", cap=")
		b.B = strconv.AppendInt(b.B, int64(val.Cap()), 10)
		_, _ = b.WriteString(")")
	case reflect.Struct:
		b.dumpStruct(val, lvl+1)
	default:
		_, _ = b.WriteString(fmt.Sprintf("%#v", val.Interface()))
	}
}

func (b *buffer) dumpValue(v interface{}, lvl int, last, isInterface bool) {
	if v == nil {
		_, _ = b.WriteString("<nil>,")
		if !last {
			_, _ = b.WriteString("\n")
		}
		return
	}

	typStr, val, kind := normalize(v)

	if isInterface {
		_, _ = b.WriteString("(")
		_, _ = b.WriteString(typStr)
		_, _ = b.WriteString(") ")
	}

	if !val.IsValid() {
		_, _ = b.WriteString("<nil>,")
		if !last {
			_, _ = b.WriteString("\n")
		}
		return
	}

	switch kind {
	case reflect.Bool:
		b.B = strconv.AppendBool(b.B, val.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b.B = strconv.AppendInt(b.B, val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		b.B = strconv.AppendUint(b.B, val.Uint(), 10)
	case reflect.String:
		_ = b.WriteByte('"')
		_, _ = b.WriteString(val.String())
		_ = b.WriteByte('"')
	case reflect.Chan:
		if val.IsNil() {
			_, _ = b.WriteString("<nil>")
			break
		}
		_, _ = b.WriteString(fmt.Sprintf("%v", val.Interface()))
		_, _ = b.WriteString("(len=")
		b.B = strconv.AppendInt(b.B, int64(val.Len()), 10)
		_, _ = b.WriteString(", cap=")
		b.B = strconv.AppendInt(b.B, int64(val.Cap()), 10)
		_, _ = b.WriteString(")")
	case reflect.Array:
		b.dumpArrayOrSlice(val, lvl)
	case reflect.Slice:
		if val.IsNil() {
			_, _ = b.WriteString("<nil>")
			break
		}

		b.dumpArrayOrSlice(val, lvl)
	case reflect.Map:
		if val.IsNil() {
			_, _ = b.WriteString("<nil>")
			break
		}
		b.dumpMap(val, lvl+1)
	case reflect.Struct:
		if val.IsNil() {
			_, _ = b.WriteString("<nil>")
			break
		}
		b.dumpStruct(val, lvl+1)
	default:
		_, _ = b.WriteString(fmt.Sprintf("%v", val.Interface()))
	}

	_, _ = b.WriteString(",")
	if !last {
		_, _ = b.WriteString("\n")
	}
}

func normalize(v interface{}) (string, reflect.Value, reflect.Kind) {
	val := reflect.ValueOf(v)
	typ := val.Type()

	typStr := typ.String()

	for typ.Kind() == reflect.Ptr && val.IsValid() {
		val = val.Elem()
		typ = typ.Elem()
	}

	return typStr, val, typ.Kind()
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return &buffer{}
	},
}

func get(indent string, maxDepth int) *buffer {
	b := bufferPool.Get().(*buffer)
	b.ByteBuffer = bytebufferpool.Get()
	b.indent = indent
	b.maxDepth = maxDepth
	return b
}

func put(b *buffer) {
	bytebufferpool.Put(b.ByteBuffer)
	b.ByteBuffer = nil
	bufferPool.Put(b)
}

func (b *buffer) writeIndex(n int) {
	b.B = strconv.AppendInt(b.B, int64(n), 10)
	_ = b.WriteByte(' ')
}
