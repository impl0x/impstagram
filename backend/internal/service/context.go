package service

import "context"

// wrapper over context.Context to include important fields
type Context struct {
	RequestId string
	context.Context
}
