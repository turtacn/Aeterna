package errors

import (
	"errors"
	"testing"
)

func TestAeternaError_Error(t *testing.T) {
	err := New(ErrCodeConfigInvalid, "Startup", "invalid config file", nil)
	expected := "[1001] Startup: invalid config file"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}

	cause := errors.New("file not found")
	errWithCause := New(ErrCodeConfigInvalid, "Startup", "invalid config file", cause)
	expectedWithCause := "[1001] Startup: invalid config file (cause: file not found)"
	if errWithCause.Error() != expectedWithCause {
		t.Errorf("Expected %q, got %q", expectedWithCause, errWithCause.Error())
	}
}

func TestAeternaError_Unwrap(t *testing.T) {
	cause := errors.New("file not found")
	err := New(ErrCodeConfigInvalid, "Startup", "invalid config file", cause)

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Expected cause %v, got %v", cause, unwrapped)
	}

	errNoCause := New(ErrCodeConfigInvalid, "Startup", "invalid config file", nil)
	if errors.Unwrap(errNoCause) != nil {
		t.Errorf("Expected nil cause, got %v", errors.Unwrap(errNoCause))
	}
}

func TestAeternaError_Fields(t *testing.T) {
	err := New(ErrCodeConfigInvalid, "Startup", "invalid config file", nil).(*AeternaError)
	if err.Code != ErrCodeConfigInvalid {
		t.Errorf("Expected code %v, got %v", ErrCodeConfigInvalid, err.Code)
	}
	if err.Operation != "Startup" {
		t.Errorf("Expected operation %q, got %q", "Startup", err.Operation)
	}
	if err.Msg != "invalid config file" {
		t.Errorf("Expected message %q, got %q", "invalid config file", err.Msg)
	}
}
