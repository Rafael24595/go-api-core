package dto

import "github.com/Rafael24595/go-api-core/src/commons/session"

type DtoSession struct {
	Username    string `json:"username"`
	Secret      string `json:"secret"`
	Timestamp   int64  `json:"timestamp"`
	Context     string `json:"context"`
	IsProtected bool   `json:"is_protected"`
	IsAdmin     bool   `json:"is_admin"`
	Count       int    `json:"count"`
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
		Context:     dto.Context,
		IsProtected: dto.IsProtected,
		IsAdmin:     dto.IsAdmin,
		Count:       dto.Count,
	}
}

func FromSession(session session.Session) *DtoSession {
	return &DtoSession{
		Username:    session.Username,
		Secret:      string(session.Secret),
		Timestamp:   session.Timestamp,
		Context:     session.Context,
		IsProtected: session.IsProtected,
		IsAdmin:     session.IsAdmin,
		Count:       session.Count,
	}
}
