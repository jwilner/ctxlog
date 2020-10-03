package ctxlog

import (
	"context"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"
)

// New returns a new Logger that writes to the provided writer and the output of which is controlled by the provided
// options.
func New(w io.Writer, opts ...Option) *Logger {
	l := Logger{w: w, bufs: makeBufs(10)}

	l.filter = levelInfo

	for _, o := range opts {
		o(&l.settings)
	}
	return &l
}

func makeBufs(l int) chan []byte {
	ch := make(chan []byte, l)
	for i := 0; i < l; i++ {
		ch <- make([]byte, 10)
	}
	return ch
}

type ctxKey uint

const (
	_ ctxKey = iota
	prefixKey
)

// Add adds contextual information to logs derived from this context.
func Add(ctx context.Context, keyVals ...interface{}) context.Context {
	if len(keyVals) == 0 {
		return ctx
	}

	// super dumb heuristic to reduce allocs
	// guess that every key / val will take eight bytes
	const termSize = 8

	lengthGuess := termSize * len(keyVals)

	var buf []byte
	if prefix, ok := ctx.Value(prefixKey).([]byte); ok {
		buf = make([]byte, len(prefix), len(prefix)+lengthGuess)
		copy(buf, prefix)
	} else {
		buf = make([]byte, 0, lengthGuess)
	}

	trimmed := len(keyVals)
	if trimmed%2 == 1 {
		trimmed--
	}

	for i := 0; i < trimmed; i += 2 {
		buf = append(buf, ',') // we want a leading comma
		buf = printPair(buf, keyVals[i], keyVals[i+1])
	}

	if trimmed < len(keyVals) {
		buf = append(buf, ',')
		buf = printPair(buf, "message", keyVals[trimmed])
	}

	return context.WithValue(ctx, prefixKey, buf)
}

// Logger writes structured logs to its underlying writer.
type Logger struct {
	bufs chan []byte

	mu sync.Mutex
	w  io.Writer

	settings
}

var (
	std     = New(os.Stderr)
	stdLock sync.RWMutex
)

// SetOptions sets options for the global logger.
func SetOptions(opts ...Option) {
	stdLock.Lock()
	defer stdLock.Unlock()

	for _, o := range opts {
		o(&std.settings)
	}
}

// SetOutput sets an alternative writer for the global logger.
func SetOutput(w io.Writer) {
	stdLock.Lock()
	defer stdLock.Unlock()

	std.w = w
}

// Debug logs a message on the global logger at the debug level.
func Debug(ctx context.Context, keyVal ...interface{}) {
	output(ctx, levelDebug, keyVal)
}

// Debug logs a message at the debug level.
func (l *Logger) Debug(ctx context.Context, keyVal ...interface{}) {
	l.output(ctx, levelDebug, 1, keyVal)
}

// Info logs a message on the global logger at the info level.
func Info(ctx context.Context, keyVal ...interface{}) {
	output(ctx, levelInfo, keyVal)
}

// Info logs a message at the info level.
func (l *Logger) Info(ctx context.Context, keyVal ...interface{}) {
	l.output(ctx, levelInfo, 1, keyVal)
}

// Warn logs a message on the global logger at the warn level.
func Warn(ctx context.Context, keyVal ...interface{}) {
	output(ctx, levelWarn, keyVal)
}

// Warn logs a message at the warn level.
func (l *Logger) Warn(ctx context.Context, keyVal ...interface{}) {
	l.output(ctx, levelWarn, 1, keyVal)
}

// Error logs a message on the global logger at the error level.
func Error(ctx context.Context, keyVal ...interface{}) {
	output(ctx, levelError, keyVal)
}

// Error logs a message at the error level.
func (l *Logger) Error(ctx context.Context, keyVal ...interface{}) {
	l.output(ctx, levelError, 1, keyVal)
}

func output(ctx context.Context, lvl int, keyVal []interface{}) {
	stdLock.RLock()
	defer stdLock.RUnlock()
	std.output(ctx, lvl, 2, keyVal)
}

