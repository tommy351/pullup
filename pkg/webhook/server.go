package webhook

import (
	"context"
	"net/http"
	"time"

	"github.com/dimfeld/httptreemux"
	"github.com/go-logr/logr"
	"github.com/justinas/alice"
	"github.com/tommy351/pullup/pkg/log"
	"github.com/tommy351/pullup/pkg/middleware"
	"golang.org/x/xerrors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	Address string `mapstructure:"address"`
}

type Server struct {
	Config    Config
	Namespace string

	client client.Client
	logger logr.Logger
}

func (s *Server) InjectClient(c client.Client) error {
	s.client = c
	return nil
}

func (s *Server) InjectLogger(l logr.Logger) error {
	s.logger = l
	return nil
}

func (s *Server) Start(stop <-chan struct{}) error {
	chain := alice.New(
		middleware.SetLogger(s.logger),
		middleware.RequestLog(func(r *http.Request, status, size int, duration time.Duration) {
			if r.RequestURI != "/" {
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
	)

	httpServer := http.Server{
		Addr:    s.Config.Address,
		Handler: chain.Then(s.newRouter()),
	}

	idleConnsClosed := make(chan struct{})

	go func() {
		<-stop

		s.logger.Info("Shutting down webhook server")

		if err := httpServer.Shutdown(context.TODO()); err != nil {
			// Error from closing listeners, or context timeout
			s.logger.Error(err, "failed to shut down the webhook server")
		}

		close(idleConnsClosed)
	}()

	s.logger.Info("Starting webhook server", "address", httpServer.Addr)
	err := httpServer.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	<-idleConnsClosed
	return nil
}

func (s *Server) newRouter() *httptreemux.ContextMux {
	mux := httptreemux.NewContextMux()

	mux.PanicHandler = func(w http.ResponseWriter, r *http.Request, err interface{}) {
		log.FromContext(r.Context()).Error(
			xerrors.Errorf("http handler panicked: %w", err),
			"HTTP handler panicked",
		)

		_ = String(w, http.StatusInternalServerError, "Internal server error")
	}

	mux.NotFoundHandler = NewHandler(func(w http.ResponseWriter, r *http.Request) error {
		return String(w, http.StatusNotFound, "Not found")
	})

	mux.GET("/", NewHandler(func(w http.ResponseWriter, r *http.Request) error {
		return String(w, http.StatusOK, "ok")
	}))

	mux.POST("/webhooks/:name", NewHandler(s.Webhook))

	return mux
}
