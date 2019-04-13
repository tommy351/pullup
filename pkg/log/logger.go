package log

import (
	"context"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	crlog "sigs.k8s.io/controller-runtime/pkg/log"
	zaplog "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	Debug  = int(zapcore.DebugLevel) * -1
	Info   = int(zapcore.InfoLevel) * -1
	Warn   = int(zapcore.WarnLevel) * -1
	Error  = int(zapcore.ErrorLevel) * -1
	DPanic = int(zapcore.DPanicLevel) * -1
	Panic  = int(zapcore.PanicLevel) * -1
	Fatal  = int(zapcore.FatalLevel) * -1

	FieldError = "error"
)

type Config struct {
	Level string `mapstructure:"level"`
}

type FlushLogger interface {
	logr.Logger
	Flush()
}

type zapFlushLogger struct {
	logr.Logger
	zapLogger *zap.Logger
}

func (z *zapFlushLogger) Flush() {
	_ = z.zapLogger.Sync()
}

func New(conf *Config) FlushLogger {
	encoderConf := zap.NewProductionEncoderConfig()
	encoderConf.TimeKey = "time"
	encoderConf.EncodeTime = zapcore.ISO8601TimeEncoder

	sink := zapcore.AddSync(os.Stderr)
	level := zap.NewAtomicLevel()

	if err := level.UnmarshalText([]byte(conf.Level)); err != nil {
		panic(err)
	}

	encoder := &zaplog.KubeAwareEncoder{
		Encoder: zapcore.NewJSONEncoder(encoderConf),
		Verbose: level.Enabled(zapcore.DebugLevel),
	}

	zapLogger := zap.New(
		zapcore.NewCore(encoder, sink, level),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(core, time.Second, 100, 100)
		}),
		zap.AddStacktrace(zapcore.WarnLevel),
		zap.AddCaller(),
		zap.ErrorOutput(sink),
	)
	logger := zapr.NewLogger(zapLogger)

	crlog.SetLogger(logger)

	return &zapFlushLogger{
		Logger:    crlog.Log.WithName("pullup"),
		zapLogger: zapLogger,
	}
}

// nolint: gochecknoglobals
var (
	contextKey = &struct{}{}
	nullLogger = crlog.NullLogger{}
)

func FromContext(ctx context.Context) logr.Logger {
	logger := ctx.Value(contextKey)

	if logger != nil {
		if logger, ok := logger.(logr.Logger); ok {
			return logger
		}
	}

	return nullLogger
}

func NewContext(ctx context.Context, logger logr.Logger) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}
