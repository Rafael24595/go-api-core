package openapi

type OpenAPI struct {
	OpenAPI    string                `json:"openapi"`
	Info       Info                  `json:"info"`
	Servers    []Server              `json:"servers"`
	Paths      map[string]PathItem   `json:"paths"`
	Components Components            `json:"components"`
	Security   []SecurityRequirement `json:"security"`
	Tags       []Tag                 `json:"tags"`
}

type Info struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Version     string  `json:"version"`
	Contact     Contact `json:"contact"`
}

type Contact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

type Operation struct {
	Tags        []string              `json:"tags"`
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []SecurityRequirement `json:"security,omitempty"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Schema      Schema `json:"schema"`
}

type RequestBody struct {
	Required bool                 `json:"required"`
	Content  map[string]MediaType `json:"content"`
	Ref      string               `json:"$ref,omitempty"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type Components struct {
	Schemas         map[string]Schema         `json:"schemas"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes"`
	RequestBodies   map[string]RequestBody    `json:"requestBodies"`
}

type SecurityRequirement map[string][]string

type SecurityScheme struct {
	Type         string `json:"type"`
	In           string `json:"in"`
	Name         string `json:"name"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

type Schema struct {
	Ref        string            `json:"$ref,omitempty"`
	Type       string            `json:"type"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Example    any               `json:"example,omitempty"`
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
