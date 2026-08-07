package responses


import "github.com/impl0x/mo"

type ErrorResp struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

// returns a c.JSON with the data
//
// json structure is from [ErrorResp]
func Error(c *mo.Context, code int, message string) error {
	return c.JSON(code, ErrorResp{code, message})
}
