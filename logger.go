package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

const (
	JobLogKey = "jobID"
)

var (
	consoleWriter = zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
	}
	baseLogger = log.With().Caller().Logger().Output(consoleWriter)
)

func ConsoleLogger() zerolog.Logger {
	return baseLogger
}
