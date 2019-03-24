package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
)

func Context(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	logger := zerolog.Ctx(ctx)
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-ch
		logger.Debug().Str("signal", sig.String()).Msg("Received signal")
		cancel()
	}()

	return ctx
}
