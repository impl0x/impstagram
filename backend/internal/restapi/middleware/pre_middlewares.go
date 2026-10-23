package middleware

import (
	"backend/internal/restapi"

	"github.com/impl0x/mo"
)

func AddTiming()

// ? ---- Add request ID ----

func AddRequestID(generator func() string) mo.PrePostMiddleware {
	return func(ctx *mo.Context) {
		ctx.Add(restapi.KeyRequestID, generator())
	}
}
