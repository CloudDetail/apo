// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse/factory"
)

func (ch *chRepo) CreateLogTable(ctx core.Context, params *request.LogTableRequest) ([]string, error) {
	sqls := factory.GetCreateTableSQL(params)
	for _, sql := range sqls {
		err := ch.GetContextDB(ctx).Exec(ctx.GetContext(), sql)
		if err != nil {
			return nil, err
		}
	}
	return sqls, nil
}

func (ch *chRepo) DropLogTable(ctx core.Context, params *request.LogTableRequest) ([]string, error) {
	sqls := factory.GetDropTableSQL(params)
	for _, sql := range sqls {
		err := ch.GetContextDB(ctx).Exec(ctx.GetContext(), sql)
		if err != nil {
			return nil, err
		}
	}
	return sqls, nil
}

func (ch *chRepo) UpdateLogTable(ctx core.Context, req *request.LogTableRequest, old []request.Field) ([]string, error) {
	sqls := factory.GetUpdateTableSQLByFields(req, old)
	for _, sql := range sqls {
		err := ch.GetContextDB(ctx).Exec(ctx.GetContext(), sql)
		if err != nil {
			return nil, err
		}
	}
	return sqls, nil
}
