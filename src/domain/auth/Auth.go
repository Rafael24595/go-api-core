package auth

type Auth struct {
	Active     bool                 `json:"active"`
	Type       Type                 `json:"type"`
	Parameters map[string]Parameter `json:"parameters"`
}
