package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Level      string
	Pretty     bool
	TimeFormat string
}

var baseLogger zerolog.Logger

func Init(cfg *Config) {
	if cfg == nil {
		cfg = &Config{
			Level:      "info",
			Pretty:     true,
			TimeFormat: time.RFC3339,
		}
	}

	zerolog.TimeFieldFormat = cfg.TimeFormat

	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var output io.Writer = os.Stdout
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
	}

	// CallerMarshalFunc shortens file paths to just filename:line
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		return short + ":" + itoa(line)
	}

	baseLogger = zerolog.New(output).With().Timestamp().Logger()
	log.Logger = baseLogger
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(b[pos:])
}

func InitDevelopment() {
	Init(&Config{
		Level:  "debug",
		Pretty: true,
	})
}

func InitProduction() {
	Init(&Config{
		Level:  "info",
		Pretty: false,
	})
}

func Info(msg string) {
	baseLogger.Info().Caller(1).Msg(msg)
}

func Infof(format string, v ...interface{}) {
	baseLogger.Info().Caller(1).Msgf(format, v...)
}

func Warnf(format string, v ...interface{}) {
	baseLogger.Warn().Caller(1).Msgf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	baseLogger.Error().Caller(1).Msgf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	baseLogger.Fatal().Caller(1).Msgf(format, v...)
}
