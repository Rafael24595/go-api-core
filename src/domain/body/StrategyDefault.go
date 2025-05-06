package body

import (
	"bytes"
)

const (
	DOCUMENT_PARAM = "document"
	PAYLOAD_PARAM = "payload"
)

func applyDefault(b *BodyRequest) *bytes.Buffer {
	body := new(bytes.Buffer)

	parameters, ok := b.Parameters[DOCUMENT_PARAM]
	if !ok {
		return body
	}

	payload, ok := parameters[PAYLOAD_PARAM]
	if !ok {
		return body
	}

	if len(payload) == 0 || payload[0].IsFile {
		return body
	}
	
	return bytes.NewBuffer([]byte(payload[0].Value))
}
