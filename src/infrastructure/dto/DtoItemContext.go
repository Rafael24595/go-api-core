package dto

type DtoItemContext struct {
	Order  int64  `json:"order"`
	Status bool   `json:"status"`
	Value  string `json:"value"`
}
