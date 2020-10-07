[![Tests](https://github.com/jwilner/ctxlog/workflows/tests/badge.svg)](https://github.com/jwilner/ctxlog/actions?query=workflow%3Atests+branch%3Amain)
[![Lint](https://github.com/jwilner/ctxlog/workflows/lint/badge.svg)](https://github.com/jwilner/ctxlog/actions?query=workflow%3Alint+branch%3Amain)
[![GoDoc](https://godoc.org/github.com/jwilner/ctxlog?status.svg)](https://godoc.org/github.com/jwilner/ctxlog)

# ctxlog

`ctxlog` provides super simple logging.

<!-- goquote .#ExampleLogger -->
**Code**:
```go
log := ctxlog.New(os.Stdout)

ctx := ctxlog.Add(context.Background(), "user_id", 23, "foo", "bar")

log.Debug(ctx, "default level is info")
log.Info(ctx, "var_val", time.Unix(0, 0).UTC(), "odd number of fields treated as message")
```
**Output**:
```
{"level":"INFO","user_id":23,"foo":"bar","var_val":"1970-01-01T00:00:00Z","message":"odd number of fields treated as message"}
```
<!-- /goquote -->

A global logger is available:

<!-- goquote .#ExampleSetOutput -->
**Code**:
```go
ctxlog.SetOutput(os.Stdout)

ctx := ctxlog.Add(context.Background(), "foo", "bar")

ctxlog.Info(ctx, "noice")
```
**Output**:
```
{"level":"INFO","foo":"bar","message":"noice"}
```
<!-- /goquote -->

Along with all the usual perks:

<!-- goquote .#ExampleOptCaller -->
**Code**:
```go
log := ctxlog.New(os.Stdout, ctxlog.OptCaller(false))

log.Error(context.Background())
```
**Output**:
```
{"level":"ERROR","caller":"example_test.go:12"}
```
<!-- /goquote -->
