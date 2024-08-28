package domain

import (
	"go-api/src/domain/body"
	"go-api/src/domain/cookie"
)

type Operation struct {
	Id      string         `json:"_id"`
	Name    string         `json:"name"`
	Method  string         `json:"method"`
	Uri     string         `json:"uri"`
	Headers Headers        `json:"headers"`
	Cookies cookie.Cookies `json:"cookies"`
	Body    body.Body      `json:"body"`
}