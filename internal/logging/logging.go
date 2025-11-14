package logging

import (
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogChannel is the log queue for the UI to consume
var LogChannel = make(chan LogRecord, 2000)

type LogRecord struct {
	Timestamp  string                 `json:"ts"`
	Name       string                 `json:"name"`
	Level      string                 `json:"level"`
	Message    string                 `json:"msg"`
	Caller     string                 `json:"caller"`
	Func       string                 `json:"func"`
	Stacktrace string                 `json:"stacktrace"`
	Params     map[string]interface{} `json:"-"`
}

type uiLogWriter struct{}

func (s *uiLogWriter) Write(p []byte) (n int, err error) {
	var lr LogRecord

	if err := json.Unmarshal(p, &lr); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal log: %v\n", err)
		return len(p), nil
	}

	var allFields map[string]interface{}
	if err := json.Unmarshal(p, &allFields); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal log fields: %v\n", err)
		return len(p), nil
	}

	delete(allFields, "ts")
	delete(allFields, "name")
	delete(allFields, "level")
	delete(allFields, "msg")
	delete(allFields, "caller")
	delete(allFields, "func")
	delete(allFields, "stacktrace")

	if len(allFields) > 0 {
		lr.Params = allFields
	}

	select {
	case LogChannel <- lr:
	default:
		println("Log channel is full!")
	}

	return len(p), nil
}

func (s *uiLogWriter) Sync() error {
	return nil
}

var RootLogger *zap.Logger

func init() {
	// JSON encoder for UI use
	jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
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
	})

	// Console encoder for CLI use
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
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
	})

	// Configure logging to write to both outputs
	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(&uiLogWriter{}), zapcore.DebugLevel),
		zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel),
	)

	RootLogger = zap.New(core, zap.AddCaller())
}

func NewLogger(name string) *zap.Logger {
	return RootLogger.Named(name)
}
