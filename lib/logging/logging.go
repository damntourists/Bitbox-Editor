package logging

import (
	"bitbox-editor/lib/events"
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"github.com/maniartech/signals"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var LogEvent signals.Signal[events.LogRecord]

type LogSink struct {
	*bytes.Buffer
}

func (s LogSink) Write(p []byte) (n int, err error) {
	lr := events.LogRecord{}
	json.Unmarshal(p, &lr)
	ctx := context.Background()
	LogEvent.Emit(ctx, lr)
	return s.Buffer.Write(p)
}
func (s LogSink) Sync() error  { return nil }
func (s LogSink) Close() error { return nil }

// Unmarshal returns decoded data as key value and reset the buffer.
func (s LogSink) Unmarshal() map[string]string {
	defer s.Reset()
	v := make(map[string]string)
	json.Unmarshal(s.Bytes(), &v)
	return v
}

var RootLogger *zap.Logger
var Logger *zap.Logger

var cfg = zap.Config{
	Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
	Development: true,
	Encoding:    "json",
	EncoderConfig: zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "caller",
		FunctionKey:    "func",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	},
	OutputPaths:      []string{"stdout", "log://"},
	ErrorOutputPaths: []string{"stderr", "log://"},
}

func init() {
	es := &LogSink{Buffer: new(bytes.Buffer)}
	err := zap.RegisterSink("log", func(u *url.URL) (zap.Sink, error) {
		return es, nil
	})
	if err != nil {
		panic(err)
	}

	RootLogger, _ = cfg.Build()

	LogEvent = signals.New[events.LogRecord]()
	Logger = NewLogger("lib")
}

func NewLogger(name string) *zap.Logger {
	l := RootLogger.Named(name)
	return l
}
