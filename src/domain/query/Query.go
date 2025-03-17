package query

type Query struct {
	Status bool   `json:"status"`
	Value  string `json:"value"`
}

func NewQuery(status bool, value string) Query {
	return Query{
		Status: status,
		Value:  value,
	}
}
