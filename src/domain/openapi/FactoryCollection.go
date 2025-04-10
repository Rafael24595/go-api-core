package openapi

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/auth"
	"github.com/Rafael24595/go-api-core/src/domain/body"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/domain/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/header"
	"github.com/Rafael24595/go-api-core/src/domain/query"
)

type FactoryCollection struct {
	owner   string
	openapi OpenAPI
	raw     map[string]any
}

func NewFactoryCollection(owner string, openapi *OpenAPI) *FactoryCollection {
	return &FactoryCollection{
		owner:   owner,
		openapi: *openapi,
		raw:     make(map[string]any),
	}
}

func (b *FactoryCollection) SetRaw(raw map[string]any) *FactoryCollection {
	b.raw = raw
	return b
}

func (b *FactoryCollection) Make() (*domain.Collection, *context.Context, []domain.Request, error) {
	now := time.Now().UnixMilli()

	ctx := context.NewContext(b.owner)
	nodes := make([]domain.Request, 0)

	server := ""
	for i, v := range b.openapi.Servers {
		key := fmt.Sprintf("server-%d", i)
		ctx.Put(context.URI, fmt.Sprintf("server-%d", i), v.URL)
		server = fmt.Sprintf("${%s}", key)
	}

	var node *domain.Request
	for path, v := range b.openapi.Paths {
		pathFull := fmt.Sprintf("%s%s", server, path)
		if operation := v.Get; operation != nil {
			ctx, node = b.MakeFromOperation(domain.GET, pathFull, operation, ctx)
			nodes = append(nodes, *node)
		}
		if operation := v.Post; operation != nil {
			ctx, node = b.MakeFromOperation(domain.POST, pathFull, operation, ctx)
			nodes = append(nodes, *node)
		}
		if operation := v.Put; operation != nil {
			ctx, node = b.MakeFromOperation(domain.PUT, pathFull, operation, ctx)
			nodes = append(nodes, *node)
		}
		if operation := v.Delete; operation != nil {
			ctx, node = b.MakeFromOperation(domain.DELETE, pathFull, operation, ctx)
			nodes = append(nodes, *node)
		}
	}

	return &domain.Collection{
		Id:        "",
		Name:      fmt.Sprintf("%s-%s", b.openapi.Info.Title, b.openapi.Info.Version),
		Timestamp: now,
		Context:   "",
		Nodes:     make([]domain.NodeReference, 0),
		Owner:     b.owner,
		Modified:  now,
	}, ctx, nodes, nil
}

func (b *FactoryCollection) MakeFromOperation(method domain.HttpMethod, path string, operation *Operation, ctx *context.Context) (*context.Context, *domain.Request) {
	now := time.Now().UnixMilli()

	name := path
	if operation.Summary != "" {
		name = operation.Summary
	}

	path, ctx, queries, headers := b.MakeFromParameters(path, operation.Parameters, ctx)
	payload := b.MakeFromRequestBody(operation.RequestBody)
	auth := b.MakeFromSecurity(operation.Security, headers)

	return ctx, &domain.Request{
		Id:        "",
		Timestamp: now,
		Name:      name,
		Method:    method,
		Uri:       path,
		Query:     *queries,
		Header:    *headers,
		Cookie:    *cookie.NewCookies(),
		Body:      *payload,
		Auth:      *auth,
		Owner:     b.owner,
		Modified:  now,
		Status:    domain.GROUP,
	}
}

func (b *FactoryCollection) MakeFromParameters(path string, parameters []Parameter, ctx *context.Context) (string, *context.Context, *query.Queries, *header.Headers) {
	queries := query.NewQueries()
	headers := header.NewHeaders()

	for _, v := range parameters {
		switch v.In {
		case "path":
			ctx.Put(context.URI, v.Name, v.Description)
			placeholder := fmt.Sprintf("{%s}", v.Name)
			replacement := fmt.Sprintf("${%s}", v.Name)
			path = strings.ReplaceAll(path, placeholder, replacement)
		case "query":
			order := int64(queries.SizeOf(v.Name))
			queries.Add(v.Name, query.NewQuery(order, true, v.Description))
		case "header":
			order := int64(headers.SizeOf(v.Name))
			headers.Add(v.Name, header.NewHeader(order, true, v.Description))
		}
	}

	return path, ctx, queries, headers
}

