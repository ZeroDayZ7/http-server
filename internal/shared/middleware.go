package shared

import (
	"fmt"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/shared/logger"
	"go.uber.org/zap"
)

// isSensitive sprawdza, czy klucz powinien być maskowany w logach
func isSensitive(key string) bool {
	sensitiveKeys := []string{"password", "token", "secret", "authorization", "cookie"}
	return slices.Contains(sensitiveKeys, key)
}

func RequestLoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {

		if c.Method() == fiber.MethodOptions {
			return c.Next()
		}
		start := time.Now()
		log := logger.GetLogger()

		// Sprawdzamy poziom logowania (czy jesteśmy w trybie debug/dev)
		isDev := log.Core().Enabled(zap.DebugLevel)

		// 1. Wyciąganie Body - używamy mapy, żeby ładnie sformatować JSON
		var bodyMap map[string]interface{}

		// Próbujemy parsować body tylko jeśli metoda na to pozwala
		if c.Method() == fiber.MethodPost || c.Method() == fiber.MethodPut {
			// Używamy BodyParser do mapy, żeby podejrzeć co idzie z frontu
			if err := c.BodyParser(&bodyMap); err != nil {
				// Jeśli nie uda się sparsować do mapy, nic nie robimy - handler sam obsłuży błąd
				bodyMap = nil
			}
		}

		if isDev {
			fmt.Printf("\n--- [DEBUG] INCOMING REQUEST ---\n")
			fmt.Printf("Method:  %s\n", c.Method())
			fmt.Printf("Path:    %s\n", c.Path())
			fmt.Printf("IP:      %s\n", c.IP())

			if bodyMap != nil {
				fmt.Printf("Body Content:\n")
				for k, v := range bodyMap {
					displayValue := v
					if isSensitive(k) {
						displayValue = "********"
					}
					fmt.Printf("  -> %s: %v\n", k, displayValue)
				}
			} else if len(c.Body()) > 0 {
				fmt.Printf("Raw Body (non-json): %s\n", string(c.Body()))
			}
			fmt.Printf("-------------------------------\n")
		}

		// Kontynuacja do handlera
		err := c.Next()

		// Logowanie po zakończeniu requestu (strukturalne)
		latency := time.Since(start)
		status := c.Response().StatusCode()

		log.Info("Request Processed",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.String("latency", latency.String()),
			zap.Any("body_payload", bodyMap), // To pokaże fingerprint w logach JSON/Zap
			zap.String("ip", c.IP()),
		)

		return err
	}
}
