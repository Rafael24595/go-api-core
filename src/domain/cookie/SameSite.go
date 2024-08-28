package cookie

type SameSite int

const (
	Strict SameSite = iota
	Lax
	None
)

func (s SameSite) String() string {
	switch s {
	case Strict:
		return "Strict"
	case Lax:
		return "Lax"
	case None:
		return "None"
	default:
		return "Unknown"
	}
}