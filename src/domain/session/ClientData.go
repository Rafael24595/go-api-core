package session

type ClientData struct {
	Owner       string `json:"username"`
	Timestamp   int64  `json:"timestamp"`
	Transient   string `json:"transient"`
	Persistent  string `json:"persistent"`
	Collections string `json:"collections"`
	Modified    int64  `json:"modified"`
}

func NewClientData(owner, transient, persistent, collections string) *ClientData {
	return &ClientData{
		Owner: owner,
		Timestamp: 0,
		Transient: transient,
		Persistent: persistent,
		Collections: collections,
		Modified: 0,
	}
}

func (r ClientData) PersistenceId() string {
	return r.Owner
}
