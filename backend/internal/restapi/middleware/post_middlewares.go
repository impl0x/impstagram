package middleware

import (
	"backend/internal/restapi"
	"strconv"
	"time"

	"github.com/impl0x/go-utils/logger"
	"github.com/impl0x/mo"
)

// ? ---- request logger ----

// post middleware, logs requests with status code and time taken
//
// assumes that request id and start time is present in context store
func RequestLogger(log logger.Logger) mo.PrePostMiddleware {
	return func(c *mo.Context) {
		r := c.Request()
		log.Trace(
			r.Method,
			"status", strconv.Itoa(c.Response().StatusCode()),
			"addr", r.RemoteAddr,
			"req", c.GetTypedMust[string](restapi.KeyRequestID),
			"time taken", time.Since(c.GetTypedMust[time.Time](restapi.KeyStartTime)).String(),
		)
	}
}
