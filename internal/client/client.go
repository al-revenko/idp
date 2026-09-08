package client

type ClientWithSecretHash struct {
	ID         string
	Name       string
	SecretHash string
}

type Client struct {
	ID   string
	Name string
}
