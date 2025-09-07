package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/cache"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/validator"
)

type InteractionHandler struct {
	service *service.InteractionService
}

func NewInteractionHandler(svc *service.InteractionService) *InteractionHandler {
	return &InteractionHandler{service: svc}
}

// RecordVisit
func (h *InteractionHandler) RecordVisit(c *fiber.Ctx) error {
	ip := c.IP()
	var userID *uint
	if uid := c.Locals("userID"); uid != nil {
		id := uid.(uint)
		userID = &id
	}

	lastVisit, err := h.service.GetLastVisitByIP(ip)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get last visit"})
	}

	if !lastVisit.IsZero() && time.Since(lastVisit) < 30*time.Minute {
		count, _ := h.service.CountByType("visit")
		return c.JSON(fiber.Map{
			"ip":      ip,
			"visits":  count,
			"message": "visit recently recorded",
		})
	}

	// Zapis wizyty w bazie
	if err := h.service.Record(ip, userID, "visit", 0, nil); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to record visit"})
	}

	count, _ := h.service.CountByType("visit")
	return c.JSON(fiber.Map{"ip": ip, "visits": count})
}

func (h *InteractionHandler) RecordLike(c *fiber.Ctx) error {
	val := c.Locals("validatedBody")
	if val == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "body not validated"})
	}
	body, ok := val.(validator.InteractionRequest)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	ip := c.IP()
	var userID *uint
	if uid := c.Locals("userID"); uid != nil {
		id := uid.(uint)
		userID = &id
	}

	cacheKey := ip + ":" + body.Type

	// 1️⃣ Sprawdzenie w cache
	if cached, found := cache.InteractionCache.Get(cacheKey); found {
		return c.JSON(cached)
	}

	// 2️⃣ Sprawdzenie ostatniego like/dislike od tego IP w bazie
	last, err := h.service.GetLastInteractionByIP(ip, body.Type)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to check last interaction"})
	}

	if !last.IsZero() && time.Since(last) < 1*time.Hour {
		// jeśli limit nie minął, możemy w cache ustawić odczyt z bazy na 1h
		likes, _ := h.service.CountByType("like")
		dislikes, _ := h.service.CountByType("dislike")
		resp := fiber.Map{
			"ip":       ip,
			"likes":    likes,
			"dislikes": dislikes,
			"message":  "interaction recently recorded",
		}
		cache.InteractionCache.Set(cacheKey, resp, 1*time.Hour)
		return c.JSON(resp)
	}

	// 3️⃣ Zapis interakcji do bazy
	if err := h.service.Record(ip, userID, body.Type, 0, nil); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to record interaction"})
	}

	// 4️⃣ Pobranie liczby like i dislike i zapis do cache
	likes, _ := h.service.CountByType("like")
	dislikes, _ := h.service.CountByType("dislike")
	resp := fiber.Map{
		"ip":       ip,
		"likes":    likes,
		"dislikes": dislikes,
		"message":  "interaction recorded",
	}
	cache.InteractionCache.Set(cacheKey, resp, 1*time.Hour)

	return c.JSON(resp)
}
