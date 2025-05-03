package body

type Body struct {
	Status      bool                     `json:"status"`
	ContentType ContentType              `json:"content_type"`
	Parameters  map[string]BodyParameter `json:"parameters"`
}

type BodyParameter struct {
	Order    int64  `json:"order"`
	Status   bool   `json:"status"`
	IsFile   bool   `json:"is_file"`
	FileName string `json:"file_name"`
	Value    string `json:"value"`
}

func NewBody(status bool, contentType ContentType, parameters map[string]BodyParameter) *Body {
	return &Body{
		Status:      status,
		ContentType: contentType,
		Parameters:  parameters,
	}
}

func NewBodyDocument(order int64, status bool, value string) BodyParameter {
	return BodyParameter{
		Order:  order,
		Status: status,
		IsFile: false,
		FileName: "",
		Value:  value,
	}
}

func NewBodyParameter(order int64, status, isFile bool, fileName, value string) BodyParameter {
	return BodyParameter{
		Order:  order,
		Status: status,
		IsFile: isFile,
		FileName: fileName,
		Value:  value,
	}
}

func (b Body) Empty() bool {
	return b.ContentType == None
}
