package context

type ItemContext struct {
	Order   int64  `json:"order"`
	Private bool   `json:"private"`
	Status  bool   `json:"status"`
	Value   string `json:"value"`
}

func NewItemContext(order int64, private, status bool, value string) ItemContext {
	return ItemContext{
		Order:   order,
		Private: private,
		Status:  status,
		Value:   value,
	}
}
