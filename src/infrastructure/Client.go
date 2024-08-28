package infrastructure

import (
	"bytes"
	"go-api-core/src/commons"
	"go-api-core/src/domain"
	"go-api-core/src/domain/body"
	"go-api-core/src/domain/cookie"
	"io"
	"net/http"
	"strings"
	"time"
)

func Fetch(request domain.Request) (*domain.Response, commons.ApiError) {
	req, err := makeRequest(request)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	
	start := time.Now().UnixMilli()
    resp, err := client.Do(req)
	end := time.Now().UnixMilli()
    if err != nil {
        return nil, commons.ApiErrorFromCause(500, "Cannot execute HTTP request", err)
    }

	response, err := makeResponse(start, end, resp)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func makeRequest(operation domain.Request) (*http.Request, commons.ApiError) {
	method := strings.ToUpper(operation.Method)
	url := operation.Uri

	var body io.Reader
	if !operation.Body.Empty() && method != "GET" && method != "HEAD" {
		body = bytes.NewBuffer(operation.Body.Bytes)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, commons.ApiErrorFromCause(500, "Cannot build HTTP request", err)
	}
	
	return req, nil
}

func makeResponse(start int64, end int64, resp *http.Response) (*domain.Response, commons.ApiError) {
	defer resp.Body.Close()

    bodyResponse, err := io.ReadAll(resp.Body)
    if err != nil {
		return nil, commons.ApiErrorFromCause(500, "Failed to read response", err)
    }

	headers := domain.Headers{}
	cookies := cookie.Cookies{}

	bodyData := body.Body{
		ContentType: body.None,
		Bytes: bodyResponse,
	}

	return &domain.Response{
		Date: start,
		Time: end - start,
		Headers: headers,
		Cookies: cookies,
		Body: bodyData,
	}, nil
}