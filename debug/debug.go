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
		b.writeNewLine()
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

const (
	strInterface      = "interface {}"
	strInterfaceKey   = strInterface + "]"
	strInterfaceValue = "]" + strInterface
	strNil            = "<nil>"
)

func (b *buffer) dump(v interface{}, lvl int) {
	if b.writeNil(v == nil) {
		return
	}

	typStr, val, kind := normalize(v)

	_, _ = b.WriteString(typStr)
	b.writeNewLine()
	b.writeIndent(lvl)

	if b.writeNil(!val.IsValid()) {
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
		b.writeString(val.String())
	case reflect.Chan:
		if b.writeNil(val.IsNil()) {
			return
		}
		b.writeInterface(val.Interface())
		b.writeLenAndCap(val.Len(), val.Cap())
	case reflect.Func:
		if b.writeNil(val.IsNil()) {
			return
		}
		b.writeInterface(val.Interface())
	case reflect.Array:
		b.dumpArrayOrSlice(val, lvl)
	case reflect.Slice:
		if b.writeNil(val.IsNil()) {
			return
		}
		b.dumpArrayOrSlice(val, lvl)
	case reflect.Map:
		if b.writeNil(val.IsNil()) {
			return
		}
		b.dumpMap(val, lvl)
	case reflect.Struct:
		b.dumpStruct(val, lvl)
	default:
		b.writeInterface(val.Interface())
	}
}

func (b *buffer) dumpArrayOrSlice(val reflect.Value, lvl int) {
	b.writeLenAndCap(val.Len(), val.Cap())

	if b.writeEllipsis(lvl, "[...]") {
		return
	}

	isInterface := strings.Contains(val.Type().String(), strInterface)

	b.writeBracket('[')
	for i, l := 0, val.Len(); i < l; i++ {
		var newLine bool
		b.dumpElem(val.Index(i).Interface(), lvl+1, isInterface, &newLine)

		if i != l-1 {
			b.writeComma()
			if !newLine {
				b.writeSpace()
			}
		}

		if i == l-1 && newLine {
			b.writeComma()
			b.writeNewLine()
			b.writeIndent(lvl)
		}
	}
	b.writeBracket(']')
}

func (b *buffer) dumpMap(val reflect.Value, lvl int) {
	l := val.Len()

	b.writeLen(l)
	b.writeSpace()

	if b.writeEllipsis(lvl, "{...}") {
		return
	}

	b.writeBracket('{')
	b.writeNewLine()

	isInterfaceKey := strings.Contains(val.Type().String(), strInterfaceKey)
	isInterfaceValue := strings.Contains(val.Type().String(), strInterfaceValue)

	var i int
	for iter := val.MapRange(); iter.Next(); {
		b.writeIndent(lvl + 1)

		b.dumpMapKey(iter.Key().Interface(), lvl+1, isInterfaceKey)

		b.writeColon()

		b.dumpValue(iter.Value().Interface(), lvl+1, isInterfaceValue)

		b.writeComma()
		if i != l-1 {
			b.writeNewLine()
		}

		i++
	}

	b.writeNewLine()
	b.writeIndent(lvl)
	b.writeBracket('}')
}

func (b *buffer) dumpStruct(val reflect.Value, lvl int) {
	if b.writeEllipsis(lvl, "{...}") {
		return
	}

	b.writeBracket('{')
	b.writeNewLine()

	typ := val.Type()

	clone := reflect.New(typ).Elem()
	clone.Set(val)

	for i, l := 0, val.NumField(); i < l; i++ {
		b.writeIndent(lvl + 1)
		_, _ = b.WriteString(typ.Field(i).Name)

		b.writeColon()

		b.writeInterfaceType(typ.Field(i).Type.String())

		f := clone.Field(i)
		/* #nosec G103 */
		f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()

		b.dumpValue(f.Interface(), lvl+1, false)

		b.writeComma()
		if i != l-1 {
			b.writeNewLine()
		}
	}

	b.writeNewLine()
	b.writeIndent(lvl)
	b.writeBracket('}')
}

