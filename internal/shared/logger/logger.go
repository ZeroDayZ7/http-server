package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type zapLogger struct {
	*zap.Logger
}

func NewLogger(env string) Logger {
	level := zap.DebugLevel
	if env == "production" {
		level = zap.InfoLevel
	}

	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	if env == "production" {
		consoleEncoderConfig = zap.NewProductionEncoderConfig()
	}
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logFile := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     7,
		Compress:   true,
	}

	fileEncoderConfig := zap.NewProductionEncoderConfig()

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

	zapInst := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	return &zapLogger{zapInst}
}

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

func (l *zapLogger) InfoObj(msg string, obj any) {
	l.Logger.Info(msg, zap.Any("data", obj))
}

func (l *zapLogger) DebugObj(msg string, obj any) {
	l.Logger.Debug(msg, zap.Any("data", obj))
}

func (l *zapLogger) Sync() error {
	return l.Logger.Sync()
}
