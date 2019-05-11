package httputil

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
)

type ServerOptions struct {
	Name            string
	Address         string
	Handler         http.Handler
	ShutdownTimeout time.Duration
	Logger          logr.Logger
	Stop            <-chan struct{}
}

func RunServer(options ServerOptions) error {
	server := http.Server{
		Addr:    options.Address,
		Handler: options.Handler,
	}

	idleConnsClosed := make(chan struct{})

	go func() {
		<-options.Stop

		options.Logger.Info(fmt.Sprintf("Shutting down %s server", options.Name))

		ctx, cancel := context.WithTimeout(context.Background(), options.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout
			options.Logger.Error(err, "failed to shut down the webhook server")
		}

		close(idleConnsClosed)
	}()

	options.Logger.Info(fmt.Sprintf("Starting %s server", options.Name), "address", server.Addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	<-idleConnsClosed
	return nil
}
