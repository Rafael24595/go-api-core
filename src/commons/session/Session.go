package session

type Session struct {
	Username  string `json:"username"`
	Secret    []byte `json:"secret"`
	Timestamp int64  `json:"timestamp"`
	Context   string `json:"context"`
}
