package logger

import (
	tag "nikki-noceps/serviceCatalogue/logger/tag"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _globalZapLogger = NewZapLogger("info")
var _globalZapLoggerMutex = sync.Mutex{}

type Tag interface {
	Key() string
	Value() interface{}
}

type zapLogger struct {
	zl *zap.Logger
}

// NewZapLogger returns a zap based logger
func NewZapLogger(level string) *zapLogger {
	zl := buildZapLogger(level)
	return &zapLogger{
		zl,
	}
}

// With instantiates a new logger object which implements the logger interface
// Adds tags which will always be logged whenever the logger object is used
func (l *zapLogger) WITH(tags ...Tag) Logger {
	fields := l.buildFields(tags)
	loggerWithFields := l.zl.With(fields...)
	return &zapLogger{zl: loggerWithFields}
}

// builds zapper compatible Fields from Tags
func (l *zapLogger) buildFields(tags []Tag) []zap.Field {
	fields := make([]zap.Field, len(tags))

	for i, t := range tags {
		if zt, ok := t.(tag.ZapTag); ok {
			fields[i] = zt.Field()
		} else {
			fields[i] = zap.Any(t.Key(), t.Value())
		}
	}

	return fields
}

func (l *zapLogger) DEBUG(msg string, tags ...Tag) {
	if l.zl.Core().Enabled(zap.DebugLevel) {
		msg = setDefaultMsg(msg)
		fields := l.buildFields(tags)
		l.zl.Debug(msg, fields...)
	}
}

func (l *zapLogger) INFO(msg string, tags ...Tag) {
	if l.zl.Core().Enabled(zap.InfoLevel) {
		msg = setDefaultMsg(msg)
		fields := l.buildFields(tags)
		l.zl.Info(msg, fields...)
	}
}

func (l *zapLogger) WARN(msg string, tags ...Tag) {
	if l.zl.Core().Enabled(zap.WarnLevel) {
		msg = setDefaultMsg(msg)
		fields := l.buildFields(tags)
		l.zl.Warn(msg, fields...)
	}
}

func (l *zapLogger) ERROR(msg string, tags ...Tag) {
	if l.zl.Core().Enabled(zap.ErrorLevel) {
		msg = setDefaultMsg(msg)
		fields := l.buildFields(tags)
		l.zl.Error(msg, fields...)
	}
}

func (l *zapLogger) FATAL(msg string, tags ...Tag) {
	if l.zl.Core().Enabled(zap.FatalLevel) {
		msg = setDefaultMsg(msg)
		fields := l.buildFields(tags)
		l.zl.Fatal(msg, fields...)
	}
}

func (l *zapLogger) ErrorWithCallerSkip(skip int, msg string, tags ...Tag) {
	if l.zl.Core().Enabled(zap.ErrorLevel) {
		msg = setDefaultMsg(msg)
		fields := l.buildFields(tags)

		if skip >= 0 {
			l.zl.WithOptions(zap.AddCallerSkip(skip)).Error(msg, fields...)
			return
		}

		l.zl.Error(msg, fields...)
	}
}

func DEBUG(msg string, tags ...Tag) {
	msg = setDefaultMsg(msg)
	fields := _globalZapLogger.buildFields(tags)
	_globalZapLogger.zl.WithOptions(zap.AddCallerSkip(0)).Debug(msg, fields...)
}

func INFO(msg string, tags ...Tag) {
	msg = setDefaultMsg(msg)
	fields := _globalZapLogger.buildFields(tags)
	_globalZapLogger.zl.WithOptions(zap.AddCallerSkip(0)).Info(msg, fields...)
}

func WARN(msg string, tags ...Tag) {
	msg = setDefaultMsg(msg)
	fields := _globalZapLogger.buildFields(tags)
	_globalZapLogger.zl.WithOptions(zap.AddCallerSkip(0)).Warn(msg, fields...)
}

func ERROR(msg string, tags ...Tag) {
	msg = setDefaultMsg(msg)
	fields := _globalZapLogger.buildFields(tags)
	_globalZapLogger.zl.WithOptions(zap.AddCallerSkip(0)).Error(msg, fields...)
}

func FATAL(msg string, tags ...Tag) {
	msg = setDefaultMsg(msg)
	fields := _globalZapLogger.buildFields(tags)
	_globalZapLogger.zl.WithOptions(zap.AddCallerSkip(0)).Fatal(msg, fields...)
}

func WITH(tags ...Tag) Logger {
	return _globalZapLogger.WITH(tags...)
}

func ReplaceGlobalZapLogger(zl *zapLogger) {
	_globalZapLoggerMutex.Lock()
	defer _globalZapLoggerMutex.Unlock()

	_globalZapLogger = zl
}

func setDefaultMsg(msg string) string {
	if msg == "" {
		return "none"
	}

	return msg
}

func buildZapLogger(level string) *zap.Logger {
	encodeConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		CallerKey:      "logging-at",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config := zap.Config{
		DisableStacktrace: true,
		Level:             zap.NewAtomicLevelAt(parseZapLevel(level)),
		Development:       false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encodeConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stdout"},
	}

	logger, _ := config.Build(zap.AddCallerSkip(1))
	return logger
}

// defaults to info level if incorrect level provided
// Pick up level provided and sets it as log level. It is case agnostic
func parseZapLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}
