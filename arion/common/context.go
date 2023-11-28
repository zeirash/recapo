package common

import (
	"context"
	"time"
)

type contextKey string

const (
	DefaultContextSeconds = 60

	UserIDKey     contextKey = "user-id"
	SystemModeKey contextKey = "system-mode"
)

func DefaultTimeoutContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultContextSeconds*time.Second)
	return ctx, cancel
}

func CustomTimeoutContext(seconds int) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
	return ctx, cancel
}
