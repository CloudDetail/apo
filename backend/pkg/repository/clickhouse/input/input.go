// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package input

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/CloudDetail/apo/backend/pkg/model/input/alert"
)

type Input interface {
	InsertExtraAlertEvent(ctx context.Context, alertEvents []alert.AlertEvent, sourceFrom alert.SourceFrom) error
}

type chRepo struct {
	conn     driver.Conn
	database string
}

func NewInputRepo(conn driver.Conn, database string) (*chRepo, error) {
	return &chRepo{
		conn:     conn,
		database: database,
	}, nil
}
