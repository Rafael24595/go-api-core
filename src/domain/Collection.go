package domain

type Collection struct {
	Id        string
	Name      string
	Timestamp int64
	Variables map[string]CollectionVariable
	Nodes     []CollectionNode
}

func NewCollection() *Collection {
	return &Collection{}
}

func (c Collection) PersistenceId() string {
	return c.Id
}