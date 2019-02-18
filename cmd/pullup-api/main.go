package main

import (
	"context"

	"github.com/tommy351/pullup/pkg/api"
	"github.com/tommy351/pullup/pkg/config"
)

func main() {
	conf, err := config.ReadConfig()

	if err != nil {
		panic(err)
	}

	s := &api.Server{
		Config: conf,
	}

	if err := s.Serve(context.Background()); err != nil {
		panic(err)
	}
}
