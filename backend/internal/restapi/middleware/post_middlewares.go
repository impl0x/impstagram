package middleware

import (
	"backend/internal/restapi"
	"strconv"

	"github.com/impl0x/go-utils/logger"
	"github.com/impl0x/mo"
)

// ? ---- request logger ----

// post middleware, logs requests with status code
func RequestLogger(log logger.Logger) mo.PrePostMiddleware {
	return func(c *mo.Context) {
		r := c.Request()
		log.Trace(
			r.Method,
			"status", strconv.Itoa(c.Response().StatusCode()),
			"addr", r.RemoteAddr,
			"req", c.GetTypedMust[string](restapi.KeyRequestID),
		)
	}
}
