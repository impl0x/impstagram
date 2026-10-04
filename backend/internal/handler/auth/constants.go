package auth
// ? ----+-----+-----Store keys-----+-----+-----
// store keys are the keys used to store items in the context storage
type storeKey = string

const (
	keyAuthToken storeKey = "at"
)