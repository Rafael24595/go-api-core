package dto

import (
	"github.com/Rafael24595/go-api-core/src/domain/body"
)

type DtoBody struct {
	Status      bool                          `json:"status"`
	ContentType body.ContentType              `json:"content_type"`
	Parameters  map[string]body.BodyParameter `json:"parameters"`
}

func ToBody(dto *DtoBody) *body.Body {
	return &body.Body{
		Status:      dto.Status,
		ContentType: dto.ContentType,
		Parameters:  dto.Parameters,
	}
}

func FromBody(body *body.Body) *DtoBody {
	return &DtoBody{
		Status:      body.Status,
		ContentType: body.ContentType,
		Parameters:  body.Parameters,
	}
}
