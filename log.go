package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"time"
)

const (
	DefaultConsoleSeparator = " | "
	DefaultTimeLayout       = time.RFC3339
	DefaultMaxSize          = 10
	DefaultMaxAge           = 7
	DefaultMaxBackups       = 3
	DefaultCompress         = true
	DefaultFile             = "logs/log.log"
	DefaultDisableConsole   = false
	DefaultLevel            = zapcore.InfoLevel
)

type Level int8

type Logger struct {
	logger         *zap.Logger
	sugar          *zap.SugaredLogger
	level          zapcore.Level
	sep            string
	file           io.Writer
	encoder        Encoder
	hooks          []Hook
	timeLayout     string
	disableConsole bool
	fields         map[string]string
}

// Hook is a collection of hooks that are synchronously
// triggered for each logging event.
type Hook = func(zapcore.Entry) error
type Encoder = zapcore.Encoder

// F is a set of fields
type F map[string]string

type Loggable interface {
	Loggable() map[string]string
}

// Loggable allows Logger.With to consume an F.
func (f F) Loggable() map[string]string {
	return f
}

func (l Logger) Fatalw(msg string, kvs ...interface{}) { l.sugar.Fatalw(msg, kvs...) }
func (l Logger) Fatalf(fmt string, kvs ...interface{}) { l.sugar.Fatalf(fmt, kvs...) }
func (l Logger) Fatalln(v ...interface{})              { l.sugar.Fatalln(v...) }

func (l Logger) DPanicw(msg string, kvs ...interface{}) { l.sugar.DPanicw(msg, kvs...) }
func (l Logger) DPanicf(fmt string, v ...interface{})   { l.sugar.DPanicf(fmt, v...) }
func (l Logger) DPanicln(v ...interface{})              { l.sugar.DPanicln(v...) }

func (l Logger) Debugw(msg string, kvs ...interface{}) { l.sugar.Debugw(msg, kvs...) }
func (l Logger) Debugf(fmt string, v ...interface{})   { l.sugar.Debugf(fmt, v...) }
func (l Logger) Debugln(v ...interface{})              { l.sugar.Debugln(v...) }

func (l Logger) Infow(msg string, kvs ...interface{}) { l.sugar.Infow(msg, kvs...) }
func (l Logger) Infof(fmt string, v ...interface{})   { l.sugar.Infof(fmt, v...) }
func (l Logger) Infoln(v ...interface{})              { l.sugar.Infoln(v...) }

func (l Logger) Warnw(msg string, kvs ...interface{}) { l.sugar.Warnw(msg, kvs...) }
func (l Logger) Warnf(fmt string, v ...interface{})   { l.sugar.Warnf(fmt, v...) }
func (l Logger) Warnln(v ...interface{})              { l.sugar.Warnln(v...) }

func (l Logger) Errorw(msg string, kvs ...interface{}) { l.sugar.Errorw(msg, kvs...) }
func (l Logger) Errorf(fmt string, v ...interface{})   { l.sugar.Errorf(fmt, v...) }
func (l Logger) Errorln(v ...interface{})              { l.sugar.Errorln(v...) }

func New(opts ...Option) *Logger {
	l := &Logger{}
	for _, f := range withDefaults(opts) {
		f(l)
	}

	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= l.level && lvl < zapcore.ErrorLevel
	})
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= l.level && lvl >= zapcore.ErrorLevel
	})

	core := zapcore.NewTee()
	stdOut := zapcore.Lock(os.Stdout)
	stdErr := zapcore.Lock(os.Stderr)
	if !l.disableConsole {
		core = zapcore.NewTee(
			zapcore.NewCore(l.encoder, stdOut, lowPriority),
			zapcore.NewCore(l.encoder, stdErr, highPriority),
		)
	}

	if l.file != nil {
		core = zapcore.NewTee(core,
			zapcore.NewCore(
				l.encoder,
				zapcore.AddSync(l.file),
				zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl >= l.level }),
			))
	}

	l.logger = zap.New(core, zap.AddCaller(), zap.ErrorOutput(stdErr))
	for k, v := range l.fields {
		l.logger = l.logger.WithOptions(zap.Fields(zapcore.Field{Key: k, Type: zapcore.StringType, String: v}))
	}
	if len(l.hooks) > 0 {
		l.logger = l.logger.WithOptions(zap.Hooks(l.hooks...))
	}
	l.sugar = l.logger.Sugar()
	return l
}

func (l *Logger) Named(name string) *Logger {
	return &Logger{
		logger: l.logger.Named(name),
		sugar:  l.sugar.Named(name),
	}
}
