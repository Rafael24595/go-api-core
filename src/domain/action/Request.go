package action

import (
	"time"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action/auth"
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	"github.com/Rafael24595/go-api-core/src/domain/action/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/action/header"
	"github.com/Rafael24595/go-api-core/src/domain/action/query"
)

const ANONYMOUS_OWNER = "anonymous"

type Request struct {
	Id        string               `json:"_id"`
	Timestamp int64                `json:"timestamp"`
	Name      string               `json:"name"`
	Method    domain.HttpMethod    `json:"method"`
	Uri       string               `json:"uri"`
	Query     query.Queries        `json:"query"`
	Header    header.Headers       `json:"header"`
	Cookie    cookie.CookiesClient `json:"cookie"`
	Body      body.BodyRequest     `json:"body"`
	Auth      auth.Auths           `json:"auth"`
	Owner     string               `json:"owner"`
	Modified  int64                `json:"modified"`
	Status    StatusRequest        `json:"status"`
}

func NewRequestEmpty() *Request {
	return NewRequest("", domain.GET, "")
}

func NewRequest(name string, method domain.HttpMethod, uri string) *Request {
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
		Cookie: cookie.CookiesClient{
			Cookies: make(map[string]cookie.CookieClient),
		},
		Body: body.BodyRequest{
			ContentType: domain.None,
			Parameters:  make(map[string]map[string][]body.BodyParameter),
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

type RequestLite struct {
	Id        string            `json:"_id"`
	Timestamp int64             `json:"timestamp"`
	Name      string            `json:"name"`
	Method    domain.HttpMethod `json:"method"`
	Uri       string            `json:"uri"`
	Owner     string            `json:"owner"`
	Modified  int64             `json:"modified"`
}

func ToLiteRequest(request *Request) *RequestLite {
	return &RequestLite{
		Id:        request.Id,
		Timestamp: request.Timestamp,
		Name:      request.Name,
		Method:    request.Method,
		Uri:       request.Uri,
		Owner:     request.Owner,
		Modified:  request.Modified,
	}
}
