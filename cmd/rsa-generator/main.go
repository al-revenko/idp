package main

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"github.com/al-revenko/idp/internal/lib/crypt/rsa"
)

func main() {
	priv, pub, err := rsa.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	privDerBytes := x509.MarshalPKCS1PrivateKey(priv)
	pubDerBytes := x509.MarshalPKCS1PublicKey(pub)

	fmt.Print("\nPrivate key: ", base64.StdEncoding.EncodeToString(privDerBytes))
	fmt.Print("\n\nPublic key: ", base64.StdEncoding.EncodeToString(pubDerBytes), "\n\n")
}
