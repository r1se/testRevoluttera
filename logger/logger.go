// Package logger предоставляет функции для кастомизации логгера.
package logger

import (
	"github.com/rs/zerolog"
	"io"
	"os"
	"strings"
)

type Logger struct {
	logger   *zerolog.Logger
	Loglevel string
}

func NewLogger(filename, logLevel string) Logger {
	var out io.Writer
	out = os.Stdout

	logFile, err := os.OpenFile(
		filename,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0755)
	if err == nil {
		out = io.MultiWriter(os.Stdout, logFile)
	}

	level, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		level = zerolog.InfoLevel
	}
	logger := zerolog.New(out).With().Timestamp().Logger().Level(level)

	return Logger{
		logger:   &logger,
		Loglevel: logLevel,
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logger.Debug().Msgf(format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logger.Error().Msgf(format, v...)
}
