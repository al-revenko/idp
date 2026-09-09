package dto

type ClientCreateRequest struct {
	Name string `validate:"required,min=3,max=255"`
}

type ClientDeleteRequest struct {
	ClientId string `validate:"required,uuid"`
}
