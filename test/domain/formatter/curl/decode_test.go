package curl_test

import (
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain"
	"github.com/Rafael24595/go-api-core/src/domain/action/auth"
	auth_strategy "github.com/Rafael24595/go-api-core/src/domain/action/auth/strategy"
	body_strategy "github.com/Rafael24595/go-api-core/src/domain/action/body/strategy"
	"github.com/Rafael24595/go-api-core/src/domain/formatter/curl"
)

func TestUnmarshal_SimpleGet(t *testing.T) {
	input := `curl https://api.example.com/users`

	req, err := curl.Unmarshal([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "https://api.example.com/users"
	if req.Uri != expected {
		t.Errorf("Found %#v, but %#v expected", req.Uri, expected)
	}

	expectedMethod := domain.GET
	if req.Method != expectedMethod {
		t.Errorf("Found %#v, but %#v expected", req.Method, expectedMethod)
	}
}

func TestUnmarshal_PostData(t *testing.T) {
	input := `curl -X POST https://api.example.com/users -d '{"name":"John"}'`

	req, err := curl.Unmarshal([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedMethod := domain.POST
	if req.Method != expectedMethod {
		t.Errorf("Found %#v, but %#v expected", req.Method, expectedMethod)
	}

	strategy := body_strategy.LoadStrategy(req.Body.ContentType)
	payload, _ := strategy(&req.Body, &req.Query)

	result := payload.String()
	expected := `{"name":"John"}`
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}

func TestUnmarshalRequest_QueryParams(t *testing.T) {
	input := `curl 'https://api.example.com/search?q=test&page=2'`

	req, err := curl.Unmarshal([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Uri != "https://api.example.com/search" {
		t.Errorf("expected URI without query, got %s", req.Uri)
	}

	expected := "test"
	query, ok := req.Query.FindIndex("q", 0)
	if !ok {
		t.Errorf("Expected value '%s', but nothing found", expected)
	}

	if query.Value != expected {
		t.Errorf("Found %#v, but %#v expected", query.Value, expected)
	}	

	expected = "2"
	query, ok = req.Query.FindIndex("page", 0)
	if !ok {
		t.Errorf("Expected value '%s', but nothing found", expected)
	}

	if query.Value != expected {
		t.Errorf("Found %#v, but %#v expected", query.Value, expected)
	}
}

func TestUnmarshal_HeaderCookie(t *testing.T) {
	input := "curl https://api.example.com/data \\\n-H 'Accept: application/json' \\\n-b 'session=abc123; token=xyz'"

	req, err := curl.Unmarshal([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "application/json"
	header, ok := req.Header.FindIndex("Accept", 0)
	if !ok {
		t.Errorf("Expected value '%s', but nothing found", expected)
	}

	if header.Value != expected {
		t.Errorf("Found %#v, but %#v expected", header.Value, expected)
	}

	expected = "abc123"
	cookie, ok := req.Cookie.Find("session")
	if !ok {
		t.Errorf("Expected value '%s', but nothing found", expected)
	}

	if cookie.Value != expected {
		t.Errorf("Found %#v, but %#v expected", cookie.Value, expected)
	}
	
	expected = "xyz"
	cookie, ok = req.Cookie.Find("token")
	if !ok {
		t.Errorf("Expected value '%s', but nothing found", expected)
	}

	if cookie.Value != expected {
		t.Errorf("Found %#v, but %#v expected", cookie.Value, expected)
	}
}

func TestUnmarshal_BasicAuth(t *testing.T) {
	input := `curl -u user:pass https://secure.example.com/`

	req, err := curl.Unmarshal([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	auth, ok := auth_strategy.FindTypeAuth(req.Auth, auth.Basic)
	if !ok {
		t.Fatal("expected basic auth")
	}

	strategy := auth_strategy.LoadStrategy(auth.Type)
	req = strategy(*auth, req)

	expected := "Basic cGFzczo="
	header, ok := req.Header.FindIndex("Authorization", 0)
	if !ok {
		t.Errorf("Expected value '%s', but nothing found", expected)
	}

	if header.Value != expected {
		t.Errorf("Found %#v, but %#v expected", header.Value, expected)
	}
}

func TestUnmarshal_FormData(t *testing.T) {
	input := "curl -F 'avatar=@/tmp/photo.png' \\\n-F 'name=Alice' \\\nhttps://upload.example.com/profile"

	req, err := curl.Unmarshal([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedMethod := domain.POST
	if req.Method != expectedMethod {
		t.Errorf("Found %#v, but %#v expected", req.Method, expectedMethod)
	}

	expected := "/tmp/photo.png"
	data, ok := body_strategy.FindFormDataParameterIndex(&req.Body, "avatar", 0)
	if !ok {
		t.Errorf("Expected value '%s', but nothing found", expected)
	}

	if !data.IsFile {
		t.Error("File expected for 'avatar' form-data")
	}

	if data.Value != expected {
		t.Errorf("Found %#v, but %#v expected", data.Value, expected)
	}

	expected = "Alice"
	data, ok = body_strategy.FindFormDataParameterIndex(&req.Body, "name", 0)
	if !ok {
		t.Errorf("Expected value '%s', but nothing found", expected)
	}

	if data.Value != expected {
		t.Errorf("Found %#v, but %#v expected", data.Value, expected)
	}
}

func TestCurlToRequest_BinaryData(t *testing.T) {
	input := `curl --data-binary '@/tmp/blob.bin' https://files.example.com/upload`

	req, err := curl.Unmarshal([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "https://files.example.com/upload"
	if req.Uri != expected {
		t.Errorf("Found %#v, but %#v expected", req.Uri, expected)
	}

	expectedMethod := domain.POST
	if req.Method != expectedMethod {
		t.Errorf("Found %#v, but %#v expected", req.Method, expectedMethod)
	}

	strategy := body_strategy.LoadStrategy(req.Body.ContentType)
	payload, _ := strategy(&req.Body, &req.Query)

	result := payload.String()
	expected = `/tmp/blob.bin`
	if result != expected {
		t.Errorf("Found %#v, but %#v expected", result, expected)
	}
}
