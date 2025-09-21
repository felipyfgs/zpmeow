package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	waLog "go.mau.fi/whatsmeow/util/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...interface{})
	Info(msg string)
	Infof(format string, args ...interface{})
	Warn(msg string)
	Warnf(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	Fatal(msg string)
	Fatalf(format string, args ...interface{})

	With() LogContext
	Sub(module string) Logger
}

type LogContext interface {
	Str(key, val string) LogContext
	Int(key string, val int) LogContext
	Bool(key string, val bool) LogContext
	Dur(key string, val time.Duration) LogContext
	Time(key string, val time.Time) LogContext
	Err(err error) LogContext
	Logger() Logger
}

type ZerologLogger struct {
	logger zerolog.Logger
	module string
}

type ZerologContext struct {
	context zerolog.Context
	base    *ZerologLogger
}

var globalLogger Logger

type LoggerConfig interface {
	GetLevel() string
	GetFormat() string
	GetConsoleColor() bool
	GetFileEnabled() bool
	GetFilePath() string
	GetFileMaxSize() int
	GetFileMaxBackups() int
	GetFileMaxAge() int
	GetFileCompress() bool
	GetFileFormat() string
}

func Initialize(config LoggerConfig) Logger {
	level := parseLogLevel(config.GetLevel())
	zerolog.SetGlobalLevel(level)

	zerolog.TimeFieldFormat = time.RFC3339

	var writers []io.Writer

	if config.GetFormat() == "console" || config.GetFormat() == "" {
		var out io.Writer = os.Stdout

		if runtime.GOOS == "windows" {
			out = colorable.NewColorableStdout()
		}

		useColor := shouldUseColor(out, config.GetConsoleColor())

		consoleWriter := zerolog.ConsoleWriter{
			Out:        out,
			TimeFormat: "02-01-2006 15:04:05",
			NoColor:    !useColor,
		}
		writers = append(writers, consoleWriter)
	}

	if config.GetFileEnabled() {
		logDir := filepath.Dir(config.GetFilePath())
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Printf("Failed to create log directory: %v\n", err)
		}

		fileWriter := &lumberjack.Logger{
			Filename:   config.GetFilePath(),
			MaxSize:    config.GetFileMaxSize(),
			MaxBackups: config.GetFileMaxBackups(),
			MaxAge:     config.GetFileMaxAge(),
			Compress:   config.GetFileCompress(),
		}
		writers = append(writers, fileWriter)
	}

	var writer io.Writer
	if len(writers) == 1 {
		writer = writers[0]
	} else if len(writers) > 1 {
		writer = io.MultiWriter(writers...)
	} else {
		writer = os.Stdout
	}

	logger := zerolog.New(writer).With().
		Timestamp().
		Logger()

	globalLogger = &ZerologLogger{
		logger: logger,
		module: "main",
	}

	return globalLogger
}

func GetLogger() Logger {
	if globalLogger == nil {
		globalLogger = &ZerologLogger{
			logger: log.Logger,
			module: "default",
		}
	}
	return globalLogger
}

func SetLogger(logger Logger) {
	globalLogger = logger
}

func GetWALogger(module string) waLog.Logger {
	return NewWALogger(module)
}

func NewWALogger(module string) waLog.Logger {
	logger := GetLogger().Sub(module)
	return &WALoggerAdapter{
		logger: logger,
		module: module,
	}
}

type WALoggerAdapter struct {
	logger Logger
	module string
}

func (w *WALoggerAdapter) Errorf(msg string, args ...interface{}) {
	w.logger.Errorf(msg, args...)
}

func (w *WALoggerAdapter) Warnf(msg string, args ...interface{}) {
	w.logger.Warnf(msg, args...)
}

func (w *WALoggerAdapter) Infof(msg string, args ...interface{}) {
	w.logger.Infof(msg, args...)
}

func (w *WALoggerAdapter) Debugf(msg string, args ...interface{}) {
	w.logger.Debugf(msg, args...)
}

func (w *WALoggerAdapter) Sub(module string) waLog.Logger {
	return NewWALogger(w.module + "." + module)
}

func TruncateID(id string) string {
	if len(id) <= 16 {
		return id
	}
	return id[:8] + "..." + id[len(id)-8:]
}

func shouldUseColor(out io.Writer, configColor bool) bool {
	if forceColor := os.Getenv("FORCE_COLOR"); forceColor != "" {
		return forceColor != "0" && forceColor != "false"
	}

	if !configColor {
		return false
	}

	if f, ok := out.(*os.File); ok {
		return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}

	if runtime.GOOS == "windows" {
		return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	}

	return false
}

func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

func (l *ZerologLogger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

func (l *ZerologLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func (l *ZerologLogger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

func (l *ZerologLogger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *ZerologLogger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

func (l *ZerologLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l *ZerologLogger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

func (l *ZerologLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

func (l *ZerologLogger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

func (l *ZerologLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

func (l *ZerologLogger) With() LogContext {
	ctx := l.logger.With()
	return &ZerologContext{
		context: ctx,
		base:    l,
	}
}

func (l *ZerologLogger) Sub(module string) Logger {
	fullModule := l.module
	if module != "" {
		if l.module != "" {
			fullModule = l.module + "." + module
		} else {
			fullModule = module
		}
	}
	return &ZerologLogger{
		logger: l.logger,
		module: fullModule,
	}
}

func (c *ZerologContext) Str(key, val string) LogContext {
	if len(val) > 50 && (strings.Contains(key, "id") || strings.Contains(key, "uuid") || strings.Contains(key, "hash")) {
		val = TruncateID(val)
	}
	c.context = c.context.Str(key, val)
	return c
}

func (c *ZerologContext) Int(key string, val int) LogContext {
	c.context = c.context.Int(key, val)
	return c
}

func (c *ZerologContext) Bool(key string, val bool) LogContext {
	c.context = c.context.Bool(key, val)
	return c
}

func (c *ZerologContext) Dur(key string, val time.Duration) LogContext {
	c.context = c.context.Dur(key, val)
	return c
}

func (c *ZerologContext) Time(key string, val time.Time) LogContext {
	c.context = c.context.Time(key, val)
	return c
}

func (c *ZerologContext) Err(err error) LogContext {
	c.context = c.context.Err(err)
	return c
}

func (c *ZerologContext) Logger() Logger {
	return &ZerologLogger{
		logger: c.context.Logger(),
		module: c.base.module,
	}
}
