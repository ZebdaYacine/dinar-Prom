package server

import (
	"log/slog"
	"os"

	"github.com/dinar-dz/dinar/internal/config"
	"github.com/dinar-dz/dinar/internal/guest"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func New(cfg config.Config, log *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "${time} ${status} ${method} ${path} request_id=${locals:requestid}\\n",
		Output:     os.Stdout,
		TimeFormat: "2006-01-02T15:04:05Z07:00",
	}))
	app.Use(cors.New(cors.Config{AllowOrigins: "http://localhost:8180,http://localhost:3143", AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Request-ID"}))

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/readyz", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ready", "environment": cfg.Environment})
	})

	v1 := app.Group("/api/v1")
	v1.Get("", func(c *fiber.Ctx) error {
		return success(c, fiber.Map{"service": "dinar-api", "version": "v1"})
	})
	guest.RegisterBranding(v1)
	guest.NewCatalog(cfg.AppURL).Register(v1)

	return app
}

func success(c *fiber.Ctx, data any) error {
	return c.JSON(fiber.Map{"data": data})
}
