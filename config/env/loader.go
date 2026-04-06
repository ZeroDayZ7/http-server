package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

var validate = validator.New()

func LoadConfig(cfg *Config) error {
	setDefaults()

	viper.BindEnv("WORKER_PORT")
	viper.BindEnv("DB_HOST")
	viper.BindEnv("ENV")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// AUTO detect pliku .env
	if fileExists(".env") {
		viper.SetConfigFile(".env")

		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("error loading config file: %w", err)
		}
	}

	// Mapowanie
	if err := viper.Unmarshal(cfg, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	)); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Walidacja
	if err := validate.Struct(cfg); err != nil {
		var errorMsgs []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMsgs = append(errorMsgs, fmt.Sprintf(
				"- Field '%s' failed on '%s' (value: %v)",
				err.Field(),
				err.Tag(),
				err.Value(),
			))
		}
		return fmt.Errorf("config validation failed:\n%s", strings.Join(errorMsgs, "\n"))
	}

	return nil
}

// helper do sprawdzania pliku
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
