package parser

type GroupCategory string

const(
	MAP GroupCategory = "MAP"
	ARR GroupCategory = "ARR"
	STR GroupCategory = "STR"
	OBJ GroupCategory = "OBJ"
)

func (m GroupCategory) String() string {
	return string(m)
}