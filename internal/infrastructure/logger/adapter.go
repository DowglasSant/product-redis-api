package logger

import (
	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"go.uber.org/zap"
)

type ZapAdapter struct {
	logger *zap.Logger
}

func NewZapAdapter(logger *zap.Logger) port.Logger {
	return &ZapAdapter{logger: logger}
}

func (z *ZapAdapter) Debug(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Debugw(msg, keysAndValues...)
}

func (z *ZapAdapter) Info(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Infow(msg, keysAndValues...)
}

func (z *ZapAdapter) Warn(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Warnw(msg, keysAndValues...)
}

func (z *ZapAdapter) Error(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Errorw(msg, keysAndValues...)
}
