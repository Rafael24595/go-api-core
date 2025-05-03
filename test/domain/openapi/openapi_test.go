package test_openapi

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/auth"
	"github.com/Rafael24595/go-api-core/src/domain/body"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/domain/header"
	"github.com/Rafael24595/go-api-core/src/domain/openapi"
	"github.com/Rafael24595/go-api-core/src/domain/query"
)

const TEST_OWNER = "anonymoys"

func makeOpenApiArguments(t *testing.T) (*openapi.OpenAPI, *map[string]any) {
	file, err := os.Open("test_openaoi.yaml")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	yaml, err := io.ReadAll(file)
	if err != nil {
		t.Error(err)
	}

	oapi, raw, err := openapi.MakeFromYaml(yaml)
	if err != nil {
		t.Error(err)
	}

	return oapi, raw
}

func TestMake(t *testing.T) {
	oapi, raw := makeOpenApiArguments(t)

	builder := openapi.NewFactoryCollection(TEST_OWNER, oapi).SetRaw(*raw)

	collection, ctx, requests, err := builder.Make()
	if err != nil {
		t.Error(err)
	}

	value := collection.Name
	expected := "Task Management API-1.0.0"
	if value != expected {
		t.Errorf("Found variable %s but %s expected", value, expected)
	}

	valideContext(t, ctx)
	valideRequests(t, requests)
}

func valideContext(t *testing.T, ctx *context.Context) {
	uriContext, ok := ctx.Dictionary.Get(context.URI.String())
	if !ok {
		t.Errorf("Uri context not found.")
	}

	key := "server-0"

	server, ok := uriContext.Get(key)
	expected := "https://api.example.com/v1"
	if !ok || server.Value != expected {
		t.Errorf("Server URI not found.")
	}
}

func valideRequests(t *testing.T, requests []domain.Request) {
	if len(requests) != 4 {
		t.Fatalf("Expected 4 request object, got %d", len(requests))
	}

	name := "Get collection data"
	method := domain.GET
	uri := "${server-0}/collection/${userId}"

	var request *domain.Request
	for _, v := range requests {
		if v.Name == name && v.Method == method && v.Uri == uri {
			request = &v
			break
		}
	}

	if request == nil {
		t.Error("Request not found")
	}
}

func TestMakeFromOperation(t *testing.T) {
	oapi, raw := makeOpenApiArguments(t)

	builder := openapi.NewFactoryCollection(TEST_OWNER, oapi).SetRaw(*raw)

	ctx := context.NewContext(TEST_OWNER)

	path := "/request"
	operation := oapi.Paths["/request"].Post
	_, request := builder.MakeFromOperation(domain.POST, path, operation, ctx)

	value := request.Name
	expected := "Create a new request"
	if value != expected {
		t.Errorf("Found variable %s but %s expected", value, expected)
	}
}

func TestMakeFromParameters(t *testing.T) {
	oapi, raw := makeOpenApiArguments(t)

	builder := openapi.NewFactoryCollection(TEST_OWNER, oapi).SetRaw(*raw)

	ctx := context.NewContext(TEST_OWNER)

	path := "/collection/{userId}"
	parameters := oapi.Paths[path].Get.Parameters
	fixPath, ctx, queries, headers := builder.MakeFromParameters(path, parameters, ctx)

	valideParametersPath(t, fixPath)
	valideParametersContext(t, ctx)
	valideParametersQuery(t, queries)
	valideParametersHeader(t, headers)
}

func valideParametersPath(t *testing.T, path string) {
	expected := "/collection/${userId}"
	if path != expected {
		t.Errorf("Found variable %s but %s expected", path, expected)
	}
}

func valideParametersContext(t *testing.T, ctx *context.Context) {
	uriContext, ok := ctx.Dictionary.Get(context.URI.String())
	if !ok {
		t.Errorf("Uri context not found.")
	}

	key := "userId"

	userId, ok := uriContext.Get(key)
	expected := "The ID of the user"
	if !ok || userId.Value != expected {
		t.Errorf("User ID not found.")
	}
}

func valideParametersQuery(t *testing.T, queries *query.Queries) {
	if len(queries.Queries) != 2 {
		t.Errorf("%d queries found but %d expected.", len(queries.Queries), 2)
	}

	key := "skip"

	query, ok := queries.Queries[key]
	if !ok || len(query) == 0 {
		t.Errorf("Query '%s' found.", key)
	}

	value := query[0].Value
	expected := "The skip of items to return"
	if value != expected {
		t.Errorf("Found variable %v but %v expected", value, expected)
	}

	key = "limit"

	query, ok = queries.Queries[key]
	if !ok || len(query) == 0 {
		t.Errorf("Query '%s' found.", key)
	}

	value = query[0].Value
	expected = "The limit of items to return"
	if value != expected {
		t.Errorf("Found variable %v but %v expected", value, expected)
	}
}

func valideParametersHeader(t *testing.T, headers *header.Headers) {
	key := "X-Request-ID"

	header, ok := headers.Headers[key]
	if !ok || len(header) == 0 {
		t.Errorf("Header '%s' found.", key)
	}

	value := header[0].Value
	expected := "X-Request-ID header"
	if value != expected {
		t.Errorf("Found variable %v but %v expected", value, expected)
	}
}

