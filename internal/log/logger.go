package log

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/google/wire"
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

// LoggerSet provides everything required for a logger.
// nolint: gochecknoglobals
var LoggerSet = wire.NewSet(
	NewEncoderConfig,
	NewZapLevel,
	NewSink,
	NewEncoder,
	NewZapLogger,
	NewLogger,
)

type Config struct {
	Level string `mapstructure:"level"`
}

func NewEncoderConfig() zapcore.EncoderConfig {
	conf := zap.NewProductionEncoderConfig()
	conf.TimeKey = "time"
	conf.EncodeTime = zapcore.ISO8601TimeEncoder

	return conf
}

func NewZapLevel(conf Config) (zap.AtomicLevel, error) {
	level := zap.NewAtomicLevel()

	if err := level.UnmarshalText([]byte(conf.Level)); err != nil {
		return level, fmt.Errorf("failed to unmarshal zap level: %w", err)
	}

	return level, nil
}

func NewEncoder(encoderConf zapcore.EncoderConfig, level zap.AtomicLevel) (zapcore.Encoder, error) {
	return &zaplog.KubeAwareEncoder{
		Encoder: zapcore.NewJSONEncoder(encoderConf),
		Verbose: level.Enabled(zapcore.DebugLevel),
	}, nil
}

func NewSink() (zapcore.WriteSyncer, func(), error) {
	return zap.Open("stderr")
}

func NewZapLogger(encoder zapcore.Encoder, level zap.AtomicLevel, sink zapcore.WriteSyncer) (*zap.Logger, func()) {
	logger := zap.New(
		zapcore.NewCore(encoder, sink, level),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
		}),
		zap.AddStacktrace(zapcore.WarnLevel),
		zap.AddCallerSkip(1),
		zap.ErrorOutput(sink),
	)

	return logger, func() {
		_ = logger.Sync()
	}
}

func NewLogger(zapLogger *zap.Logger) logr.Logger {
	logger := zapr.NewLogger(zapLogger)
	crlog.SetLogger(logger)

	return crlog.Log.WithName("pullup")
}
