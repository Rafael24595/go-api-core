package infrastructure

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/body"
	"github.com/Rafael24595/go-api-core/src/domain/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/header"
)

type HttpClient struct {
}

func Client() *HttpClient {
	return &HttpClient{}
}

func WarmUp() (*domain.Response, commons.ApiError) {
	println("Warming up HTTP client...")
	start := time.Now().UnixMilli()
	response, result := Client().Fetch(domain.Request{
		Method: domain.GET,
		Uri:    "https://www.google.es",
	})
	end := time.Now().UnixMilli()
	println(fmt.Sprintf("Client initialized successfully in: %d ms", end-start))
	return response, result
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
	if !operation.Body.Empty() && operation.Body.Status && method != "GET" && method != "HEAD" {
		body = bytes.NewBuffer(operation.Body.Bytes)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, commons.ApiErrorFromCause(500, "Cannot build HTTP request", err)
	}

	req = c.applyQuery(operation, req)
	req = c.applyHeader(operation, req)
	req = c.applyAuth(operation, req)

	return req, nil
}

func (c *HttpClient) applyQuery(operation domain.Request, req *http.Request) *http.Request {
	query := req.URL.Query()
	for k, q := range operation.Query.Queries {
		for _, v := range q {
			if !v.Status {
				continue
			}
			query.Add(k, v.Value)
		}
	}
	req.URL.RawQuery = query.Encode()
	return req
}

func (c *HttpClient) applyHeader(operation domain.Request, req *http.Request) *http.Request {
	headers := map[string][]string{}
	for k, h := range operation.Header.Headers {
		for _, v := range h {
			if !v.Status {
				continue
			}
			if _, ok := headers[k]; !ok {
				headers[k] = make([]string, 0)
			}
			headers[k] = append(headers[k], v.Value)
		}
	}

	req.Header = headers

	return req
}

func (c *HttpClient) applyAuth(operation domain.Request, req *http.Request) *http.Request {
	if !operation.Auth.Status {
		return req
	}
	for _, a := range operation.Auth.Auths {
		if !a.Status {
			continue
		}
		strategy := a.Type.LoadStrategy()
		req = strategy(a, req)
	}
	return req
}

func (c *HttpClient) makeResponse(start int64, end int64, req domain.Request, resp http.Response) (*domain.Response, commons.ApiError) {
	defer resp.Body.Close()

	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, commons.ApiErrorFromCause(500, "Failed to read response", err)
	}

	headers := c.makeHeaders(resp)

	cookies, err := c.makeCookies(headers)
	if err != nil {
		return nil, commons.ApiErrorFromCause(500, "Failed to read response", err)
	}

	contentType := body.Text
	sContentType := resp.Header.Get("content-type")
	if oContentType, ok := body.ContentTypeFromHeader(sContentType); ok {
		contentType = oContentType
	}

	bodyData := body.Body{
		Status:      true,
		ContentType: contentType,
		Bytes:       bodyResponse,
	}

	return &domain.Response{
		Request: req.Id,
		Date:    start,
		Time:    end - start,
		Status:  int16(resp.StatusCode),
		Headers: *headers,
		Cookies: *cookies,
		Body:    bodyData,
		Size:    len(bodyResponse),
	}, nil
}

func (c *HttpClient) makeHeaders(resp http.Response) *header.Headers {
	headersResponse := map[string][]header.Header{}
	for k, h := range resp.Header {
		if _, ok := headersResponse[k]; !ok {
			headersResponse[k] = make([]header.Header, 0)
		}
		for _, v := range h {
			headersResponse[k] = append(headersResponse[k], header.Header{
				Status: true,
				Value:  v,
			})
		}
	}

	return &header.Headers{
		Headers: headersResponse,
	}
}

func (c *HttpClient) makeCookies(headers *header.Headers) (*cookie.Cookies, error) {
	setCookie, ok := headers.Headers["Set-Cookie"]
	if !ok && len(setCookie) > 0 {
		return &cookie.Cookies{
			Cookies: make(map[string]cookie.Cookie),
		}, nil
	}

	cookies := map[string]cookie.Cookie{}
	for _, c := range setCookie {
		parsed, err := cookie.CookieFromString(c.Value)
		if err != nil {
			return nil, err
		}
		cookies[parsed.Code] = *parsed
	}

	return &cookie.Cookies{
		Cookies: cookies,
	}, nil
}
