package query

type Query struct {
	Active bool     `json:"active"`
	Key    string   `json:"key"`
	Query  []string `json:"query"`
}

func NewQuery(active bool, key string, query ...string) Query {
	return Query{
		Active: active,
		Key: key,
		Query: query,
	}
}