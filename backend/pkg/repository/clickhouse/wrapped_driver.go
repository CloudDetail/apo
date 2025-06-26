// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"
)

type WrappedConn struct {
	driver.Conn
	logger *zap.Logger
}

func (c *WrappedConn) Select(ctx context.Context, dest any, query string, args ...any) error {
	startTime := time.Now()
	err := c.Conn.Select(ctx, dest, query, args...)
	endTime := time.Now()
	c.logger.Debug("Clickhouse Select",
		zap.String("query", query),
		zap.Any("args", args),
		zap.Int64("cost(ms)", endTime.UnixMilli()-startTime.UnixMilli()))
	return err
}

func (c *WrappedConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	startTime := time.Now()
	rows, err := c.Conn.Query(ctx, query, args...)
	endTime := time.Now()
	c.logger.Debug("Clickhouse Query",
		zap.String("query", query),
		zap.Any("args", args),
		zap.Int64("cost(ms)", endTime.UnixMilli()-startTime.UnixMilli()))
	return rows, err
}

func (c *WrappedConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	startTime := time.Now()
	rows := c.Conn.QueryRow(ctx, query, args...)
	endTime := time.Now()
	c.logger.Debug("Clickhouse QueryRow",
		zap.String("query", query),
		zap.Any("args", args),
		zap.Int64("cost(ms)", endTime.UnixMilli()-startTime.UnixMilli()))
	return rows
}

func (c *WrappedConn) Exec(ctx context.Context, query string, args ...any) error {
	startTime := time.Now()
	err := c.Conn.Exec(ctx, query, args...)
	endTime := time.Now()
	c.logger.Debug("Clickhouse Exec: {query=%s, args=%v}, cost: %d ms",
		zap.String("query", query),
		zap.Any("args", args),
		zap.Int64("cost(ms)", endTime.UnixMilli()-startTime.UnixMilli()))
	return err
}

func (c *WrappedConn) Ping(ctx context.Context) error {
	startTime := time.Now()
	err := c.Conn.Ping(ctx)
	endTime := time.Now()
	c.logger.Debug("Clickhouse Ping",
		zap.Int64("cost(ms)", endTime.UnixMilli()-startTime.UnixMilli()))
	return err
}
