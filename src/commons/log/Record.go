package log

type Category string

const (
	MESSAGE Category = "MESSAGE"
	WARNING Category = "WARNING"
	ERROR Category = "ERROR"
	PANIC Category = "PANIC"
)

type Record struct {
	Category Category `json:"category" bson:"category"`
	Message string `json:"message" bson:"message"`
	Timestamp int64 `json:"timestamp" bson:"timestamp"`
}
