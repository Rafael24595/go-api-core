package action

import (
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	"github.com/Rafael24595/go-api-core/src/domain/action/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/action/header"
)

type Response struct {
	Id        string               `json:"_id"`
	Timestamp int64                `json:"timestamp"`
	Request   string               `json:"request"`
	Date      int64                `json:"date"`
	Time      int64                `json:"time"`
	Status    int16                `json:"status"`
	Headers   header.Headers       `json:"headers"`
	Cookies   cookie.CookiesServer `json:"cookies"`
	Body      body.BodyResponse    `json:"body"`
	Size      int                  `json:"size"`
	Owner     string               `json:"owner"`
}

func NewResponseDefault() *Response {
	return &Response{}
}

func (r Response) PersistenceId() string {
	return r.Id
}
