package werror

import (
	"errors"
	"fmt"
	h "net/http"
	"strings"
)

// Stringer interface for types with a String() method.
type Stringer interface {
	String() string
}

// ToError converts any value to an error.
func ToError(x interface{}) error {
	switch v := x.(type) {
	case error:
		return v
	case Stringer:
		return errors.New(v.String())
	default:
		return ErrInternalError
	}
}

// Err is the base error type.
type Err struct { //nolint:errname // lib
	error

	HttpStatus int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

// NewBaseErr creates a new base Err.
func NewBaseErr(httpStatus int, code, msg string) *Err {
	return &Err{
		error:      fmt.Errorf("%s %s", code, msg),
		HttpStatus: httpStatus,
		Code:       code,
		Message:    msg,
	}
}

// NewBaseErrFrom creates a new base Err from another base Err.
func NewBaseErrFrom(base Err, code, msg string) *Err {
	if strings.TrimSpace(code) == "" {
		code = base.Code
	}
	if strings.TrimSpace(msg) == "" {
		msg = base.Message
	}
	err := &Err{
		error:      fmt.Errorf("%w: %s %s", base.error, code, msg),
		HttpStatus: base.HttpStatus,
		Code:       code,
		Message:    msg,
	}
	return err
}

// NewErr creates a new Err from a base Err.
func NewErr(base Err, msg, msgDetail string) *Err {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		msg = base.Message
	}
	msgDetail = strings.TrimSpace(msgDetail)
	if msgDetail != "" {
		msg = msg + ": " + msgDetail
	}
	return &Err{
		error:      fmt.Errorf("%w: %s", base.error, msg),
		HttpStatus: base.HttpStatus,
		Code:       base.Code,
		Message:    msg,
	}
}

func (e *Err) Error() string {
	return fmt.Sprintf("%v: %s", e.HttpStatus, e.error.Error())
}

func (e *Err) Is(target error) bool {
	return errors.Is(e.error, target)
}

func (e *Err) As(target any) bool {
	return errors.As(e.error, target)
}

// IsErrOf checks if err wraps *Err and has the given code.
func IsErrOf(err error, code string) bool {
	var e *Err
	ok := errors.As(err, &e)
	return ok && e.Code == code
}

// References:
// https://github.com/microsoft/api-guidelines/blob/vNext/azure/Guidelines.md#handling-errors
// https://docs.microsoft.com/en-us/rest/api/storageservices/common-rest-api-error-codes
// https://docs.azure.cn/en-us/cdn/cdn-api-get-endpoint

const (
	StatusClientClosedRequest = 499
)

// Base Errs.
var (
	ErrBadRequest       = NewBaseErr(h.StatusBadRequest, "BadRequest", "Bad request")
	ErrBadArgument      = NewBaseErr(h.StatusBadRequest, "BadArgument", "Bad argument")
	ErrInvalidInput     = NewBaseErr(h.StatusBadRequest, "InvalidInput", "Some request inputs are not valid")
	ErrInvalidOperation = NewBaseErr(
		h.StatusBadRequest,
		"InvalidOperation",
		"The attempted operation is invalid",
	)
	ErrPasswordTooWeak        = NewBaseErr(h.StatusBadRequest, "PasswordTooWeak", "The specified password is too weak")
	ErrUnauthorized           = NewBaseErr(h.StatusUnauthorized, "Unauthorized", "Unauthorized")
	ErrInvalidLoginCredential = NewBaseErr(
		h.StatusUnauthorized,
		"InvalidLoginCredential",
		"The login credential is invalid",
	)
	ErrAlreadyLoggedIn = NewBaseErr(
		h.StatusUnauthorized,
		"AlreadyLoggedIn",
		"User already logged in in another place",
	)
	ErrInvalidAuthenticationInfo = NewBaseErr(
		h.StatusUnauthorized,
		"InvalidAuthenticationInfo",
		"The authentication information is invalid",
	)
	ErrForbidden            = NewBaseErr(h.StatusForbidden, "Forbidden", "Forbidden")
	ErrAuthenticationFailed = NewBaseErr(
		h.StatusForbidden,
		"AuthenticationFailed",
		"Server failed to authenticate the request. Make sure the authentication information is correct",
	)
	ErrInsufficientAccountPermissions = NewBaseErr(
		h.StatusForbidden,
		"InsufficientAccountPermissions",
		"The account being accessed does not have sufficient permissions to execute this operation",
	)
	ErrNotFound              = NewBaseErr(h.StatusNotFound, "NotFound", "Not found")
	ErrEndpointNotFound      = NewBaseErr(h.StatusNotFound, "EndpointNotFound", "The requested endpoint does not exist")
	ErrResourceNotFound      = NewBaseErr(h.StatusNotFound, "ResourceNotFound", "The specified resource does not exist")
	ErrMethodNotSupported    = NewBaseErr(h.StatusMethodNotAllowed, "MethodNotAllowed", "Method not allowed")
	ErrTimeout               = NewBaseErr(h.StatusRequestTimeout, "RequestTimeout", "Request timeout")
	ErrConflict              = NewBaseErr(h.StatusConflict, "Conflict", "Conflict")
	ErrResourceAlreadyExists = NewBaseErr(
		h.StatusConflict,
		"ResourceAlreadyExists",
		"The specified resource already exists",
	)
	ErrAccountAlreadyExists = NewBaseErr(
		h.StatusConflict,
		"AccountAlreadyExists",
		"The specified account already exists",
	)
	ErrPreconditionFailed = NewBaseErr(h.StatusPreconditionFailed, "PreconditionFailed", "Precondition failed")
	ErrPayloadTooLarge    = NewBaseErr(
		h.StatusRequestEntityTooLarge,
		"RequestEntityTooLarge",
		"Request entity too large",
	)
	ErrTooManyRequests     = NewBaseErr(h.StatusTooManyRequests, "TooManyRequests", "Too many requests")
	ErrClientClosedRequest = NewBaseErr(StatusClientClosedRequest, "ClientClosedRequest", "Client closed request")
	ErrInternalError       = NewBaseErr(
		h.StatusInternalServerError,
		"InternalServerError",
		"The server encountered an internal error, please retry the request",
	)
	ErrServiceUnavailable = NewBaseErr(h.StatusServiceUnavailable, "ServiceUnavailable", "Service unavailable")
	ErrServerBusy         = NewBaseErr(
		h.StatusServiceUnavailable,
		"ServerBusy",
		"The server is currently unable to receive requests. Please retry your request",
	)
)
