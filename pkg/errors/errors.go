package errors

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorResponse struct {
	Status       int           `json:"status"`
	Message      string        `json:"message"`
	Details      []ErrorDetail `json:"details,omitempty"`
	LogTimestamp string        `json:"log_timestamp"`
	RequestID    string        `json:"request_id"`
}

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

func NewErrorResponse(statusCode int, message string, details []ErrorDetail) *ErrorResponse {
	return &ErrorResponse{
		Status:       statusCode,
		Message:      message,
		Details:      details,
		LogTimestamp: time.Now().Format(time.RFC3339),
		RequestID:    uuid.New().String(),
	}
}

func NewErrorResponseWithRequestID(statusCode int, message string, details []ErrorDetail, requestID string) *ErrorResponse {
	return &ErrorResponse{
		Status:       statusCode,
		Message:      message,
		Details:      details,
		LogTimestamp: time.Now().Format(time.RFC3339),
		RequestID:    requestID,
	}
}

type ErrImportShipmentsFailed struct {
	err error
}

func (e *ErrImportShipmentsFailed) Error() string {
	return fmt.Sprintf("failed to import shipments: %v", e.err)
}

func (e *ErrImportShipmentsFailed) Unwrap() error {
	return e.err
}

func (e *ErrImportShipmentsFailed) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

func NewErrImportShipmentsFailed(err error) *ErrImportShipmentsFailed {
	return &ErrImportShipmentsFailed{err: err}
}

type ErrImportTerminalsFailed struct {
	err error
}

func (e *ErrImportTerminalsFailed) Error() string {
	return fmt.Sprintf("failed to import terminals: %v", e.err)
}

func (e *ErrImportTerminalsFailed) Unwrap() error {
	return e.err
}

func (e *ErrImportTerminalsFailed) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

func NewErrImportTerminalsFailed(err error) *ErrImportTerminalsFailed {
	return &ErrImportTerminalsFailed{err: err}
}

type ErrImportDistancesFailed struct {
	err error
}

func (e *ErrImportDistancesFailed) Error() string {
	return fmt.Sprintf("failed to import distances: %v", e.err)
}

func (e *ErrImportDistancesFailed) Unwrap() error {
	return e.err
}

func (e *ErrImportDistancesFailed) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

func NewErrImportDistancesFailed(err error) *ErrImportDistancesFailed {
	return &ErrImportDistancesFailed{err: err}
}

type ErrImportRatesFailed struct {
	err error
}

func (e *ErrImportRatesFailed) Error() string {
	return fmt.Sprintf("failed to import rates: %v", e.err)
}

func (e *ErrImportRatesFailed) Unwrap() error {
	return e.err
}

func (e *ErrImportRatesFailed) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

func NewErrImportRatesFailed(err error) *ErrImportRatesFailed {
	return &ErrImportRatesFailed{err: err}
}

func NewBadRequestError(message string) *ErrorResponse {
	return NewErrorResponse(400, message, nil)
}

func NewUnprocessableEntityError(message string) *ErrorResponse {
	return NewErrorResponse(422, message, nil)
}

func NewInternalServerError(message string) *ErrorResponse {
	return NewErrorResponse(500, message, nil)
}

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidData    = errors.New("invalid data")
	ErrInternalServer = errors.New("internal server error")
	ErrAlreadyExists  = errors.New("already exists")
)

type ErrInternal struct {
	msg string
	err error
}

func (e *ErrInternal) Error() string {
	return e.msg
}

func (e *ErrInternal) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

func (e *ErrInternal) Unwrap() error {
	return e.err
}

func NewErrInternal(err error, msg ...string) *ErrInternal {
	m := "internal server error"

	if len(msg) > 0 {
		m = msg[0]
	} else if err != nil {
		m = err.Error()
	}

	return &ErrInternal{msg: m, err: err}
}

func NewErrInvalidArgument(err error, msg string) *ErrorResponse {
	return &ErrorResponse{
		Status:       400,
		Message:      msg,
		Details:      nil,
		LogTimestamp: time.Now().Format(time.RFC3339),
		RequestID:    uuid.New().String(),
	}
}

func NewErrInvalidArgumentWithDetails(details []ErrorDetail) *ErrorResponse {
	return &ErrorResponse{
		Status:       400,
		Message:      "invalid request parameters",
		Details:      details,
		LogTimestamp: time.Now().Format(time.RFC3339),
		RequestID:    uuid.New().String(),
	}
}

func NewErrOptimizationFailed(format string, args ...interface{}) error {
	return fmt.Errorf("optimization failed: "+format, args...)
}
