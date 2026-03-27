package config

import (
	"github.com/zerodayz7/http-server/config/env"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

var AppConfig env.Config

func LoadConfigGlobal(log logger.Logger) error {
	if err := env.LoadConfig(&AppConfig); err != nil {
		log.Error("Failed to load configuration", zap.Error(err))
		return err
	}

	log.Info("Configuration loaded successfully",
		zap.String("env", AppConfig.Server.Env),
		zap.String("app", AppConfig.Server.AppName),
		zap.String("port", AppConfig.Server.Port),
	)

	return nil
}
