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

func DocumentBody(status bool, contentType ContentType, document string) *BodyRequest {
	parameters := make(map[string]map[string][]BodyParameter)

	parameters[DOCUMENT_PARAM] = make(map[string][]BodyParameter)
	parameters[DOCUMENT_PARAM][PAYLOAD_PARAM] = []BodyParameter{
		NewBodyDocument(0, true, document),
	}

	return NewBody(status, contentType, parameters)
}

func FormDataBody(status bool, contentType ContentType, builder *BuilderFormDataBody) *BodyRequest {
	parameters := make(map[string]map[string][]BodyParameter)
	parameters[FORM_DATA_PARAM] = builder.formData

	return NewBody(status, contentType, parameters)
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

func NewParameter(order int64, status bool, value string) *BodyParameter {
	return NewBodyParameter(order, status, false, "", "", value)
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

type BuilderFormDataBody struct {
	formData map[string][]BodyParameter
}

func NewBuilderFromDataBody() *BuilderFormDataBody {
	return &BuilderFormDataBody{
		formData: make(map[string][]BodyParameter),
	}
}

func (b *BuilderFormDataBody) Add(key string, parameter *BodyParameter) *BuilderFormDataBody {
	var parameters []BodyParameter
	if exists, ok := b.formData[key]; ok {
		parameters = exists
	} else {
		parameters = make([]BodyParameter, 0)
	}

	b.formData[key] = append(parameters, *parameter)
	
	return b
}
