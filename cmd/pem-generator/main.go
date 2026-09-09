package main

import (
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"

	"github.com/al-revenko/idp/internal/lib/crypt/rsa"
)

func main() {
	var pemFileName string
	flag.StringVar(&pemFileName, "name", "private_key.pem", "name of pem file")
	flag.Parse()

	privateKey, err := rsa.GenerateKey()

	derBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		panic(err)
	}

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: derBytes,
	}

	file, err := os.Create(fmt.Sprintf("certs/%s", pemFileName))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = pem.Encode(file, pemBlock)
	if err != nil {
		panic(err)
	}
}
