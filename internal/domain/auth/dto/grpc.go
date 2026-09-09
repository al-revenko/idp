package dto

type RegisterRequest struct {
	Username string `validate:"required,max=50,alphanum"`
	Password string `validate:"required,max=128,password"`
}

type LoginRequest struct {
	Username string `validate:"required"`
	Password string `validate:"required"`
}
