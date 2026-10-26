package email

import (
	"context"
)

// Email sender
type Sender interface {
	// Sends a single email to the provided email id
	SendSingle(ctx context.Context, sub string, to string, html string) error
	// Sends a email to multiple people
	SendMultiple(ctx context.Context, sub string, to []string, html string) error
}
