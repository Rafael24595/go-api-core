package query

type Query struct {
	Active bool     `json:"active"`
	Query  []string `json:"query"`
}
