package body

import (
	"bytes"
)

const (
	DOCUMENT_PARAM = "$DOCUMENT"
)

func applyDefault(b *Body) *bytes.Buffer {
	var body *bytes.Buffer
	parameter, ok := b.Parameters[DOCUMENT_PARAM]
	if !ok || parameter.IsFile {
		return body
	}
	return bytes.NewBuffer([]byte(parameter.Value))
}
