package factory

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

var (
	distributedlogSQL = `
CREATE TABLE IF NOT EXISTS %s.%s %s
(
timestamp          DateTime64(9) CODEC(Delta(8), ZSTD(1)),
content            String CODEC(ZSTD(1)),
source             String CODEC(ZSTD(1)),
container_id       String CODEC(ZSTD(1)),
pid                String CODEC(ZSTD(1)),
container_name 	   LowCardinality(String) CODEC(ZSTD(1)),
host_ip            LowCardinality(String) CODEC(ZSTD(1)),
host_name          LowCardinality(String) CODEC(ZSTD(1)),
k8s_namespace_name LowCardinality(String) CODEC(ZSTD(1)),
k8s_pod_name       LowCardinality(String) CODEC(ZSTD(1))%s
)
ENGINE = Distributed('%s', '%s', '%s_local', rand());
`

	logSQL = `
CREATE TABLE IF NOT EXISTS %s.%s %s
(
timestamp          DateTime64(9) CODEC(Delta(8), ZSTD(1)),
content            String CODEC(ZSTD(1)),
source             String CODEC(ZSTD(1)),
container_id       String CODEC(ZSTD(1)),
pid                String CODEC(ZSTD(1)),
container_name 	   LowCardinality(String) CODEC(ZSTD(1)),
host_ip            LowCardinality(String) CODEC(ZSTD(1)),
host_name          LowCardinality(String) CODEC(ZSTD(1)),
k8s_namespace_name LowCardinality(String) CODEC(ZSTD(1)),
k8s_pod_name       LowCardinality(String) CODEC(ZSTD(1)),
%s
INDEX idx_content content TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 1
)
%s
PARTITION BY toDate(timestamp)
ORDER BY (host_ip, timestamp)
%s
SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1;
`

	dropLogSQL = `DROP TABLE IF EXISTS %s.%s %s`

	updateLogSQL = `
ALTER TABLE %s.%s %s
%s;
`
)

const (
	distributedEngine = "ENGINE = ReplicatedMergeTree('/clickhouse/tables/{uuid}/{shard}', '{replica}')"
	mergeTreeEngine   = "ENGINE = MergeTree()"
)

var _ SQLFactory = (*LogTableFactory)(nil)

type LogTableFactory struct {
}

// CreateTableSQL implements Factory.
func (l *LogTableFactory) CreateTableSQL(params *request.LogTableRequest) string {
	var ttlExpr string
	if params.TTL > 0 {
		ttlExpr = fmt.Sprintf(`TTL toDateTime(timestamp) + toIntervalDay(%d)`, params.TTL)
	}
	var AnalyzerFiles string
	for _, field := range params.Fields {
		AnalyzerFiles += fmt.Sprintf("%s Nullable(%s),\n", field.Name, field.Type)
	}
	cluster := params.ClusterString()
	var engine string
	tablename := params.TableName
	if cluster != "" {
		tablename += "_local"
		engine = distributedEngine
	} else {
		engine = mergeTreeEngine
	}

	if !params.Replica {
		engine = mergeTreeEngine
	}

	return fmt.Sprintf(logSQL, params.DataBase, tablename, cluster,
		AnalyzerFiles, engine, ttlExpr)
}

// DropTableSQL implements Factory.
func (l *LogTableFactory) DropTableSQL(params *request.LogTableRequest) string {
	cluster := params.ClusterString()
	tablename := params.TableName
	if cluster != "" {
		tablename += "_local"
	}
	return fmt.Sprintf(dropLogSQL, params.DataBase, tablename, cluster)
}

func (l *LogTableFactory) CreateDistributedTableSQL(params *request.LogTableRequest) string {
	var AnalyzerFiles string
	for _, field := range params.Fields {
		AnalyzerFiles += fmt.Sprintf(",\n%s Nullable(%s)", field.Name, field.Type)
	}
	return fmt.Sprintf(distributedlogSQL, params.DataBase, params.TableName, params.ClusterString(),
		AnalyzerFiles, params.Cluster, params.DataBase, params.TableName)
}

func (l *LogTableFactory) DropDistributedTableSQL(params *request.LogTableRequest) string {
	return fmt.Sprintf(dropLogSQL, params.DataBase, params.TableName, params.ClusterString())
}

func (l *LogTableFactory) UpdateTableSQL(params *request.LogTableRequest, distrubted bool) string {
	var upateFiles string
	size := len(params.Fields)
	for i, field := range params.Fields {
		upateFiles += fmt.Sprintf("ADD COLUMN %s Nullable(%s)", field.Name, field.Type)
		if i < size-1 {
			upateFiles += ",\n"
		}
	}
	tablename := params.TableName
	cluster := params.ClusterString()
	if cluster != "" && !distrubted {
		tablename += "_local"
	}
	return fmt.Sprintf(updateLogSQL, params.DataBase, tablename, cluster, upateFiles)
}
