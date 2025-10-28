package context

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/domain/action/auth"
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	"github.com/Rafael24595/go-api-core/src/domain/action/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/action/header"
	"github.com/Rafael24595/go-api-core/src/domain/action/query"
	"github.com/Rafael24595/go-collections/collection"
)

type DictionaryVariables = collection.Dictionary[string, ItemContext]
type DictionaryCategory = collection.Dictionary[string, DictionaryVariables]

type ContextCategoy string

const (
	URI     ContextCategoy = "uri"
	QUERY   ContextCategoy = "query"
	HEADER  ContextCategoy = "header"
	COOKIE  ContextCategoy = "cookie"
	PAYLOAD ContextCategoy = "payload"
	AUTH    ContextCategoy = "auth"
)

func (s ContextCategoy) String() string {
	return string(s)
}

type Context struct {
	Id         string             `json:"_id"`
	Status     bool               `json:"status"`
	Timestamp  int64              `json:"timestamp"`
	Dictionary DictionaryCategory `json:"dictionary"`
	Owner      string             `json:"owner"`
	Collection string             `json:"collection"`
	Modified   int64              `json:"modified"`
}

func NewContext(owner string) *Context {
	return &Context{
		Id:         "",
		Status:     true,
		Timestamp:  time.Now().UnixMilli(),
		Dictionary: *collection.DictionaryEmpty[string, DictionaryVariables](),
		Owner:      owner,
		Collection: "",
		Modified:   time.Now().UnixMilli(),
	}
}

func (c *Context) Put(category ContextCategoy, key, value string, private bool) *Context {
	variables, ok := c.Dictionary.Get(category.String())
	if !ok {
		c.Dictionary.Put(category.String(), *collection.DictionaryEmpty[string, ItemContext]())
		variables, _ = c.Dictionary.Get(category.String())
	}
	variables.Put(key, ItemContext{
		Order:   int64(variables.Size()),
		Private: private,
		Status:  true,
		Value:   value,
	})
	return c
}

func (c *Context) PutAll(category string, context map[string]ItemContext) *Context {
	variables, ok := c.Dictionary.Get(category)
	if !ok {
		c.Dictionary.Put(category, *collection.DictionaryEmpty[string, ItemContext]())
		variables, _ = c.Dictionary.Get(category)
	}
	variables.PutAll(context)
	return c
}

func (c Context) Apply(category, source string) string {
	fix := source
	for _, v := range c.IdentifyVariables(category, source) {
		value := ""
		categoryGroup, ok := c.Dictionary.Get(v.Key())
		if ok {
			keyValue, ok := categoryGroup.Get(v.Value())
			if keyValue.Status && ok {
				value = keyValue.Value
			}
		}
		fix = strings.ReplaceAll(fix, fmt.Sprintf("${%s}", v.Value()), value)
		fix = strings.ReplaceAll(fix, fmt.Sprintf("${%s.%s}", v.Key(), v.Value()), value)
	}
	return fix
}

func (c Context) IdentifyVariables(category, source string) []collection.Pair[string, string] {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(source, -1)

	results := collection.DictionaryEmpty[string, collection.Pair[string, string]]()
	for _, match := range matches {
		if len(match) == 0 {
			continue
		}

		category := category
		key := match[1]

		fragments := strings.Split(key, ".")
		if len(fragments) > 1 {
			category = fragments[0]
			key = fragments[1]
		}

		results.PutIfAbsent(fmt.Sprintf("%s.%s", category, key), collection.NewPair(category, key))
	}

	return results.Values()
}

func ProcessRequest(request *action.Request, context *Context) *action.Request {
	if !context.Status {
		return request
	}

	return &action.Request{
		Id:        request.Id,
		Timestamp: request.Timestamp,
		Name:      request.Name,
		Method:    request.Method,
		Uri:       context.Apply("uri", request.Uri),
		Query:     *processQuery(request.Query, context),
		Header:    *processHeader(request.Header, context),
		Cookie:    *processCookie(request.Cookie, context),
		Body:      *processBody(request.Body, context),
		Auth:      *processAuth(request.Auth, context),
		Owner:     request.Owner,
		Modified:  request.Modified,
		Status:    request.Status,
	}
}

func processQuery(queries query.Queries, context *Context) *query.Queries {
	queryCategory := map[string][]query.Query{}
	for k, qs := range queries.Queries {
		key := context.Apply("query", k)
		queryCollection := []query.Query{}
		for _, q := range qs {
			value := context.Apply("query", q.Value)
			queryCollection = append(queryCollection, query.Query{
				Status: q.Status,
				Value:  value,
			})
		}
		queryCategory[key] = queryCollection
	}

	return &query.Queries{
		Queries: queryCategory,
	}
}

func processHeader(headers header.Headers, context *Context) *header.Headers {
	headerCategory := map[string][]header.Header{}
	for k, hs := range headers.Headers {
		key := context.Apply("header", k)
		headerCollection := []header.Header{}
		for _, h := range hs {
			value := context.Apply("header", h.Value)
			headerCollection = append(headerCollection, header.Header{
				Status: h.Status,
				Value:  value,
			})
		}
		headerCategory[key] = headerCollection
	}

	return &header.Headers{
		Headers: headerCategory,
	}
}

func processCookie(cookies cookie.CookiesClient, context *Context) *cookie.CookiesClient {
	cookieCategory := map[string]cookie.CookieClient{}
	for k, c := range cookies.Cookies {
		key := context.Apply("cookie", k)
		cookieCategory[key] = cookie.CookieClient{
			Order:  c.Order,
			Status: c.Status,
			Value:  context.Apply("cookie", c.Value),
		}
	}

	return &cookie.CookiesClient{
		Cookies: cookieCategory,
	}
}

func processBody(payload body.BodyRequest, context *Context) *body.BodyRequest {
	for k, v := range payload.Parameters {
		for j, v := range v {
			for i, v := range v {
				if !v.Status || v.IsFile {
					continue
				}
				v.Value = context.Apply("payload", v.Value)
				payload.Parameters[k][j][i] = v
			}
		}
	}

	payload.ContentType = body.ContentType(context.Apply("payload", string(payload.ContentType)))

	return &body.BodyRequest{
		Status:      payload.Status,
		ContentType: payload.ContentType,
		Parameters:  payload.Parameters,
	}
}

func processAuth(auths auth.Auths, context *Context) *auth.Auths {
	authCategory := map[string]auth.Auth{}

	for k, a := range auths.Auths {
		key := context.Apply("auth", k)
		parameters := map[string]string{}
		for k, p := range a.Parameters {
			key := context.Apply("auth", k)
			parameters[key] = context.Apply("auth", p)
		}
		authCategory[key] = auth.Auth{
			Status:     a.Status,
			Type:       a.Type,
			Parameters: parameters,
		}
	}

	return &auth.Auths{
		Status: auths.Status,
		Auths:  authCategory,
	}
}
