package email

import (
	"context"
)

type Sender interface {
	SendSingle(ctx context.Context, sub string, to string, html string) error
	SendMultiple(ctx context.Context, sub string, to string, html string) error
}
