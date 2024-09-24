package auth

type Type string

const (
	Basic  Type = "BASIC"
	Bearer Type = "BEARER"
	Cookie Type = "COOKIE"
)

func (t Type) String() string {
	return string(t)
}
