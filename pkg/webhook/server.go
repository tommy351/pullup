package webhook

import (
	"context"
	"net/http"
	"time"

	"github.com/dimfeld/httptreemux"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/tommy351/pullup/pkg/client/clientset/versioned"
	"golang.org/x/xerrors"
)

type Config struct {
	Address string `mapstructure:"address"`
}

type Server struct {
	Config    Config
	Client    versioned.Interface
	Namespace string
}

func (s *Server) Serve(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	logger := zerolog.Ctx(ctx)

	chain := alice.New(
		hlog.NewHandler(*logger),
		hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			if r.RequestURI != "/" {
				hlog.FromRequest(r).Debug().
					Int("status", status).
					Int("size", size).
					Dur("duration", duration).
					Msg("")
			}
		}),
		hlog.MethodHandler("method"),
		hlog.URLHandler("url"),
		hlog.RemoteAddrHandler("remoteIp"),
		hlog.UserAgentHandler("userAgent"),
	)

	httpServer := http.Server{
		Addr:    s.Config.Address,
		Handler: chain.Then(s.newRouter()),
	}

	go func() {
		logger.Info().Str("address", httpServer.Addr).Msg("Starting webhook server")
		err = httpServer.ListenAndServe()
		cancel()
	}()

	<-ctx.Done()

	if err != nil {
		return
	}

	logger.Info().Msg("Shutting down webhook server")
	return httpServer.Shutdown(context.Background())
}

func (s *Server) newRouter() *httptreemux.ContextMux {
	mux := httptreemux.NewContextMux()

	mux.PanicHandler = func(w http.ResponseWriter, r *http.Request, err interface{}) {
		hlog.FromRequest(r).Error().Stack().
			Err(xerrors.Errorf("http handler panicked: %w", err)).
			Msg("HTTP handler panicked")

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
