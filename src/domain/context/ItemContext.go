package context

type ItemContext struct {
	Order  int64  `json:"order"`
	Status bool   `json:"status"`
	Value  string `json:"value"`
}

func NewItemContext(order int64, status bool, value string) ItemContext {
	return ItemContext{
		Order:  order,
		Status: status,
		Value:  value,
	}
}
