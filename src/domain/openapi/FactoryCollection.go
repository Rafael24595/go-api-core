package openapi

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/Rafael24595/go-api-core/src/commons/utils"
	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	"github.com/Rafael24595/go-api-core/src/domain/action/auth"
	auth_strategy "github.com/Rafael24595/go-api-core/src/domain/action/auth/strategy"
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	body_strategy "github.com/Rafael24595/go-api-core/src/domain/action/body/strategy"
	"github.com/Rafael24595/go-api-core/src/domain/action/cookie"
	"github.com/Rafael24595/go-api-core/src/domain/action/header"
	"github.com/Rafael24595/go-api-core/src/domain/action/query"
	"github.com/Rafael24595/go-api-core/src/domain/collection"
	"github.com/Rafael24595/go-api-core/src/domain/context"
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

func (b *FactoryCollection) Make() (*collection.Collection, *context.Context, []action.Request, error) {
	now := time.Now().UnixMilli()

	ctx := context.NewContext(b.owner)
	nodes := make([]action.Request, 0)

	server := ""
	for i, v := range b.openapi.Servers {
		key := fmt.Sprintf("server-%d", i)
		ctx.Put(context.URI, fmt.Sprintf("server-%d", i), v.URL, false)
		server = fmt.Sprintf("${%s}", key)
	}

	var node *action.Request
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

	return &collection.Collection{
		Id:        "",
		Name:      fmt.Sprintf("%s-%s", b.openapi.Info.Title, b.openapi.Info.Version),
		Timestamp: now,
		Context:   "",
		Nodes:     make([]domain.NodeReference, 0),
		Owner:     b.owner,
		Modified:  now,
		Status:    collection.FREE,
	}, ctx, nodes, nil
}

func (b *FactoryCollection) MakeFromOperation(method domain.HttpMethod, path string, operation *Operation, ctx *context.Context) (*context.Context, *action.Request) {
	now := time.Now().UnixMilli()

	name := path
	if operation.Summary != "" {
		name = operation.Summary
	}

	path, ctx, queries, headers, cookies := b.MakeFromParameters(path, operation.Parameters, ctx)
	payload := b.MakeFromRequestBody(operation.RequestBody)
	auth := b.MakeFromSecurity(operation.Security, headers)

	return ctx, &action.Request{
		Id:        "",
		Timestamp: now,
		Name:      name,
		Method:    method,
		Uri:       path,
		Query:     *queries,
		Header:    *headers,
		Cookie:    *cookies,
		Body:      *payload,
		Auth:      *auth,
		Owner:     b.owner,
		Modified:  now,
		Status:    action.GROUP,
	}
}

func (b *FactoryCollection) MakeFromParameters(path string, parameters []Parameter, ctx *context.Context) (string, *context.Context, *query.Queries, *header.Headers, *cookie.CookiesClient) {
	queries := query.NewQueries()
	headers := header.NewHeaders()
	cookies := cookie.NewCookiesClient()

	for _, v := range parameters {
		switch v.In {
		case "path":
			ctx.Put(context.URI, v.Name, v.Description, false)
			placeholder := fmt.Sprintf("{%s}", v.Name)
			replacement := fmt.Sprintf("${%s}", v.Name)
			path = strings.ReplaceAll(path, placeholder, replacement)
		case "query":
			queries.Add(v.Name, v.Description)
		case "header":
			headers.Add(v.Name, v.Description)
		case "cookie":
			cookies.Put(v.Name, v.Description)
		}
	}

	return path, ctx, queries, headers, cookies
}

func (b *FactoryCollection) MakeFromRequestBody(requestBody *RequestBody) *body.BodyRequest {
	if requestBody == nil {
		return body.EmptyBody(false, domain.None)
	}

	if requestBody.Ref != "" {
		reference, err := b.findRequestBodyReference(requestBody.Ref)
		if reference == nil || err != nil {
			return body.EmptyBody(false, domain.None)
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

	return body.EmptyBody(false, domain.Text)
}

func (b *FactoryCollection) fromExample(content string, schema *Schema) *body.BodyRequest {
	example, err := json.Marshal(schema.Example)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}

	var bodyType domain.ContentType
	switch content {
	case "multipart/form-data":
		bodyType = domain.Form
	case "application/json":
		bodyType = domain.Json
	case "application/xml":
		bodyType = domain.Xml
	default:
		bodyType = domain.Text
	}

	if bodyType != domain.Form {
		return body_strategy.DocumentBody(false, bodyType, string(example))
	}

	return body.EmptyBody(false, bodyType)
}

func (b *FactoryCollection) toFormData(parameters map[string]BuildParameter) *body.BodyRequest {
	builder := body_strategy.NewBuilderFromDataBody()

	count := int64(0)
	for k, v := range parameters {
		parameter := &body.BodyParameter{
			Order:    count,
			Status:   true,
			IsFile:   v.Binary,
			FileType: "",
			FileName: fmt.Sprintf("%s file", k),
			Value:    v.Value,
		}

		builder.Add(k, parameter)

		count++
	}

	return body_strategy.FormDataBody(false, domain.Form, builder)
}

func (b *FactoryCollection) toDocument(content string, schema *Schema, parameters map[string]BuildParameter) *body.BodyRequest {
	payload, contenType := b.formatDocument(content, schema, parameters)
	return body_strategy.DocumentBody(false, contenType, payload)
}

func (b *FactoryCollection) formatDocument(content string, schema *Schema, parameters map[string]BuildParameter) (string, domain.ContentType) {
	vector := schema.Type == "array"

	switch content {
	case "application/json":
		return b.formatJson(parameters, vector), domain.Json
	case "application/xml":
		return b.formatXml(parameters), domain.Xml
	default:
		return b.formatJson(parameters, vector), domain.Text
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
				queries.AddHeader(schema.Name, header)
			}

			switch schema.Scheme {
			case "basic":
				auths.PutAuth(*auth_strategy.BasicAuth(true,
					auth_strategy.BASIC_PARAM_USER,
					auth_strategy.BASIC_PARAM_PASSWORD))
			case "bearer":
				auths.PutAuth(*auth_strategy.BearerAuth(true,
					schema.BearerFormat,
					auth_strategy.BEARER_PARAM_TOKEN))
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
