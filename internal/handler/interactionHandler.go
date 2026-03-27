package handler

import (
	"github.com/gofiber/fiber/v2"
	apperrors "github.com/zerodayz7/http-server/internal/errors"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/validator"
)

type InteractionHandler struct {
	service service.InteractionServiceInterface
}

func NewInteractionHandler(service service.InteractionServiceInterface) *InteractionHandler {
	return &InteractionHandler{
		service: service,
	}
}

func (h *InteractionHandler) getFP(c *fiber.Ctx) string {
	ip := c.IP()
	ua := c.Get("User-Agent")
	lang := c.Get("Accept-Language")

	return h.service.GenerateFingerprint(ip, ua, lang)
}

func (h *InteractionHandler) HandleVisit(c *fiber.Ctx) error {
	fp := h.getFP(c)

	stats, err := h.service.ProcessInitialVisit(c.UserContext(), fp)
	if err != nil {
		return apperrors.ErrInternal.WithErr(err)
	}

	return c.JSON(stats)
}

func (h *InteractionHandler) HandleInteraction(c *fiber.Ctx) error {
	fp := h.getFP(c)

	body := c.Locals("validatedBody").(validator.InteractionRequest)

	stats, err := h.service.HandleInteraction(c.UserContext(), fp, body.Type)
	if err != nil {
		return apperrors.ErrInvalidRequest.WithErr(err)
	}

	return c.JSON(stats)
}

func (h *InteractionHandler) GetStats(c *fiber.Ctx) error {
	fp := h.getFP(c)

	stats, err := h.service.GetStats(c.UserContext(), fp)
	if err != nil {
		return apperrors.ErrInternal.WithErr(err)
	}

	return c.JSON(stats)
}
