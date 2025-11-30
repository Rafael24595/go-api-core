package dto

import "github.com/Rafael24595/go-api-core/src/commons/session"

type DtoSession struct {
	Username    string         `json:"username"`
	Secret      string         `json:"secret"`
	Timestamp   int64          `json:"timestamp"`
	IsProtected bool           `json:"is_protected"`
	IsAdmin     bool           `json:"is_admin"`
	Count       int            `json:"count"`
	Refresh     string         `json:"refresh"`
	Roles       []session.Role `json:"roles"`
}

func NewDtoSessionDefault() *DtoSession {
	return &DtoSession{}
}

func (s DtoSession) PersistenceId() string {
	return s.Username
}

func ToSession(dto DtoSession) *session.Session {
	return &session.Session{
		Username:  dto.Username,
		Secret:    []byte(dto.Secret),
		Timestamp: dto.Timestamp,
		Count:     dto.Count,
		Refresh:   dto.Refresh,
		Roles:     dto.Roles,
	}
}

func FromSession(session session.Session) *DtoSession {
	return &DtoSession{
		Username:  session.Username,
		Secret:    string(session.Secret),
		Timestamp: session.Timestamp,
		Count:     session.Count,
		Refresh:   session.Refresh,
		Roles:     session.Roles,
	}
}
