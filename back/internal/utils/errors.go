package utils

import (
	"encoding/json"
	"net/http"

	"github.com/finkabaj/squid/back/internal/logger"
	"github.com/pkg/errors"
)

type ErrorType struct {
	Status  int
	Message string
}

var (
	ErrorTypeBadRequest = ErrorType{
		Status:  http.StatusBadRequest,
		Message: "bad request",
	}
	ErrorTypeUnauthorized = ErrorType{
		Status:  http.StatusUnauthorized,
		Message: "unauthorized",
	}
	ErrorTypeNotFound = ErrorType{
		Status:  http.StatusNotFound,
		Message: "not found",
	}
	ErrorTypeInternal = ErrorType{
		Status:  http.StatusInternalServerError,
		Message: "internal server error",
	}
	ErrorTypeValidation = ErrorType{
		Status:  http.StatusBadRequest,
		Message: "validation error",
	}
)

type AppError struct {
	Type          ErrorType
	OriginalError error
	Fields        map[string]string // For validation errors
}

func (e AppError) Error() string {
	if e.OriginalError != nil {
		return e.OriginalError.Error()
	}
	return e.Type.Message
}

func NewValidationError(fields map[string]string) error {
	return AppError{
		Type:   ErrorTypeValidation,
		Fields: fields,
	}
}

func NewBadRequestError(err error) error {
	return AppError{
		Type:          ErrorTypeBadRequest,
		OriginalError: errors.WithStack(err),
	}
}

func NewInternalError(err error) error {
	return AppError{
		Type:          ErrorTypeInternal,
		OriginalError: errors.WithStack(err),
	}
}

func NewUnauthorizedError(err error) error {
	return AppError{
		Type:          ErrorTypeUnauthorized,
		OriginalError: errors.WithStack(err),
	}
}

func NewNotFoundError(err error) error {
	return AppError{
		Type:          ErrorTypeNotFound,
		OriginalError: errors.WithStack(err),
	}
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Status  int               `json:"status"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// HandleError Single error handling function
func HandleError(w http.ResponseWriter, err error) {
	var response ErrorResponse

	switch e := err.(type) {
	case AppError:
		errMsg := e.Error()
		if e.Type == ErrorTypeInternal {
			errMsg = "Internal server error"
		}
		response = ErrorResponse{
			Error:   errMsg,
			Message: e.Type.Message,
			Status:  e.Type.Status,
			Fields:  e.Fields,
		}
		if e.Type == ErrorTypeInternal {
			if stackTracer, ok := e.OriginalError.(interface{ StackTrace() errors.StackTrace }); ok {
				logger.Logger.Debug().
					Stack().
					Err(e.OriginalError).
					Interface("stack", stackTracer.StackTrace()).
					Msg("Application error occurred")
			} else {
				logger.Logger.Debug().
					Stack().
					Err(errors.WithStack(e.OriginalError)).
					Msg("Application error occurred (no stack trace)")
			}
		}
	default:
		response = ErrorResponse{
			Error:   "internal server error",
			Message: ErrorTypeInternal.Message,
			Status:  http.StatusInternalServerError,
		}
		logger.Logger.Error().Stack().Err(errors.WithStack(err)).Msg("Unexpected error")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.Status)
	json.NewEncoder(w).Encode(response)
}
