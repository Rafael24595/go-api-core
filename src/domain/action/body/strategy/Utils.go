package body_strategy

import (
	"bytes"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	"github.com/Rafael24595/go-api-core/src/domain/action/query"
)

func LoadStrategy(typ domain.ContentType) func(a *body.BodyRequest, q *query.Queries) (*bytes.Buffer, *query.Queries) {
	switch typ {
	case domain.Form:
		return applyFormData
	default:
		return applyDefault
	}
}
