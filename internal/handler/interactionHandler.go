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
	// Pobierz fingerprint z body lub query (zależnie jak wysyłasz wizytę)
	// Jeśli wizyta to POST przy wejściu:
	var body validator.InteractionRequest
	if err := c.BodyParser(&body); err != nil {
		// Fallback do IP jeśli body puste, ale lepiej wysłać FP z frontu
		fp := c.Query("fp")
		if fp == "" {
			fp = c.IP()
		}
		resp, _ := h.service.HandleInteraction(c.Context(), fp, service.TypeVisit)
		return c.JSON(resp)
	}

	resp, err := h.service.HandleInteraction(c.Context(), body.Fingerprint, service.TypeVisit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(resp)
}

func (h *InteractionHandler) GetStats(c *fiber.Ctx) error {
	// Skoro mamy middleware ValidateBody[validator.FingerprintRequest],
	// to body jest już sparsowane i bezpieczne.
	var body validator.FingerprintRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	resp, err := h.service.GetStats(c.Context(), body.Fingerprint)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(resp)
}

func (h *InteractionHandler) RecordLike(c *fiber.Ctx) error {
	// Body z frontu: { type: 'like', fingerprint: 'HASH' }
	var body validator.InteractionRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	// Używamy fingerprintu z body zamiast IP!
	resp, err := h.service.HandleInteraction(c.Context(), body.Fingerprint, body.Type)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(resp)
}
