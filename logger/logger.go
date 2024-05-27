package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

type Args struct {
	Level  Level
	Msg    string
	Time   time.Time
	Buffer buffer
	PC     uintptr
}

type IPart interface {
	Output(args *Args)
}

type Logger struct {
	parts  []IPart
	output io.Writer
	pool   sync.Pool
	Level  Level
}

type OptionFunc func(l *Logger)

func OptOutput(o io.Writer) OptionFunc {
	return func(l *Logger) {
		l.output = o
	}
}

func OptLevel(lv Level) OptionFunc {
	return func(l *Logger) {
		l.Level = lv
	}
}

func OptPart(parts ...IPart) OptionFunc {
	return func(l *Logger) {
		l.parts = parts
	}
}

func NewLogger(options ...OptionFunc) *Logger {
	l := &Logger{}
	l.parts = []IPart{}
	l.output = os.Stdout

	for _, o := range options {
		o(l)
	}

	l.pool.New = func() any {
		return &Args{}
	}

	return l
}

const badKey = "!BADKEY"

func argsToAttr(args []any) (IPart, []any) {
	switch x := args[0].(type) {
	default:
		if len(args) == 1 {
			return PartKV(badKey, x), nil
		}
		return PartKV(x, args[1]), args[2:]

	case IPart:
		return x, args[1:]
	}
}

func argsToAttrSlice(args []any) []IPart {
	var attr IPart
	var attrs []IPart

	for len(args) > 0 {
		attr, args = argsToAttr(args)
		attrs = append(attrs, attr)
	}

	return attrs
}

func (l *Logger) With(args ...any) ILogger {
	c := &Logger{}
	c.output = l.output
	c.pool.New = l.pool.New
	c.Level = l.Level

	for _, p := range l.parts {

		c.parts = append(c.parts, p)

		if _, ok := p.(*partLevel); ok {
			parts := argsToAttrSlice(args)
			c.parts = append(c.parts, parts...)
		}
	}

	return c
}

func (l *Logger) Write(p []byte) (n int, err error) {
	return l.output.Write(p)
}

func (l *Logger) Log(level Level, format string, msgs []any) {

	if l.Level >= level {
		return
	}

	var msg string
	if format != "" {
		msg = fmt.Sprintf(format, msgs...)
	} else {
		msg = fmt.Sprintln(msgs...)
	}

	args := l.pool.Get().(*Args)
	args.Level = level
	args.Msg = msg
	args.Time = time.Now()
	args.Buffer.Reset()

	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(3, pcs[:])
	args.PC = pcs[0]

	for i, part := range l.parts {
		part.Output(args)
		if i != len(l.parts)-1 {
			args.Buffer.Write(' ')
		}
	}

	bf := args.Buffer.buf
	if len(bf) == 0 || bf[len(bf)-1] != '\n' {
		args.Buffer.Write('\n')
	}

	l.output.Write(args.Buffer.buf)

	l.pool.Put(args)
}

func (l *Logger) Debug(i ...any)                 { l.Log(LevelDebug, "", i) }
func (l *Logger) Debugf(format string, i ...any) { l.Log(LevelDebug, format, i) }
func (l *Logger) Info(i ...any)                  { l.Log(LevelInfo, "", i) }
func (l *Logger) Infof(format string, i ...any)  { l.Log(LevelInfo, format, i) }
func (l *Logger) Warn(i ...any)                  { l.Log(LevelWarn, "", i) }
func (l *Logger) Warnf(format string, i ...any)  { l.Log(LevelWarn, format, i) }
func (l *Logger) Err(i ...any)                   { l.Log(LevelError, "", i) }
func (l *Logger) Errf(format string, i ...any)   { l.Log(LevelError, format, i) }

type ILogger interface {
	Debug(i ...any)
	Debugf(format string, i ...any)
	Info(i ...any)
	Infof(format string, i ...any)
	Warn(i ...any)
	Warnf(format string, i ...any)
	Err(i ...any)
	Errf(format string, i ...any)
	With(parts ...any) ILogger
}
