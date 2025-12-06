package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	Pretty     bool   // Use pretty console output (for development)
	TimeFormat string // Time format (default: RFC3339)
}

// baseLogger is the underlying logger without caller skip
var baseLogger zerolog.Logger

// Init initializes the global logger with the given configuration
func Init(cfg *Config) {
	if cfg == nil {
		cfg = &Config{
			Level:      "info",
			Pretty:     true,
			TimeFormat: time.RFC3339,
		}
	}

	// Set time format
	zerolog.TimeFieldFormat = cfg.TimeFormat

	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer = os.Stdout
	if cfg.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
	}

	// Set global logger with CallerWithSkipFrameCount to skip wrapper functions
	// Skip 2 frames: the wrapper function and the actual Log call
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

// itoa converts int to string (simple implementation)
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

// InitDevelopment initializes logger for development environment
func InitDevelopment() {
	Init(&Config{
		Level:  "debug",
		Pretty: true,
	})
}

// InitProduction initializes logger for production environment
func InitProduction() {
	Init(&Config{
		Level:  "info",
		Pretty: false,
	})
}

// Logger returns a new logger instance with optional context fields
func Logger() zerolog.Logger {
	return log.Logger
}

// Debug logs a debug message
func Debug(msg string) {
	baseLogger.Debug().Caller(1).Msg(msg)
}

// Debugf logs a formatted debug message
func Debugf(format string, v ...interface{}) {
	baseLogger.Debug().Caller(1).Msgf(format, v...)
}

// Info logs an info message
func Info(msg string) {
	baseLogger.Info().Caller(1).Msg(msg)
}

// Infof logs a formatted info message
func Infof(format string, v ...interface{}) {
	baseLogger.Info().Caller(1).Msgf(format, v...)
}

// Warn logs a warning message
func Warn(msg string) {
	baseLogger.Warn().Caller(1).Msg(msg)
}

// Warnf logs a formatted warning message
func Warnf(format string, v ...interface{}) {
	baseLogger.Warn().Caller(1).Msgf(format, v...)
}

// Error logs an error message
func Error(msg string) {
	baseLogger.Error().Caller(1).Msg(msg)
}

// Errorf logs a formatted error message
func Errorf(format string, v ...interface{}) {
	baseLogger.Error().Caller(1).Msgf(format, v...)
}

// ErrorErr logs an error with error object
func ErrorErr(err error, msg string) {
	baseLogger.Error().Caller(1).Err(err).Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	baseLogger.Fatal().Caller(1).Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, v ...interface{}) {
	baseLogger.Fatal().Caller(1).Msgf(format, v...)
}

// WithField returns a logger with a field added
func WithField(key string, value interface{}) zerolog.Logger {
	return baseLogger.With().Interface(key, value).Logger()
}

// WithFields returns a logger with multiple fields added
func WithFields(fields map[string]interface{}) zerolog.Logger {
	ctx := baseLogger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return ctx.Logger()
}

// WithError returns a logger with error field added
func WithError(err error) zerolog.Logger {
	return baseLogger.With().Err(err).Logger()
}

// RequestLogger returns a logger for HTTP requests
func RequestLogger(method, path, ip string) zerolog.Logger {
	return baseLogger.With().
		Str("method", method).
		Str("path", path).
		Str("ip", ip).
		Logger()
}
