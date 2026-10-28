package httpserver

type ContextKey = string

const (
	KeyRequestID ContextKey = "req"
	KeyStartTime ContextKey = "time"
)
