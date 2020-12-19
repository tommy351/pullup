package log

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/wire"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	Debug = int(zapcore.DebugLevel) * -1
)

// LoggerSet provides everything required for a logger.
// nolint: gochecknoglobals
var LoggerSet = wire.NewSet(
	NewZapLevelEnabler,
	NewLogger,
)

type Config struct {
	Level string `mapstructure:"level"`
	Dev   bool   `mapstructure:"dev"`
}

func NewZapLevelEnabler(conf Config) (zapcore.LevelEnabler, error) {
	var level zapcore.Level

	if err := level.Set(conf.Level); err != nil {
		return nil, fmt.Errorf("invalid level: %w", err)
	}

	return &level, nil
}

func NewLogger(conf Config, level zapcore.LevelEnabler) logr.Logger {
	options := []zap.Opts{
		zap.UseDevMode(conf.Dev),
		zap.Level(level),
	}

	if conf.Dev {
		options = append(options, zap.ConsoleEncoder())
	} else {
		options = append(options, zap.JSONEncoder(func(z *zapcore.EncoderConfig) {
			z.EncodeTime = zapcore.ISO8601TimeEncoder
			z.TimeKey = "time"
		}))
	}

	logger := zap.New(options...)

	log.SetLogger(logger)

	return logger
}
