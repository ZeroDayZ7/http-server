package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// zapLogger implementuje interfejs Logger
type zapLogger struct {
	*zap.Logger
}

// NewLogger to teraz Twój "Provider" dla Wire.
// Przyjmuje env (np. z flag startowych lub zmiennej środowiskowej), aby wiedzieć jak się skonfigurować.
func NewLogger(env string) Logger {
	level := zap.DebugLevel
	if env == "production" {
		level = zap.InfoLevel
	}

	// Konfiguracja Konsoli
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	if env == "production" {
		consoleEncoderConfig = zap.NewProductionEncoderConfig()
	}
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// Konfiguracja Pliku (Lumberjack)
	logFile := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     7, // days
		Compress:   true,
	}

	fileEncoderConfig := zap.NewProductionEncoderConfig()

	// Tworzenie Core
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(fileEncoderConfig),
		zapcore.AddSync(logFile),
		zap.InfoLevel,
	)

	core := zapcore.NewTee(consoleCore, fileCore)

	// Tworzenie finalnej instancji zap.Logger
	zapInst := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	return &zapLogger{zapInst}
}

// --- Implementacja metod interfejsu Logger ---

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

func (l *zapLogger) InfoObj(msg string, obj any) {
	l.Logger.Info(msg, zap.Any("data", obj))
}

func (l *zapLogger) DebugObj(msg string, obj any) {
	l.Logger.Debug(msg, zap.Any("data", obj))
}

func (l *zapLogger) Sync() error {
	return l.Logger.Sync()
}
