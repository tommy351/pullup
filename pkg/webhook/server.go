package webhook

import (
	"context"
	"net/http"

	"github.com/dimfeld/httptreemux"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/k8s"
	"github.com/tommy351/pullup/pkg/middleware"
)

type Config struct {
	Address string `mapstructure:"address"`
}

type Server struct {
	http.Server

	Client *k8s.Client
	Config Config
}

func (s *Server) Serve(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	logger := zerolog.Ctx(ctx)
	s.Server = http.Server{
		Addr:    s.Config.Address,
		Handler: s.newRouter(ctx),
	}

	go func() {
		logger.Info().Str("address", s.Server.Addr).Msg("Starting webhook server")
		err = s.Server.ListenAndServe()
		cancel()
	}()

	<-ctx.Done()

	if err != nil {
		return
	}

	logger.Info().Msg("Shutting down webhook server")
	return s.Server.Shutdown(context.Background())
}

func (s *Server) newRouter(ctx context.Context) *httptreemux.ContextMux {
	mux := httptreemux.NewContextMux()

	mux.NotFoundHandler = func(w http.ResponseWriter, r *http.Request) {
		if err := String(w, http.StatusNotFound, "Not found"); err != nil {
			panic(err)
		}
	}

	chain := alice.New(
		middleware.Logger(zerolog.Ctx(ctx)),
		middleware.Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := String(w, http.StatusInternalServerError, "Internal server error"); err != nil {
				panic(err)
			}
		})),
	)

	mux.Handler(http.MethodGet, "/", chain.Then(NewHandler(func(w http.ResponseWriter, r *http.Request) error {
		return String(w, http.StatusOK, "ok")
	})))

	mux.Handler(http.MethodPost, "/webhooks/:name", chain.Then(NewHandler(s.Webhook)))

	return mux
}
