package primer

import (
	"fmt"
)

// Error SDK 错误类型
type Error struct {
	Message    string
	StatusCode int
	ErrorID    string
	Err        error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

// NewError 创建新错误
func NewError(message string) *Error {
	return &Error{Message: message}
}

// WrapError 包装错误
func WrapError(err error, message string) *Error {
	return &Error{
		Message: message,
		Err:     err,
	}
}

// WrapErrorf 格式化包装错误
func WrapErrorf(err error, format string, args ...any) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}
