package openapi

import (
	"encoding/json"
	"fmt"
	"maps"
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

type BuildParameter struct {
	Value    string
	Children *map[string]BuildParameter
	Vector   bool
	Binary   bool
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
		Status:    domain.FREE,
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
		Cookie:    *cookie.NewCookiesClient(),
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
			order := len(queries.Queries)
			queries.Add(v.Name, query.NewQuery(int64(order), true, v.Description))
		case "header":
			order := len(headers.Headers)
			headers.Add(v.Name, header.NewHeader(int64(order), true, v.Description))
		}
	}

	return path, ctx, queries, headers
}

func (b *FactoryCollection) MakeFromRequestBody(requestBody *RequestBody) *body.BodyRequest {
	if requestBody == nil {
		return body.NewBody(false, body.None, make(map[string]map[string][]body.BodyParameter))
	}

	if requestBody.Ref != "" {
		reference, err := b.findRequestBodyReference(requestBody.Ref)
		if reference == nil || err != nil {
			return body.NewBody(false, body.None, make(map[string]map[string][]body.BodyParameter))
		}
		return b.MakeFromRequestBody(reference)
	}

	for k, v := range requestBody.Content {
		schema, err := b.findSchema(&v.Schema)
		if schema == nil {
			continue
		}
		if err != nil {
			break
		}

		if schema.Example != nil && schema.Example != "" {
			return b.fromExample(k, schema)
		}

		parameters, _ := b.MakeFromSchema(k, schema, make(map[string]int))

		switch k {
		case "multipart/form-data":
			return b.toFormData(parameters)
		default:
			return b.toDocument(k, schema, parameters)
		}
	}

	return body.NewBody(false, body.None, make(map[string]map[string][]body.BodyParameter))
}

func (b *FactoryCollection) fromExample(content string, schema *Schema) *body.BodyRequest {
	example, err := json.Marshal(schema.Example)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}

	data := make(map[string]map[string][]body.BodyParameter)

	bodyType := body.None
	switch content {
	case "multipart/form-data":
		bodyType = body.Form
	case "application/json":
		bodyType = body.Json
	case "application/xml":
		bodyType = body.Xml
	default:
		bodyType = body.Text
	}

	if bodyType != body.Form {
		data[body.DOCUMENT_PARAM] = make(map[string][]body.BodyParameter)
		data[body.DOCUMENT_PARAM][body.PAYLOAD_PARAM] = []body.BodyParameter{
			body.NewBodyDocument(0, true, string(example)),
		}
	}

	return body.NewBody(false, bodyType, data)
}

func (b *FactoryCollection) toFormData(parameters map[string]BuildParameter) *body.BodyRequest {
	data := make(map[string]map[string][]body.BodyParameter)

	count := int64(0)
	for k, v := range parameters {
		data[body.FORM_DATA_PARAM] = make(map[string][]body.BodyParameter)
		data[body.FORM_DATA_PARAM][k] = []body.BodyParameter{
			{
				Order:    count,
				Status:   true,
				IsFile:   v.Binary,
				FileType: "",
				FileName: fmt.Sprintf("%s file", k),
				Value:    v.Value,
			},
		}
		count++
	}

	return body.NewBody(false, body.Form, data)
}

func (b *FactoryCollection) toDocument(content string, schema *Schema, parameters map[string]BuildParameter) *body.BodyRequest {
	payload, contenType := b.formatDocument(content, schema, parameters)
	data := make(map[string]map[string][]body.BodyParameter)
	data[body.DOCUMENT_PARAM] = make(map[string][]body.BodyParameter)
	data[body.DOCUMENT_PARAM][body.PAYLOAD_PARAM] = []body.BodyParameter{
		body.NewBodyDocument(0, true, payload),
	}
	return body.NewBody(false, contenType, data)
}

func (b *FactoryCollection) formatDocument(content string, schema *Schema, parameters map[string]BuildParameter) (string, body.ContentType) {
	vector := schema.Type == "array"

	switch content {
	case "application/json":
		return b.formatJson(parameters, vector), body.Json
	case "application/xml":
		return b.formatXml(parameters), body.Xml
	default:
		return b.formatJson(parameters, vector), body.Text
	}
}

func (b *FactoryCollection) formatXml(parameters map[string]BuildParameter) string {
	lines := make([]string, 0)

	for k, v := range parameters {
		value := v.Value
		if v.Children != nil {
			value = b.formatJson(*v.Children, v.Vector)
		}

		line := fmt.Sprintf("<%s>%v<%s>", k, value, k)
		lines = append(lines, line)
	}

	return strings.Join(lines, " ")
}

