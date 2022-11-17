package except

import (
	"fmt"
	"github.com/hostfactor/api/go/exception"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

var (
	ErrNotFound      = NewBlank(exception.Reason_REASON_NOT_FOUND)
	ErrAlreadyExists = NewBlank(exception.Reason_REASON_ALREADY_EXISTS)
	ErrInternal      = NewBlank(exception.Reason_REASON_INTERNAL)
	ErrTimeout       = NewBlank(exception.Reason_REASON_TIMEOUT)
	ErrUnauthorized  = NewBlank(exception.Reason_REASON_UNAUTHORIZED)
	ErrInvalid       = NewBlank(exception.Reason_REASON_INVALID)
)

func New(reason exception.Reason, msg string, args ...any) error {
	return &Error{
		Message: fmt.Sprintf(msg, args...),
		Reason:  reason,
	}
}

func NewNotFound(msg string, args ...any) error {
	return New(exception.Reason_REASON_NOT_FOUND, msg, args...)
}

func NewNotAlreadyExists(msg string, args ...any) error {
	return New(exception.Reason_REASON_ALREADY_EXISTS, msg, args...)
}

func NewInternal(msg string, args ...any) error {
	return New(exception.Reason_REASON_INTERNAL, msg, args...)
}

func NewTimeout(msg string, args ...any) error {
	return New(exception.Reason_REASON_TIMEOUT, msg, args...)
}

func NewUnauthorized(msg string, args ...any) error {
	return New(exception.Reason_REASON_UNAUTHORIZED, msg, args...)
}

func NewInvalid(msg string, args ...any) error {
	return New(exception.Reason_REASON_INVALID, msg, args...)
}

func NewBlank(reason exception.Reason) error {
	return &Error{
		Reason: reason,
	}
}

func NewFromGRPCStatus(s *status.Status) error {
	return &Error{
		Message: s.Message(),
		Reason:  CodeToReason(s.Code()),
	}
}

func ToGRPC(err error) error {
	v, ok := err.(*Error)
	if !ok {
		return status.Error(ReasonToCode(exception.Reason_REASON_INTERNAL), err.Error())
	}
	return status.Error(ReasonToCode(v.Reason), v.Message)
}

func ReasonToHttpStatus(reason exception.Reason) int {
	switch reason {
	case exception.Reason_REASON_NOT_FOUND:
		return http.StatusNotFound
	case exception.Reason_REASON_ALREADY_EXISTS, exception.Reason_REASON_INVALID:
		return http.StatusBadRequest
	case exception.Reason_REASON_TIMEOUT:
		return http.StatusRequestTimeout
	case exception.Reason_REASON_UNAUTHORIZED:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

func ReasonFromHttpStatus(status int) exception.Reason {
	if status >= 500 {
		return exception.Reason_REASON_INTERNAL
	} else if status >= 400 {
		switch status {
		case http.StatusNotFound:
			return exception.Reason_REASON_NOT_FOUND
		case http.StatusUnauthorized:
			return exception.Reason_REASON_UNAUTHORIZED
		case http.StatusRequestTimeout:
			return exception.Reason_REASON_TIMEOUT
		case http.StatusBadRequest:
			return exception.Reason_REASON_INVALID
		case http.StatusConflict:
			return exception.Reason_REASON_ALREADY_EXISTS
		default:
			return exception.Reason_REASON_INTERNAL
		}
	}

	return exception.Reason_REASON_UNKNOWN
}

func Is(err error, reasons ...exception.Reason) bool {
	reason := ReasonFromErr(err)
	for _, v := range reasons {
		if reason == v {
			return true
		}
	}

	return false
}

func ReasonFromErr(err error) exception.Reason {
	if err == nil {
		return exception.Reason_REASON_UNKNOWN
	}

	v, ok := err.(*Error)
	if ok {
		return v.Reason
	}
	return exception.Reason_REASON_INTERNAL
}

func NewFromGRPC(err error) error {
	if err == nil {
		return nil
	}

	r := &Error{
		Message: "Internal error",
		Reason:  exception.Reason_REASON_INTERNAL,
	}

	v, ok := status.FromError(err)
	if ok {
		return NewFromGRPCStatus(v)
	}
	return r
}

func ReasonToCode(reason exception.Reason) codes.Code {
	switch reason {
	case exception.Reason_REASON_NOT_FOUND:
		return codes.NotFound
	case exception.Reason_REASON_ALREADY_EXISTS:
		return codes.AlreadyExists
	case exception.Reason_REASON_TIMEOUT:
		return codes.DeadlineExceeded
	case exception.Reason_REASON_INVALID:
		return codes.InvalidArgument
	case exception.Reason_REASON_UNAUTHORIZED:
		return codes.Unauthenticated
	default:
		return codes.Internal
	}
}

func CodeToReason(code codes.Code) exception.Reason {
	switch code {
	case codes.Unauthenticated, codes.PermissionDenied:
		return exception.Reason_REASON_UNAUTHORIZED
	case codes.NotFound, codes.Unknown:
		return exception.Reason_REASON_NOT_FOUND
	case codes.AlreadyExists:
		return exception.Reason_REASON_ALREADY_EXISTS
	case codes.DeadlineExceeded:
		return exception.Reason_REASON_TIMEOUT
	case codes.InvalidArgument:
		return exception.Reason_REASON_INVALID
	default:
		return exception.Reason_REASON_INTERNAL
	}
}

var _ error = &Error{}
var _ ErrorIser = &Error{}
var _ fmt.Stringer = &Error{}

type ErrorIser interface {
	Is(err error) bool
}

type Error exception.Error

func (e *Error) String() string {
	if e == nil {
		return ""
	}

	if e.Message == "" {
		return Reason(e.Reason).String()
	}

	return fmt.Sprintf("%s: %s", e.Message, Reason(e.Reason).String())
}

func (e *Error) Is(err error) bool {
	if e == nil {
		return false
	}

	v, ok := err.(*Error)
	if !ok {
		return false
	}

	return v.Reason == e.Reason
}

func (e *Error) ToException() *exception.Error {
	return &exception.Error{
		Message: e.Message,
		Reason:  e.Reason,
	}
}

func (e *Error) Error() string {
	return e.String()
}

var _ fmt.Stringer = Reason(0)

type Reason exception.Reason

func (r Reason) String() string {
	switch exception.Reason(r) {
	case exception.Reason_REASON_NOT_FOUND:
		return "not found"
	case exception.Reason_REASON_ALREADY_EXISTS:
		return "already exists"
	case exception.Reason_REASON_TIMEOUT:
		return "timeout"
	case exception.Reason_REASON_INVALID:
		return "invalid"
	case exception.Reason_REASON_UNAUTHORIZED:
		return "unauthorized"
	case exception.Reason_REASON_INTERNAL:
		return "internal"
	}
	return "internal"
}
