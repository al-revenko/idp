package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/al-revenko/idp/internal/app/transport"
)

type pattern = string
type methods = []string

func TraceMiddleware(log *slog.Logger, excludePaths map[pattern]methods) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if methods, ok := excludePaths[r.Pattern]; ok {
				for _, method := range methods {
					if method == r.Method {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			traceLog := log.With(slog.String("method", r.Method), slog.String("path", r.URL.Path))

			transport.Trace(transport.TraceParams{
				Ctx:       r.Context(),
				Log:       traceLog,
				Transport: transport.HTTP,
			}, func(ctx context.Context, log *slog.Logger) *slog.Logger {
				recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

				next.ServeHTTP(recorder, r.WithContext(ctx))

				return log.With(slog.Uint64("code", uint64(recorder.status)), slog.String("status", http.StatusText(recorder.status)))
			})
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
