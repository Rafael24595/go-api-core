package body_strategy

import (
	"bytes"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	"github.com/Rafael24595/go-api-core/src/domain/action/query"
)

const (
	FORM_ENCODE_PARAM = "form-encode"
)

func applyFormEncode(b *body.BodyRequest, q *query.Queries) (*bytes.Buffer, *query.Queries) {
	for k, p := range b.Parameters[FORM_DATA_PARAM] {
		for _, v := range p {
			if v.IsFile {
				log.Warning("Files are not supported in form-encode")
				continue
			}
			if v.Status {
				q.AddQuery(k, query.Query{
					Order:  0,
					Status: true,
					Value:  v.Value,
				})
			}
		}
	}
	return new(bytes.Buffer), q
}

func hasFiles(b *body.BodyRequest) bool {
	for _, p := range b.Parameters[FORM_DATA_PARAM] {
		for _, v := range p {
			if v.IsFile {
				return true
			}
		}
	}
	return false
}
