package auth

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func Register(v1 fiber.Router, service *Service) {
	v1.Get("/auth/google/login", func(c *fiber.Ctx) error {
		target, state, err := service.BeginOAuth(c.Query("next", "/ar/onboarding/personal"))
		if err != nil {
			return errorJSON(c, fiber.StatusServiceUnavailable, "oauth_unavailable", "Google sign-in is not configured", nil)
		}
		c.Cookie(&fiber.Cookie{Name: "dinar_oauth_state", Value: state, HTTPOnly: true, Secure: service.cfg.CookieSecure || service.cfg.Environment == "production", SameSite: "Lax", Path: "/", MaxAge: 600})
		return c.Redirect(target, fiber.StatusFound)
	})
	v1.Get("/auth/google/callback", func(c *fiber.Ctx) error {
		state := c.Query("state")
		if state == "" || state != c.Cookies("dinar_oauth_state") {
			return errorJSON(c, fiber.StatusUnauthorized, "unauthorized", "Google sign-in could not be completed", nil)
		}
		user, next, err := service.finishOAuth(state, c.Query("code"))
		if err != nil {
			return errorJSON(c, fiber.StatusUnauthorized, "unauthorized", "Google sign-in could not be completed", nil)
		}
		if err := service.StartSession(user.ID, c); err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, "internal", "could not create session", nil)
		}
		c.ClearCookie("dinar_oauth_state")
		return c.Redirect(service.cfg.AppURL+service.LoginRedirect(user.ID, next), fiber.StatusFound)
	})
	v1.Post("/auth/refresh", requireCSRF(service), func(c *fiber.Ctx) error {
		access, err := service.Refresh(c)
		if err != nil {
			return errorJSON(c, 401, "unauthorized", "session expired", nil)
		}
		return success(c, fiber.Map{"access_token": access, "token_type": "Bearer", "expires_in": int(service.cfg.AccessTTL.Seconds())})
	})
	v1.Post("/auth/logout", requireCSRF(service), func(c *fiber.Ctx) error { service.Logout(c); return c.SendStatus(fiber.StatusNoContent) })

	protected := v1.Group("", service.RequireAuth())
	protected.Get("/me", func(c *fiber.Ctx) error {
		me, err := service.Current(c.Locals("user_id").(string))
		if err != nil {
			return errorJSON(c, 404, "not_found", "user not found", nil)
		}
		return success(c, me)
	})
	merchant := protected.Group("/merchant")
	merchant.Put("/profile", jsonBody(service.saveProfile))
	merchant.Put("/shop", jsonBody(service.saveShop))
	merchant.Post("/documents", func(c *fiber.Ctx) error {
		typ := c.FormValue("type")
		file, err := c.FormFile("file")
		if err != nil {
			return errorJSON(c, 422, "validation_failed", "file is required", nil)
		}
		if file.Size > 8<<20 {
			return errorJSON(c, 422, "validation_failed", "file is too large", nil)
		}
		opened, err := file.Open()
		if err != nil {
			return errorJSON(c, 422, "validation_failed", "file cannot be read", nil)
		}
		defer opened.Close()
		data, err := io.ReadAll(io.LimitReader(opened, 8<<20+1))
		if err != nil {
			return errorJSON(c, 422, "validation_failed", "file cannot be read", nil)
		}
		doc, err := service.SaveDocument(c.Locals("user_id").(string), typ, data)
		if err != nil {
			return errorJSON(c, 422, "validation_failed", err.Error(), nil)
		}
		return c.Status(201).JSON(fiber.Map{"data": fiber.Map{"id": doc.ID, "type": doc.Type, "content_type": doc.ContentType, "sha256": doc.SHA256, "size": doc.Size}})
	})
	merchant.Post("/submit", func(c *fiber.Ctx) error {
		var input struct {
			ConsentAccepted bool `json:"consent_accepted"`
		}
		if err := c.BodyParser(&input); err != nil {
			return errorJSON(c, 422, "validation_failed", "invalid request", nil)
		}
		merchantData, fields := service.Submit(c.Locals("user_id").(string), input.ConsentAccepted)
		if len(fields) > 0 {
			return errorJSON(c, 422, "validation_failed", "onboarding is incomplete", fields)
		}
		return success(c, merchantData)
	})
}

func (s *Service) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := strings.TrimSpace(c.Get("Authorization"))
		if !strings.HasPrefix(raw, "Bearer ") {
			return errorJSON(c, 401, "unauthorized", "authentication required", nil)
		}
		userID, err := s.VerifyAccess(strings.TrimSpace(strings.TrimPrefix(raw, "Bearer ")))
		if err != nil {
			return errorJSON(c, 401, "unauthorized", "authentication required", nil)
		}
		c.Locals("user_id", userID)
		return c.Next()
	}
}
func requireCSRF(s *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cookie, header := c.Cookies(s.cfg.CSRFCookieName), c.Get(s.cfg.CSRFHeaderName)
		if cookie == "" || header == "" || cookie != header {
			return errorJSON(c, 403, "forbidden", "csrf validation failed", nil)
		}
		return c.Next()
	}
}
func jsonBody(handler func(*fiber.Ctx, []byte) (any, map[string]string)) fiber.Handler {
	return func(c *fiber.Ctx) error {
		value, fields := handler(c, c.Body())
		if len(fields) > 0 {
			return errorJSON(c, 422, "validation_failed", "check the fields", fields)
		}
		return success(c, value)
	}
}
func (s *Service) saveProfile(c *fiber.Ctx, body []byte) (any, map[string]string) {
	var input Profile
	if json.Unmarshal(body, &input) != nil {
		return input, map[string]string{"body": "invalid request"}
	}
	return s.SaveProfile(c.Locals("user_id").(string), input)
}
func (s *Service) saveShop(c *fiber.Ctx, body []byte) (any, map[string]string) {
	var input Shop
	if json.Unmarshal(body, &input) != nil {
		return input, map[string]string{"body": "invalid request"}
	}
	return s.SaveShop(c.Locals("user_id").(string), input)
}
func success(c *fiber.Ctx, value any) error { return c.JSON(fiber.Map{"data": value}) }
func errorJSON(c *fiber.Ctx, status int, code, message string, fields map[string]string) error {
	payload := fiber.Map{"code": code, "message": message}
	if fields != nil {
		payload["fields"] = fields
	}
	return c.Status(status).JSON(fiber.Map{"error": payload})
}
