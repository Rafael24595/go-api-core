package collection

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/body"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/domain/header"
	"github.com/Rafael24595/go-api-core/src/domain/openapi"
	"github.com/Rafael24595/go-api-core/src/domain/query"
)

type BuilderCollection struct {
	owner   string
	openapi openapi.OpenAPI
	raw     map[string]any
}

func NewBuilderCollection(owner string, openapi *openapi.OpenAPI) *BuilderCollection {
	return &BuilderCollection{
		owner:   owner,
		openapi: *openapi,
		raw:     make(map[string]any),
	}
}

func (b *BuilderCollection) SetRaw(raw map[string]any) *BuilderCollection {
	b.raw = raw
	return b
}

func (b *BuilderCollection) Make() (*domain.Collection, *context.Context, []domain.Request, error) {
	now := time.Now().UnixMilli()

	ctx := context.NewContext(b.owner)
	nodes := make([]domain.Request, 0)

	for i, v := range b.openapi.Servers {
		ctx.Put(context.URI, fmt.Sprintf("server-%d", i), v.URL)
	}

	for path, v := range b.openapi.Paths {
		if operation := v.Get; operation != nil {
			var node *domain.Request
			ctx, node = b.makeFromOperation(domain.GET, path, operation, ctx)
			nodes = append(nodes, *node)
		}
		if operation := v.Post; operation != nil {
			var node *domain.Request
			ctx, node = b.makeFromOperation(domain.POST, path, operation, ctx)
			nodes = append(nodes, *node)
		}
		if operation := v.Put; operation != nil {
			var node *domain.Request
			ctx, node = b.makeFromOperation(domain.PUT, path, operation, ctx)
			nodes = append(nodes, *node)
		}
		if operation := v.Delete; operation != nil {
			var node *domain.Request
			ctx, node = b.makeFromOperation(domain.DELETE, path, operation, ctx)
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

func (b *BuilderCollection) makeFromOperation(method domain.HttpMethod, path string, operation *openapi.Operation, ctx *context.Context) (*context.Context, *domain.Request) {
	now := time.Now().UnixMilli()

	name := path
	if operation.Summary != "" {
		name = operation.Summary
	}

	path, ctx, queries, headers := b.makeFromParameters(path, operation.Parameters, ctx)
	payload := b.MakeFromRequestBody(operation.RequestBody)

	return ctx, &domain.Request{
		Id:        "",
		Timestamp: now,
		Name:      name,
		Method:    method,
		Uri:       path,
		Query:     *queries,
		Header:    *headers,
		//Cookie: ,
		Body: *payload,
		//Auth: ,
		Owner:    b.owner,
		Modified: now,
		Status:   domain.GROUP,
	}
}

func (b *BuilderCollection) makeFromParameters(path string, parameters []openapi.Parameter, ctx *context.Context) (string, *context.Context, *query.Queries, *header.Headers) {
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

func (b *BuilderCollection) MakeFromRequestBody(requestBody *openapi.RequestBody) *body.Body {
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

func (b *BuilderCollection) MakeFromSchema(schema *openapi.Schema) string {
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

func (b *BuilderCollection) makeFromExample(schema *openapi.Schema) string {
	example, err := json.Marshal(schema.Example) 
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	return string(example)
}


func (b *BuilderCollection) makeFromReference(schema *openapi.Schema) string {
	ref, err := b.findReference(schema)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	if ref != nil {
		return b.MakeFromSchema(ref)
	}

	return ""
}

func (b *BuilderCollection) makeFromProperties(schema *openapi.Schema) string {
	lines := []string{}

	for key, v := range schema.Properties {
		value := v.Example

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

func (b *BuilderCollection) findSchema(schema *openapi.Schema) (*openapi.Schema, error) {
	if schema.Ref == "" {
		return schema, nil
	}

	return b.findReference(schema)
}

func (b *BuilderCollection) findReference(schema *openapi.Schema) (*openapi.Schema, error) {
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
