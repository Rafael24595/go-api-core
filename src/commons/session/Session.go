package session

type Session struct {
	Username    string `json:"username"`
	Secret      []byte `json:"secret"`
	Timestamp   int64  `json:"timestamp"`
	History     string `json:"history"`
	Collection  string `json:"collection"`
	Group       string `json:"group"`
	IsProtected bool   `json:"is_protected"`
	IsAdmin     bool   `json:"is_admin"`
	Count       int    `json:"count"`
	Refresh     string `json:"refresh"`
}

func (s Session) IsVerified() bool {
	return s.Count >= 0
}

func (s Session) IsNotVerified() bool {
	return s.Count < 0
}
