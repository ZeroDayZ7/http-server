package logger

import (
	"go.uber.org/zap"
)

// zapLogger to prywatna implementacja interfejsu Logger
type zapLogger struct {
	*zap.Logger
}

// NewNop zwraca loggera, który nic nie robi (przydatne w bardzo szybkich testach jednostkowych)
func NewNop() Logger {
	return &zapLogger{
		Logger: zap.NewNop(),
	}
}

// --- IMPLEMENTACJA METOD INTERFEJSU ---

func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

// --- SUGAR METHODS (Wygodne logowanie klucz-wartość) ---

func (l *zapLogger) Infow(msg string, keysAndValues ...any) {
	l.Logger.Sugar().Infow(msg, keysAndValues...)
}

func (l *zapLogger) Debugw(msg string, keysAndValues ...any) {
	l.Logger.Sugar().Debugw(msg, keysAndValues...)
}

func (l *zapLogger) Warnw(msg string, keysAndValues ...any) {
	l.Logger.Sugar().Warnw(msg, keysAndValues...)
}

func (l *zapLogger) Errorw(msg string, keysAndValues ...any) {
	l.Logger.Sugar().Errorw(msg, keysAndValues...)
}

func (l *zapLogger) Fatalw(msg string, keysAndValues ...any) {
	l.Logger.Sugar().Fatalw(msg, keysAndValues...)
}

// --- OBJECT LOGGING ---

func (l *zapLogger) InfoObj(msg string, obj any) {
	l.Logger.Info(msg, zap.Any("data", obj))
}

func (l *zapLogger) DebugObj(msg string, obj any) {
	l.Logger.Debug(msg, zap.Any("data", obj))
}

func (l *zapLogger) Sync() error {
	return l.Logger.Sync()
}
