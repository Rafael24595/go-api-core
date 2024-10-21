package domain

import (
	"time"

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
	Auths     auth.Auths     `json:"auth"`
	Status    Status         `json:"status"`
}

func NewRequestDefault() *Request {
	return &Request{}
}

func NewRequestEmpty() *Request {
	return NewRequest("", GET, "")
}

func NewRequest(name string, method HttpMethod, uri string) *Request {
	return &Request{
		Id:        "",
		Timestamp: time.Now().UnixMilli(),
		Name:      name,
		Method:    method,
		Uri:       uri,
		Queries: query.Queries{
			Queries: make(map[string]query.Query),
		},
		Headers: header.Headers{
			Headers: make(map[string]header.Header),
		},
		Cookies: cookie.Cookies{
			Cookies: make(map[string]cookie.Cookie),
		},
		Body: body.Body{
			ContentType: body.None,
			Bytes:       make([]byte, 0),
		},
		Auths: auth.Auths{
			Auths: make(map[string]auth.Auth),
		},
		Status: Historic,
	}
}

func (r Request) PersistenceId() string {
	return r.Id
}
