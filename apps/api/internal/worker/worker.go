package worker

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/dinar-dz/dinar/internal/config"
	"github.com/dinar-dz/dinar/internal/notifications"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Run is the lifecycle-safe worker bootstrap. Jobs are added behind this loop
// as their PostgreSQL/Redis contracts are implemented in later epics.
func Run(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	if cfg.DatabaseURL == "" {
		return sql.ErrConnDone
	}
	if cfg.Environment == "production" && cfg.SMTPHost == "" {
		return errors.New("SMTP_HOST is required in production")
	}
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	mailer := notifications.NewSMTPMailer(cfg)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	log.Info("worker started")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			for {
				processed, err := notifications.ProcessOne(ctx, db, mailer)
				if err != nil {
					log.Warn("notification delivery failed", "error", err)
				}
				if !processed || err != nil {
					break
				}
			}
		}
	}
}
