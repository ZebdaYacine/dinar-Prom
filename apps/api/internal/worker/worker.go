package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/dinar-dz/dinar/internal/config"
)

// Run is the lifecycle-safe worker bootstrap. Jobs are added behind this loop
// as their PostgreSQL/Redis contracts are implemented in later epics.
func Run(ctx context.Context, _ config.Config, log *slog.Logger) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	log.Info("worker started")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			log.Info("worker heartbeat")
		}
	}
}
