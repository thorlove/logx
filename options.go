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

// WithLevel 设置日志级别。
func WithLevel(lvl zapcore.Level) Option {
	return func(log *Logger) {
		log.level = lvl
	}
}

// WithTimeLayout 设置时间格式。
func WithTimeLayout(f string) Option {
	return func(log *Logger) {
		log.timeLayout = f
	}
}

// WithDisableConsole 禁用控制台输出。
func WithDisableConsole(disableConsole bool) Option {
	return func(log *Logger) {
		log.disableConsole = disableConsole
	}
}

// WithHooks 添加日志钩子。
func WithHooks(hs ...Hook) Option {
	if hs == nil {
		panic("hook can not be nil")
	}
	return func(log *Logger) {
		log.hooks = hs
	}
}

// WithConsoleSeparator 设置控制台分隔符。
func WithConsoleSeparator(sep string) Option {
	return func(log *Logger) {
		log.sep = sep
	}
}

// WithField 设置自定义字段。
func WithField(v Loggable) Option {
	return func(log *Logger) {
		log.fields = v.Loggable()
	}
}

// WithFileWriter 创建一个文件写入器。
// 参数:
//   - file: 文件路径。
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

// WithEncoder 设置日志编码器。
func WithEncoder(encoder Encoder) Option {
	return func(log *Logger) {
		log.encoder = encoder
	}
}

// WithSampleEncoder 设置日志编码器。
func WithSampleEncoder(t EncoderType) Option {
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
	return func(log *Logger) {
		if log.sep != "" {
			encoderConfig.ConsoleSeparator = log.sep
		}
		if t == JSON {
			log.encoder = zapcore.NewJSONEncoder(encoderConfig)
		} else if t == Console {
			log.encoder = zapcore.NewConsoleEncoder(encoderConfig)
		} else {
			panic(t)
		}
	}
}
func withDefaults(opt []Option) []Option {
	return append([]Option{
		WithLevel(DefaultLevel),
		WithTimeLayout(DefaultTimeLayout),
		WithDisableConsole(DefaultDisableConsole),
		WithConsoleSeparator(DefaultConsoleSeparator),
		WithSampleEncoder(JSON),
		WithRotationFileWriter(DefaultFile, DefaultMaxSize, DefaultMaxAge, DefaultMaxBackups, DefaultCompress),
	}, opt...)
}
