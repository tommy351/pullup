package webhook

import (
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/justinas/alice"
	"github.com/tommy351/pullup/pkg/httputil"
	"github.com/tommy351/pullup/pkg/log"
	"github.com/tommy351/pullup/pkg/middleware"
	"github.com/tommy351/pullup/pkg/webhook/github"
)

const healthCheckPath = "/healthz"

// ServerSet provides handlers and the server.
// nolint: gochecknoglobals
var ServerSet = wire.NewSet(github.NewHandler, NewServer)

type Config struct {
	Address string `mapstructure:"address"`
}

type Server struct {
	address       string
	logger        logr.Logger
	githubHandler *github.Handler
}

func NewServer(conf Config, logger logr.Logger, githubHandler *github.Handler) *Server {
	return &Server{
		address:       conf.Address,
		logger:        logger.WithName("webhook"),
		githubHandler: githubHandler,
	}
}

func (s *Server) Start(stop <-chan struct{}) error {
	chain := alice.New(
		middleware.SetLogger(s.logger),
		middleware.RequestLog(func(r *http.Request, status, size int, duration time.Duration) {
			if r.RequestURI != healthCheckPath {
				log.FromContext(r.Context()).V(log.Debug).Info("",
					"requestMethod", r.Method,
					"requestUrl", r.RequestURI,
					"remoteAddr", r.RemoteAddr,
					"userAgent", r.UserAgent(),
					"responseStatus", status,
					"responseSize", size,
					"duration", duration,
				)
			}
		}),
		middleware.Recovery(func(w http.ResponseWriter, r *http.Request, err error) {
			log.FromContext(r.Context()).Error(err, "Webhook server error",
				"requestMethod", r.Method,
				"requestUrl", r.RequestURI,
			)
			_ = httputil.String(w, http.StatusInternalServerError, "Internal server error")
		}),
	)

	return httputil.RunServer(httputil.ServerOptions{
		Name:            "webhook",
		Address:         s.address,
		Handler:         chain.Then(s.newRouter()),
		ShutdownTimeout: time.Second * 5,
		Stop:            stop,
		Logger:          s.logger,
	})
}

func (s *Server) newRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle(healthCheckPath, httputil.NewHandler(func(w http.ResponseWriter, r *http.Request) error {
		return httputil.String(w, http.StatusOK, "ok")
	}))

	mux.Handle("/webhooks/github", s.githubHandler)

	return mux
}
