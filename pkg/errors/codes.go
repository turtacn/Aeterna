package errors

import "fmt"

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

type AeternaError struct {
	Code      ErrorCode
	Msg       string
	Operation string
	Err       error
}

func (e *AeternaError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %s (cause: %v)", e.Code, e.Operation, e.Msg, e.Err)
	}
	return fmt.Sprintf("[%d] %s: %s", e.Code, e.Operation, e.Msg)
}

func (e *AeternaError) Unwrap() error {
	return e.Err
}

func New(code ErrorCode, op, msg string, err error) error {
	return &AeternaError{
		Code:      code,
		Msg:       msg,
		Operation: op,
		Err:       err,
	}
}

// Personal.AI order the ending
