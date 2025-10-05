package common

import (
	"fmt"
	"strconv"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorCode int

const (
	CodeOK ErrorCode = 0

	// Invalid : 1xx
	CodeInvalidRequest ErrorCode = 100
	CodeUnauthorized   ErrorCode = 101
	CodeNotFound       ErrorCode = 102

	// Biz: 2xx
	CodeInvalidLogin       ErrorCode = 200
	CodeExistedUsername    ErrorCode = 201
	CodeNotExistedUsername ErrorCode = 202
	CodeNotExistedUserID   ErrorCode = 203
	CodeNotImplemented     ErrorCode = 204

	// Internal: 9xx
	CodeInternal      ErrorCode = 900
	CodeDatabaseError ErrorCode = 901
)

type AppError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func NewError(code ErrorCode, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func WrapError(code ErrorCode, msg string, cause error) *AppError {
	wrapMsg := fmt.Sprintf("%s: %s", msg, cause.Error())
	return &AppError{Code: code, Message: wrapMsg, Cause: cause}
}

func (e *AppError) Error() string {
	return e.Message
}

func ToGRPCError(appErr *AppError) error {
	st := status.New(codes.Unknown, appErr.Message)

	detail := &errdetails.ErrorInfo{
		Reason: strconv.Itoa(int(appErr.Code)),
		Metadata: map[string]string{
			"msg": appErr.Message,
		},
	}

	stWithDetails, err := st.WithDetails(detail)
	if err != nil {
		return st.Err()
	}

	return stWithDetails.Err()
}

func FromGRPCError(err error) *AppError {
	st, ok := status.FromError(err)
	if !ok {
		return NewError(CodeInternal, err.Error())
	}

	appErr := &AppError{Message: st.Message()}

	for _, d := range st.Details() {
		switch info := d.(type) {
		case *errdetails.ErrorInfo:
			errCode, parseErr := strconv.Atoi(info.Reason)
			if parseErr != nil {
				errCode = int(CodeInvalidRequest)
			}
			appErr.Code = ErrorCode(errCode)
		}
	}

	return appErr
}
