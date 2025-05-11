package logging_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexGoLyceum/calculator-service/pkg/logging"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLogger(t *testing.T) {
	t.Run("successful creation with file logging", func(t *testing.T) {
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "test.log")

		config := &logging.LoggerConfig{
			FilePath:          logPath,
			Level:             "info",
			EnableFileLogging: true,
		}

		logger, err := logging.NewLogger(config)
		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})

	t.Run("successful creation without file logging", func(t *testing.T) {
		config := &logging.LoggerConfig{
			Level:             "info",
			EnableFileLogging: false,
		}

		logger, err := logging.NewLogger(config)
		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})

	t.Run("invalid level", func(t *testing.T) {
		config := &logging.LoggerConfig{
			FilePath:          "test.log",
			Level:             "invalid",
			EnableFileLogging: true,
		}

		logger, err := logging.NewLogger(config)
		assert.Error(t, err)
		assert.Nil(t, logger)
		assert.Contains(t, err.Error(), "unknown log level")
	})

	t.Run("failed to create directory when file logging enabled", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping test when running as root")
		}

		config := &logging.LoggerConfig{
			FilePath:          "/nonexistent/path/test.log",
			Level:             "info",
			EnableFileLogging: true,
		}

		logger, err := logging.NewLogger(config)
		assert.Error(t, err)
		assert.Nil(t, logger)
		assert.Contains(t, err.Error(), "failed to create log directory")
	})

	t.Run("no error when file logging disabled with invalid path", func(t *testing.T) {
		config := &logging.LoggerConfig{
			FilePath:          "/nonexistent/path/test.log",
			Level:             "info",
			EnableFileLogging: false,
		}

		logger, err := logging.NewLogger(config)
		assert.NoError(t, err)
		assert.NotNil(t, logger)
	})

	t.Run("failed to open log file", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping test when running as root")
		}

		tmpDir := t.TempDir()
		unwritable := filepath.Join(tmpDir, "unwritable")
		os.Mkdir(unwritable, 0444)

		config := &logging.LoggerConfig{
			FilePath:          filepath.Join(unwritable, "test.log"),
			Level:             "info",
			EnableFileLogging: true,
		}

		logger, err := logging.NewLogger(config)
		assert.Error(t, err)
		assert.Nil(t, logger)
		assert.Contains(t, err.Error(), "failed to open log file")
	})
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected zapcore.Level
		hasError bool
	}{
		{"debug level", "debug", zapcore.DebugLevel, false},
		{"info level", "info", zapcore.InfoLevel, false},
		{"warn level", "warn", zapcore.WarnLevel, false},
		{"error level", "error", zapcore.ErrorLevel, false},
		{"case insensitive", "INFO", zapcore.InfoLevel, false},
		{"unknown level", "unknown", zapcore.InfoLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, err := logging.ParseLevel(tt.level)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestFieldHelpers(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() logging.Field
		expected logging.Field
	}{
		{"String", func() logging.Field { return logging.String("key", "value") }, logging.Field{Key: "key", Value: "value"}},
		{"Error", func() logging.Field { return logging.Error(errors.New("test")) }, logging.Field{Key: "error", Value: "test"}},
		{"Int", func() logging.Field { return logging.Int("key", 42) }, logging.Field{Key: "key", Value: 42}},
		{"Duration", func() logging.Field { return logging.Duration("key", time.Second) }, logging.Field{Key: "key", Value: time.Second}},
		{"Time", func() logging.Field { return logging.Time("key", time.Time{}) }, logging.Field{Key: "key", Value: time.Time{}}},
		{"Bool", func() logging.Field { return logging.Bool("key", true) }, logging.Field{Key: "key", Value: true}},
		{"Any", func() logging.Field { return logging.Any("key", "any") }, logging.Field{Key: "key", Value: "any"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := tt.fn()
			assert.Equal(t, tt.expected.Key, field.Key)
			assert.Equal(t, tt.expected.Value, field.Value)
		})
	}

	t.Run("Error with nil", func(t *testing.T) {
		field := logging.Error(nil)
		assert.Equal(t, logging.Field{}, field)
	})
}

func TestConvertFields(t *testing.T) {
	tests := []struct {
		name     string
		fields   []logging.Field
		expected int
	}{
		{"nil value", []logging.Field{{Key: "key", Value: nil}}, 0},
		{"valid fields", []logging.Field{
			{Key: "key1", Value: "value1"},
			{Key: "key2", Value: 42},
		}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logging.ConvertFields(tt.fields)
			assert.Len(t, result, tt.expected)
		})
	}
}

func setupTestLogger(level zapcore.Level) (*logging.ZapLogger, *observer.ObservedLogs) {
	core, logs := observer.New(level)
	logger := zap.New(core)
	return &logging.ZapLogger{Logger: logger}, logs
}

func TestZapLogger_Debug(t *testing.T) {
	l, logs := setupTestLogger(zapcore.DebugLevel)
	l.Debug("debug-msg", logging.String("key", "value"))
	assert.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, "debug-msg", entry.Message)
	assert.Equal(t, zapcore.DebugLevel, entry.Level)
	assert.Equal(t, "value", entry.ContextMap()["key"])
}

func TestZapLogger_Info(t *testing.T) {
	l, logs := setupTestLogger(zapcore.InfoLevel)
	l.Info("info-msg", logging.Int("code", 200))
	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, "info-msg", logs.All()[0].Message)
}

func TestZapLogger_Warn(t *testing.T) {
	l, logs := setupTestLogger(zapcore.WarnLevel)
	l.Warn("warn-msg", logging.Bool("flag", true))
	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, "warn-msg", logs.All()[0].Message)
}

func TestZapLogger_Error(t *testing.T) {
	l, logs := setupTestLogger(zapcore.ErrorLevel)
	l.Error("error-msg", logging.Error(errors.New("something went wrong")))
	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, "error-msg", logs.All()[0].Message)
	assert.Contains(t, logs.All()[0].ContextMap()["error"], "something went wrong")
}

func TestZapLogger_With(t *testing.T) {
	l, logs := setupTestLogger(zapcore.InfoLevel)
	l2 := l.With(logging.String("tag", "child"))
	l2.Info("child-msg")
	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, "child-msg", logs.All()[0].Message)
	assert.Equal(t, "child", logs.All()[0].ContextMap()["tag"])
}

func TestZapLogger_Sync(t *testing.T) {
	l, _ := setupTestLogger(zapcore.InfoLevel)
	err := l.Sync()
	assert.NoError(t, err)
}
