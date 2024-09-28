package body

import "strings"

type ContentType string

const (
	None ContentType = "None"
	Text ContentType = "Text"
	Form ContentType = "Form"
	Json ContentType = "Json"
	Xml  ContentType = "Xml"
	Html ContentType = "Html"
)

func (s ContentType) String() string {
	return string(s)
}

func RequestContentTypes() []ContentType {
	return []ContentType{
		Text,
		Json,
		Xml,
		Form,
	}
}

func FromContentType(contentType string) (ContentType, bool) {
	contentType = strings.ToLower(contentType)

	if strings.Contains(contentType, "text/plain") {
		return Text, true
	}
	if strings.Contains(contentType, "multipart/form-data") {
		return Form, true
	}
	if strings.Contains(contentType, "application/json") {
		return Json, true
	}
	if strings.Contains(contentType, "text/xml") {
		return Xml, true
	}
	if strings.Contains(contentType, "text/html") {
		return Html, true
	}

	return None, false
}