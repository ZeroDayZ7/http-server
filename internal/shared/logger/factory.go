package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger(env Env) Logger {
	var core zapcore.Core

	switch env {
	case EnvProduction:
		core = createProductionCore()
	case EnvStaging:
		core = createStagingCore()
	case EnvTest:
		core = createTestCore()
	default: // Development
		core = createDevelopmentCore()
	}

	zapInst := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	return &zapLogger{zapInst}
}

// --- POMOCNICZE FUNKCJE KREUJĄCE ---

func createProductionCore() zapcore.Core {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zap.InfoLevel)
}

func createStagingCore() zapcore.Core {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zap.DebugLevel)
}

func createDevelopmentCore() zapcore.Core {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(config)

	logFile := zapcore.AddSync(&lumberjack.Logger{
		Filename: "logs/dev.log",
		MaxSize:  5,
	})

	return zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.DebugLevel),
		zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), logFile, zap.DebugLevel),
	)
}

func createTestCore() zapcore.Core {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder

	return zapcore.NewCore(
		zapcore.NewConsoleEncoder(config),
		zapcore.AddSync(os.Stdout),
		zap.WarnLevel,
	)
}
