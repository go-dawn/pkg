package rand

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Rand_Int(t *testing.T) {
	for i := 1; i <= 1000000; i++ {
		assert.True(t, Int(i) < i)
	}
}

func Benchmark_Rand_Int(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			Int(1000000)
		}
	})
}

func Test_Rand_Uint32(t *testing.T) {
	for i := 1; i <= 1000000; i++ {
		assert.True(t, Uint32() < math.MaxUint32)
	}
}

func Benchmark_Rand_Uint32(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			Uint32()
		}
	})
}

func Test_Rand_Float64(t *testing.T) {
	assert.True(t, Float64() < 1)
}

func Benchmark_Rand_Float64(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			Float64()
		}
	})
}

func Test_Rand_Float64Range(t *testing.T) {
	for i := 1; i <= 1000000; i++ {
		f := Float64Range(0, 1000000)
		assert.True(t, f >= 0 && f < 1000000)
	}
}

func Benchmark_Rand_Float64Range(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			Float64Range(0, 1000000)
		}
	})
}

func Test_Rand_String(t *testing.T) {
	for i := 1; i <= 1000; i++ {
		assert.Equal(t, i, len(String(i)))
	}
}

func Benchmark_Rand_String_16(b *testing.B) {
	randString(b, 16)
}

func Benchmark_Rand_String_32(b *testing.B) {
	randString(b, 32)
}

func Benchmark_Rand_String_64(b *testing.B) {
	randString(b, 64)
}

func randString(b *testing.B, i int) {
	b.RunParallel(func(pb *testing.PB) {
		b.ReportAllocs()
		b.ResetTimer()
		for pb.Next() {
			String(i)
		}
	})
}
