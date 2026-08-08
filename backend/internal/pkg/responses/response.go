package responses

import "backend/internal/utils/errorcodes"

type Response struct {
	Code    errorcodes.Code `json:"code"`
	Message string          `json:"message"`
	Data    any             `json:"data,omitempty"`
}

// returns [Response] with the data
func Success(code errorcodes.Code, message string, data any) Response {
	return Response{code, message, data}
}

// returns [Response] without the data
func Error(code errorcodes.Code, message string) Response {
	return Response{code, message, nil}
}
