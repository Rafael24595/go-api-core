package curl_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action"
	auth_strategy "github.com/Rafael24595/go-api-core/src/domain/action/auth/strategy"
	"github.com/Rafael24595/go-api-core/src/domain/action/body"
	body_strategy "github.com/Rafael24595/go-api-core/src/domain/action/body/strategy"
	"github.com/Rafael24595/go-api-core/src/domain/context"
	"github.com/Rafael24595/go-api-core/src/domain/formatter/curl"
)

func TestMarshalContext_InvalidInput(t *testing.T) {
	ctx := context.NewContext("tester")

	req1 := action.NewRequest("_test_001", "", "http://example.com")
	_, err := curl.MarshalContext(ctx, req1, false)
	if err == nil {
		t.Error("Error expected for invalid method")
	}

	req2 := action.NewRequest("_test_002", domain.GET, "")
	_, err = curl.MarshalContext(ctx, req2, false)
	if err == nil {
		t.Error("Error expected for invalid URI")
	}
}

func TestMarshalContext_NoQueries(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_001", domain.GET, "http://example.com")

	curl, err := curl.MarshalContext(ctx, req, false)

	if err != nil {
		t.Error(err)
	}

	expected := "curl -X GET http://example.com"
	if curl != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, curl)
	}
}

func TestMarshalContext_WithQueries(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_004", domain.GET, "http://example.com")
	req.Query.AddStatus("user", "username", true)
	req.Query.AddStatus("id", "001", true)
	req.Query.AddStatus("id", "002", false)

	curl, err := curl.MarshalContext(ctx, req, false)

	if err != nil {
		t.Error(err)
	}

	expected1 := "curl -X GET http://example.com?user=username&id=001"
	expected2 := "curl -X GET http://example.com?id=001&user=username"
	if curl != expected1 && curl != expected2 {
		t.Errorf("Expected '%s', but got '%s'", expected1, curl)
	}
}

func TestMarshalContext_WithHeaders(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_headers_001", domain.GET, "http://example.com")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer 123")

	curl, err := curl.MarshalContext(ctx, req, true)

	if err != nil {
		t.Error(err)
	}

	expected1 := `-H "Content-Type: application/json"`
	expected2 := `-H "Authorization: Bearer 123"`

	if !strings.Contains(curl, expected1) {
		t.Errorf("Expected header '%s' not found in curl: %s", expected1, curl)
	}
	if !strings.Contains(curl, expected2) {
		t.Errorf("Expected header '%s' not found in curl: %s", expected2, curl)
	}
}

func TestMarshalContext_HeadersDisabled(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_headers_002", domain.GET, "http://example.com")
	req.Header.AddStatus("Authorization", "Bearer 123", false)

	curl, err := curl.MarshalContext(ctx, req, true)

	if err != nil {
		t.Error(err)
	}

	unexpected := `-H "Authorization: Bearer 123"`
	if strings.Contains(curl, unexpected) {
		t.Errorf("Unexpected header '%s' found in curl: %s", unexpected, curl)
	}
}

func TestMarshalContext_WithCookies(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_headers_001", domain.GET, "http://example.com")
	req.Cookie.Put("sessionid", "123abc")
	req.Cookie.Put("theme", "marinego")

	curl, err := curl.MarshalContext(ctx, req, true)

	if err != nil {
		t.Error(err)
	}

	expected1 := `-H "Cookie: sessionid=123abc; theme=marinego"`
	expected2 := `-H "Cookie: theme=marinego; sessionid=123abc"`
	if !strings.Contains(curl, expected1) && !strings.Contains(curl, expected2) {
		t.Errorf("Expected header '%s' not found in curl: %s", expected1, curl)
	}
}

func TestMarshalContext_WithDisabledCookies(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_headers_001", domain.GET, "http://example.com")
	req.Cookie.PutStatus("sessionid", "123abc", false)

	curl, err := curl.MarshalContext(ctx, req, true)

	if err != nil {
		t.Error(err)
	}

	expected := `-H "Cookie: sessionid=123abc"`
	if strings.Contains(curl, expected) {
		t.Errorf("Expected header '%s' not found in curl: %s", expected, curl)
	}
}

func TestMarshalContext_WithBasicAuth(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_headers_001", domain.GET, "http://example.com")

	req.Auth.Status = true
	req.Auth.PutAuth(*auth_strategy.BasicAuth(true, "username", "123"))

	curl, err := curl.MarshalContext(ctx, req, true)

	if err != nil {
		t.Error(err)
	}

	expected := `-H "Authorization: Basic MTIzOg=="`
	if !strings.Contains(curl, expected) {
		t.Errorf("Expected header '%s' not found in curl: %s", expected, curl)
	}
}

