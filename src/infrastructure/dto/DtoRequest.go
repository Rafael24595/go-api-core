package dto

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/auth"
	"github.com/Rafael24595/go-api-core/src/domain/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/header"
	"github.com/Rafael24595/go-api-core/src/domain/query"
)

const ANONYMOUS_OWNER = "anonymous"

type DtoRequest struct {
	Id        string            `json:"_id"`
	Timestamp int64             `json:"timestamp"`
	Name      string            `json:"name"`
	Method    domain.HttpMethod `json:"method"`
	Uri       string            `json:"uri"`
	Query     query.Queries     `json:"query"`
	Header    header.Headers    `json:"header"`
	Cookie    cookie.Cookies    `json:"cookie"`
	Body      DtoBody           `json:"body"`
	Auth      auth.Auths        `json:"auth"`
	Owner     string            `json:"owner"`
	Modified  int64             `json:"modified"`
	Status    domain.Status     `json:"status"`
}

func ToRequest(dto *DtoRequest) *domain.Request {
	return &domain.Request{
		Id: dto.Id,
		Timestamp: dto.Timestamp,
		Name: dto.Name,
		Method: dto.Method,
		Uri: dto.Uri,
		Query: dto.Query,
		Header: dto.Header,
		Cookie: dto.Cookie,
		Body: *ToBody(&dto.Body),
		Auth: dto.Auth,
		Owner: dto.Owner,
		Modified: dto.Modified,
		Status: dto.Status,
	}
}

func FromRequest(request *domain.Request) *DtoRequest {
	return &DtoRequest{
		Id: request.Id,
		Timestamp: request.Timestamp,
		Name: request.Name,
		Method: request.Method,
		Uri: request.Uri,
		Query: request.Query,
		Header: request.Header,
		Cookie: request.Cookie,
		Body: *FromBody(&request.Body),
		Auth: request.Auth,
		Owner: request.Owner,
		Modified: request.Modified,
		Status: request.Status,
	}
}
