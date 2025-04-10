package exception

import "fmt"

type ApiError interface {
	error
}

type apiErrorImpl struct {
	Status  int16
	Message string
	Cause   error
}

func ApiErrorFrom(status int16, message string) ApiError {
	return ApiErrorFromCause(status, message, nil)
}

func ApiErrorFromCause(status int16, message string, cause error) ApiError {
	return &apiErrorImpl{
		Status: status,
		Message: message,
		Cause: cause,
	}
}

func (e *apiErrorImpl) Error() string {
	message := e.Message
	if e.Cause != nil {
		message = fmt.Sprintf("%s -> %s", message, e.Cause)
	}
	return message
}
