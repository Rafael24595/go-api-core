package session

type Session struct {
	Username    string `json:"username"`
	Secret      []byte `json:"secret"`
	Timestamp   int64  `json:"timestamp"`
	Context     string `json:"context"`
	IsProtected bool   `json:"is_protected"`
	IsAdmin     bool   `json:"is_admin"`
	Count       int    `json:"count"`
}

func (s Session) IsVerified() bool {
	 return s.Count >= 0
}

func (s Session) IsNotVerified() bool {
	return s.Count < 0
}