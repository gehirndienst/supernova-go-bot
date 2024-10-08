package botapi

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	o sync.Once
	l zerolog.Logger
)

func GetLogger() zerolog.Logger {
	o.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel := strToLevel(os.Getenv("LOG_LEVEL"))

		// default pretty-print console writer to stdout for APP_ENV=dev
		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		if os.Getenv("GO_ENV") != "dev" {
			var lf = os.Getenv("LOG_FILE")

			if lf == "" {
				output = os.Stderr
			} else {
				fileLogger := &lumberjack.Logger{
					Filename:   lf,
					MaxSize:    20,
					MaxBackups: 10,
					MaxAge:     14,
					Compress:   true,
				}
				output = zerolog.MultiLevelWriter(os.Stderr, fileLogger)
			}
		}

		l = zerolog.New(output).
			Level(logLevel).
			With().
			Timestamp().
			Caller().
			Logger()

		l.Info().Msg("Logger get initialized!")

	})

	return l
}

func strToLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "disabled":
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}
