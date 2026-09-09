package app

import (
	"context"
	"fmt"
	"log/slog"
)

type redisSlogAdapter struct {
	log *slog.Logger
}

func (r *redisSlogAdapter) Printf(ctx context.Context, format string, v ...interface{}) {
	r.log.InfoContext(ctx, fmt.Sprintf(format, v...))
}
