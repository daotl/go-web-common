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
func ToError(x interface{}) error {
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
func ConvertToWError(x interface{}) *Err {
	if x == nil {
		return nil
	}

	if werr, ok := x.(*Err); ok {
		return werr
	}

	var rawErr error
	switch v := x.(type) {
	case error:
		rawErr = v
	default:
		rawErr = fmt.Errorf("%v", x)
	}

	detailErr := &Err{
		Message: rawErr.Error(),
	}

	return &Err{
		error:      rawErr,
		HttpStatus: ErrInternalServerError.HttpStatus,
		Code:       ErrInternalServerError.Code,
		Message:    ErrInternalServerError.Message,
		Details:    []WError{detailErr}, // 有技术详情，非nil
		params:     nil,
	}
}

type WError interface {
	error
	Is(error) bool
	As(any) bool
	GetHttpStatus() int
	GetCode() string
	GetMessage() string
	GetDetails() []WError
	SetDetails(details []WError)
	GetParams() map[string]any
	SetParams(params map[string]any)
	SetMessage(msg string)
}

// Err is the base error type.
// Reference: https://github.com/microsoft/api-guidelines/blob/vNext/azure/Guidelines.md#handling-errors
type Err struct { //nolint:errname // lib
	error
	// HTTP status code
	HttpStatus int `json:"-"`
	// One of a server-defined set of error codes.
	Code string `json:"code"              v:"required" dc:"Error code"`
	// A human-readable representation of the error.
	Message string `json:"message"           v:"required" dc:"Error message"`
	// An array of details about specific errors that led to this reported error.
	Details []WError `json:"details,omitempty"              dc:"Error details"`
	params  map[string]any
}

// ToErr converts any value to an *Err.
// If x is not an *Err, the base will be ErrInternalServerError.
func ToErr(x interface{}) WError {
	if x == nil {
		return nil
	}
	err := ToError(x)
	werr := &Err{}
	if errors.As(err, &werr) {
		return werr
	}
	return NewErrFromError(ErrInternalServerError, err)
}

// NewBaseErr creates a new base Err.
func NewBaseErr(httpStatus int, code, msg string) *Err {
	return &Err{
		error:      fmt.Errorf("%s %s", code, msg),
		HttpStatus: httpStatus,
		Code:       code,
		Message:    msg,
		Details:    nil,
		params:     nil,
	}
}

// NewBaseErrFrom creates a new base Err from another base Err.
func NewBaseErrFrom(base WError, code, msg string) WError {
	if strings.TrimSpace(code) == "" {
		code = base.GetCode()
	}
	if strings.TrimSpace(msg) == "" {
		msg = base.GetMessage()
	}
	err := &Err{
		error:      fmt.Errorf("%w: %s %s", base, code, msg),
		HttpStatus: base.GetHttpStatus(),
		Code:       code,
		Message:    msg,
	}
	return err
}

// NewErr creates a new Err from a base Err.
func NewErr(base WError, msg, msgDetail string) *Err {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		msg = base.GetMessage()
	}
	msgDetail = strings.TrimSpace(msgDetail)
	if msgDetail != "" {
		msg = msg + ": " + msgDetail
	}
	return &Err{
		error:      fmt.Errorf("%w: %s", base, msg),
		HttpStatus: base.GetHttpStatus(),
		Code:       base.GetCode(),
		Message:    msg,
	}
}

// NewErrFromError creates a new Err from an error.
func NewErrFromError(base WError, err error) *Err {
	msgDetail := err.Error()
	werr := &Err{}
	if errors.As(err, &werr) {
		if werr.Code == base.GetCode() && werr.Message == base.GetMessage() {
			return werr
		}
		msgDetail = werr.Message
	}
	return &Err{
		error:      err,
		HttpStatus: base.GetHttpStatus(),
		Code:       base.GetCode(),
		Message:    base.GetMessage() + ": " + msgDetail,
	}
}

func NewErrWithParams(base *Err, code string, params map[string]any, msgDetail string) *Err {
	if strings.TrimSpace(code) == "" {
		code = base.Code
	}
	msg := base.Message

	detailErr := &Err{
		Message: base.Message + msgDetail,
	}
	return &Err{
		error:      fmt.Errorf("%w: %s", base.error, msg),
		HttpStatus: base.HttpStatus,
		Code:       code,
		Message:    msg,
		Details:    []WError{detailErr},
		params:     params,
	}
}
func (e *Err) Error() string {
	return fmt.Sprintf("%v: %s", e.HttpStatus, e.error.Error())
}

func (e *Err) Is(target error) bool {
	if t, ok := target.(*Err); ok {
		return t.Code == e.Code
	}
	return errors.Is(e.error, target)
}

func (e *Err) As(target any) bool {
	return errors.As(e.error, target)
}

func (e *Err) GetHttpStatus() int {
	return e.HttpStatus
}

func (e *Err) GetCode() string {
	return e.Code
}

func (e *Err) GetMessage() string {
	return e.Message
}
func (e *Err) SetMessage(msg string) {
	e.Message = msg
}
func (e *Err) GetDetails() []WError {
	return e.Details
}
func (e *Err) SetDetails(details []WError) {
	e.Details = details
}

func (e *Err) GetParams() map[string]any {
	if e.params == nil {
		return nil
	}
	paramsCopy := make(map[string]any, len(e.params))
	for k, v := range e.params {
		paramsCopy[k] = v
	}
	return paramsCopy
}

func (e *Err) SetParams(params map[string]any) {
	if params == nil {
		e.params = make(map[string]any)
		return
	}
	e.params = make(map[string]any, len(params))
	for k, v := range params {
		e.params[k] = v
	}
}

// IsErrOf checks if err wraps *Err and has the given code.
func IsErrOf(err error, code string) bool {
	var e *Err
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
	ErrNotFound              = NewBaseErr(h.StatusNotFound, "NotFound", "Not found")
	ErrEndpointNotFound      = NewBaseErr(h.StatusNotFound, "EndpointNotFound", "The requested endpoint does not exist")
	ErrResourceNotFound      = NewBaseErr(h.StatusNotFound, "ResourceNotFound", "The specified resource does not exist")
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

var HttpStatus2ErrMap = map[int]*Err{
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
