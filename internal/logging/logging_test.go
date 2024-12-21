package logging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexGoLyceum/calculator-service/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger_FileCreation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-logger-*.log")
	assert.NoError(t, err, "Failed to create temporary file")
	defer os.Remove(tmpFile.Name())

	tmpFile.Close()

	cfg := config.LogConfig{
		Level: "debug",
		Path:  tmpFile.Name(),
	}

	logger, err := NewLogger(&cfg)
	assert.NoError(t, err, "Failed to create logger")
	assert.NotNil(t, logger, "Logger should not be nil")
	defer logger.Sync()

	_, err = os.Stat(tmpFile.Name())
	assert.NoError(t, err, "Log file should exist after logger creation")
}

func TestNewLogger_InvalidFilePath(t *testing.T) {
	invalidPath := filepath.Join("/nonexistent", "invalid.log")

	cfg := config.LogConfig{
		Level: "debug",
		Path:  invalidPath,
	}

	logger, err := NewLogger(&cfg)
	assert.Error(t, err, "Expected error for invalid file path")
	assert.Nil(t, logger, "Logger should not be nil")
}

func TestNewLogger_LogLevelProduction(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-logger-*.log")
	assert.NoError(t, err, "Failed to create temporary file")
	defer os.Remove(tmpFile.Name())

	tmpFile.Close()

	cfg := config.LogConfig{
		Level: "production",
		Path:  tmpFile.Name(),
	}

	logger, err := NewLogger(&cfg)
	assert.NoError(t, err, "Failed to create logger")
	assert.NotNil(t, logger, "Logger should not be nil")
	defer logger.Sync()

	fileInfo, err := os.Stat(tmpFile.Name())
	assert.NoError(t, err, "Failed to stat log file")
	assert.Equal(t, int64(0), fileInfo.Size(), "Log file should be empty")
}

func TestNewLogger_DebugLevel(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-logger-*.log")
	assert.NoError(t, err, "Failed to create temporary file")
	defer os.Remove(tmpFile.Name())

	tmpFile.Close()

	cfg := config.LogConfig{
		Level: "debug",
		Path:  tmpFile.Name(),
	}

	logger, err := NewLogger(&cfg)
	assert.NoError(t, err, "Failed to create logger")
	assert.NotNil(t, logger, "Logger should not be nil")
	defer logger.Sync()

	logger.Debug("Debug message")

	fileInfo, err := os.Stat(tmpFile.Name())
	assert.NoError(t, err, "Failed to stat log file")
	assert.Greater(t, fileInfo.Size(), int64(0), "Log file should contain logs for debug level")
}

func TestNewLogger_OpenFileError(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-logger-*.log")
	assert.NoError(t, err, "Failed to create temporary file")
	defer os.Remove(tmpFile.Name())

	tmpFile.Close()
	err = os.Chmod(tmpFile.Name(), 0400)
	assert.NoError(t, err, "Failed to change file permissions")

	cfg := config.LogConfig{
		Level: "debug",
		Path:  tmpFile.Name(),
	}

	logger, err := NewLogger(&cfg)
	assert.Error(t, err, "Expected error for invalid file path")
	assert.Nil(t, logger, "Logger should not be nil")
}
