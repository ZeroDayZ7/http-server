package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/service"
)

type InteractionHandler struct {
	service service.InteractionServiceInterface
}

func NewInteractionHandler(service service.InteractionServiceInterface) *InteractionHandler {
	return &InteractionHandler{
		service: service,
	}
}

// Pomocnicza funkcja do wyciągania danych i tworzenia FP
func (h *InteractionHandler) getFP(c *fiber.Ctx) string {
	ip := c.IP()
	ua := c.Get("User-Agent")
	lang := c.Get("Accept-Language")

	// Wywołujemy serwis tożsamości poprzez główny serwis
	return h.service.GenerateFingerprint(ip, ua, lang)
}

func (h *InteractionHandler) HandleVisit(c *fiber.Ctx) error {
	fp := h.getFP(c) // <--- Generujemy na backendzie, nie z Query!

	stats, err := h.service.ProcessInitialVisit(c.Context(), fp)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(stats)
}

func (h *InteractionHandler) HandleLike(c *fiber.Ctx) error {
	fp := h.getFP(c)

	stats, err := h.service.HandleInteraction(c.Context(), fp, service.TypeLike)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(stats)
}

func (h *InteractionHandler) HandleDislike(c *fiber.Ctx) error {
	fp := h.getFP(c)

	stats, err := h.service.HandleInteraction(c.Context(), fp, service.TypeDislike)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(stats)
}

func (h *InteractionHandler) GetStats(c *fiber.Ctx) error {
	fp := h.getFP(c)

	stats, err := h.service.GetStats(c.Context(), fp)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(stats)
}
