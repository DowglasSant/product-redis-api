package port

// Logger define a interface de logging para a camada de aplicação.
// Isso permite desacoplar os usecases de implementações específicas de logger.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}
