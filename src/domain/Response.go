package domain

import (
	"github.com/Rafael24595/go-api-core/src/domain/body"
	"github.com/Rafael24595/go-api-core/src/domain/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/header"
)

type Response struct {
	Id      string         `json:"_id"`
	Request string         `json:"request"`
	Date    int64          `json:"date"`
	Time    int64          `json:"time"`
	Status  int16          `json:"status"`
	Headers header.Headers `json:"headers"`
	Cookies cookie.Cookies `json:"cookies"`
	Body    body.Body      `json:"body"`
	Size    int            `json:"size"`
	Owner   string         `json:"owner"`
}

func NewResponseDefault() *Response {
	return &Response{}
}

func (r Response) PersistenceId() string {
	return r.Id
}