func TestMarshalContext_WithBearerAuth(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_headers_001", domain.GET, "http://example.com")

	req.Auth.Status = true
	req.Auth.PutAuth(*auth_strategy.BearerAuth(true, "Bearer", "123"))

	curl, err := curl.MarshalContext(ctx, req, true)

	if err != nil {
		t.Error(err)
	}

	expected := `-H "Authorization: Bearer 123"`
	if !strings.Contains(curl, expected) {
		t.Errorf("Expected header '%s' not found in curl: %s", expected, curl)
	}
}

func TestMarshalContext_WithDisabledAuth(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_headers_001", domain.GET, "http://example.com")

	req.Auth.Status = true
	req.Auth.PutAuth(*auth_strategy.BearerAuth(false, "Bearer", "123"))

	curl, err := curl.MarshalContext(ctx, req, true)

	if err != nil {
		t.Error(err)
	}

	unexpected := `-H "Authorization: Bearer 123"`
	if strings.Contains(curl, unexpected) {
		t.Errorf("Unexpected header '%s' found in curl: %s", unexpected, curl)
	}
}

func TestMarshalContext_WithDocumentBody(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_headers_001", domain.GET, "http://example.com")

	req.Body.Status = true
	req.Body = *body_strategy.DocumentBody(true, domain.Json, `{"id": 001, "user": "username"}`)

	curl, err := curl.MarshalContext(ctx, req, true)

	if err != nil {
		t.Error(err)
	}

	expected := `-d '{"id": 001, "user": "username"}'`
	if !strings.Contains(curl, expected) {
		t.Errorf("Expected header '%s' not found in curl: %s", expected, curl)
	}
}

func TestMarshalContext_WithDisabledDocumentBody(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_headers_001", domain.GET, "http://example.com")

	req.Body.Status = true
	req.Body = *body_strategy.DocumentBody(false, domain.Json, `{"id": 001, "user": "username"}`)

	curl, err := curl.MarshalContext(ctx, req, true)

	if err != nil {
		t.Error(err)
	}

	unexpected := `-d '{"id": 001, "user": "username"}'`
	if strings.Contains(curl, unexpected) {
		t.Errorf("Unexpected header '%s' found in curl: %s", unexpected, curl)
	}
}

func TestMarshalContext_WithFormDataBody(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_headers_001", domain.GET, "http://example.com")

	builder := body_strategy.NewBuilderFromDataBody()
	builder.Add("user", body.NewParameter(0, true, "username"))
	builder.Add("id", body.NewParameter(0, true, "001"))

	fileKey := "log_file"
	fileName := "exceptions.log"
	base64 := "SXQncyBqdXN0IHdvcmtz"
	builder.Add(fileKey, body.NewFileParameter(0, true, "logs", fileName, base64))

	req.Body.Status = true
	req.Body = *body_strategy.FormDataBody(true, domain.Form, builder)

	curl, err := curl.MarshalContext(ctx, req, true)

	if err != nil {
		t.Error(err)
	}

	expected1 := `-F "user=username"`
	if !strings.Contains(curl, expected1) {
		t.Errorf("Expected header '%s' not found in curl: %s", expected1, curl)
	}

	expected2 := `-F "id=001"`
	if !strings.Contains(curl, expected2) {
		t.Errorf("Expected header '%s' not found in curl: %s", expected2, curl)
	}

	expected3 := fmt.Sprintf(`-F "%s=@<(echo '%s' | base64 --decode);filename=%s"`, fileKey, base64, fileName)
	if !strings.Contains(curl, expected3) {
		t.Errorf("Expected header '%s' not found in curl: %s", expected3, curl)
	}
}

func TestMarshalContext_WithDisabledFormDataBody(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_headers_001", domain.GET, "http://example.com")

	builder := body_strategy.NewBuilderFromDataBody()
	builder.Add("user", body.NewParameter(0, false, "username"))

	req.Body.Status = true
	req.Body = *body_strategy.FormDataBody(true, domain.Form, builder)

	curl, err := curl.MarshalContext(ctx, req, true)

	if err != nil {
		t.Error(err)
	}

	unexpected := `-F "user=username"`
	if strings.Contains(curl, unexpected) {
		t.Errorf("Unexpected header '%s' found in curl: %s", unexpected, curl)
	}
}

func TestMarshalContext_InlineVsMultiline(t *testing.T) {
	ctx := context.NewContext("tester")

	req := action.NewRequest("_test_005", domain.POST, "http://api.test")
	req.Header.Add("Content-Type", "application/json")

	curlMulti, err := curl.MarshalContext(ctx, req, false)
	if err != nil {
		t.Error(err)
	}

	expectedMult := "curl -X POST http://api.test \\\n -H \"Content-Type: application/json\""
	if curlMulti != expectedMult {
		t.Errorf("Expected '%s', but got '%s'", expectedMult, curlMulti)
	}

	curlInline, err := curl.MarshalContext(ctx, req, true)
	if err != nil {
		t.Error(err)
	}

	expectedLine := "curl -X POST http://api.test -H \"Content-Type: application/json\""
	if curlInline != expectedLine {
		t.Errorf("Expected '%s', but got '%s'", expectedLine, curlInline)
	}
}
