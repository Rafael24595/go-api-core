package dto

import "github.com/Rafael24595/go-api-core/src/commons/session"

type DtoSession struct {
	Username    string `json:"username"`
	Secret      string `json:"secret"`
	Timestamp   int64  `json:"timestamp"`
	History     string `json:"history"`
	Collection  string `json:"collection"`
	Group       string `json:"group"`
	IsProtected bool   `json:"is_protected"`
	IsAdmin     bool   `json:"is_admin"`
	Count       int    `json:"count"`
	Refresh     string `json:"refresh"`
}

func NewDtoSessionDefault() *DtoSession {
	return &DtoSession{}
}

func (s DtoSession) PersistenceId() string {
	return s.Username
}

func ToSession(dto DtoSession) *session.Session {
	return &session.Session{
		Username:    dto.Username,
		Secret:      []byte(dto.Secret),
		Timestamp:   dto.Timestamp,
		History:     dto.History,
		Collection:  dto.Collection,
		Group:       dto.Group,
		IsProtected: dto.IsProtected,
		IsAdmin:     dto.IsAdmin,
		Count:       dto.Count,
		Refresh:     dto.Refresh,
	}
}

func FromSession(session session.Session) *DtoSession {
	return &DtoSession{
		Username:    session.Username,
		Secret:      string(session.Secret),
		Timestamp:   session.Timestamp,
		History:     session.History,
		Collection:  session.Collection,
		Group:       session.Group,
		IsProtected: session.IsProtected,
		IsAdmin:     session.IsAdmin,
		Count:       session.Count,
		Refresh:     session.Refresh,
	}
}
