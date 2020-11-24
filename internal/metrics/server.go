package metrics

import (
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tommy351/pullup/internal/httputil"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// ServerSet provides a server.
// nolint: gochecknoglobals
var ServerSet = wire.NewSet(
	wire.Struct(new(Server), "*"),
)

type Server struct {
	Logger logr.Logger
}

func (s *Server) Start(stop <-chan struct{}) error {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))

	return httputil.RunServer(httputil.ServerOptions{
		Name:    "metrics",
		Address: ":9100",
		Handler: router,
		Stop:    stop,
		Logger:  s.Logger,
	})
}
