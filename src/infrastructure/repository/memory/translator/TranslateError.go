package translator

import "fmt"

type TranslateError interface {
	error
}

type translateError struct {
	Message string
	Cause   error
}

func TranslateErrorFrom(message string) TranslateError {
	return TranslateErrorFromCause(message, nil)
}

func TranslateErrorFromCause(message string, cause error) TranslateError {
	return &translateError{
		Message: message,
		Cause: cause,
	}
}

func (e *translateError) Error() string {
	message := e.Message
	if e.Cause != nil {
		message = fmt.Sprintf("%s -> %s", message, e.Cause)
	}
	return message
}
