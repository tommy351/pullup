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
	logger := zap.New(
		zap.UseDevMode(conf.Dev),
		zap.Level(level),
	)

	log.SetLogger(logger)

	return logger
}
