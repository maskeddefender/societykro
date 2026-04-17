package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func Init(env string) {
	if env == "development" {
		Log = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Caller().Logger()
	} else {
		Log = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if env == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
