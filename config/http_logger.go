// config/fiber_logger.go
package config

import (
	"io"
	"os"

	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gopkg.in/natefinch/lumberjack.v2"
)

// FiberLoggerMiddleware zwraca gotowe middleware Fiber do logów HTTP (do pliku + konsoli)
func FiberLoggerMiddleware() fiber.Handler {
	// Request ID middleware
	requestIDMiddleware := requestid.New()

	// Rotacja plików z lumberjack
	logFile := &lumberjack.Logger{
		Filename:   "./logs/http.log",
		MaxSize:    10,   // MB
		MaxBackups: 5,    // ilość backupów
		MaxAge:     7,    // dni
		Compress:   true, // kompresja
	}

	// MultiWriter: zapis jednocześnie do konsoli i do pliku
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	format := "${pid} | ${locals:requestid} | ${status} | ${latency} | ${ip} | ${method} | ${path}\n"
	timeFormat := "2006-01-02 15:04:05.000"

	loggerConfig := fiberlogger.Config{
		Output:     multiWriter,
		TimeFormat: timeFormat,
		Format:     format,
		TimeZone:   "Europe/Warsaw",
		// Callback po każdym logu, np. wysyłka 5xx na Slack
		Done: func(c *fiber.Ctx, logBytes []byte) {
			if c.Response().StatusCode() >= 500 {
				// reporter.SendToSlack(logBytes)
			}
		},
	}

	// Łączymy middleware: requestID + logger
	return func(c *fiber.Ctx) error {
		_ = requestIDMiddleware(c)
		return fiberlogger.New(loggerConfig)(c)
	}
}
