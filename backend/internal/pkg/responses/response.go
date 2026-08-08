package responses

import "backend/internal/utils/codes"

type Response struct {
	Code    codes.Code `json:"code"`
	Message string     `json:"message"`
	Data    any        `json:"data,omitempty"`
}

// returns [Response] with the data
func Success(code codes.Code, message string, data any) Response {
	return Response{code, message, data}
}

// returns [Response] without the data
func Error(code codes.Code, message string) Response {
	return Response{code, message, nil}
}
