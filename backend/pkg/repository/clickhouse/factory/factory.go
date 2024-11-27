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

// 先删除view,然后调整log，再创建view
// 分布式表先调整本地表，然后调分布式表
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
