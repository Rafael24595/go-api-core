package domain

import (
	"fmt"
	"strings"

	"github.com/Rafael24595/go-api-core/src/commons"
)

type HttpMethod string

const (
	GET     HttpMethod = "GET"
	POST    HttpMethod = "POST"
	PUT     HttpMethod = "PUT"
	DELETE  HttpMethod = "DELETE"
	PATCH   HttpMethod = "PATCH"
	HEAD    HttpMethod = "HEAD"
	OPTIONS HttpMethod = "OPTIONS"
	TRACE   HttpMethod = "TRACE"
	CONNECT HttpMethod = "CONNECT"
)

func (m HttpMethod) String() string {
	return string(m)
}

func HttpMethodFromString(value string) (*HttpMethod, commons.ApiError) {
	switch strings.ToUpper(value) {
	case string(GET):
		GET := GET
		return &GET, nil
	case string(POST):
		POST := POST
		return &POST, nil
	case string(PUT):
		PUT := PUT
		return &PUT, nil
	case string(DELETE):
		DELETE := DELETE
		return &DELETE, nil
	case string(PATCH):
		PATCH := PATCH
		return &PATCH, nil
	case string(HEAD):
		HEAD := HEAD
		return &HEAD, nil
	case string(OPTIONS):
		OPTIONS := OPTIONS
		return &OPTIONS, nil
	case string(TRACE):
		TRACE := TRACE
		return &TRACE, nil
	case string(CONNECT):
		CONNECT := CONNECT
		return &CONNECT, nil
	default:
		return nil, commons.ApiErrorFrom(422, fmt.Sprintf("Unknown method value: '%s'", value))
	}
}

func HttpMethods() []HttpMethod {
	return []HttpMethod{
		GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, TRACE, CONNECT,
	}
}