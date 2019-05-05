package metrics

import (
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tommy351/pullup/pkg/httputil"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type Server struct {
	logger logr.Logger
}

func NewServer(logger logr.Logger) *Server {
	return &Server{logger: logger}
}

func (s *Server) Start(stop <-chan struct{}) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))

	return httputil.RunServer(httputil.ServerOptions{
		Name:            "metrics",
		Address:         ":9100",
		Handler:         mux,
		ShutdownTimeout: time.Second * 5,
		Stop:            stop,
		Logger:          s.logger,
	})
}
