package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/dinar-dz/dinar/internal/config"
	"github.com/dinar-dz/dinar/internal/http/server"
	"github.com/dinar-dz/dinar/internal/worker"
)

var version = "dev"

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("version", version)

	if len(os.Args) > 1 && os.Args[1] == "worker" {
		if err := worker.Run(context.Background(), cfg, logger); err != nil {
			logger.Error("worker stopped", "error", err)
			os.Exit(1)
		}
		return
	}

	app := server.New(cfg, logger)
	logger.Info("api listening", "addr", cfg.Addr, "environment", cfg.Environment)
	if err := app.Listen(cfg.Addr); err != nil {
		logger.Error("api stopped", "error", err)
		os.Exit(1)
	}
}
