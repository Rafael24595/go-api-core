package body

type BodyRequest struct {
	Status      bool                                  `json:"status"`
	ContentType ContentType                           `json:"content_type"`
	Parameters  map[string]map[string][]BodyParameter `json:"parameters"`
}

type BodyResponse struct {
	ContentType ContentType `json:"content_type"`
	Payload     string      `json:"payload"`
}

type BodyParameter struct {
	Order    int64  `json:"order"`
	Status   bool   `json:"status"`
	IsFile   bool   `json:"is_file"`
	FileType string `json:"file_type"`
	FileName string `json:"file_name"`
	Value    string `json:"value"`
}

func NewResponseBody(contentType ContentType, payload string) *BodyResponse {
	return &BodyResponse{
		ContentType: contentType,
		Payload:     payload,
	}
}

func EmptyResponseBody(contentType ContentType) *BodyResponse {
	return NewResponseBody(contentType, "")
}

func EmptyBody(status bool, contentType ContentType) *BodyRequest {
	return NewBody(status, contentType, make(map[string]map[string][]BodyParameter))
}

func NewBody(status bool, contentType ContentType, parameters map[string]map[string][]BodyParameter) *BodyRequest {
	return &BodyRequest{
		Status:      status,
		ContentType: contentType,
		Parameters:  parameters,
	}
}

func NewBodyDocument(order int64, status bool, value string) BodyParameter {
	return BodyParameter{
		Order:    order,
		Status:   status,
		IsFile:   false,
		FileType: "",
		FileName: "",
		Value:    value,
	}
}

func NewParameterActive(value string) *BodyParameter {
	return NewParameter(0, true, value)
}

func NewParameter(order int64, status bool, value string) *BodyParameter {
	return NewBodyParameter(order, status, false, "", "", value)
}

func NewFileParameterActive(fileType, fileName, value string) *BodyParameter {
	return NewFileParameter(0, true, fileType, fileName, value)
}

func NewFileParameter(order int64, status bool, fileType, fileName, value string) *BodyParameter {
	return NewBodyParameter(order, status, true, fileType, fileName, value)
}

func NewBodyParameter(order int64, status, isFile bool, fileType, fileName, value string) *BodyParameter {
	return &BodyParameter{
		Order:    order,
		Status:   status,
		IsFile:   isFile,
		FileType: fileType,
		FileName: fileName,
		Value:    value,
	}
}

func (b BodyRequest) Empty() bool {
	return b.ContentType == None
}
