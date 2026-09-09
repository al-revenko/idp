package dto

type ClientRegisterRequest struct {
	Name string `validate:"required,min=3,max=255"`
}

type ClientDeleteRequest struct {
	ClientId          string `validate:"required,uuid"`
	ClientSecretToken string `validate:"required"`
}
