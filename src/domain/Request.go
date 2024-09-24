package domain

import (
	"github.com/Rafael24595/go-api-core/src/domain/auth"
	"github.com/Rafael24595/go-api-core/src/domain/body"
	"github.com/Rafael24595/go-api-core/src/domain/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/header"
	"github.com/Rafael24595/go-api-core/src/domain/query"
)

type Request struct {
	Id        string         `json:"_id"`
	Timestamp int64          `json:"timestamp"`
	Name      string         `json:"name"`
	Method    HttpMethod     `json:"method"`
	Uri       string         `json:"uri"`
	Queries   query.Queries  `json:"queries"`
	Headers   header.Headers `json:"headers"`
	Cookies   cookie.Cookies `json:"cookies"`
	Body      body.Body      `json:"body"`
	Auth      auth.Auths     `json:"auth"`
	Status    Status         `json:"status"`
}
