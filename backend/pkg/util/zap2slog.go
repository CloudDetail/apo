// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ slog.Handler = &ZapSlogHandler{}

type ZapSlogHandler struct {
	logger *zap.Logger
}

func NewZapSlogHandler(zapLogger *zap.Logger) *slog.Logger {
	return slog.New(&ZapSlogHandler{logger: zapLogger})
}

func (h *ZapSlogHandler) Enabled(_ context.Context, level slog.Level) bool {
	zapLevel := slogToZapLevel(level)
	return h.logger.Core().Enabled(zapLevel)
}

func (h *ZapSlogHandler) Handle(ctx context.Context, record slog.Record) error {
	fields := make([]zap.Field, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, zap.Any(attr.Key, attr.Value.Any()))
		return true
	})

	zapLevel := slogToZapLevel(record.Level)

	ce := h.logger.Check(zapLevel, record.Message)
	if ce != nil {
		ce.Write(fields...)
	}
	return nil
}

func (h *ZapSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := make([]zap.Field, len(attrs))
	for i, attr := range attrs {
		fields[i] = zap.Any(attr.Key, attr.Value.Any())
	}
	return &ZapSlogHandler{
		logger: h.logger.With(fields...),
	}
}

func (h *ZapSlogHandler) WithGroup(name string) slog.Handler {
	return &ZapSlogHandler{
		logger: h.logger.Named(name),
	}
}

func slogToZapLevel(level slog.Level) zapcore.Level {
	switch {
	case level < slog.LevelInfo:
		return zapcore.DebugLevel
	case level < slog.LevelWarn:
		return zapcore.InfoLevel
	case level < slog.LevelError:
		return zapcore.WarnLevel
	default:
		return zapcore.ErrorLevel
	}
}
