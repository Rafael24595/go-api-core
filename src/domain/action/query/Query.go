package query

type Query struct {
	Order  int64  `json:"order"`
	Status bool   `json:"status"`
	Value  string `json:"value"`
}

func NewQuery(order int64, status bool, value string) Query {
	return Query{
		Order:  order,
		Status: status,
		Value:  value,
	}
}