func TestMakeFromRequestBody(t *testing.T) {
	oapi, raw := makeOpenApiArguments(t)

	builder := openapi.NewFactoryCollection(TEST_OWNER, oapi).SetRaw(*raw)

	payload := oapi.Paths["/request"].Post.RequestBody
	result := builder.MakeFromRequestBody(payload)

	bodyParameter, ok := result.Parameters[body.DOCUMENT_PARAM]
	if !ok || bodyParameter.IsFile {
		t.Fatal("Body should be JSON")
	}

	var request map[string]any
	if err := json.Unmarshal([]byte(bodyParameter.Value), &request); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	value := request["method"]
	expected := "Request method"
	if value != expected {
		t.Errorf("Found variable %s but %s expected", value, expected)
	}

	value = request["uri"]
	expected = "Request Uri"
	if value != expected {
		t.Errorf("Found variable %s but %s expected", value, expected)
	}
}

func TestMakeFromSchema(t *testing.T) {
	oapi, raw := makeOpenApiArguments(t)

	builder := openapi.NewFactoryCollection(TEST_OWNER, oapi).SetRaw(*raw)

	schema := oapi.Paths["/collection/{userId}"].Get.Responses["200"].Content["application/json"].Schema
	example, _ := builder.MakeFromSchema("application/json", &schema, make(map[string]int))

	valideSchemaCollections(t, example)
}

func valideSchemaCollections(t *testing.T, example map[string]openapi.BuildParameter) {
	value := example["id"].Value
	expected := "\"000A1\""
	if value != expected {
		t.Errorf("Found variable %s but %s expected", value, expected)
	}

	valideSchemaCollectionContext(t, example)
	valideSchemaCollectionRequests(t, example)
}

func valideSchemaCollectionContext(t *testing.T, example map[string]openapi.BuildParameter) {
	context, ok := example["context"]
	if !ok || context.Children == nil {
		t.Errorf("Context not found or incorrect format.")
	}

	children := *context.Children

	value := children["status"].Value
	expected := "\"pending\""
	if value != expected {
		t.Errorf("Found variable %v but %v expected", value, expected)
	}
}

func valideSchemaCollectionRequests(t *testing.T, example map[string]openapi.BuildParameter) {
	requests, ok := example["requests"]
	if !ok || requests.Children == nil || !requests.Vector {
		t.Fatalf("Expected 1 request object, got %v", requests)
	}

	children := *requests.Children

	value := children["method"].Value
	expected := "\"Request method\""
	if value != expected {
		t.Errorf("Found variable %v but %v expected", value, expected)
	}

	fValue := children["timestamp"].Value
	fExpected := "1743433941068"
	if fValue != fExpected {
		t.Errorf("Found variable %v but %v expected", fValue, fExpected)
	}
}

func TestMakeFromSecurityBasic(t *testing.T) {
	oapi, raw := makeOpenApiArguments(t)

	builder := openapi.NewFactoryCollection(TEST_OWNER, oapi).SetRaw(*raw)

	security := oapi.Paths["/login"].Post.Security
	result := builder.MakeFromSecurity(security, header.NewHeaders())

	if len(result.Auths) > 1 {
		t.Error("More than one authentication found.")
	}

	authResult, ok := result.Auths[auth.Basic.String()]
	if !ok {
		t.Error("Basic authentication not found.")
	}

	value := authResult.Parameters[auth.BASIC_PARAM_USER]
	expected := auth.BASIC_PARAM_USER
	if value != expected {
		t.Errorf("Found variable %v but %v expected", value, expected)
	}

	value = authResult.Parameters[auth.BASIC_PARAM_PASSWORD]
	expected = auth.BASIC_PARAM_PASSWORD
	if value != expected {
		t.Errorf("Found variable %v but %v expected", value, expected)
	}
}

func TestMakeFromSecurityBearer(t *testing.T) {
	oapi, raw := makeOpenApiArguments(t)

	builder := openapi.NewFactoryCollection(TEST_OWNER, oapi).SetRaw(*raw)

	security := oapi.Paths["/request"].Post.Security
	result := builder.MakeFromSecurity(security, header.NewHeaders())

	if len(result.Auths) > 1 {
		t.Error("More than one authentication found.")
	}

	authResult, ok := result.Auths[auth.Bearer.String()]
	if !ok {
		t.Error("Bearer authentication not found.")
	}

	value := authResult.Parameters[auth.BEARER_PARAM_PREFIX]
	expected := "JWT"
	if value != expected {
		t.Errorf("Found variable %v but %v expected", value, expected)
	}
}

func TestMakeFromSecurityApiKey(t *testing.T) {
	oapi, raw := makeOpenApiArguments(t)

	builder := openapi.NewFactoryCollection(TEST_OWNER, oapi).SetRaw(*raw)

	headers := header.NewHeaders()

	security := oapi.Paths["/collection/{userId}"].Get.Security
	result := builder.MakeFromSecurity(security, headers)

	if len(result.Auths) > 0 {
		t.Error("Authentication found.")
	}

	key := "X-API-Key"

	header, ok := headers.Headers[key]
	if !ok || len(header) == 0 {
		t.Errorf("Header '%s' found.", key)
	}

	value := header[0].Value
	if value != key {
		t.Errorf("Found variable %v but %v expected", value, key)
	}
}
