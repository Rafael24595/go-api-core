package infrastructure

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/body"
	"github.com/Rafael24595/go-api-core/src/domain/cookie"
)

type HttpClient struct {
}

func Client() *HttpClient {
	return &HttpClient{}
}

func (c *HttpClient) Fetch(request domain.Request) (*domain.Response, commons.ApiError) {
	req, err := c.makeRequest(request)
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

	response, err := c.makeResponse(start, end, request, *resp)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *HttpClient) makeRequest(operation domain.Request) (*http.Request, commons.ApiError) {
	method := operation.Method.String()
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

func (c *HttpClient) makeResponse(start int64, end int64, req domain.Request, resp http.Response) (*domain.Response, commons.ApiError) {
	defer resp.Body.Close()

	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, commons.ApiErrorFromCause(500, "Failed to read response", err)
	}

	headers := domain.Headers{
		Headers: resp.Header,
	}

	cookies := cookie.Cookies{
		Cookies: make(map[string]cookie.Cookie),
	}

	setCookie := headers.Headers["Set-Cookie"]
	if len(setCookie) > 0 {
		for _, c := range setCookie {
			parsed, err := cookie.CookieFromString(c)
			if err != nil {
				return nil, err
			}
			cookies.Cookies[parsed.Code] = *parsed
		}
	}

	bodyData := body.Body{
		ContentType: body.None,
		Bytes:       bodyResponse,
	}

	return &domain.Response{
		Request: req.Id,
		Date:    start,
		Time:    end - start,
		Status:  int16(resp.StatusCode),
		Headers: headers,
		Cookies: cookies,
		Body:    bodyData,
		Size:    len(bodyResponse),
	}, nil
}
