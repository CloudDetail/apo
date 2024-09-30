package model

import (
	"github.com/CloudDetail/apo/backend/config"
	"strings"
)

// type TableTTLMap struct {
// 	Name          string
// 	TTLExpression string
// 	OriginalDays  int
// }

type ModifyTableTTLMap struct {
	Name          string `json:"name"`
	TTLExpression string `json:"TTLExpression"`
	OriginalDays  *int   `json:"originalDays"`
}

type TablesQuery struct {
	Name             string `ch:"name" json:"name"`
	CreateTableQuery string `ch:"create_table_query" json:"createTableQuery"`
}

type TableType struct {
	Typ    string
	Tables []Table
}
type Table struct {
	IsMaterialView bool // 暂时没用
	Name           string
}

var tableType = map[string][]Table{
	"logs": {
		{Name: "ilogtail_logs"},
	},
	"trace": {
		{Name: "span_trace"},
		{Name: "jaeger_index_local"},
		{Name: "jaeger_spans_archive_local"},
		{Name: "jaeger_spans_local"},
	},
	"k8s": {
		{Name: "k8s_events"},
	},
	"topology": {
		{Name: "service_relation"},
		{Name: "service_topology"},
	},
	"other": {
		{Name: "agent_log"},
		{Name: "alert_event"},
		{Name: "error_propagation"},
		{Name: "error_report"},
		{Name: "jvm_gc"},
		{Name: "onoff_metric"},
		{Name: "onstack_profiling"},
		{Name: "profiling_event"},
		{Name: "report_metric"},
		{Name: "slo_record"},
		{Name: "slow_report"},
	},
}

func (t Table) TableName() string {
	suffix := "_local"
	if len(config.GetCHCluster()) > 0 && !strings.HasSuffix(t.Name, suffix) {
		return t.Name + suffix
	}
	return t.Name
}

func GetAllTables() []Table {
	var tables []Table
	totalTables := 0
	for _, table := range tableType {
		totalTables += len(table)
	}

	tables = make([]Table, 0, totalTables)

	for _, table := range tableType {
		tables = append(tables, table...)
	}

	return tables
}

func IsTableExists(tableName string) bool {
	for _, table := range GetAllTables() {
		if table.Name == tableName {
			return true
		}
	}
	return false
}

func GetTables(typ string) []Table {
	return tableType[typ]
}

func TableToType() map[string]string {
	tableToType := make(map[string]string)
	for typ, tables := range tableType {
		for _, table := range tables {
			tableToType[table.TableName()] = typ
		}
	}
	return tableToType
}
