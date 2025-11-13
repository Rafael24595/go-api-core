package mock

type Response struct {
	Status  int      `json:"status"`
	Name    string   `json:"name"`
	Headers []Header `json:"headers"`
	Body    string   `json:"body"`
}

type Header struct {
	Status bool   `json:"status"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

type ResponseFull struct {
	Status    int             `json:"status"`
	Condition []ConditionStep `json:"condition"`
	Name      string          `json:"name"`
	Headers   []Header        `json:"headers"`
	Body      string          `json:"body"`
}

type ConditionStep struct {
	Order int      `json:"order"`
	Type  StepType `json:"type"`
	Value string   `json:"value"`
}

func NewConditionStep(typ StepType, value string) *ConditionStep {
	return &ConditionStep{
		Order: 0,
		Type:  typ,
		Value: value,
	}
}
