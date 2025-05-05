package body

type BodyRequest struct {
	Status      bool                                  `json:"status"`
	ContentType ContentType                           `json:"content_type"`
	Parameters  map[string]map[string][]BodyParameter `json:"parameters"`
}

type BodyResponse struct {
	Status      bool        `json:"status"`
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

func NewBodyParameter(order int64, status, isFile bool, fileType, fileName, value string) BodyParameter {
	return BodyParameter{
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
