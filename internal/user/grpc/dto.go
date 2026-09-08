package grpc

type UserRegisterRequest struct {
	Username string `validate:"required,max=50,alphanum"`
	Password string `validate:"required,min=8,max=128"`
}

type UserRegisterResponse struct {
	UserId string `validate:"required"`
}

type UserLoginRequest struct {
	ClientId string `validate:"required,uuid"`
	Username string `validate:"required,max=50,alphanum"`
	Password string `validate:"required,min=8,max=128"`
}
