package domain

type Collection struct {
	Id        string          `json:"_id"`
	Name      string          `json:"name"`
	Timestamp int64           `json:"timestamp"`
	Context   string          `json:"context"`
	Nodes     []NodeReference `json:"nodes"`
	Owner     string          `json:"owner"`
	Modified  int64           `json:"modified"`
}

func NewCollectionDefault() *Collection {
	return &Collection{}
}

func (c Collection) PersistenceId() string {
	return c.Id
}
