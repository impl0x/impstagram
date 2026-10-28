package middleware

import (
	"backend/internal/restapi"
	"time"

	"github.com/impl0x/mo"
)

// Adds request start time to context store for calculating request time taken in post mw
//
// uses context key [restapi.KeyStartTime]
func AddStartTime() mo.PrePostMiddleware {
	return func(c *mo.Context) {
		c.Add(restapi.KeyStartTime, time.Now())
	}
}

// ? ---- Add request ID ----

// Adds a request id for tracing and logging purposes to each request
//
// uses context key [restapi.KeyRequestID]
func AddRequestID(generator func() string) mo.PrePostMiddleware {
	return func(c *mo.Context) {
		c.Add(restapi.KeyRequestID, generator())
	}
}
