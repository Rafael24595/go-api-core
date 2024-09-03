package body

type ContentType string

const (
	None ContentType = "None"
	Text ContentType = "Text"
	Form ContentType = "Form"
	Json ContentType = "Json"
	Xml  ContentType = "Xml"
)

func (s ContentType) String() string {
	return string(s)
}
