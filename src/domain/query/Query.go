package query

type Query struct {
	Active bool   `json:"active"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

func NewQuery(active bool, key string, value string) Query {
	return Query{
		Active: active,
		Key:    key,
		Value:  value,
	}
}
