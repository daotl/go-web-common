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

// ToError converts any value to an error, default to ErrInternalServerError.
func ToError(x any) error {
	if x == nil {
		return nil
	}
	switch v := x.(type) {
	case error:
		return v
	case Stringer:
		return errors.New(v.String())
	default:
		return ErrInternalServerError
	}
}

// Err is the standard error interface.
// Code using this package should return Err interface in function signatures
// instead of the Serr struct type.
// Err.SetXxx methods return the Err itself.
type Err interface {
	error
	Is(error) bool
	As(any) bool
	GetHttpStatus() int
	GetCode() string
	SetCode(code string)
	GetMessage() string
	SetMessage(msg string)
	GetSubErrors() []Err
	SetSubErrors(errs []Err)
	// AddSubErrors will append errs to the current sub-errors slice
	AddSubErrors(errs ...Err)
	GetMetadata() map[string]any
	SetMetadata(meta map[string]any)
	// AddMetadata will merge meta map to the current metadata map
	AddMetadata(meta map[string]any)
}

// Serr is the base error struct type.
// Reference: https://github.com/microsoft/api-guidelines/blob/vNext/azure/Guidelines.md#handling-errors
type Serr struct { //nolint:errname // lib
	error

	// HTTP status code
	HttpStatus int `json:"-"`
	// One of a server-defined set of error codes.
	Code string `json:"code"                v:"required" dc:"Error code"`
	// A human-readable representation of the error.
	Message string `json:"message"             v:"required" dc:"Error message"`
	// An array of specific errors that led to this error.
	SubErrors []Err `json:"subErrors,omitempty"              dc:"Sub-errors that led to this error"`
	// Error metadata, useful for debugging, logging, generating i18n error messages etc.
	Metadata map[string]any `json:"metadata,omitempty"               dc:"Error metadata"`
}

// ToErr converts any value to an *Err.
// If x is not an *Err, the base will be ErrInternalServerError.
func ToErr(x any) Err {
	if x == nil {
		return nil
	}

	var err error
	switch v := x.(type) {
	case Err:
		return v
	case error:
		err = v
	default:
		err = fmt.Errorf("%v", v)
	}

	return NewErrFromError(ErrInternalServerError, err)
}

// NewBaseErr creates a new base Err.
func NewBaseErr(httpStatus int, code, msg string) Err {
	return &Serr{
		error:      fmt.Errorf("%s %s", code, msg),
		HttpStatus: httpStatus,
		Code:       code,
		Message:    msg,
	}
}

// NewBaseErrFrom creates a new base Err from another base Err.
func NewBaseErrFrom(base Err, code, msg string) Err {
	if strings.TrimSpace(code) == "" {
		code = base.GetCode()
	}
	if strings.TrimSpace(msg) == "" {
		msg = base.GetMessage()
	}
	err := &Serr{
		error:      fmt.Errorf("%w: %s %s", base, code, msg),
		HttpStatus: base.GetHttpStatus(),
		Code:       code,
		Message:    msg,
	}
	return err
}

// NewErr creates a new Err from a base Err.
func NewErr(base Err, msg, msgDetail string) Err {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		msg = base.GetMessage()
	}
	msgDetail = strings.TrimSpace(msgDetail)
	if msgDetail != "" {
		msg = msg + ": " + msgDetail
	}
	return &Serr{
		error:      fmt.Errorf("%w: %s", base, msg),
		HttpStatus: base.GetHttpStatus(),
		Code:       base.GetCode(),
		Message:    msg,
	}
}

// NewErrFromError creates a new Err from an error.
func NewErrFromError(base Err, err error) Err {
	msgDetail := err.Error()
	werr := &Serr{}
	if errors.As(err, &werr) {
		if werr.Code == base.GetCode() && werr.Message == base.GetMessage() {
			return werr
		}
		msgDetail = werr.Message
	}
	return &Serr{
		error:      err,
		HttpStatus: base.GetHttpStatus(),
		Code:       base.GetCode(),
		Message:    base.GetMessage() + ": " + msgDetail,
	}
}

