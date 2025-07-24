package infrastructure

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/exception"
	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/body"
	"github.com/Rafael24595/go-api-core/src/domain/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/header"

	"golang.org/x/net/html/charset"
)

type HttpClient struct {
}

func Client() *HttpClient {
	return &HttpClient{}
}

func WarmUp() (*domain.Response, error) {
	log.Message("Warming up the HTTP client...")

	start := time.Now().UnixMilli()
	response, result := Client().Fetch(domain.Request{
		Method: domain.GET,
		Uri:    "https://www.google.es",
	})

	if result != nil {
		return nil, result
	}

	end := time.Now().UnixMilli()
	log.Messagef("The client has been initialized successfully in: %d ms", end-start)
	return response, nil
}

func (c *HttpClient) Fetch(request domain.Request) (*domain.Response, *exception.ApiError) {
	req, err := c.makeRequest(request)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	start := time.Now().UnixMilli()
	resp, respErr := client.Do(req)
	end := time.Now().UnixMilli()
	if respErr != nil {
		return nil, exception.NewCauseApiError(500, "Cannot execute HTTP request", respErr)
	}

	response, err := c.makeResponse(request.Owner, start, end, request, resp)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *HttpClient) makeRequest(operation domain.Request) (*http.Request, *exception.ApiError) {
	method := operation.Method.String()
	url := strings.TrimSpace(operation.Uri)

	body := new(bytes.Buffer)
	if !operation.Body.Empty() && operation.Body.Status && method != "GET" && method != "HEAD" {
		strategy := operation.Body.ContentType.LoadStrategy()
		body, _ = strategy(&operation.Body, &operation.Query)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, exception.NewCauseApiError(http.StatusUnprocessableEntity, "Cannot build the HTTP request", err)
	}

	req = c.applyQuery(operation, req)
	req = c.applyHeader(operation, req)
	req = c.applyCookies(operation, req)
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

func (c *HttpClient) applyCookies(operation domain.Request, req *http.Request) *http.Request {
	cookies := []string{}
	for k, c := range operation.Cookie.Cookies {
		if !c.Status {
			continue
		}
		cookies = append(cookies, fmt.Sprintf("%s=%s", k, c.Value))
	}

	req.Header["Cookie"] = []string{
		strings.Join(cookies, "; "),
	}

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

func (c *HttpClient) makeResponse(owner string, start int64, end int64, req domain.Request, resp *http.Response) (*domain.Response, *exception.ApiError) {
	headers := c.makeHeaders(resp)

	cookies, err := c.makeCookies(headers)
	if err != nil {
		return nil, exception.NewCauseApiError(http.StatusInternalServerError, "Failed to read the cookies", err)
	}

	bodyData, size, err := c.makeBody(resp)
	if err != nil {
		return nil, exception.NewCauseApiError(http.StatusInternalServerError, "Failed to read the cookies", err)
	}

	return &domain.Response{
		Id:      req.Id,
		Request: req.Id,
		Date:    start,
		Time:    end - start,
		Status:  int16(resp.StatusCode),
		Headers: *headers,
		Cookies: *cookies,
		Body:    *bodyData,
		Size:    size,
		Owner:   owner,
	}, nil
}

func (c *HttpClient) makeHeaders(resp *http.Response) *header.Headers {
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

func (c *HttpClient) makeCookies(headers *header.Headers) (*cookie.CookiesServer, error) {
	setCookie, ok := headers.Headers["Set-Cookie"]
	if !ok && len(setCookie) > 0 {
		return &cookie.CookiesServer{
			Cookies: make(map[string]cookie.CookieServer),
		}, nil
	}

	cookies := map[string]cookie.CookieServer{}
	for _, c := range setCookie {
		parsed, err := cookie.CookieServerFromString(c.Value)
		if err != nil {
			return nil, err
		}
		cookies[parsed.Code] = *parsed
	}

	return &cookie.CookiesServer{
		Cookies: cookies,
	}, nil
}

func (c *HttpClient) makeBody(resp *http.Response) (*body.BodyResponse, int, error) {
	contentTypeHeader := resp.Header.Get("Content-Type")

	reader, err := charset.NewReader(resp.Body, contentTypeHeader)
	if err != nil {
		return nil, 0, err
	}

	bodyResponse, err := io.ReadAll(reader)
	if err != nil {
		return nil, 0, err
	}

	contentType := body.Text
	if oContentType, ok := body.ContentTypeFromHeader(contentTypeHeader); ok {
		contentType = oContentType
	}

	if err := resp.Body.Close(); err != nil {
		return nil, 0, err
	}

	return &body.BodyResponse{
		Status:      true,
		ContentType: contentType,
		Payload:     string(bodyResponse),
	}, len(bodyResponse), nil
}
