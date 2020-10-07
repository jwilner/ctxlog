package ctxlog_test

import (
	"context"
	"os"
	"time"

	"github.com/jwilner/ctxlog"
)

func ExampleLogger() {
	log := ctxlog.New(os.Stdout)

	ctx := ctxlog.Add(context.Background(), "user_id", 23, "foo", "bar")

	log.Debug(ctx, "default level is info")
	log.Info(ctx, "var_val", time.Unix(0, 0).UTC(), "odd number of fields treated as message")
	// Output:
	// {"level":"INFO","user_id":23,"foo":"bar","var_val":"1970-01-01T00:00:00Z","message":"odd number of fields treated as message"}
}

func ExampleOptCaller() {
	log := ctxlog.New(os.Stdout, ctxlog.OptCaller(false))

	log.Error(context.Background())
	// Output:
	// {"level":"ERROR","caller":"example_test.go:25"}
}

func ExampleAdd() {
	log := ctxlog.New(os.Stdout)

	ctx := context.Background()
	ctx = ctxlog.Add(ctx, "key", 23)
	ctx = ctxlog.Add(ctx, "key2", true)

	log.Error(ctx, "got an error")
	// Output:
	// {"level":"ERROR","key":23,"key2":true,"message":"got an error"}
}

func ExampleOptInfo() {
	log := ctxlog.New(os.Stdout, ctxlog.OptInfo)

	ctx := context.Background()

	log.Debug(ctx, "hi") // debug is ignored
	log.Info(ctx, "cool")
	// Output:
	// {"level":"INFO","message":"cool"}
}

func ExampleSetOutput() {
	ctxlog.SetOutput(os.Stdout)

	ctx := ctxlog.Add(context.Background(), "foo", "bar")

	ctxlog.Info(ctx, "noice")
	// Output:
	// {"level":"INFO","foo":"bar","message":"noice"}
}
