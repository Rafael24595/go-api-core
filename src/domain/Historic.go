package domain

import (
	"time"
)

type Historic struct {
	Id        string `json:"_id"`
	Owner     string `json:"owner"`
	Timestamp int64  `json:"timestamp"`
}

func NewHistoricDefault() *Historic {
	return &Historic{}
}

func NewHistoric(id, owner string) *Historic {
	return &Historic{
		Id:        id,
		Owner:     owner,
		Timestamp: time.Now().UnixMilli(),
	}
}

func (r Historic) PersistenceId() string {
	return r.Id
}
