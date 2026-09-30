package response

type Response struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// returns [Response] with the data
func Success(code Code, message string, data any) Response {
	return Response{code, message, data}
}

// returns [Response] without the data
func Error(code Code, message string) Response {
	return Response{code, message, nil}
}
