package mock

import (
	"github.com/Rafael24595/go-api-core/src/domain/body"
	"github.com/Rafael24595/go-api-core/src/domain/header"
)

type Response struct {
	Status  int16
	Headers header.Headers
	Body    body.BodyResponse
}
