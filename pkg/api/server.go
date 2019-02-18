package api

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tommy351/pullup/pkg/config"
	"github.com/tommy351/pullup/pkg/kubernetes"
)

type Server struct {
	Config           *config.Config
	KubernetesClient *kubernetes.Client
}

func (s *Server) Serve(ctx context.Context) error {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/webhooks/github", s.GitHubWebhook)

	return e.Start(s.Config.Server.Addr)
}
