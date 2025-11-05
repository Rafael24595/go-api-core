package assert

import (
	"cmp"
	"fmt"
	"testing"
)

func Equal[T comparable](t *testing.T, want, have T, message ...any) {
	t.Helper()

	if want == have {
		return
	}

	if len(message) > 0 {
		if format, ok := message[0].(string); ok {
			msg := fmt.Sprintf(format, message[1:]...)
			t.Errorf("%s — expected: '%v', got: '%v'", msg, want, have)
			return
		}
	}

	t.Errorf("Expected '%v', but got '%v'", want, have)
}

func GreaterOrEqual[T cmp.Ordered](t *testing.T, want, have T, message ...any) {
	t.Helper()

	if have >= want {
		return
	}

	if len(message) > 0 {
		if format, ok := message[0].(string); ok {
			msg := fmt.Sprintf(format, message[1:]...)
			t.Errorf("%s — expected greater or equal than %v, got: %v", msg, want, have)
			return
		}
	}

	t.Errorf("Expected greater or equal than %v, but got %v", want, have)
}

func LessOrEqual[T cmp.Ordered](t *testing.T, want, have T, message ...any) {
	t.Helper()

	if have <= want {
		return
	}

	if len(message) > 0 {
		if format, ok := message[0].(string); ok {
			msg := fmt.Sprintf(format, message[1:]...)
			t.Errorf("%s — expected less or equal than %v, got: %v", msg, want, have)
			return
		}
	}

	t.Errorf("Expected less or equal than %v, but got %v", want, have)
}
