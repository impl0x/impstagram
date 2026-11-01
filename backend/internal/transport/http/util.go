package httpserver

import (
	"backend/internal/service"
	"backend/internal/transport/http/response"
	"backend/pkg/apperr"

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

// Converts a apperr.AppErr to http status code and response.Response object
func AppErrToHttp(err apperr.AppErr, data any) (int, response.Response) {
	return response.HttpStatuscodeFromAppErrKind(err.Kind), response.FromAppErr(err, data)
}

// if err is of type [apperr.AppErr] then converts it to [response.Response]
// struct using the method [response.FromAppErr] and adds the extra data to
// the response struct and sends a http json response with it, statuscode is
// derived using [response.HttpStatuscodeFromAppErrKind].
//
// if err is not of that type then it is returned directly.
// this is meant to be a wrapper function over this functionality to
// not keep rewriting the same snippet of code
func SendJSONWithAppErrMaybe(c *mo.Context, err error, data any) error {
	e, ok := err.(apperr.AppErr)
	if !ok {
		return err
	}
	return c.JSON(AppErrToHttp(e, data))
}
