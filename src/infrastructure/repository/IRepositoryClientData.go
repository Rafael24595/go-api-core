package repository

import (
	"github.com/Rafael24595/go-api-core/src/domain/client"
)

type IRepositoryClientData interface {
	Find(owner string) (*client.ClientData, bool)
	Insert(data *client.ClientData) *client.ClientData
	Update(data *client.ClientData) (*client.ClientData, bool)
	Delete(data *client.ClientData) *client.ClientData
}
