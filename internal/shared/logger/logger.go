package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	*zap.Logger
}

var (
	instance *Logger
	once     sync.Once
)

func InitLogger(env string) (*Logger, error) {
	var err error
	once.Do(func() {
		var cfg zap.Config

		if env == "production" {
			cfg = zap.NewProductionConfig()
			cfg.OutputPaths = []string{"stdout"}
		} else {
			cfg = zap.NewDevelopmentConfig()
			cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			cfg.OutputPaths = []string{"stdout"}
		}

		logFile := &lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    10, // MB
			MaxBackups: 5,
			MaxAge:     7, // days
			Compress:   true,
		}

		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(cfg.EncoderConfig),
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		)
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(logFile),
			zapcore.InfoLevel,
		)

		core := zapcore.NewTee(consoleCore, fileCore)

		zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		instance = &Logger{zapLogger}
	})
	return instance, err
}

func GetLogger() *Logger {
	if instance == nil {
		InitLogger("development")
	}
	return instance
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

func (l *Logger) InfoObj(msg string, obj any) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).Info(msg, zap.Any("data", obj))
}

func (l *Logger) DebugObj(msg string, obj any) {
	l.Logger.WithOptions(zap.AddCallerSkip(1)).Debug(msg, zap.Any("data", obj))
}
