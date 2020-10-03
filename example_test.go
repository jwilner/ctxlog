package ctxlog_test

import (
	"context"
	"github.com/jwilner/ctxlog"
	"os"
)

func ExampleOptCaller() {
	log := ctxlog.New(os.Stdout, ctxlog.OptCaller(false))

	log.Error(context.Background())

	// Output:
	// {"level":"ERROR","caller":"example_test.go:12"}
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
