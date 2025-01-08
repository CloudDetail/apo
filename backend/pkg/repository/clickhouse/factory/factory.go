// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package factory

import "github.com/CloudDetail/apo/backend/pkg/model/request"

type SQLFactory interface {
	CreateTableSQL(params *request.LogTableRequest) string
	DropTableSQL(params *request.LogTableRequest) string
}

func GetCreateTableSQL(params *request.LogTableRequest) []string {
	sqlfactorys := []SQLFactory{
		&NullTableFactory{},
		&BufferTableFactory{},
		&LogTableFactory{},
		&ViewTableFactory{},
	}
	sqls := []string{}
	for _, factory := range sqlfactorys {
		sqls = append(sqls, factory.CreateTableSQL(params))
	}
	if params.Cluster != "" {
		sqls = append(sqls, sqlfactorys[2].(*LogTableFactory).CreateDistributedTableSQL(params))
	}
	return sqls
}

func GetDropTableSQL(params *request.LogTableRequest) []string {
	sqlfactorys := []SQLFactory{
		&NullTableFactory{},
		&BufferTableFactory{},
		&LogTableFactory{},
		&ViewTableFactory{},
	}
	sqls := []string{}
	if params.Cluster != "" {
		sqls = append(sqls, sqlfactorys[2].(*LogTableFactory).DropDistributedTableSQL(params))
	}
	for _, factory := range sqlfactorys {
		sqls = append(sqls, factory.DropTableSQL(params))

	}
	return sqls
}

// Delete view first, then adjust log, and then create view
// The distributed table adjusts the local table first, and then the distributed table.
func GetUpdateTableSQLByFields(params *request.LogTableRequest, old []request.Field) []string {
	var sqls []string
	viewfactory := &ViewTableFactory{}
	logfactory := &LogTableFactory{}
	sqls = append(sqls, viewfactory.DropTableSQL(params))
	logSql := logfactory.UpdateTableSQL(params, old, false)
	if len(logSql) > 0 {
		sqls = append(sqls, logSql)
	}
	if params.Cluster != "" && len(logSql) > 0 {
		sqls = append(sqls, logfactory.UpdateTableSQL(params, old, true))
	}
	sqls = append(sqls, viewfactory.CreateTableSQL(params))
	return sqls
}
