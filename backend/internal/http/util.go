package restapi

import (
	"backend/internal/service"

	"github.com/impl0x/mo"
)

// converts a [mo.Context] into a [service.Context]
// by getting the request id from the context and using the request context
//
// warning: uses [mo.Context.GetTypedMust] to retrieve the request id, 
// therefore it can panic if request id is not present in context store.
func ToServiceContext(c *mo.Context) service.Context {
	return service.Context{
		ReqId:   c.GetTypedMust[string](KeyRequestID),
		Context: c.Request().Context(),
	}
}
