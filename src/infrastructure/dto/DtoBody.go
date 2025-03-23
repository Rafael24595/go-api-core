package dto

import (
	"github.com/Rafael24595/go-api-core/src/domain/body"
)

type DtoBody struct {
	Status      bool             `json:"status"`
	ContentType body.ContentType `json:"content_type"`
	Payload     string           `json:"payload"`
}

func ToBody(dto *DtoBody) *body.Body {
	return &body.Body{
		Status:      dto.Status,
		ContentType: dto.ContentType,
		Payload:     []byte(dto.Payload),
	}
}

func FromBody(body *body.Body) *DtoBody {
	return &DtoBody{
		Status:      body.Status,
		ContentType: body.ContentType,
		Payload:     string(body.Payload),
	}
}
