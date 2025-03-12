package query

type Query struct {
	Status bool   `json:"status"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

func NewQuery(status bool, key string, value string) Query {
	return Query{
		Status: status,
		Key:    key,
		Value:  value,
	}
}
