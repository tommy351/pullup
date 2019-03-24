package log

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
)

type Config struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// nolint: gochecknoinits
func init() {
	zerolog.ErrorStackMarshaler = func(err error) interface{} {
		return fmt.Sprintf("%+v", err)
	}
}

func New(conf *Config) *zerolog.Logger {
	var writer io.Writer
	lv, _ := zerolog.ParseLevel(conf.Level)

	switch conf.Format {
	case "console":
		writer = zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.Out = os.Stderr
			w.TimeFormat = "15:04:05"
		})

	default:
		writer = os.Stderr
	}

	logger := zerolog.New(writer).
		Level(lv).
		With().
		Timestamp().
		Logger()

	return &logger
}
