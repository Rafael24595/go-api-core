package body

type ContentType int

const (
	None ContentType = iota
	Text
	Form
	Json
	Xml
)