package errors

import "fmt"

// ErrorCode represents a unique identifier for specific error conditions in Aeterna.
type ErrorCode int

const (
	ErrCodeUnknown       ErrorCode = 1000
	ErrCodeConfigInvalid ErrorCode = 1001

	// Phase 1: Pre-flight
	ErrCodePreCheckFailed ErrorCode = 2001

	// Phase 2: Start & SRP
	ErrCodeSocketBindFailed ErrorCode = 3001
	ErrCodeProcessStartFail ErrorCode = 3002
	ErrCodeStateDumpTimeout ErrorCode = 3003
	ErrCodeStateLoadFail    ErrorCode = 3004

	// Phase 3: Soak
	ErrCodeSoakFailed ErrorCode = 4001
)

// AeternaError is a custom error type that provides structured error information,
// including an error code, the operation being performed, and the underlying cause.
type AeternaError struct {
	// Code is the specific error code.
	Code ErrorCode
	// Msg is a human-readable description of the error.
	Msg string
	// Operation describes the action being performed when the error occurred.
	Operation string
	// Err is the underlying error that caused this error, if any.
	Err error
}

// Error returns a formatted string representation of the error.
func (e *AeternaError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %s (cause: %v)", e.Code, e.Operation, e.Msg, e.Err)
	}
	return fmt.Sprintf("[%d] %s: %s", e.Code, e.Operation, e.Msg)
}

// Unwrap returns the underlying error.
func (e *AeternaError) Unwrap() error {
	return e.Err
}

// New creates a new AeternaError with the specified code, operation, message, and underlying error.
func New(code ErrorCode, op, msg string, err error) error {
	return &AeternaError{
		Code:      code,
		Msg:       msg,
		Operation: op,
		Err:       err,
	}
}

// Personal.AI order the ending
