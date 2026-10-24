package restapi

type ContextKey = string

const (
	KeyRequestID ContextKey = "req"
	KeyStartTime ContextKey = "time"
)
