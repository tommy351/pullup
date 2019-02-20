package cmd

import (
	"io"
	"os"

	"github.com/ansel1/merry"
	"github.com/rs/zerolog"
	"github.com/tommy351/pullup/pkg/config"
)

func init() {
	zerolog.ErrorStackMarshaler = func(err error) interface{} {
		return merry.Stacktrace(err)
	}
}

func NewLogger(conf *config.LogConfig) *zerolog.Logger {
	var writer io.Writer
	lv, _ := zerolog.ParseLevel(conf.Level)

	switch conf.Format {
	case "console":
		writer = zerolog.NewConsoleWriter(
			setConsoleLoggerOutput,
			setConsoleLoggerTimeFormat,
		)

	default:
		writer = os.Stderr
	}

	logger := zerolog.New(writer).
		Level(lv).
		With().
		Timestamp().
		Stack().
		Logger()

	return &logger
}

func setConsoleLoggerOutput(w *zerolog.ConsoleWriter) {
	w.Out = os.Stderr
}

func setConsoleLoggerTimeFormat(w *zerolog.ConsoleWriter) {
	w.TimeFormat = "15:04:05"
}
