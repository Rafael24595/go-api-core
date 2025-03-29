package domain

type Node struct {
	Order   int     `json:"order"`
	Request Request `json:"request"`
}
