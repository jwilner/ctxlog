package ctxlog

import (
	"fmt"
	"testing"
	"time"
)

func Benchmark_printPair(b *testing.B) {
	type named struct {
		name string
		val  interface{}
	}
	for _, v := range []interface{}{
		named{"empty string", ""},
		named{"ten byte string", "abcdefghij"},
		named{"quoted string", "abcde\"fghij"},
		true,
		uint(0),
		1,
		1.,
		time.Time{},
	} {
		name := fmt.Sprintf("%T", v)
		if n, ok := v.(named); ok {
			name = n.name
			v = n.val
		}
		buf := make([]byte, 100)
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				printPair(buf, "key", v)
			}
		})
	}
}
