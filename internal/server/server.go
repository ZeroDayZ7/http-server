package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/config"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

func New(cfg config.Config) *fiber.App {
	log := logger.GetLogger()

	return fiber.New(fiber.Config{
		AppName:       "http-server",
		ServerHeader:  "ZeroDayZ7",
		Prefork:       false,
		CaseSensitive: true,
		StrictRouting: true,
		BodyLimit:     cfg.BodyLimitMB * 1024 * 1024,
		IdleTimeout:   30 * time.Second,
		ReadTimeout:   15 * time.Second,
		WriteTimeout:  15 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if e, ok := err.(*fiber.Error); ok {
				log.Error("HTTP error", zap.Error(err), zap.String("path", c.Path()))
				return c.Status(e.Code).JSON(fiber.Map{"error": e.Message})
			}
			log.Error("Server error", zap.Error(err), zap.String("path", c.Path()))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		},
	})
}
