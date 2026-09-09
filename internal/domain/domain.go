package domain

type Client struct {
	ID         string
	Name       string
	SecretHash string
}

type User struct {
	ID           string
	Username     string
	PasswordHash string
}
