package app

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/lestrrat-go/jwx/v4/jwk"
)

type JWKSProvider interface {
	JWKS() (jwk.Set, error)
}

func registerHTTP(mux *http.ServeMux, jwks JWKSProvider, log *slog.Logger) {
	mux.HandleFunc("/.well-known/jwks", func(w http.ResponseWriter, r *http.Request) {
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
	})
}
