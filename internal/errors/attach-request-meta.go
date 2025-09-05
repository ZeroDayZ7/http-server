package errors

import (
	"github.com/gofiber/fiber/v2"
)

func AttachRequestMeta(c *fiber.Ctx, err *AppError, keysToInclude ...string) {
	if err.Meta == nil {
		err.Meta = make(map[string]any)
	}

	requestID := c.Locals("requestid")
	if requestID != nil {
		err.Meta["requestID"] = requestID.(string)
	} else {
		err.Meta["requestID"] = ""
	}

	err.Meta["ip"] = c.IP()
	err.Meta["path"] = c.Path()
	err.Meta["method"] = c.Method()

	if len(keysToInclude) > 0 {
		var body map[string]any
		if parseErr := c.BodyParser(&body); parseErr == nil {
			for _, k := range keysToInclude {
				if v, ok := body[k]; ok {
					err.Meta[k] = v
				}
			}
		}
	}
}
