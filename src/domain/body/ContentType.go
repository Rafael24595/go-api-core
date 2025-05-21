package body

import (
	"bytes"
	"strings"
)

type ContentType string

const (
	None ContentType = "none"
	Text ContentType = "text"
	Form ContentType = "form"
	Json ContentType = "json"
	Xml  ContentType = "xml"
	Html ContentType = "html"
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

func (c ContentType) LoadStrategy() func(a *BodyRequest) *bytes.Buffer {
	switch c {
	case Form:
		return applyFormData
	default:
		return applyDefault
	}
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
	if strings.Contains(contentType, "text/xml") || 
		strings.Contains(contentType, "application/xml") ||
		strings.Contains(contentType, "application/xhtml+xml") {
		return Xml, true
	}
	if strings.Contains(contentType, "text/html") {
		return Html, true
	}

	return None, false
}
