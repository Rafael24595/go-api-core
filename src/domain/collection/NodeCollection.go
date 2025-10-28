package collection

type NodeCollection struct {
	Order      int        `json:"order"`
	Collection Collection `json:"collection"`
}
