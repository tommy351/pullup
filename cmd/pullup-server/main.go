package main

import (
	"context"

	"github.com/tommy351/pullup/internal/cmd"
	"github.com/tommy351/pullup/pkg/config"
	"github.com/tommy351/pullup/pkg/k8s"
	"github.com/tommy351/pullup/pkg/server"
)

func main() {
	conf := config.MustReadConfig()
	ctx := context.Background()
	logger := cmd.NewLogger(&conf.Log)
	kubeClient, err := k8s.NewClient(&conf.Kubernetes)

	if err != nil {
		logger.Fatal().Stack().Err(err).Msg("Failed to create a Kubernetes client")
	}

	ctx = logger.WithContext(ctx)

	s := &server.Server{
		Config:           conf,
		KubernetesClient: kubeClient,
	}

	if err := s.Serve(ctx); err != nil {
		logger.Fatal().Stack().Err(err).Msg("Failed to start the server")
	}
}