func (b *buffer) dumpElem(v interface{}, lvl int, isInterface bool, newLine *bool) {
	if b.writeNil(v == nil) {
		return
	}

	typStr, val, kind := normalize(v)

	if kind == reflect.Slice || kind == reflect.Map || kind == reflect.Struct {
		b.writeNewLine()
		b.writeIndent(lvl)
		*newLine = true
	}

	if isInterface {
		b.writeInterfaceType(typStr)
	}

	if b.writeNil(!val.IsValid()) {
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
		b.writeString(val.String())
	case reflect.Chan:
		if b.writeNil(val.IsNil()) {
			return
		}
		b.writeInterface(val.Interface())
		b.writeLenAndCap(val.Len(), val.Cap())
	case reflect.Func:
		if b.writeNil(val.IsNil()) {
			return
		}
		b.writeInterface(val.Interface())
	case reflect.Array:
		b.dumpArrayOrSlice(val, lvl)
	case reflect.Slice:
		if b.writeNil(val.IsNil()) {
			return
		}
		b.dumpArrayOrSlice(val, lvl)
	case reflect.Map:
		if !b.writeNil(val.IsNil()) {
			b.dumpMap(val, lvl)
		}
	case reflect.Struct:
		b.dumpStruct(val, lvl)
	default:
		b.writeInterface(val.Interface())
	}
}

func (b *buffer) dumpMapKey(key interface{}, lvl int, isInterface bool) {
	if b.writeNil(key == nil) {
		return
	}

	typStr, val, kind := normalize(key)

	if isInterface {
		b.writeInterfaceType(typStr)
	}

	if b.writeNil(!val.IsValid()) {
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
		b.writeString(val.String())
	case reflect.Chan:
		if b.writeNil(val.IsNil()) {
			return
		}
		b.writeInterface(val.Interface())
		b.writeLenAndCap(val.Len(), val.Cap())
	case reflect.Func:
		if b.writeNil(val.IsNil()) {
			return
		}
		b.writeInterface(val.Interface())
	case reflect.Array:
		b.dumpArrayOrSlice(val, lvl)
	case reflect.Struct:
		b.dumpStruct(val, lvl)
	default:
		b.writeInterface(val.Interface())
	}
}

func (b *buffer) dumpValue(v interface{}, lvl int, isInterface bool) {
	if b.writeNil(v == nil) {
		return
	}

	typStr, val, kind := normalize(v)

	if isInterface {
		b.writeInterfaceType(typStr)
	}

	if b.writeNil(!val.IsValid()) {
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
		b.writeString(val.String())
	case reflect.Chan:
		if b.writeNil(val.IsNil()) {
			return
		}
		b.writeInterface(val.Interface())
		b.writeLenAndCap(val.Len(), val.Cap())
	case reflect.Array:
		b.dumpArrayOrSlice(val, lvl)
	case reflect.Slice:
		if b.writeNil(val.IsNil()) {
			return
		}
		b.dumpArrayOrSlice(val, lvl)
	case reflect.Map:
		if b.writeNil(val.IsNil()) {
			return
		}
		b.dumpMap(val, lvl)
	case reflect.Struct:
		b.dumpStruct(val, lvl)
	default:
		b.writeInterface(val.Interface())
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

func (b *buffer) writeNil(condition bool) bool {
	if condition {
		_, _ = b.WriteString(strNil)
	}
	return condition
}

func (b *buffer) writeNewLine() {
	_ = b.WriteByte('\n')
}

func (b *buffer) writeInterfaceType(s string) {
	_ = b.WriteByte('(')
	_, _ = b.WriteString(s)
	_, _ = b.WriteString(") ")
}

func (b *buffer) writeLen(l int) {
	_, _ = b.WriteString("(len=")
	b.B = strconv.AppendInt(b.B, int64(l), 10)
	_ = b.WriteByte(')')
}

func (b *buffer) writeLenAndCap(l, c int) {
	_, _ = b.WriteString("(len=")
	b.B = strconv.AppendInt(b.B, int64(l), 10)
	_, _ = b.WriteString(", cap=")
	b.B = strconv.AppendInt(b.B, int64(c), 10)
	_ = b.WriteByte(')')
}

func (b *buffer) writeString(s string) {
	_ = b.WriteByte('"')
	_, _ = b.WriteString(s)
	_ = b.WriteByte('"')
}

func (b *buffer) writeIndent(lvl int) {
	_, _ = b.WriteString(strings.Repeat(b.indent, lvl))
}

func (b *buffer) writeColon() {
	_, _ = b.WriteString(" : ")
}

func (b *buffer) writeComma() {
	_ = b.WriteByte(',')
}

func (b *buffer) writeSpace() {
	_ = b.WriteByte(' ')
}

func (b *buffer) writeBracket(c byte) {
	_ = b.WriteByte(c)
}

func (b *buffer) writeInterface(v interface{}) {
	_, _ = b.WriteString(fmt.Sprintf("%v", v))
}

func (b *buffer) writeEllipsis(lvl int, s string) bool {
	if lvl > b.maxDepth {
		_, _ = b.WriteString(s)
		return true
	}
	return false
}
