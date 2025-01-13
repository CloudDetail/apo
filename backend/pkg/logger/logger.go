// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	// DefaultLevel the default log level
	DefaultLevel = zapcore.InfoLevel

	// DefaultTimeLayout the default time layout;
	DefaultTimeLayout = time.RFC3339
)

var (
	levelMap = map[string]zapcore.Level{
		"WARN":  zap.WarnLevel,
		"ERROR": zap.ErrorLevel,
		"FATAL": zap.FatalLevel,
		"INFO":  zap.InfoLevel,
		"DEBUG": zap.DebugLevel,
	}
)

func NewLogger(opts ...Option) *zap.Logger {
	opt := &option{level: DefaultLevel, fields: make(map[string]string)}
	for _, f := range opts {
		f(opt)
	}

	timeLayout := DefaultTimeLayout
	if opt.timeLayout != "" {
		timeLayout = opt.timeLayout
	}
	// similar to zap.NewProductionEncoderConfig()
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger", // used by logger.Named(key); optional; useless
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace", // use by zap.AddStacktrace; optional; useless
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder, // lowercase encoder
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(timeLayout))
		},
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // full path encoder
	}

	core := zapcore.NewTee()
	if !opt.disableConsole {
		core = zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), opt.level)
	}
	if opt.file != nil {
		core = zapcore.NewTee(core,
			zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(opt.file),
				zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
					return lvl >= opt.level
				}),
			),
		)
	}
	logger := zap.New(core,
		zap.AddCaller(),
		zap.ErrorOutput(zapcore.AddSync(os.Stderr)),
	)

	for key, value := range opt.fields {
		logger = logger.WithOptions(zap.Fields(zapcore.Field{Key: key, Type: zapcore.StringType, String: value}))
	}
	return logger
}

// Option custom setup config
type Option func(*option)

type option struct {
	level          zapcore.Level
	fields         map[string]string
	file           io.Writer
	timeLayout     string
	disableConsole bool
}

func WithLevel(level string) Option {
	return func(opt *option) {
		if lvl, ok := levelMap[strings.ToUpper(level)]; ok {
			opt.level = lvl
		}
	}
}

// WithDebugLevel only greater than 'level' will output
func WithDebugLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.DebugLevel
	}
}

// WithInfoLevel only greater than 'level' will output
func WithInfoLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.InfoLevel
	}
}

// WithWarnLevel only greater than 'level' will output
func WithWarnLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.WarnLevel
	}
}

// WithErrorLevel only greater than 'level' will output
func WithErrorLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.ErrorLevel
	}
}

// WithField add some field(s) to log
func WithField(key, value string) Option {
	return func(opt *option) {
		opt.fields[key] = value
	}
}

// WithFileRotationP write log to some file with rotation
func WithFileRotationP(enable bool, file string, size int, num int) Option {
	return func(opt *option) {
		if enable {
			dir := filepath.Dir(file)
			if err := os.MkdirAll(dir, 0766); err != nil {
				panic(err)
			}
			opt.file = &lumberjack.Logger{ // concurrent-safed
				Filename:   file, // file path
				MaxSize:    size, // maximum size of a single file, default unit M
				MaxBackups: num,  // Keep up to 10 backups
				LocalTime:  true, // use local time
				Compress:   true, // whether to compress the disabled by default
			}
		}
	}
}

// WithTimeLayout custom time format
func WithTimeLayout(timeLayout string) Option {
	return func(opt *option) {
		opt.timeLayout = timeLayout
	}
}

// WithConsole write log to os.Stdout or os.Stderr
func WithConsole(enable bool) Option {
	return func(opt *option) {
		if !enable {
			opt.disableConsole = true
		}
	}
}
