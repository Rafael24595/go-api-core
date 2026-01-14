package session

import "slices"

type Role string

const SYSTEM_USER string = "system"

const (
	ROLE_ADMIN     Role = "admin"
	ROLE_PROTECTED Role = "protected"
	ROLE_ANONYMOUS Role = "anonymous"
)

var PUBLIC_ROLES = []Role{
	ROLE_ADMIN,
}

func RoleFromString(role string) (Role, bool) {
	switch role {
	case "admin":
		return ROLE_ADMIN, true
	case "protected":
		return ROLE_PROTECTED, true
	case "anonymous":
		return ROLE_ANONYMOUS, true
	default:
		return "", false
	}
}

func IsPrivateRole(role Role) bool {
	return !slices.Contains(PUBLIC_ROLES, role)
}

type Session struct {
	Username  string `json:"username"`
	Secret    []byte `json:"secret"`
	Timestamp int64  `json:"timestamp"`
	Publisher string `json:"publisher"`
	Count     int    `json:"count"`
	Refresh   string `json:"refresh"`
	Roles     []Role `json:"roles"`
}

type SessionLite struct {
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
	Publisher string `json:"publisher"`
	Count     int    `json:"count"`
	Roles     []Role `json:"roles"`
}

type SessionSafe struct {
	Username  string `json:"username"`
	Timestamp int64  `json:"timestamp"`
	Count     int    `json:"count"`
	Roles     []Role `json:"roles"`
}


func ToLite(session Session) SessionLite {
	return SessionLite{
		Username:  session.Username,
		Timestamp: session.Timestamp,
		Publisher: session.Publisher,
		Count:     session.Count,
		Roles:     session.Roles,
	}
}

func ToSafe(session Session) SessionSafe {
	return SessionSafe{
		Username:  session.Username,
		Timestamp: session.Timestamp,
		Count:     session.Count,
		Roles:     session.Roles,
	}
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
