package infrastructure

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/log"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	auth_strategy "github.com/Rafael24595/go-api-core/src/domain/action/auth/strategy"
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	body_strategy "github.com/Rafael24595/go-api-core/src/domain/action/body/strategy"
	"github.com/Rafael24595/go-api-core/src/domain/action/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/action/header"
	"github.com/Rafael24595/go-api-core/src/domain/action/query"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"golang.org/x/net/html/charset"
)

type HttpClient struct {
}

func Client() *HttpClient {
	return &HttpClient{}
}

func WarmUp() (*action.Response, error) {
	log.Message("Warming up the HTTP client...")

	start := time.Now().UnixMilli()
	response, result := Client().Fetch(&action.Request{
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

func (c *HttpClient) FetchWithContext(ctx *context.Context, request *action.Request) (*action.Response, error) {
	request = context.ProcessRequest(request, ctx)
	return c.Fetch(request)
}

func (c *HttpClient) Fetch(request *action.Request) (*action.Response, error) {
	req, err := c.makeRequest(request)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	start := time.Now().UnixMilli()
	resp, respErr := client.Do(req)
	end := time.Now().UnixMilli()
	if respErr != nil {
		return nil, fmt.Errorf("cannot execute HTTP request: %s", respErr.Error())
	}

	response, err := c.makeResponse(request.Owner, start, end, request, resp)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *HttpClient) makeRequest(operation *action.Request) (*http.Request, error) {
	method := operation.Method.String()
	uri := strings.TrimSpace(operation.Uri)

	payload := new(bytes.Buffer)
	if !operation.Body.Empty() && operation.Body.Status && method != "GET" && method != "HEAD" {
		strategy := body_strategy.LoadStrategy(operation.Body.ContentType)

		var queries *query.Queries
		payload, queries = strategy(&operation.Body, &operation.Query)

		operation.Query = *queries
	}

	req, err := http.NewRequest(method, uri, payload)
	if err != nil {
		return nil, fmt.Errorf("cannot build the HTTP request: %s", err.Error())
	}

	operation = auth_strategy.ApplyAuth(operation)

	req = c.applyQuery(operation, req)
	req = c.applyHeader(operation, req)
	req = c.applyCookies(operation, req)

	return req, nil
}

func (c *HttpClient) applyQuery(operation *action.Request, req *http.Request) *http.Request {
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

func (c *HttpClient) applyHeader(operation *action.Request, req *http.Request) *http.Request {
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

func (c *HttpClient) applyCookies(operation *action.Request, req *http.Request) *http.Request {
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

func (c *HttpClient) makeResponse(owner string, start int64, end int64, req *action.Request, resp *http.Response) (*action.Response, error) {
	headers := c.makeHeaders(resp)

	cookies, err := c.makeCookies(headers)
	if err != nil {
		return nil, fmt.Errorf("failed to read the cookies: %s", err.Error())
	}

	bodyData, size, err := c.makeBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read the body: %s", err.Error())
	}

	return &action.Response{
		Id:        req.Id,
		Timestamp: end,
		Request:   req.Id,
		Date:      start,
		Time:      end - start,
		Status:    int16(resp.StatusCode),
		Headers:   *headers,
		Cookies:   *cookies,
		Body:      *bodyData,
		Size:      size,
		Owner:     owner,
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

	contentType := body.Text
	if oContentType, ok := body.ContentTypeFromHeader(contentTypeHeader); ok {
		contentType = oContentType
	}

	reader, err := charset.NewReader(resp.Body, contentTypeHeader)
	switch {
	case err == io.EOF:
		return body.EmptyResponseBody(contentType), 0, nil
	case err != nil:
		return nil, 0, err
	}

	bodyResponse, err := io.ReadAll(reader)
	if err != nil {
		return nil, 0, err
	}

	if err := resp.Body.Close(); err != nil {
		return nil, 0, err
	}

	return body.NewResponseBody(contentType, string(bodyResponse)),
		len(bodyResponse), nil
}
