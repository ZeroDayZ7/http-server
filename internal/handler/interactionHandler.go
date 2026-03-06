package handler

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
	"github.com/zerodayz7/http-server/internal/errors"
	"github.com/zerodayz7/http-server/internal/service"
	"github.com/zerodayz7/http-server/internal/validator"
)

type InteractionHandler struct {
	service *service.InteractionService
	salt    string
}

func NewInteractionHandler(svc *service.InteractionService, salt string) *InteractionHandler {
	return &InteractionHandler{
		service: svc,
		salt:    salt,
	}
}

func HandleServiceError(c *fiber.Ctx, err error) error {
	if appErr, ok := err.(*errors.AppError); ok {
		return errors.SendAppError(c, appErr)
	}

	return errors.SendAppError(c, errors.ErrInternal)
}

func GetValidated[T any](c *fiber.Ctx, key string) (T, error) {
	var zero T

	v, ok := c.Locals(key).(T)
	if !ok {
		return zero, errors.ErrInvalidRequest
	}

	return v, nil
}

func (h *InteractionHandler) getFingerprint(c *fiber.Ctx) string {
	ip := c.IP()
	ua := c.Get("User-Agent")
	lang := c.Get("Accept-Language")

	raw := ip + "|" + ua + "|" + lang + "|" + h.salt

	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func (h *InteractionHandler) InitializeSession(c *fiber.Ctx) error {
	fp := h.getFingerprint(c)

	resp, err := h.service.ProcessInitialVisit(c.Context(), fp)
	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(resp)
}

func (h *InteractionHandler) RecordVisit(c *fiber.Ctx) error {

	body, _ := GetValidated[validator.InteractionRequest](c, "validatedBody")

	fp := body.Fingerprint
	if fp == "" {
		fp = h.getFingerprint(c)
	}

	resp, err := h.service.HandleInteraction(
		c.Context(),
		fp,
		service.TypeVisit,
	)

	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(resp)
}

func (h *InteractionHandler) GetStats(c *fiber.Ctx) error {

	query, err := GetValidated[validator.FingerprintRequest](c, "validatedQuery")
	if err != nil {
		return HandleServiceError(c, err)
	}

	resp, err := h.service.GetStats(
		c.Context(),
		query.Fingerprint,
	)

	if err != nil {
		return HandleServiceError(c, err)
	}

	return c.JSON(resp)
}

func (h *InteractionHandler) RecordLike(c *fiber.Ctx) error {
	body, err := GetValidated[validator.InteractionRequest](c, "validatedBody")
	if err != nil {
		return HandleServiceError(c, err)
	}

	fp := h.getFingerprint(c)

	resp, err := h.service.HandleInteraction(c.Context(), fp, body.Type)
	if err != nil {
		return HandleServiceError(c, err)
	}
	return c.JSON(resp)
}
