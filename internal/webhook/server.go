package webhook

import (
	"context"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/gorilla/mux"
	"github.com/tommy351/pullup/internal/controller"
	"github.com/tommy351/pullup/internal/httputil"
	"github.com/tommy351/pullup/internal/middleware"
	"github.com/tommy351/pullup/internal/webhook/github"
	"github.com/tommy351/pullup/internal/webhook/hookutil"
	httphook "github.com/tommy351/pullup/internal/webhook/http"
)

// ServerSet provides handlers and the server.
// nolint: gochecknoglobals
var ServerSet = wire.NewSet(
	NewLogger,
	hookutil.NewEventRecorder,
	hookutil.NewFieldIndexer,
	controller.NewClient,
	hookutil.TriggerHandlerSet,
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

func (s *Server) Start(ctx context.Context) error {
	router := mux.NewRouter()

	router.Use(middleware.SetLogger(s.Logger))

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")
			next.ServeHTTP(w, r)
		})
	})

	router.Use(middleware.Recovery(func(w http.ResponseWriter, r *http.Request, err error) {
		logr.FromContextOrDiscard(r.Context()).Error(err, "Webhook server error",
			"requestMethod", r.Method,
			"requestUrl", r.RequestURI,
		)
		_ = httputil.JSON(w, http.StatusInternalServerError, &httputil.Response{
			Errors: []httputil.Error{
				{Description: "internal server error"},
			},
		})
	}))

	handlers := map[string]Handler{
		"github": s.GithubHandler,
		"http":   s.HTTPHandler,
	}

	for name, handler := range handlers {
		router.
			Handle("/webhooks/"+name, hookutil.NewHandler(handler.Handle)).
			Methods(http.MethodPost)
	}

	router.PathPrefix("/").Handler(httputil.NewHandler(func(w http.ResponseWriter, r *http.Request) error {
		return httputil.JSON(w, http.StatusNotFound, &httputil.Response{
			Errors: []httputil.Error{
				{Description: "Not found"},
			},
		})
	}))

	return httputil.RunServer(ctx, httputil.ServerOptions{
		Name:    "webhook",
		Address: s.Config.Address,
		Handler: router,
		Logger:  s.Logger,
	})
}
