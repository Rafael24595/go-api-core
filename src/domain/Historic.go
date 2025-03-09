package domain

import (
	"time"
)

type Historic struct {
	Id        string `json:"_id"`
	Timestamp int64  `json:"timestamp"`
}

func NewHistoricDefault() *Historic {
	return &Historic{}
}

func NewHistoric(id string) *Historic {
	return &Historic{
		Id:        id,
		Timestamp: time.Now().UnixMilli(),
	}
}

func (r Historic) PersistenceId() string {
	return r.Id
}
