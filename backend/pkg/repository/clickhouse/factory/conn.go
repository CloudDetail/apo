// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package factory

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/CloudDetail/apo/backend/pkg/core"
)

type Conn struct {
	driver.Conn
}

func (c *Conn) GetContextDB(ctx core.Context) driver.Conn {
	return c.Conn
}
