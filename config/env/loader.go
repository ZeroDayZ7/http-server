package env

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

var validate = validator.New()

func LoadConfig(cfg *Config) error {
	setDefaults()

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error loading config file: %w", err)
		}
	}

	// Magia mapowania
	if err := viper.Unmarshal(cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	))); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Walidacja
	if err := validate.Struct(cfg); err != nil {
		var errorMsgs []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMsgs = append(errorMsgs, fmt.Sprintf("- Field '%s' failed on '%s' (value: %v)", err.Field(), err.Tag(), err.Value()))
		}
		return fmt.Errorf("config validation failed:\n%s", strings.Join(errorMsgs, "\n"))
	}

	return nil
}
