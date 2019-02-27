package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/config"
	"github.com/tommy351/pullup/pkg/k8s"
	"github.com/tommy351/pullup/pkg/middleware"
)

type Server struct {
	Config           *config.Config
	KubernetesClient k8s.Client
}

func (s *Server) Serve(ctx context.Context) error {
	handler := s.newRouter(ctx)
	logger := zerolog.Ctx(ctx)
	address := s.Config.Server.Address

	logger.Info().Str("address", address).Msg("Starting server")
	return http.ListenAndServe(address, handler)
}

func (s *Server) newRouter(ctx context.Context) *mux.Router {
	r := mux.NewRouter()

	r.NotFoundHandler = NewHandler(func(w http.ResponseWriter, r *http.Request) error {
		return ErrNotFound
	})

	r.MethodNotAllowedHandler = r.NotFoundHandler

	r.Use(
		middleware.Context(ctx),
		middleware.Logger(),
		middleware.Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := Error(w, ErrUnknown); err != nil {
				panic(err)
			}
		})),
	)

	r.Methods(http.MethodGet).Path("/").Handler(NewHandler(func(w http.ResponseWriter, r *http.Request) error {
		return String(w, http.StatusOK, "ok")
	}))

	r.Methods(http.MethodPost).Path("/webhooks/github").Handler(NewHandler(s.GitHubWebhook))

	return r
}
