package service

import "context"

// wrapper over context.Context to include required fields by services
type Context struct {
	RequestId string // usually used for logging
	context.Context
}
