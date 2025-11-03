package session

import "slices"

type Role string

const (
	ROLE_ADMIN     Role = "admin"
	ROLE_PROTECTED Role = "protected"
	ROLE_ANONYMOUS Role = "anonymous"
)

type Session struct {
	Username   string   `json:"username"`
	Secret     []byte   `json:"secret"`
	Timestamp  int64    `json:"timestamp"`
	History    string   `json:"history"`
	Collection string   `json:"collection"`
	Group      string   `json:"group"`
	Count      int      `json:"count"`
	Refresh    string   `json:"refresh"`
	Roles      []Role `json:"roles"`
}

func (s Session) HasRole(role Role) bool {
	return slices.Contains(s.Roles, role)
}

func (s Session) IsVerified() bool {
	return s.Count >= 0
}

func (s Session) IsNotVerified() bool {
	return s.Count < 0
}
