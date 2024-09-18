package domain

import (
	"github.com/Rafael24595/go-api-core/src/domain/body"
	"github.com/Rafael24595/go-api-core/src/domain/cookie"
)

type Request struct {
	Id        string         `json:"_id"`
	Timestamp int64          `json:"timestamp"`
	Name      string         `json:"name"`
	Method    HttpMethod     `json:"method"`
	Uri       string         `json:"uri"`
	Headers   Headers        `json:"headers"`
	Cookies   cookie.Cookies `json:"cookies"`
	Body      body.Body      `json:"body"`
	Status    Status         `json:"status"`
}
