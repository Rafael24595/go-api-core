package context

type ItemContext struct {
	Status   bool   `json:"status"`
	Value    string `json:"value"`
}

func NewItemContext(status bool, value string) ItemContext {
	return ItemContext{
		Status: status,
		Value:  value,
	}
}
