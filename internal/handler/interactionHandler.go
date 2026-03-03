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

func (h *InteractionHandler) getFingerprint(c *fiber.Ctx) string {
	fp := c.Get("X-Fingerprint")
	if fp == "" {
		return c.IP()
	}
	return fp
}

func (h *InteractionHandler) RecordVisit(c *fiber.Ctx) error {
	fp := h.getFingerprint(c)

	resp, err := h.service.HandleInteraction(c.Context(), fp, service.TypeVisit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "nie udało się zarejestrować wizyty",
		})
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

	fp := h.getFingerprint(c)

	resp, err := h.service.HandleInteraction(c.Context(), fp, body.Type)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "nie udało się przetworzyć polubienia",
		})
	}

	return c.JSON(resp)
}

func (h *InteractionHandler) GetStats(c *fiber.Ctx) error {
	resp, err := h.service.GetStats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "nie udało się pobrać statystyk",
		})
	}

	return c.JSON(resp)
}
