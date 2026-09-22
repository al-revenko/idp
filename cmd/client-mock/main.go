package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"uuid"

	"github.com/al-revenko/idp/internal/lib/valid"
	"github.com/joho/godotenv"
	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jwt"
)

const (
	jwksPort = 8081
	keyID    = "jwk-keyid"
)

func main() {
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}

	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	audience := []string{os.Getenv("SERVICE_NAME")}
	if audience[0] == "" {
		return fmt.Errorf("SERVICE_NAME environment variable is not set")
	}

	v := valid.New()

	flag.Parse()
	if flag.NArg() < 1 {
		panic("clientId argument is required: go run ./cmd/jwks <clientId>")
	}
	clientID := flag.Arg(0)
	if err := v.UUID(clientID); err != nil {
		return fmt.Errorf("clientId must be UUID: %s", clientID)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	privateJWK, err := jwk.Import[jwk.RSAPrivateKey](privateKey)
	if err != nil {
		return fmt.Errorf("failed to create JWK from private key: %w", err)
	}

	privateJWK.Set(jwk.KeyIDKey, keyID)

	publicJWK, err := privateJWK.PublicKey()
	if err != nil {
		return fmt.Errorf("failed to create JWK from public key: %w", err)
	}

	publicJWK.Set(jwk.AlgorithmKey, jwa.RS256())

	jwks := jwk.NewSet()
	err = jwks.AddKey(publicJWK)
	if err != nil {
		return fmt.Errorf("failed to add JWK to set: %w", err)
	}

	jwksBytes, _ := json.Marshal(jwks)

	http.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/jwk-set+json")
		w.Write(jwksBytes)
	})

	http.HandleFunc("/bearer", func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		token, err := jwt.NewBuilder().
			Issuer(clientID).
			Subject(clientID).
			Expiration(now.Add(1 * time.Hour)).
			IssuedAt(now).
			JwtID(uuid.NewV7().String()).
			Audience(audience).
			Build()
		if err != nil {
			http.Error(w, "failed to build token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.RS256(), privateJWK))
		if err != nil {
			http.Error(w, "failed to sign token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tokenStr := string(signedToken)
		creds := clientID + ":" + tokenStr
		encodedCreds := base64.StdEncoding.EncodeToString([]byte(creds))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"Bearer": encodedCreds,
		})
	})

	addr := fmt.Sprintf(":%d", jwksPort)
	fmt.Printf("Mock JWKS server starting with clientId %s\n\n", clientID)
	fmt.Printf("JWKS:   http://localhost%s/.well-known/jwks.json\n", addr)
	fmt.Printf("Bearer: http://localhost%s/bearer\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	return nil
}
