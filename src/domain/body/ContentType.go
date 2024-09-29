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

func (s ContentType) lower() string {
	return strings.ToLower(s.String())
}

func RequestContentTypes() []ContentType {
	return []ContentType{
		Text,
		Json,
		Xml,
		Form,
	}
}

func ContentTypeFromString(contentType string) (ContentType, bool) {
	contentType = strings.ToLower(contentType)

	switch strings.ToLower(contentType) {
	case Text.lower():
		return Text, true
	case Form.lower():
		return Form, true
	case Json.lower():
		return Json, true
	case Xml.lower():
		return Xml, true
	case Html.lower():
		return Html, true
	}

	return None, false
}

func ContentTypeFromHeader(contentType string) (ContentType, bool) {
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
