package app

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/lestrrat-go/jwx/v4/jwk"
)

var ExcludeTraceHTTP = map[string][]string{
	"/health": {http.MethodGet},
}

type Mux interface {
	HandleFunc(path string, handler func(http.ResponseWriter, *http.Request))
}

type JWKSProvider interface {
	JWKS() (jwk.Set, error)
}

func RegisterHTTP(mux Mux, jwks JWKSProvider, serviceName string, log *slog.Logger) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		resp, err := json.Marshal(struct {
			Name string `json:"name"`
		}{
			Name: serviceName,
		})
		if err != nil {
			log.Error("failed to marshal health response", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	})

	jwksHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/jwk-set+json")

		jwksSet, err := jwks.JWKS()
		if err != nil {
			log.Error("failed to get JWKS", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		jwksBytes, err := json.Marshal(jwksSet)
		if err != nil {
			log.Error("failed to marshal JWKS", slog.String("error", err.Error()))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Write(jwksBytes)
	}
	mux.HandleFunc("/.well-known/jwks.json", jwksHandler)
	mux.HandleFunc("/.well-known/jwks", jwksHandler)
}
