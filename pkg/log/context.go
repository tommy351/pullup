package log

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// nolint: gochecknoglobals
var (
	contextKey = &struct{}{}
	nullLogger = log.NullLogger{}
)

func FromContext(ctx context.Context) logr.Logger {
	logger := ctx.Value(contextKey)

	if logger != nil {
		if logger, ok := logger.(logr.Logger); ok {
			return logger
		}
	}

	return nullLogger
}

func NewContext(ctx context.Context, logger logr.Logger) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}
