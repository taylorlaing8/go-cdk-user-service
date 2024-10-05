package enums

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

type ErrorCode int

const (
	ErrorCodeNotFound ErrorCode = iota
	ErrorCodeValidationFailed
	ErrorCodeAccessDenied
	ErrorCodeUnhandledException
)

func (priority ErrorCode) String() string {
	return [...]string{
		"NOT_FOUND",
		"VALIDATION_FAILED",
		"ACCESS_DENIED",
		"UNHANDLED_EXCEPTION",
	}[priority]
}

type ResponseError struct {
	ErrorMessage string   `json:"errorMessage"`
	ErrorStatus  int      `json:"-"`
	ErrorCode    string   `json:"errorCode"`
	Errors       []string `json:"errors"`
}

func (resError *ResponseError) AddData(value string) {
	resError.Errors = append(resError.Errors, value)
}

func (resError ResponseError) Error() string {
	val, _ := json.Marshal(resError)

	return string(val)
}

func (resError ResponseError) ApiResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: resError.ErrorStatus,
		Body:       resError.Error(),
	}
}

func ErrorNotFound() ResponseError {
	return ResponseError{
		ErrorMessage: "Requested resource was not found or does not exist.",
		ErrorStatus:  http.StatusNotFound,
		ErrorCode:    ErrorCodeNotFound.String(),
		Errors:       make([]string, 0),
	}
}

func ErrorValidation(msg string) ResponseError {
	return ResponseError{
		ErrorMessage: fmt.Sprintf("Validation Failed: %v", msg),
		ErrorStatus:  http.StatusBadRequest,
		ErrorCode:    ErrorCodeValidationFailed.String(),
		Errors:       make([]string, 0),
	}
}

func ErrorAuthorization(msg string) ResponseError {
	return ResponseError{
		ErrorMessage: fmt.Sprintf("Authorization Failed: %v", msg),
		ErrorStatus:  http.StatusForbidden,
		ErrorCode:    ErrorCodeAccessDenied.String(),
		Errors:       make([]string, 0),
	}
}

func ErrorUnhandled(msg string) ResponseError {
	return ResponseError{
		ErrorMessage: fmt.Sprintf("Unhandled Exception: %v", msg),
		ErrorStatus:  http.StatusInternalServerError,
		ErrorCode:    ErrorCodeUnhandledException.String(),
		Errors:       make([]string, 0),
	}
}

func ErrorGetOrDefault(e error) ResponseError {
	knownError, ok := e.(ResponseError)
	if ok {
		return knownError
	} else {
		unknownError := ErrorUnhandled(e.Error())
		return unknownError
	}
}
