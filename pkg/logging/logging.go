package logging

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
	Sync() error
}

type Field struct {
	Key   string
	Value interface{}
}

type ZapLogger struct {
	Logger *zap.Logger
}

type LoggerConfig struct {
	FilePath          string
	Level             string
	EnableFileLogging bool
}

func NewLogger(config *LoggerConfig) (Logger, error) {
	level, err := ParseLevel(config.Level)
	if err != nil {
		return nil, err
	}

	consoleEncoder := newConsoleEncoder()
	cores := []zapcore.Core{
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	}

	if config.EnableFileLogging && config.FilePath != "" {
		if err := os.MkdirAll(filepath.Dir(config.FilePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %v", err)
		}

		logFile, err := os.OpenFile(path.Clean(config.FilePath), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}

		fileEncoder := newFileEncoder()
		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), level))
	}

	opts := []zap.Option{
		zap.AddCallerSkip(1),
	}

	logger := zap.New(
		zapcore.NewTee(cores...),
		opts...,
	)

	return &ZapLogger{
		Logger: logger,
	}, nil
}

func ParseLevel(level string) (zapcore.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

func newConsoleEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func newFileEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	return zapcore.NewJSONEncoder(encoderConfig)
}

func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.log(zapcore.DebugLevel, msg, fields...)
}

func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.log(zapcore.InfoLevel, msg, fields...)
}

func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.log(zapcore.WarnLevel, msg, fields...)
}

func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.log(zapcore.ErrorLevel, msg, fields...)
}

func (l *ZapLogger) log(level zapcore.Level, msg string, fields ...Field) {
	if ce := l.Logger.Check(level, msg); ce != nil {
		ce.Write(ConvertFields(fields)...)
	}
}

func (l *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{
		Logger: l.Logger.With(ConvertFields(fields)...),
	}
}

func (l *ZapLogger) Sync() error {
	return l.Logger.Sync()
}

func ConvertFields(fields []Field) []zapcore.Field {
	zapFields := make([]zapcore.Field, 0, len(fields))
	for _, f := range fields {
		if f.Value != nil {
			zapFields = append(zapFields, zap.Any(f.Key, f.Value))
		}
	}
	return zapFields
}

func String(key, val string) Field {
	return Field{Key: key, Value: val}
}

func Error(err error) Field {
	if err == nil {
		return Field{}
	}
	return Field{Key: "error", Value: err.Error()}
}

func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val}
}

func Time(key string, val time.Time) Field {
	return Field{Key: key, Value: val}
}

func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

func Any(key string, val interface{}) Field {
	return Field{Key: key, Value: val}
}
