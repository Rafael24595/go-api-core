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
	"github.com/Rafael24595/go-collections/collection"
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
		Uri: "https://www.google.es",
	})
	end := time.Now().UnixMilli()
	println(fmt.Sprintf("Client initialized successfully in: %d ms", end - start))
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
	if !operation.Body.Empty() && method != "GET" && method != "HEAD" {
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
	for k, q := range operation.Queries.Queries {
		if !q.Active {
			continue
		}
		for _, v := range q.Query {
			query.Add(k, v)
		}
	}
	req.URL.RawQuery = query.Encode()
	return req
}

func (c *HttpClient) applyHeader(operation domain.Request, req *http.Request) *http.Request {
	filtered := collection.DictionaryFromMap(operation.Headers.Headers).
		FilterSelf(func(s string, h header.Header) bool {
			return h.Active
	})

	req.Header = collection.DictionaryMap(filtered, func(key string, value header.Header) []string {
		return value.Header
	}).Collect()
	
	return req
}

func (c *HttpClient) applyAuth(operation domain.Request, req *http.Request) *http.Request {
	if !operation.Auths.Status {
		return req
	}
	for _, a := range operation.Auths.Auths {
		if !a.Active {
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

	headersResponse := collection.DictionaryMap(collection.DictionaryFromMap(resp.Header), func(key string, value []string) header.Header {
		return header.Header{
			Active: true,
			Key:    key,
			Header: value,
		}
	}).Collect()

	headers := header.Headers{
		Headers: headersResponse,
	}

	cookies := cookie.Cookies{
		Cookies: make(map[string]cookie.Cookie),
	}

	if setCookie, ok := headers.Headers["Set-Cookie"]; ok && len(setCookie.Header) > 0 {
		for _, c := range setCookie.Header {
			parsed, err := cookie.CookieFromString(c)
			if err != nil {
				return nil, err
			}
			cookies.Cookies[parsed.Code] = *parsed
		}
	}

	contentType := body.Text
	sContentType := resp.Header.Get("content-type")
	if oContentType, ok := body.ContentTypeFromHeader(sContentType); ok {
		contentType = oContentType
	}

	bodyData := body.Body{
		ContentType: contentType,
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