func (e *Serr) Error() string {
	return fmt.Sprintf("%v: %s", e.HttpStatus, e.error.Error())
}

func (e *Serr) Is(target error) bool {
	if t, ok := target.(*Serr); ok {
		return t.Code == e.Code
	}
	return errors.Is(e.error, target)
}

func (e *Serr) As(target any) bool {
	return errors.As(e.error, target)
}

func (e *Serr) GetHttpStatus() int {
	return e.HttpStatus
}

func (e *Serr) GetCode() string {
	return e.Code
}

func (e *Serr) SetCode(code string) {
	e.Code = code
}

func (e *Serr) GetMessage() string {
	return e.Message
}

func (e *Serr) SetMessage(msg string) {
	e.Message = msg
}

func (e *Serr) GetSubErrors() []Err {
	return e.SubErrors
}

func (e *Serr) SetSubErrors(errs []Err) {
	e.SubErrors = errs
}

// AddSubErrors will append errs to the current sub-errors slice.
func (e *Serr) AddSubErrors(errs ...Err) {
	e.SubErrors = append(e.SubErrors, errs...)
}

func (e *Serr) GetMetadata() map[string]any {
	return e.Metadata
}

func (e *Serr) SetMetadata(meta map[string]any) {
	e.Metadata = meta
}

// AddMetadata will merge meta map to the current metadata map.
func (e *Serr) AddMetadata(meta map[string]any) {
	if e.Metadata == nil {
		e.Metadata = meta
	} else {
		for k, v := range meta {
			e.Metadata[k] = v
		}
	}
}

// IsErrOf checks if err wraps *Err and has the given code.
func IsErrOf(err error, code string) bool {
	var e *Serr
	ok := errors.As(err, &e)
	return ok && e.GetCode() == code
}

// References:
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
	ErrNotFound         = NewBaseErr(h.StatusNotFound, "NotFound", "Not found")
	ErrEndpointNotFound = NewBaseErr(h.StatusNotFound, "EndpointNotFound",
		"The requested endpoint does not exist")
	ErrResourceNotFound = NewBaseErr(h.StatusNotFound, "ResourceNotFound",
		"The specified resource does not exist")
	ErrMethodNotAllowed      = NewBaseErr(h.StatusMethodNotAllowed, "MethodNotAllowed", "Method not allowed")
	ErrTimeout               = NewBaseErr(h.StatusRequestTimeout, "Timeout", "Timeout")
	ErrRequestTimeout        = NewBaseErr(h.StatusRequestTimeout, "RequestTimeout", "Request timeout")
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
		"PayloadTooLarge",
		"Payload too large",
	)
	ErrRequestEntityTooLarge = NewBaseErr(
		h.StatusRequestEntityTooLarge,
		"RequestEntityTooLarge",
		"Request entity too large",
	)
	ErrTooManyRequests     = NewBaseErr(h.StatusTooManyRequests, "TooManyRequests", "Too many requests")
	ErrClientClosedRequest = NewBaseErr(StatusClientClosedRequest, "ClientClosedRequest", "Client closed request")
	ErrInternalError       = NewBaseErr(
		h.StatusInternalServerError,
		"InternalError",
		"The system encountered an internal error",
	)
	ErrInternalServerError = NewBaseErr(
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

var HttpStatus2ErrMap = map[int]Err{
	h.StatusBadRequest:            ErrBadRequest,
	h.StatusUnauthorized:          ErrUnauthorized,
	h.StatusForbidden:             ErrForbidden,
	h.StatusNotFound:              ErrNotFound,
	h.StatusMethodNotAllowed:      ErrMethodNotAllowed,
	h.StatusRequestTimeout:        ErrRequestTimeout,
	h.StatusConflict:              ErrConflict,
	h.StatusPreconditionFailed:    ErrPreconditionFailed,
	h.StatusRequestEntityTooLarge: ErrRequestEntityTooLarge,
	h.StatusTooManyRequests:       ErrTooManyRequests,
	h.StatusInternalServerError:   ErrInternalServerError,
	h.StatusServiceUnavailable:    ErrServiceUnavailable,
}
