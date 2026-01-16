package session

type RepositorySessionData interface {
	Find(owner string) (*ClientData, bool)
	Insert(data *ClientData) *ClientData
	Update(data *ClientData) (*ClientData, bool)
	Delete(data *ClientData) *ClientData
}
