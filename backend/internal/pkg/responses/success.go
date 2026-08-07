package responses

import "github.com/impl0x/mo"

type SuccessResp struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

// returns a c.JSON with the data
//
// json structure is from [SuccessResp]
func Success(c *mo.Context, code int, data any) error {
	return c.JSON(code, SuccessResp{code, data})
}
