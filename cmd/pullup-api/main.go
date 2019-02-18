package main

import (
	"context"

	"github.com/tommy351/pullup/pkg/api"
	"github.com/tommy351/pullup/pkg/config"
	"github.com/tommy351/pullup/pkg/kubernetes"
)

func main() {
	conf, err := config.ReadConfig()

	if err != nil {
		panic(err)
	}

	kubeClient, err := kubernetes.NewClient(&conf.Kubernetes)

	if err != nil {
		panic(err)
	}

	s := &api.Server{
		Config:           conf,
		KubernetesClient: kubeClient,
	}

	if err := s.Serve(context.Background()); err != nil {
		panic(err)
	}
}
