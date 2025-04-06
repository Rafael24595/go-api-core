package dto

type DtoNode struct {
	Order   int        `json:"order"`
	Request DtoRequest `json:"request"`
}
