package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"unsafe"

	"github.com/valyala/fastrand"
)

// Int returns a random value in [0, max).
func Int(max int) int {
	return int(fastrand.Uint32n(uint32(max)))
}

// Uint32 returns a random value in [0, 2^32].
func Uint32() uint32 {
	return fastrand.Uint32()
}

// Float64 returns a random float64 value in [0, 1).
func Float64() float64 {
again:
	f := float64(fastrand.Uint32()) / (1 << 32)
	if f == 1 {
		goto again
	}
	return f
}

// Float64Range returns a random float64 value in [min, max).
func Float64Range(min, max float64) float64 {
	return Float64()*(max-min) + min
}

var encoder = base64.URLEncoding.WithPadding(base64.NoPadding)

// String returns a random string value with given len.
func String(len int) string {
	return b2s(Bytes(len))
}

// Bytes returns a random byte slice value (visible characters) with given len.
func Bytes(len int) []byte {
	l1 := (len*6-5)/8 + 1
	l2 := encoder.EncodedLen(l1)
	b := make([]byte, l1+l2)
	_, err := rand.Read(b[:l1])
	if err != nil {
		panic(fmt.Sprintf("rand: failed to generate random string: %s", err))
	}
	encoder.Encode(b[l1:], b[:l1])

	return b[l1 : l1+len]
}

// b2s converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func b2s(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

// s2b converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
//func s2b(s string) (b []byte) {
//	/* #nosec G103 */
//	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
//	/* #nosec G103 */
//	sh := *(*reflect.StringHeader)(unsafe.Pointer(&s))
//	bh.Data = sh.Data
//	bh.Len = sh.Len
//	bh.Cap = sh.Len
//	return b
//}
