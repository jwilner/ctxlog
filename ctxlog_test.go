package ctxlog_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"reflect"
	"testing"
	"testing/quick"
	"time"

	"github.com/jwilner/ctxlog"
)

func equaler(t *testing.T) func(expected, actual interface{}, formatAndArgs ...interface{}) {
	return func(expected, actual interface{}, formatAndArgs ...interface{}) {
		if !reflect.DeepEqual(expected, actual) {
			if len(formatAndArgs) > 0 {
				t.Fatalf(formatAndArgs[0].(string), formatAndArgs[1:]...)
			}
			t.Fatalf("expected %v to be equal to %v", expected, actual)
		}
	}
}

func TestLogger(t *testing.T) {
	requireEQ := equaler(t)

	var buf bytes.Buffer

	log := ctxlog.New(&buf)
	ctx := context.Background()

	log.Info(ctx, "hi")
	requireEQ(`{"level":"INFO","message":"hi"}`+"\n", buf.String())

	buf.Reset()
	log.Info(ctx, "key", 23, "blah")
	requireEQ(`{"level":"INFO","key":23,"message":"blah"}`+"\n", buf.String())

	buf.Reset()
	log.Info(ctx, "moment", time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC))
	requireEQ(`{"level":"INFO","moment":"2014-01-01T00:00:00Z"}`+"\n", buf.String())

	buf.Reset()
	log.Info(ctx, "float", 23.2)
	requireEQ(`{"level":"INFO","float":23.2}`+"\n", buf.String())

	buf.Reset()
	log.Info(ctx, "stringer", net.IPv4(255, 255, 255, 255))
	requireEQ(`{"level":"INFO","stringer":"255.255.255.255"}`+"\n", buf.String())

	buf.Reset()
	log.Info(ctxlog.Add(ctx, "hi", 1), "🎂\"\n ")
	requireEQ(`{"level":"INFO","hi":1,"message":"🎂\"\n "}`+"\n", buf.String())
}

func BenchmarkAdd(b *testing.B) {
	b.ReportAllocs()

	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		ctxlog.Add(ctx, "abc", "def")
	}
}

func BenchmarkWrite(b *testing.B) {
	log := ctxlog.New(ioutil.Discard)
	ctx := context.Background()

	rand.Seed(time.Now().UnixNano())

	for _, val := range []interface{}{
		func() interface{} {
			b := make([]byte, len(time.RFC3339Nano))
			rand.Read(b)
			for i := range b {
				b[i] = b[i]%(126-32) + 32
			}
			s := string(b)
			return s
		},
		int32(0),
		0,
		uint32(0),
		uint(0),
		float32(0),
		float64(0),
		true,
		func() interface{} { return time.Now() },
		func() interface{} { return sql.ErrNoRows },
		func() interface{} {
			ip := make(net.IP, 16)
			rand.Read(ip)
			return ip
		},
	} {
		var f func() interface{}
		if valF, ok := val.(func() interface{}); ok {
			f = valF
		} else {
			f = func() interface{} {
				val, ok := quick.Value(reflect.TypeOf(val), rand.New(rand.NewSource(time.Now().UnixNano())))
				if !ok {
					panic(fmt.Sprintf("%T is not quickable", val))
				}
				return val.Interface()
			}
		}
		b.Run(fmt.Sprintf("%T", f()), func(b *testing.B) {
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				v := f()
				for pb.Next() {
					log.Info(ctx, "key", v)
				}
			})
		})
	}
}

func BenchmarkWriteTen(b *testing.B) {
	log := ctxlog.New(ioutil.Discard)

	ctx := context.Background()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info(
				ctx,
				"a", 1,
				"b", 2,
				"c", 3,
				"d", 4,
				"e", 5,
				"f", 6,
				"g", 7,
				"h", 8,
				"i", 9,
				"j", 10,
			)
		}
	})
}
