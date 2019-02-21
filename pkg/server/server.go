package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/config"
	"github.com/tommy351/pullup/pkg/kubernetes"
	"github.com/tommy351/pullup/pkg/middleware"
)

type Server struct {
	Config           *config.Config
	KubernetesClient kubernetes.Client
}

func (s *Server) Serve(ctx context.Context) error {
	handler := s.newRouter(ctx)
	logger := zerolog.Ctx(ctx)
	addr := s.Config.Server.Addr

	logger.Info().Str("addr", addr).Msg("Starting server")
	return http.ListenAndServe(addr, handler)
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
			_ = Error(w, ErrUnknown)
		})),
	)

	r.Methods("POST").Path("/webhooks/github").Handler(NewHandler(s.GitHubWebhook))

	return r
}
