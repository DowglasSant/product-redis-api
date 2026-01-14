package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(level string, environment string) (*zap.Logger, *zap.AtomicLevel, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, nil, fmt.Errorf("invalid log level: %w", err)
	}

	var config zap.Config
	if environment == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	atomicLevel := zap.NewAtomicLevelAt(zapLevel)
	config.Level = atomicLevel

	logger, err := config.Build(
		zap.AddCallerSkip(0),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, &atomicLevel, nil
}
