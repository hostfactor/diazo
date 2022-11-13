package except

import (
	"fmt"
	"github.com/hostfactor/api/go/exception"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	return &except{
		Message: fmt.Sprintf(msg, args...),
		Reason:  reason,
	}
}

func NewBlank(reason exception.Reason) error {
	return &except{
		Reason: reason,
	}
}

func NewFromGRPCStatus(s *status.Status) error {
	return &except{
		Message: s.Message(),
		Reason:  CodeToReason(s.Code()),
	}
}

func Is(err error, reasons ...exception.Reason) bool {
	for _, v := range reasons {
		if ReasonFromErr(err) == v {
			return true
		}
	}

	return false
}

func ReasonFromErr(err error) exception.Reason {
	if err != nil {
		return exception.Reason_REASON_UNKNOWN
	}

	v, ok := err.(*except)
	if ok {
		return v.Reason
	}
	return exception.Reason_REASON_INTERNAL
}

func NewFromGRPC(err error) error {
	if err == nil {
		return nil
	}

	r := &except{
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

var _ error = &except{}
var _ ErrorIser = &except{}
var _ fmt.Stringer = &except{}

type ErrorIser interface {
	Is(err error) bool
}

type except exception.Error

func (e *except) String() string {
	if e == nil {
		return ""
	}

	if e.Message == "" {
		return Reason(e.Reason).String()
	}

	return fmt.Sprintf("%s: %s", e.Message, Reason(e.Reason).String())
}

func (e *except) Is(err error) bool {
	if e == nil {
		return false
	}

	v, ok := err.(*except)
	if !ok {
		return false
	}

	return v.Reason == e.Reason
}

func (e *except) ToException() *exception.Error {
	return &exception.Error{
		Message: e.Message,
		Reason:  e.Reason,
	}
}

func (e *except) Error() string {
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