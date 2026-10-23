package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/impl0x/go-utils/logger"
	"github.com/impl0x/mo"
)

// ? ---- Recover ----

// set logError to true to log the panics in your code. it is recommended to do so.
func Recover(logError bool, log logger.Logger) mo.Middleware {
	return func(next mo.HandlerFunc) mo.HandlerFunc {
		return func(c *mo.Context) error {
			defer func() {
				if err := recover(); err != nil {
					if logError {
						log.Error("PANIC recovered" + tryParsePanicReason(err))
						println("[STACK TRACE]\n" + string(debug.Stack()))
					}
					c.JSON(http.StatusInternalServerError, mo.ErrInternalServerError)
				}
			}()
			return next(c)
		}
	}
}

func tryParsePanicReason(err any) string {
	var str string
	switch e := err.(type) {
	case string:
		str = e
	case error:
		str = e.Error()
	}
	if str != "" {
		str = ", reason: " + str
	}
	return str
}