func (l *Logger) output(ctx context.Context, lvl int, callDepth int, keyVal []interface{}) {
	if l.filter > lvl {
		return
	}

	buf := <-l.bufs
	defer func() { l.bufs <- buf }()

	buf = append(buf[:0], '{')
	switch lvl {
	case levelDebug:
		buf = append(buf, `"level":"DEBUG"`...)
	case levelInfo:
		buf = append(buf, `"level":"INFO"`...)
	case levelWarn:
		buf = append(buf, `"level":"WARN"`...)
	case levelError:
		buf = append(buf, `"level":"ERROR"`...)
	}
	if l.flags&flagTime != 0 {
		buf = append(buf, `,"timestamp":"`...)
		buf = time.Now().UTC().AppendFormat(buf, time.RFC3339Nano)
		buf = append(buf, '"')
	}
	if l.flags&(flagCallerShort|flagCallerLong) != 0 {
		_, file, line, _ := runtime.Caller(callDepth + 1)
		if l.flags&flagCallerShort != 0 {
			for end, width := len(file), 0; end > 0; end -= width {
				var r rune
				r, width = utf8.DecodeLastRuneInString(file[:end])
				if r == os.PathSeparator { // exclude it
					file = file[end:]
					break
				}
				if width == 0 {
					break
				}
			}
		}
		buf = append(buf, `,"caller":`...)
		buf = simpleQuote(buf, file+":"+strconv.Itoa(line))
	}
	if prefix, ok := ctx.Value(prefixKey).([]byte); ok {
		buf = append(buf, prefix...)
	}
	trimmed := len(keyVal)
	if trimmed%2 == 1 {
		trimmed--
	}
	for i := 0; i < trimmed; i += 2 {
		buf = append(buf, ',')
		buf = printPair(buf, keyVal[i], keyVal[i+1])
	}
	if trimmed != len(keyVal) {
		buf = append(buf, ',')
		buf = printPair(buf, "message", keyVal[trimmed])
	}
	buf = append(buf, '}', '\n')

	l.mu.Lock()
	defer l.mu.Unlock()

	_, _ = l.w.Write(buf)
}

func byteEncTable() (table [utf8.RuneSelf][]byte) {
	const hex = "0123456789abcdef"
	for b := byte(0); b < utf8.RuneSelf; b++ {
		switch b {
		case '\\':
			table[b] = []byte(`\\`)
		case '"':
			table[b] = []byte(`\"`)
		case '\a':
			table[b] = []byte(`\a`)
		case '\b':
			table[b] = []byte(`\b`)
		case '\f':
			table[b] = []byte(`\f`)
		case '\n':
			table[b] = []byte(`\n`)
		case '\r':
			table[b] = []byte(`\r`)
		case '\t':
			table[b] = []byte(`\t`)
		case '\v':
			table[b] = []byte(`\v`)
		default:
			if b < ' ' {
				table[b] = []byte{'\\', 'x', hex[b>>4], hex[b&0b1111]}
			}
		}
	}
	return
}

var quoteTable = byteEncTable()

func simpleQuote(buf []byte, s string) []byte {
	buf = append(buf, '"')

	for i, width := 0, 1; i < len(s); i += width {
		if b := s[i]; b < utf8.RuneSelf {
			width = 1

			if quoted := quoteTable[b]; quoted != nil {
				buf = append(buf, quoted...)
				continue
			}

			buf = append(buf, b)
			continue
		}

		var r rune
		r, width = utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError {
			width = 1

			const hex = "0123456789abcdef"
			buf = append(buf, '\\', 'x', hex[s[i]>>4], hex[s[i]&0b1111])

			continue
		}

		buf = append(buf, s[i:i+width]...)
	}

	return append(buf, '"')
}

func printPair(buf []byte, k, v interface{}) []byte {
	if k, ok := k.(string); ok {
		buf = simpleQuote(buf, k)
		buf = append(buf, ':')
	} else {
		buf = append(buf, "\"INVALID-KEY\":"...)
	}

	switch v := v.(type) {

	case string:
		buf = simpleQuote(buf, v)

	case bool:
		buf = strconv.AppendBool(buf, v)

	case int8:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int16:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int32:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int:
		buf = strconv.AppendInt(buf, int64(v), 10)
	case int64:
		buf = strconv.AppendInt(buf, v, 10)

	case uint8:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint16:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint32:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint:
		buf = strconv.AppendUint(buf, uint64(v), 10)
	case uint64:
		buf = strconv.AppendUint(buf, v, 10)

	case float32:
		buf = strconv.AppendFloat(buf, float64(v), 'f', -1, 32)
	case float64:
		buf = strconv.AppendFloat(buf, v, 'f', -1, 64)

	case time.Time:
		buf = v.AppendFormat(buf, "\""+time.RFC3339Nano+"\"")

	case error:
		buf = simpleQuote(buf, v.Error())

	case interface{ String() string }:
		buf = simpleQuote(buf, v.String())

	default:
		buf = append(buf, "\"INVALID-VALUE\""...)
	}
	return buf
}

const (
	_ = iota
	levelDebug
	levelInfo
	levelWarn
	levelError
)

const (
	flagTime uint = 1 << iota
	flagCallerShort
	flagCallerLong
)

type settings struct {
	filter int
	flags  uint
}
