package dto

type DtoCollection struct {
	Id        string     `json:"_id"`
	Name      string     `json:"name"`
	Timestamp int64      `json:"timestamp"`
	Context   DtoContext `json:"context"`
	Nodes     []DtoNode  `json:"nodes"`
	Owner     string     `json:"owner"`
	Modified  int64      `json:"modified"`
}
