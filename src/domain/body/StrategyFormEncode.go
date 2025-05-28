package body

import (
	"bytes"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/domain/query"
)

const (
	FORM_ENCODE_PARAM = "form-encode"
)

func applyFormEncode(b *BodyRequest, q *query.Queries) (*bytes.Buffer, *query.Queries) {
	for k, p := range b.Parameters[FORM_DATA_PARAM] {
		for _, v := range p {
			if v.IsFile {
				log.Warning("Files are not supported in form-encode")
				continue
			}
			if v.Status {
				q.Add(k, query.Query{
					Order:  0,
					Status: true,
					Value:  v.Value,
				})
			}
		}
	}
	return new(bytes.Buffer), q
}

func hasFiles(b *BodyRequest) bool {
	for _, p := range b.Parameters[FORM_DATA_PARAM] {
		for _, v := range p {
			if v.IsFile {
				return true
			}
		}
	}
	return false
}
