package pkg

type SuccessResponse struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

func NewSuccessResponse(code int, data any)SuccessResponse{
	return SuccessResponse{code, data}
}