func (b *FactoryCollection) formatJson(parameters map[string]BuildParameter, vector bool) string {
	lines := make([]string, 0)

	for k, v := range parameters {
		value := v.Value
		if v.Children != nil {
			value = b.formatJson(*v.Children, v.Vector)
		}

		line := fmt.Sprintf("\"%s\": %v", k, value)
		lines = append(lines, line)
	}

	payload := strings.Join(lines, ", ")
	payload = fmt.Sprintf("{ %s }", payload)

	if vector {
		payload = fmt.Sprintf("[ %s ]", payload)
	}

	return payload
}

func (b *FactoryCollection) MakeFromSchema(content string, schema *Schema, visited map[string]int) (map[string]BuildParameter, map[string]int) {
	if schema.Ref != "" {
		return b.makeFromReference(content, schema, visited)
	}

	if schema.Items != nil {
		return b.makeFromReference(content, schema.Items, visited)
	}

	if len(schema.Properties) > 0 {
		return b.makeFromProperties(content, schema, visited)
	}

	return make(map[string]BuildParameter), visited
}

func (b *FactoryCollection) makeFromReference(content string, schema *Schema, visited map[string]int) (map[string]BuildParameter, map[string]int) {
	ref, err := b.findSchemaReference(schema.Ref)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}

	_, hasVisited := visited[schema.Ref]
	if hasVisited {
		data := make(map[string]BuildParameter)
		data["$Circular"] = BuildParameter{
			Value:  fmt.Sprintf("\"Circular schema '%s'.\"", schema.Ref),
			Binary: false,
		}
		return data, visited
	}

	if ref != nil {
		if schema.Ref != "" {
			visited[schema.Ref] = 0
		}
		return b.MakeFromSchema(content, ref, visited)
	}

	return make(map[string]BuildParameter), visited
}

func (b *FactoryCollection) makeFromProperties(content string, schema *Schema, visited map[string]int) (map[string]BuildParameter, map[string]int) {
	parameters := make(map[string]BuildParameter)

	for key, v := range schema.Properties {
		parameter, exists := b.makeFromPrimitiveProperty(key, &v)
		if exists {
			parameters[key] = *parameter
		}

		if exists || (v.Ref == "" && v.Items == nil) {
			continue
		}

		var children map[string]BuildParameter
		children, visited = b.MakeFromSchema(content, &v, visited)

		if content == "multipart/form-data" {
			maps.Copy(parameters, children)
		} else {
			parameters[key] = BuildParameter{
				Value:    "",
				Children: &children,
				Vector:   v.Type == "array",
				Binary:   false,
			}
		}

	}

	return parameters, visited
}

func (b *FactoryCollection) makeFromPrimitiveProperty(field string, schema *Schema) (*BuildParameter, bool) {
	example := schema.Example

	switch schema.Type {
	case "string":
		if example == nil {
			example = field
		}
		example = fmt.Sprintf("\"%s\"", example)
	case "number":
		if example == nil {
			example = "0.0"
		}
	case "integer":
		if example == nil {
			example = "0"
		}
	case "boolean":
		if example == nil {
			example = "false"
		}
	default:
		return nil, false
	}

	return &BuildParameter{
		Value:    fmt.Sprintf("%v", example),
		Children: nil,
		Vector:   false,
		Binary:   schema.Format == "binary",
	}, true
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

	return b.findSchemaReference(schema.Ref)
}

func (b *FactoryCollection) findRequestBodyReference(ref string) (*RequestBody, error) {
	if strings.HasPrefix(ref, "#/components/requestBodies/") {
		requestBodyName := strings.TrimPrefix(ref, "#/components/requestBodies/")
		requestBody, exists := b.openapi.Components.RequestBodies[requestBodyName]
		if !exists {
			return nil, fmt.Errorf("schema not found: %s", requestBodyName)
		}
		return &requestBody, nil
	}

	fixRef := strings.TrimPrefix(ref, "#")
	fixRef = strings.ReplaceAll(fixRef, "/", ".")

	var requestBody *RequestBody
	err := utils.FindJson(fixRef, b.raw, requestBody)

	return requestBody, err
}

func (b *FactoryCollection) findSchemaReference(ref string) (*Schema, error) {
	if strings.HasPrefix(ref, "#/components/schemas/") {
		schemaName := strings.TrimPrefix(ref, "#/components/schemas/")
		schema, exists := b.openapi.Components.Schemas[schemaName]
		if !exists {
			return nil, fmt.Errorf("schema not found: %s", schemaName)
		}
		return &schema, nil
	}

	fixRef := strings.TrimPrefix(ref, "#")
	fixRef = strings.ReplaceAll(fixRef, "/", ".")

	var schema *Schema
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
