package dto

import (
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/header"
)

type DtoResponse struct {
	Id      string               `json:"_id"`
	Request string               `json:"request"`
	Date    int64                `json:"date"`
	Time    int64                `json:"time"`
	Status  int16                `json:"status"`
	Headers header.Headers       `json:"headers"`
	Cookies cookie.CookiesServer `json:"cookies"`
	Body    DtoBody              `json:"body"`
	Size    int                  `json:"size"`
	Owner   string               `json:"owner"`
}

func ToResponse(dto *DtoResponse) *domain.Response {
	return &domain.Response{
		Id:      dto.Id,
		Request: dto.Request,
		Date:    dto.Date,
		Time:    dto.Time,
		Status:  dto.Status,
		Headers: dto.Headers,
		Cookies: dto.Cookies,
		Body:    *ToBody(&dto.Body),
		Size:    dto.Size,
		Owner:   dto.Owner,
	}
}

func FromResponse(request *domain.Response) *DtoResponse {
	return &DtoResponse{
		Id:      request.Id,
		Request: request.Request,
		Date:    request.Date,
		Time:    request.Time,
		Status:  request.Status,
		Headers: request.Headers,
		Cookies: request.Cookies,
		Body:    *FromBody(&request.Body),
		Size:    request.Size,
		Owner:   request.Owner,
	}
}
