package webhook

import (
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/gorilla/mux"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/httputil"
	"github.com/tommy351/pullup/internal/log"
	"github.com/tommy351/pullup/internal/middleware"
	"github.com/tommy351/pullup/internal/webhook/github"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	httphook "github.com/tommy351/pullup/internal/webhook/http"
)

const healthCheckPath = "/healthz"

// ServerSet provides handlers and the server.
// nolint: gochecknoglobals
var ServerSet = wire.NewSet(
	NewLogger,
	hookutil.NewEventRecorder,
	hookutil.NewFieldIndexer,
	controller.NewClient,
	hookutil.ResourceTemplateClientSet,
	github.HandlerSet,
	httphook.HandlerSet,
	wire.Struct(new(Server), "*"),
)

type Logger logr.Logger

func NewLogger(logger logr.Logger) Logger {
	return logger.WithName("webhook")
}

type Handler interface {
	Handle(w http.ResponseWriter, r *http.Request) error
}

type Config struct {
	Address string `mapstructure:"address"`
}

type Server struct {
	Config        Config
	Logger        logr.Logger
	GithubHandler *github.Handler
	HTTPHandler   *httphook.Handler
}

func (s *Server) Start(stop <-chan struct{}) error {
	router := mux.NewRouter()

	router.Use(middleware.SetLogger(s.Logger))

	router.Use(middleware.RequestLog(func(r *http.Request, status, size int, duration time.Duration) {
		if r.RequestURI != healthCheckPath {
			logr.FromContextOrDiscard(r.Context()).V(log.Debug).Info("",
				"requestMethod", r.Method,
				"requestUrl", r.RequestURI,
				"remoteAddr", r.RemoteAddr,
				"userAgent", r.UserAgent(),
				"responseStatus", status,
				"responseSize", size,
				"duration", duration,
			)
		}
	}))

	router.Use(middleware.Recovery(func(w http.ResponseWriter, r *http.Request, err error) {
		logr.FromContextOrDiscard(r.Context()).Error(err, "Webhook server error",
			"requestMethod", r.Method,
			"requestUrl", r.RequestURI,
		)
		_ = httputil.JSON(w, http.StatusInternalServerError, &httputil.Response{
			Errors: []httputil.Error{
				{Description: "Internal server error"},
			},
		})
	}))

	router.Handle(healthCheckPath, httputil.NewHandler(func(w http.ResponseWriter, r *http.Request) error {
		return httputil.JSON(w, http.StatusOK, &httputil.Response{})
	})).Methods(http.MethodGet)

	handlers := map[string]Handler{
		"github": s.GithubHandler,
		"http":   s.HTTPHandler,
	}

	for name, handler := range handlers {
		router.
			Handle("/webhooks/"+name, httputil.NewHandler(handler.Handle)).
			Methods(http.MethodPost)
	}

	router.PathPrefix("/").Handler(httputil.NewHandler(func(w http.ResponseWriter, r *http.Request) error {
		return httputil.JSON(w, http.StatusNotFound, &httputil.Response{
			Errors: []httputil.Error{
				{Description: "Not found"},
			},
		})
	}))

	return httputil.RunServer(httputil.ServerOptions{
		Name:    "webhook",
		Address: s.Config.Address,
		Handler: router,
		Stop:    stop,
		Logger:  s.Logger,
	})
}
