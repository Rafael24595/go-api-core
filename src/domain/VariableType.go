package domain

type VariableType string

const (
	INT     VariableType = "INT"
	FLOAT   VariableType = "FLOAT"
	STRING  VariableType = "STRING"
	BOOLEAN VariableType = "BOOLEAN"
)

func (v VariableType) String() string {
	return string(v)
}

func VariableTypes() []VariableType {
	return []VariableType{
		INT, FLOAT, STRING, BOOLEAN,
	}
}
