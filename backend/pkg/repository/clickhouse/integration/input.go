// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse/factory"
)

type Input interface {
	InsertAlertEvent(ctx core.Context, alertEvents []alert.AlertEvent, sourceFrom alert.SourceFrom) error

	InsertIncident2AlertEvent(ctx core.Context, incidentID string, alertEventID string) error
}

var _ Input = &chRepo{}

type chRepo struct {
	factory.Conn
	database string
}

func NewInputRepo(conn driver.Conn, database string) (*chRepo, error) {
	c := factory.Conn{Conn: conn}
	return &chRepo{
		Conn:     c,
		database: database,
	}, nil
}
