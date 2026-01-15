package logger

import (
	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"go.uber.org/zap"
)

// ZapAdapter adapta *zap.Logger para a interface port.Logger.
// Isso permite que os usecases usem uma abstração de logger.
type ZapAdapter struct {
	logger *zap.Logger
}

// NewZapAdapter cria um novo adapter do zap para a interface Logger.
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
