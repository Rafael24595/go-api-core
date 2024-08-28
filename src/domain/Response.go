package domain

import (
	"go-api-core/src/domain/body"
	"go-api-core/src/domain/cookie"
)

type Response struct {
	Date    int64          `json:"date"`
	Time    int64          `json:"time"`
	Status  int16          `json:"status"`
	Headers Headers        `json:"headers"`
	Cookies cookie.Cookies `json:"cookies"`
	Body    body.Body      `json:"body"`
	Size    int64          `json:"size"`
}