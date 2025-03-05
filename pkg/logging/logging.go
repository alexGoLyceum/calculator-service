package logging

import (
	"os"
	"path"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	Level string
	Path  string
}

func NewLogger(cfg *LogConfig) (*zap.Logger, error) {
	logFilePath := path.Clean(cfg.Path)
	var logFile *os.File
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		logFile, err = os.Create(logFilePath)
		if err != nil {
			return nil, err
		}
	} else {
		logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return nil, err
		}
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	var atomicLevel zap.AtomicLevel
	if strings.ToLower(cfg.Level) == "production" {
		atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else {
		atomicLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), atomicLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), atomicLevel),
	)
	logger := zap.New(core)
	return logger, nil
}
