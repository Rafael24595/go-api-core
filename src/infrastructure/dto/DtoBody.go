package dto

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
)

type DtoBody struct {
	Status      bool                                       `json:"status"`
	ContentType domain.ContentType                         `json:"content_type"`
	Parameters  map[string]map[string][]body.BodyParameter `json:"parameters"`
}

func ToBody(dto *DtoBody) *body.BodyRequest {
	return &body.BodyRequest{
		Status:      dto.Status,
		ContentType: dto.ContentType,
		Parameters:  dto.Parameters,
	}
}

func FromBody(body *body.BodyRequest) *DtoBody {
	return &DtoBody{
		Status:      body.Status,
		ContentType: body.ContentType,
		Parameters:  body.Parameters,
	}
}