func (b *FactoryCollection) MakeFromRequestBody(requestBody *RequestBody) *body.Body {
	if requestBody == nil {
		return body.NewBody(false, body.None, make([]byte, 0))
	}

	for _, v := range requestBody.Content {
		schema, err := b.findSchema(&v.Schema)
		if schema == nil {
			continue
		}
		if err != nil {
			break
		}

		payload := b.MakeFromSchema(schema)
		return body.NewBody(false, body.None, []byte(payload))
	}

	return body.NewBody(false, body.None, make([]byte, 0))
}

func (b *FactoryCollection) MakeFromSchema(schema *Schema) string {
	if schema.Example != nil && schema.Example != "" {
		return b.makeFromExample(schema)
	}

	payload := ""

	if schema.Ref != "" {
		payload = b.makeFromReference(schema)
	}

	if schema.Items != nil {
		payload = b.makeFromReference(schema.Items)
	}

	if len(schema.Properties) > 0 {
		payload = b.makeFromProperties(schema)
	}

	if schema.Type == "array" {
		payload = fmt.Sprintf("[ %s ]", payload)
	}

	return payload
}

func (b *FactoryCollection) makeFromExample(schema *Schema) string {
	example, err := json.Marshal(schema.Example)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	return string(example)
}

func (b *FactoryCollection) makeFromReference(schema *Schema) string {
	ref, err := b.findReference(schema)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	if ref != nil {
		return b.MakeFromSchema(ref)
	}

	return ""
}

func (b *FactoryCollection) makeFromProperties(schema *Schema) string {
	lines := []string{}

	for key, v := range schema.Properties {
		value := v.Example
		if value == nil {
			if v.Type == "integer" {
				value = "0"
			}
			if v.Type == "boolean" {
				value = "false"
			}
		}

		if value == nil {
			value = key
		}

		if v.Ref != "" || v.Items != nil {
			value = b.MakeFromSchema(&v)
		}

		if v.Type == "string" {
			value = fmt.Sprintf("\"%s\"", value)
		}

		line := fmt.Sprintf("\"%s\": %s", key, value)
		lines = append(lines, line)
	}

	body := strings.Join(lines, ", ")
	body = fmt.Sprintf("{ %s }", body)

	return body
}

func (b *FactoryCollection) MakeFromSecurity(security []SecurityRequirement, queries *header.Headers) *auth.Auths {
	auths := auth.NewAuths(false)

	for _, v := range security {
		for k := range v {
			schema, err := b.findAuth(k)
			if err != nil {
				fmt.Printf("%s", err.Error())
			}

			if schema.In == "header" {
				name := schema.Name
				order := queries.SizeOf(name)
				header := header.NewHeader(int64(order), true, name)
				queries.Add(schema.Name, header)
			}

			switch schema.Scheme {
			case "basic":
				auths.PutAuth(*auth.NewAuth(true, auth.Basic, map[string]string{
					auth.BASIC_PARAM_USER:     auth.BASIC_PARAM_USER,
					auth.BASIC_PARAM_PASSWORD: auth.BASIC_PARAM_PASSWORD,
				}))
			case "bearer":
				bearer := schema.BearerFormat
				if bearer == "" {
					bearer = auth.BEARER_PARAM_PREFIX
				}

				auths.PutAuth(*auth.NewAuth(true, auth.Bearer, map[string]string{
					auth.BEARER_PARAM_PREFIX: schema.BearerFormat,
					auth.BEARER_PARAM_TOKEN:  auth.BEARER_PARAM_TOKEN,
				}))
			}
		}
	}

	if len(auths.Auths) > 0 {
		auths.Status = true
	}

	return auths
}

func (b *FactoryCollection) findSchema(schema *Schema) (*Schema, error) {
	if schema.Ref == "" {
		return schema, nil
	}

	return b.findReference(schema)
}

func (b *FactoryCollection) findReference(schema *Schema) (*Schema, error) {
	if strings.HasPrefix(schema.Ref, "#/components/schemas/") {
		schemaName := strings.TrimPrefix(schema.Ref, "#/components/schemas/")
		schema, exists := b.openapi.Components.Schemas[schemaName]
		if !exists {
			return nil, fmt.Errorf("schema not found: %s", schemaName)
		}
		return &schema, nil
	}

	fixRef := strings.TrimPrefix(schema.Ref, "#")
	fixRef = strings.ReplaceAll(fixRef, "/", ".")

	err := utils.FindJson(fixRef, b.raw, schema)

	return schema, err
}

func (b *FactoryCollection) findAuth(auth string) (*SecurityScheme, error) {
	schema, exists := b.openapi.Components.SecuritySchemes[auth]
	if !exists {
		return nil, fmt.Errorf("schema not found: %s", auth)
	}
	return &schema, nil
}
