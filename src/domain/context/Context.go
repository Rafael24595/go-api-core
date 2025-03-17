package context

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/auth"
	"github.com/Rafael24595/go-api-core/src/domain/body"
	"github.com/Rafael24595/go-api-core/src/domain/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/header"
	"github.com/Rafael24595/go-api-core/src/domain/query"
	"github.com/Rafael24595/go-collections/collection"
)

type DictionaryVariables = collection.Dictionary[string, string]
type DictionaryCategory = collection.Dictionary[string, DictionaryVariables]

type Context struct {
	Id         string             `json:"_id"`
	Status     bool               `json:"status"`
	Timestamp  int64              `json:"timestamp"`
	Dictionary DictionaryCategory `json:"dictionary"`
	Owner      string             `json:"owner"`
	Modified   int64              `json:"modified"`
}

func NewContext(owner string) *Context {
	return &Context{
		Id:         "",
		Status:     true,
		Timestamp:  time.Now().UnixMilli(),
		Dictionary: *collection.DictionaryEmpty[string, DictionaryVariables](),
		Owner:      owner,
		Modified:   time.Now().UnixMilli(),
	}
}

func (c *Context) PutAll(category string, context map[string]string) *Context {
	variables, ok := c.Dictionary.Get(category)
	if !ok {
		c.Dictionary.Put(category, *collection.DictionaryEmpty[string, string]())
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
			if ok {
				value = *keyValue
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

func ProcessRequest(request *domain.Request, context Context) *domain.Request {
	if !context.Status {
		return request
	}

	return &domain.Request{
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

func processQuery(queries query.Queries, context Context) *query.Queries {
	queryCategory := map[string][]query.Query{}
	for k, qs := range queries.Queries {
		key := context.Apply("query", k)
		queryCollection := []query.Query{}
		for _, q := range qs {
			key := context.Apply("query", q.Key)
			value := context.Apply("query", q.Value)
			queryCollection = append(queryCollection, query.Query{
				Status: q.Status,
				Key:    key,
				Value:  value,
			})
		}
		queryCategory[key] = queryCollection
	}

	return &query.Queries{
		Queries: queryCategory,
	}
}

func processHeader(headers header.Headers, context Context) *header.Headers {
	headerCategory := map[string][]header.Header{}
	for k, hs := range headers.Headers {
		key := context.Apply("header", k)
		headerCollection := []header.Header{}
		for _, h := range hs {
			key := context.Apply("header", h.Key)
			value := context.Apply("header", h.Value)
			headerCollection = append(headerCollection, header.Header{
				Status: h.Status,
				Key:    key,
				Value:  value,
			})
		}
		headerCategory[key] = headerCollection
	}

	return &header.Headers{
		Headers: headerCategory,
	}
}

func processCookie(cookies cookie.Cookies, context Context) *cookie.Cookies {
	cookieCategory := map[string]cookie.Cookie{}
	for k, c := range cookies.Cookies {
		key := context.Apply("cookie", k)
		cookieCategory[key] = cookie.Cookie{
			Status:     c.Status,
			Code:       context.Apply("cookie", c.Code),
			Value:      context.Apply("cookie", c.Value),
			Domain:     c.Domain,
			Path:       c.Path,
			Expiration: c.Expiration,
			MaxAge:     c.MaxAge,
			Secure:     c.Secure,
			HttpOnly:   c.HttpOnly,
			SameSite:   c.SameSite,
		}
	}

	return &cookie.Cookies{
		Cookies: cookieCategory,
	}
}

func processBody(payload body.Body, context Context) *body.Body {
	bodyFixed := context.Apply("payload", string(payload.Bytes))
	return &body.Body{
		Status:      payload.Status,
		ContentType: payload.ContentType,
		Bytes:       []byte(bodyFixed),
	}
}

func processAuth(auths auth.Auths, context Context) *auth.Auths {
	authCategory := map[string]auth.Auth{}

	for k, a := range auths.Auths {
		key := context.Apply("auth", k)
		parameters := map[string]auth.Parameter{}
		for k, p := range a.Parameters {
			key := context.Apply("auth", k)
			parameters[key] = auth.Parameter{
				Key:   context.Apply("auth", p.Key),
				Value: context.Apply("auth", p.Value),
			}
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
