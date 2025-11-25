package dto

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/domain/action/auth"
	"github.com/Rafael24595/go-api-core/src/domain/action/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/action/header"
	"github.com/Rafael24595/go-api-core/src/domain/action/query"
)

const ANONYMOUS_OWNER = "anonymous"

type DtoRequest struct {
	Id        string               `json:"_id"`
	Timestamp int64                `json:"timestamp"`
	Name      string               `json:"name"`
	Method    domain.HttpMethod    `json:"method"`
	Uri       string               `json:"uri"`
	Query     query.Queries        `json:"query"`
	Header    header.Headers       `json:"header"`
	Cookie    cookie.CookiesClient `json:"cookie"`
	Body      DtoBody              `json:"body"`
	Auth      auth.Auths           `json:"auth"`
	Owner     string               `json:"owner"`
	Modified  int64                `json:"modified"`
	Status    action.StatusRequest `json:"status"`
}

func ToRequests(dtos ...DtoRequest) []action.Request {
	reqs := make([]action.Request, len(dtos))
	for i, v := range dtos {
		reqs[i] = *ToRequest(&v)
	}
	return reqs
}

func ToRequest(dto *DtoRequest) *action.Request {
	return &action.Request{
		Id:        dto.Id,
		Timestamp: dto.Timestamp,
		Name:      dto.Name,
		Method:    dto.Method,
		Uri:       dto.Uri,
		Query:     dto.Query,
		Header:    dto.Header,
		Cookie:    dto.Cookie,
		Body:      *ToBody(&dto.Body),
		Auth:      dto.Auth,
		Owner:     dto.Owner,
		Modified:  dto.Modified,
		Status:    dto.Status,
	}
}

func FromRequests(reqs ...action.Request) []DtoRequest {
	dtos := make([]DtoRequest, len(reqs))
	for i, v := range reqs {
		dtos[i] = *FromRequest(&v)
	}
	return dtos
}

func FromRequest(request *action.Request) *DtoRequest {
	return &DtoRequest{
		Id:        request.Id,
		Timestamp: request.Timestamp,
		Name:      request.Name,
		Method:    request.Method,
		Uri:       request.Uri,
		Query:     request.Query,
		Header:    request.Header,
		Cookie:    request.Cookie,
		Body:      *FromBody(&request.Body),
		Auth:      request.Auth,
		Owner:     request.Owner,
		Modified:  request.Modified,
		Status:    request.Status,
	}
}

type DtoLiteRequest struct {
	Id        string            `json:"_id"`
	Timestamp int64             `json:"timestamp"`
	Name      string            `json:"name"`
	Method    domain.HttpMethod `json:"method"`
	Uri       string            `json:"uri"`
	Owner     string            `json:"owner"`
	Modified  int64             `json:"modified"`
}

func ToLiteRequest(request *action.Request) *DtoLiteRequest {
	return &DtoLiteRequest{
		Id:        request.Id,
		Timestamp: request.Timestamp,
		Name:      request.Name,
		Method:    request.Method,
		Uri:       request.Uri,
		Owner:     request.Owner,
		Modified:  request.Modified,
	}
}
