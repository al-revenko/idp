package dto

type ClientCreateRequest struct {
	Name      string `validate:"required,min=3,max=255"`
	PubKeyUrl string `validate:"required,http_url"`
}
