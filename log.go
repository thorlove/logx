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
	DefaultMaxSize          = 25
	DefaultMaxAge           = 7
	DefaultMaxBackups       = 3
	DefaultCompress         = true
	DefaultFile             = "logs/log.log"
	DefaultDisableConsole   = false
	DefaultLevel            = InfoLevel
)

const (
	JSON    EncoderType = "json"
	Console EncoderType = "console"
)
const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

type EncoderType string
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

func New(opts ...Option) *zap.SugaredLogger {
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
	return l.logger.Sugar()
}

func (l *Logger) Named(name string) *Logger {
	return &Logger{
		logger: l.logger.Named(name),
		sugar:  l.sugar.Named(name),
	}
}
