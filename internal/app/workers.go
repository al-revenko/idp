package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/al-revenko/idp/internal/lib/crypt/keys"
)

func keyRotationWorker(ctx context.Context, keyManager *keys.Manager, interval time.Duration, log *slog.Logger) {
	op := pkg.Op("keyRotationWorker")
	log = log.With(slog.String("op", op.Name))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("key rotation worker stopped")
			return
		case <-ticker.C:
			if err := keyManager.RotateKeys(); err != nil {
				log.Error("failed to rotate keys", slog.String("error", err.Error()))
			} else {
				log.Info("keys rotated")
			}
		}
	}
}
