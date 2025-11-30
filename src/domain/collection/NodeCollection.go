package collection

type NodeCollection struct {
	Order      int        `json:"order"`
	Collection Collection `json:"collection"`
}

type NodeCollectionLite struct {
	Order      int            `json:"order"`
	Collection CollectionLite `json:"collection"`
}
