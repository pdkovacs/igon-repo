package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type LogLevel = string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
)

func CreateRootLogger(levelArg LogLevel) zerolog.Logger {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	var level zerolog.Level
	if levelArg == InfoLevel {
		level = zerolog.InfoLevel
	} else if levelArg == DebugLevel {
		level = zerolog.DebugLevel
	}
	fmt.Printf("Log level: %v\n", level)
	return logger.Level(level).With().Timestamp().Logger()
}

func CreateUnitLogger(logger zerolog.Logger, unitName string) zerolog.Logger {
	return logger.With().Str("unit", unitName).Logger()
}

func CreateMethodLogger(logger zerolog.Logger, unitName string) zerolog.Logger {
	return logger.With().Str("method", unitName).Logger()
}
