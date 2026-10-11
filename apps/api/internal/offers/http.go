package offers

import (
	"errors"
	"io"
	"strings"

	"github.com/dinar-dz/dinar/internal/auth"
	"github.com/gofiber/fiber/v2"
)

func Register(v1 fiber.Router, identity *auth.Service, service *Service) {
	protected := v1.Group("", identity.RequireAuth())
	merchant := protected.Group("/merchant", identity.RequirePermission("merchant_offers", "write"))
	merchant.Get("/offers", func(c *fiber.Ctx) error {
		items, err := service.List(c.UserContext(), userID(c))
		if err != nil {
			return respondError(c, err)
		}
		return ok(c, items)
	})
	merchant.Post("/offers", func(c *fiber.Ctx) error {
		item, err := service.Create(c.UserContext(), userID(c))
		if err != nil {
			return respondError(c, err)
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": item})
	})
	merchant.Get("/offers/:id", func(c *fiber.Ctx) error {
		item, err := service.Get(c.UserContext(), userID(c), c.Params("id"))
		if err != nil {
			return respondError(c, err)
		}
		return ok(c, item)
	})
	merchant.Put("/offers/:id", func(c *fiber.Ctx) error {
		var input Input
		if err := c.BodyParser(&input); err != nil {
			return validation(c, map[string]string{"body": "invalid request"})
		}
		item, fields, err := service.Update(c.UserContext(), userID(c), c.Params("id"), input)
		if err != nil {
			return respondError(c, err)
		}
		if len(fields) > 0 {
			return validation(c, fields)
		}
		return ok(c, item)
	})
	merchant.Delete("/offers/:id", func(c *fiber.Ctx) error {
		if err := service.Delete(c.UserContext(), userID(c), c.Params("id")); err != nil {
			return respondError(c, err)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})
	merchant.Post("/offers/:id/submit", action(func(c *fiber.Ctx) (any, error) {
		item, fields, err := service.Submit(c.UserContext(), userID(c), c.Params("id"))
		if len(fields) > 0 {
			return nil, &fieldError{fields}
		}
		return item, err
	}))
	merchant.Post("/offers/:id/pause", action(func(c *fiber.Ctx) (any, error) { return service.Pause(c.UserContext(), userID(c), c.Params("id")) }))
	merchant.Post("/offers/:id/resume", action(func(c *fiber.Ctx) (any, error) { return service.Resume(c.UserContext(), userID(c), c.Params("id")) }))
	merchant.Post("/offers/:id/archive", action(func(c *fiber.Ctx) (any, error) { return service.Archive(c.UserContext(), userID(c), c.Params("id")) }))
	merchant.Post("/offers/:id/duplicate", func(c *fiber.Ctx) error {
		item, err := service.Duplicate(c.UserContext(), userID(c), c.Params("id"))
		if err != nil {
			return respondError(c, err)
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": item})
	})
	merchant.Post("/offers/:id/media", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return validation(c, map[string]string{"file": "image is required"})
		}
		if file.Size > maxUploadBytes {
			return validation(c, map[string]string{"file": "image is too large"})
		}
		opened, err := file.Open()
		if err != nil {
			return validation(c, map[string]string{"file": "image cannot be read"})
		}
		defer opened.Close()
		data, err := io.ReadAll(io.LimitReader(opened, maxUploadBytes+1))
		if err != nil {
			return validation(c, map[string]string{"file": "image cannot be read"})
		}
		item, err := service.AddMedia(c.UserContext(), userID(c), c.Params("id"), c.FormValue("alt_text"), data)
		if err != nil {
			return respondError(c, err)
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": item})
	})
	merchant.Delete("/offers/:id/media/:mediaID", func(c *fiber.Ctx) error {
		if err := service.DeleteMedia(c.UserContext(), userID(c), c.Params("id"), c.Params("mediaID")); err != nil {
			return respondError(c, err)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})
	merchant.Put("/offers/:id/media/order", func(c *fiber.Ctx) error {
		var input struct {
			IDs []string `json:"ids"`
		}
		if err := c.BodyParser(&input); err != nil {
			return validation(c, map[string]string{"ids": "invalid media order"})
		}
		if err := service.ReorderMedia(c.UserContext(), userID(c), c.Params("id"), input.IDs); err != nil {
			return respondError(c, err)
		}
		return ok(c, fiber.Map{"ordered": true})
	})
	merchant.Get("/stats", func(c *fiber.Ctx) error {
		stats, err := service.Stats(c.UserContext(), userID(c))
		if err != nil {
			return respondError(c, err)
		}
		return ok(c, stats)
	})

	admin := protected.Group("/admin", identity.RequirePermission("offer_review", "write"))
	admin.Get("/offers", func(c *fiber.Ctx) error {
		items, err := service.AdminQueue(c.UserContext())
		if err != nil {
			return respondError(c, err)
		}
		return ok(c, items)
	})
	admin.Post("/offers/:id/review", func(c *fiber.Ctx) error {
		var input struct {
			Status string `json:"status"`
			Reason string `json:"reason"`
		}
		if err := c.BodyParser(&input); err != nil {
			return validation(c, map[string]string{"body": "invalid request"})
		}
		item, err := service.Review(c.UserContext(), userID(c), c.Params("id"), strings.TrimSpace(input.Status), strings.TrimSpace(input.Reason))
		if err != nil {
			return respondError(c, err)
		}
		return ok(c, item)
	})
}

type fieldError struct{ fields map[string]string }

func (e *fieldError) Error() string    { return "validation failed" }
func userID(c *fiber.Ctx) string       { return c.Locals("user_id").(string) }
func ok(c *fiber.Ctx, value any) error { return c.JSON(fiber.Map{"data": value}) }
func validation(c *fiber.Ctx, fields map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": fiber.Map{"code": "validation_failed", "message": "check the fields", "fields": fields}})
}
func action(handler func(*fiber.Ctx) (any, error)) fiber.Handler {
	return func(c *fiber.Ctx) error {
		value, err := handler(c)
		if err != nil {
			var fields *fieldError
			if errors.As(err, &fields) {
				return validation(c, fields.fields)
			}
			return respondError(c, err)
		}
		return ok(c, value)
	}
}
func respondError(c *fiber.Ctx, err error) error {
	status, code := fiber.StatusInternalServerError, "internal"
	switch {
	case errors.Is(err, ErrNotFound):
		status, code = fiber.StatusNotFound, "not_found"
	case errors.Is(err, ErrForbidden), errors.Is(err, ErrMerchantApproval):
		status, code = fiber.StatusForbidden, "forbidden"
	case errors.Is(err, ErrLocked), errors.Is(err, ErrActiveLimit):
		status, code = fiber.StatusConflict, "conflict"
	case strings.Contains(err.Error(), "image"), strings.Contains(err.Error(), "media order"), strings.Contains(err.Error(), "rejection reason"), strings.Contains(err.Error(), "maximum number"):
		status, code = fiber.StatusUnprocessableEntity, "validation_failed"
	}
	message := err.Error()
	if status == fiber.StatusInternalServerError {
		message = "request could not be completed"
	}
	return c.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": code, "message": message}})
}
