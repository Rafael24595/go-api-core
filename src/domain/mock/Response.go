package mock

import (
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	"github.com/Rafael24595/go-api-core/src/domain/action/header"
)

type Response struct {
	Status  int16             `json:"status"`
	Headers header.Headers    `json:"headers"`
	Body    body.BodyResponse `json:"body"`
}
