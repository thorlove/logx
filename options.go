package log

import (
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Option func(log *Logger)

func WithLevel(lvl zapcore.Level) Option {
	return func(log *Logger) {
		log.level = lvl
	}
}
func WithTimeLayout(f string) Option {
	return func(log *Logger) {
		log.timeLayout = f
	}
}
func WithEncoder(encoder Encoder) Option {
	return func(log *Logger) {
		log.encoder = encoder
	}
}
func WithDisableConsole(disableConsole bool) Option {
	return func(log *Logger) {
		log.disableConsole = disableConsole
	}
}

func WithHooks(hs ...Hook) Option {
	if hs == nil {
		panic("hook can not be nil")
	}
	return func(log *Logger) {
		log.hooks = hs
	}
}

func WithConsoleSeparator(sep string) Option {
	return func(log *Logger) {
		log.sep = sep
	}
}

func WithField(v Loggable) Option {
	return func(log *Logger) {
		log.fields = v.Loggable()
	}
}

func WithFileWriter(file string) Option {
	if file == "" {
		panic("file path can not be empty")
	}
	dir := path.Dir(file)
	if err := os.MkdirAll(dir, 0766); err != nil {
		panic(err)
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0766)
	if err != nil {
		panic(err)
	}
	return func(log *Logger) {
		log.file = f
	}
}

// WithRotationFileWriter 创建一个带有日志滚动策略的文件写入器。
//
// 参数:
//   - file: 日志文件的路径和名称。
//   - maxSize: 单个日志文件的最大大小（以MB为单位）。
//   - maxBackups: 保留的旧日志文件的最大数量。
//   - maxAge: 旧日志文件的最大保存天数。
//   - compress: 是否压缩旧日志文件。
//
// 返回值:
//   - io.Writer: 用于写入日志。
func WithRotationFileWriter(file string, maxSize, maxAge, maxBackups int, compress bool) Option {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0766); err != nil {
		panic(err)
	}
	return func(log *Logger) {
		log.file = &lumberjack.Logger{
			Filename:   file,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   compress,
		}
	}
}
func withDefaults(opt []Option) []Option {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "app",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(DefaultTimeLayout))
		},
		ConsoleSeparator: DefaultConsoleSeparator,
		EncodeDuration:   zapcore.MillisDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder, // 全路径编码器
	}
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
	return append([]Option{
		WithLevel(DefaultLevel),
		WithTimeLayout(DefaultTimeLayout),
		WithEncoder(jsonEncoder),
		WithDisableConsole(DefaultDisableConsole),
		WithRotationFileWriter(DefaultFile, DefaultMaxSize, DefaultMaxAge, DefaultMaxBackups, DefaultCompress),
		WithConsoleSeparator(DefaultConsoleSeparator),
	}, opt...)
}
