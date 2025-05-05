package dto

type DtoItemContext struct {
	Order   int64  `json:"order"`
	Private bool   `json:"private"`
	Status  bool   `json:"status"`
	Value   string `json:"value"`
}
