// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger operation object to implement gormlogger.Interface
type GormLogger struct {
	ZapLogger     *zap.Logger
	SlowThreshold time.Duration
}

// NewGormLogger external calls. Instantiate a GormLogger object, example:
//
//	DB, err := gorm.Open(dbConfig, &gorm.Config{
//	    Logger: logger.NewGormLogger(),
//	})
func NewGormLogger(logger *zap.Logger) GormLogger {
	return GormLogger{
		ZapLogger:     logger,
		SlowThreshold: 200 * time.Millisecond, //slow query threshold, unit: 1‰ seconds
	}
}

// LogMode to implement the LogMode method of gormlogger.Interface
func (l GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return GormLogger{
		ZapLogger:     l.ZapLogger,
		SlowThreshold: l.SlowThreshold,
	}
}

// Info Implement the Info method of gormlogger.Interface
func (l GormLogger) Info(ctx context.Context, str string, args ...interface{}) {
	l.logger().Sugar().Debugf(str, args...)
}

// Warn implements the Warn method of gormlogger.Interface
func (l GormLogger) Warn(ctx context.Context, str string, args ...interface{}) {
	l.logger().Sugar().Warnf(str, args...)
}

// Error Implement gormlogger.Interface Error method
func (l GormLogger) Error(ctx context.Context, str string, args ...interface{}) {
	l.logger().Sugar().Errorf(str, args...)
}

// Trace implements gormlogger.Interface Trace method
func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {

	// Get run time
	elapsed := time.Since(begin)
	// Get the number of SQL requests and returns
	sql, rows := fc()

	// Common Field
	logFields := []zap.Field{
		zap.String("sql", sql),
		zap.String("time", MicrosecondsStr(elapsed)),
		zap.Int64("rows", rows),
	}

	// Gorm error
	if err != nil {
		// log not found error using warning level
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.logger().Warn("Database ErrRecordNotFound", logFields...)
		} else {
			// Other errors using the error level
			logFields = append(logFields, zap.Error(err))
			l.logger().Error("Database Error", logFields...)
		}
	}

	// Slow query log
	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		l.logger().Warn("Database Slow Log", logFields...)
	}

	// Log all SQL requests
	l.logger().Debug("Database Query", logFields...)
}

// auxiliary methods used in logger to ensure the accuracy of Zap's built-in information Caller (e.g. paginator/paginator.go:148)
func (l GormLogger) logger() *zap.Logger {

	// Skip gorm built-in calls
	var (
		gormPackage    = filepath.Join("gorm.io", "gorm")
		zapgormPackage = filepath.Join("moul.io", "zapgorm2")
	)

	// subtract a package and add zap.AddCallerSkip(1) to logger initialization.
	clone := l.ZapLogger.WithOptions(zap.AddCallerSkip(-2))

	for i := 2; i < 15; i++ {
		_, file, _, ok := runtime.Caller(i)
		switch {
		case !ok:
		case strings.HasSuffix(file, "_test.go"):
		case strings.Contains(file, gormPackage):
		case strings.Contains(file, zapgormPackage):
		default:
			// returns a new zap logger with a skipped line number
			return clone.WithOptions(zap.AddCallerSkip(i))
		}
	}
	return l.ZapLogger
}
