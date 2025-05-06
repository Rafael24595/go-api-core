package exception

import "fmt"

type ApiError struct {
	Status  int
	Message string
	Cause   error
}

func NewApiError(status int, message string) *ApiError {
	return NewCauseApiError(status, message, nil)
}

func NewCauseApiError(status int, message string, cause error) *ApiError {
	return &ApiError{
		Status:  status,
		Message: message,
		Cause:   cause,
	}
}

func (e *ApiError) Error() string {
	message := e.Message
	if e.Cause != nil {
		message = fmt.Sprintf("%s -> %s", message, e.Cause)
	}
	return message
}
