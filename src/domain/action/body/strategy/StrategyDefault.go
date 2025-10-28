package body_strategy

import (
	"bytes"

	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	"github.com/Rafael24595/go-api-core/src/domain/action/query"
)

const (
	DOCUMENT_PARAM = "document"
	PAYLOAD_PARAM  = "payload"
)

func DocumentBody(status bool, contentType body.ContentType, document string) *body.BodyRequest {
	parameters := make(map[string]map[string][]body.BodyParameter)

	parameters[DOCUMENT_PARAM] = make(map[string][]body.BodyParameter)
	parameters[DOCUMENT_PARAM][PAYLOAD_PARAM] = []body.BodyParameter{
		body.NewBodyDocument(0, true, document),
	}

	return body.NewBody(status, contentType, parameters)
}

func applyDefault(b *body.BodyRequest, q *query.Queries) (*bytes.Buffer, *query.Queries) {
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
