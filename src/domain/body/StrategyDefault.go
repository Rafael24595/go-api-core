package body

import (
	"bytes"

	"github.com/Rafael24595/go-api-core/src/domain/query"
)

const (
	DOCUMENT_PARAM = "document"
	PAYLOAD_PARAM = "payload"
)

func applyDefault(b *BodyRequest, q *query.Queries) (*bytes.Buffer, *query.Queries) {
	body := new(bytes.Buffer)

	parameters, ok := b.Parameters[DOCUMENT_PARAM]
	if !ok {
		return body, q
	}

	payload, ok := parameters[PAYLOAD_PARAM]
	if !ok {
		return body, q
	}

	if len(payload) == 0 || payload[0].IsFile {
		return body, q
	}
	
	return bytes.NewBuffer([]byte(payload[0].Value)), q
}
