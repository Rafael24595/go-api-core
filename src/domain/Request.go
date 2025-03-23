package domain

import (
	"time"

	"github.com/Rafael24595/go-api-core/src/domain/auth"
	"github.com/Rafael24595/go-api-core/src/domain/body"
	"github.com/Rafael24595/go-api-core/src/domain/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/header"
	"github.com/Rafael24595/go-api-core/src/domain/query"
)

const ANONYMOUS_OWNER = "anonymous"

type Request struct {
	Id        string         `json:"_id"`
	Timestamp int64          `json:"timestamp"`
	Name      string         `json:"name"`
	Method    HttpMethod     `json:"method"`
	Uri       string         `json:"uri"`
	Query     query.Queries  `json:"query"`
	Header    header.Headers `json:"header"`
	Cookie    cookie.Cookies `json:"cookie"`
	Body      body.Body      `json:"body"`
	Auth      auth.Auths     `json:"auth"`
	Owner     string         `json:"owner"`
	Modified  int64          `json:"modified"`
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
		Query: query.Queries{
			Queries: make(map[string][]query.Query),
		},
		Header: header.Headers{
			Headers: make(map[string][]header.Header),
		},
		Cookie: cookie.Cookies{
			Cookies: make(map[string]cookie.Cookie),
		},
		Body: body.Body{
			ContentType: body.None,
			Payload:     make([]byte, 0),
		},
		Auth: auth.Auths{
			Auths: make(map[string]auth.Auth),
		},
		Owner:    ANONYMOUS_OWNER,
		Modified: time.Now().UnixMilli(),
		Status:   DRAFT,
	}
}

func (r Request) PersistenceId() string {
	return r.Id
}
