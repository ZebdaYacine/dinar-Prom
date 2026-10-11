package server

import (
	"database/sql"
	"errors"
	"log/slog"
	"os"

	"github.com/dinar-dz/dinar/internal/auth"
	"github.com/dinar-dz/dinar/internal/config"
	"github.com/dinar-dz/dinar/internal/guest"
	"github.com/dinar-dz/dinar/internal/offers"
	"github.com/dinar-dz/dinar/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(cfg config.Config, log *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true, BodyLimit: 9 << 20})
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "${time} ${status} ${method} ${path} request_id=${locals:requestid}\\n",
		Output:     os.Stdout,
		TimeFormat: "2006-01-02T15:04:05Z07:00",
	}))
	allowedOrigins := cfg.AllowedOrigins
	if allowedOrigins == "" {
		allowedOrigins = "http://127.0.0.1:8180,http://localhost:8180,http://127.0.0.1:3143,http://localhost:3143"
	}
	csrfHeader := cfg.CSRFHeaderName
	if csrfHeader == "" {
		csrfHeader = "X-CSRF-Token"
	}
	app.Use(cors.New(cors.Config{AllowOrigins: allowedOrigins, AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Request-ID, " + csrfHeader, AllowCredentials: true}))

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
	var database *sql.DB
	if cfg.DatabaseURL != "" {
		opened, err := sql.Open("pgx", cfg.DatabaseURL)
		if err != nil {
			log.Warn("guest analytics database disabled", "error", err)
		} else {
			database = opened
		}
	}
	guest.NewCatalog(cfg.AppURL, database).Register(v1)
	privateStore := storage.PrivateStore(storage.NewMemoryStore())
	publicStore := storage.ObjectStore(storage.NewMemoryStore())
	var storageErr error
	if cfg.MinIOEndpoint != "" {
		privateStore, storageErr = storage.NewS3Store(storage.S3Config{Endpoint: cfg.MinIOEndpoint, AccessKey: cfg.MinIOAccessKey, SecretKey: cfg.MinIOSecretKey, Bucket: cfg.MinIOBucketPrivate, Region: cfg.MinIORegion, UseSSL: cfg.MinIOUseSSL})
		if storageErr == nil {
			publicStore, storageErr = storage.NewS3Store(storage.S3Config{Endpoint: cfg.MinIOEndpoint, AccessKey: cfg.MinIOAccessKey, SecretKey: cfg.MinIOSecretKey, Bucket: cfg.MinIOBucketPublic, Region: cfg.MinIORegion, UseSSL: cfg.MinIOUseSSL})
		}
	} else if cfg.Environment == "production" {
		storageErr = errors.New("private object storage is required in production")
	}
	identity, err := auth.New(auth.Config{Database: database, AppURL: cfg.AppURL, Environment: cfg.Environment, GoogleClientID: cfg.GoogleClientID, GoogleClientSecret: cfg.GoogleClientSecret, GoogleRedirectURL: cfg.GoogleRedirectURL, Issuer: cfg.JWTIssuer, Audience: cfg.JWTAudience, RefreshCookieName: cfg.RefreshCookieName, CSRFCookieName: cfg.CSRFCookieName, CSRFHeaderName: cfg.CSRFHeaderName, JWTPrivateKeyFile: cfg.JWTPrivateKeyFile, JWTPublicKeyFile: cfg.JWTPublicKeyFile, CookieSecure: cfg.CookieSecure, AccessTTL: cfg.JWTAccessTTL, RefreshTTL: cfg.JWTRefreshTTL, IPHashSalt: cfg.IPHashSalt, PrivateStore: privateStore})
	if storageErr != nil {
		err = storageErr
	} else if cfg.Environment == "production" && database == nil {
		err = errors.New("merchant identity database is required in production")
	}
	if err != nil {
		log.Error("authentication disabled", "error", err)
	} else {
		auth.Register(v1, identity)
		if database != nil {
			offers.Register(v1, identity, offers.New(database, publicStore, cfg.PublicMediaURL))
		}
	}

	return app
}

func success(c *fiber.Ctx, data any) error {
	return c.JSON(fiber.Map{"data": data})
}
