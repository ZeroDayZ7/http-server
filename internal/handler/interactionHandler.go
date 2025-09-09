package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/validator"
)

type InteractionHandler struct {
	service *service.InteractionService
}

func NewInteractionHandler(svc *service.InteractionService) *InteractionHandler {
	return &InteractionHandler{
		service: svc,
	}
}

// getUserIPAndID extracts IP and optional userID from context
func (h *InteractionHandler) getUserIPAndID(c *fiber.Ctx) (string, *uint) {
	ip := c.IP()
	var userID *uint
	if uid := c.Locals("userID"); uid != nil {
		id := uid.(uint)
		userID = &id
	}
	return ip, userID
}

// Public Endpoints
func (h *InteractionHandler) RecordVisit(c *fiber.Ctx) error {
	ip, userID := h.getUserIPAndID(c)
	resp, err := h.service.HandleInteraction(ip, userID, service.TypeVisit, service.VisitCooldown, true)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "nie udało się przetworzyć interakcji"})
	}
	return c.JSON(resp)
}

func (h *InteractionHandler) RecordLike(c *fiber.Ctx) error {
	val := c.Locals("validatedBody")
	if val == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "nieprawidłowe dane"})
	}
	body, ok := val.(validator.InteractionRequest)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "nieprawidłowy format danych"})
	}
	ip, userID := h.getUserIPAndID(c)
	resp, err := h.service.HandleInteraction(ip, userID, body.Type, service.LikeCooldown, true)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "nie udało się przetworzyć interakcji"})
	}
	return c.JSON(resp)
}

func (h *InteractionHandler) GetStats(c *fiber.Ctx) error {
	ip, userID := h.getUserIPAndID(c)
	resp, err := h.service.HandleInteraction(ip, userID, service.TypeVisit, 0, false) // No record, no cooldown
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "nie udało się pobrać statystyk"})
	}
	return c.JSON(resp)
}
