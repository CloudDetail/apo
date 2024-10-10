package factory

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

var (
	viewSQL = `
CREATE MATERIALIZED VIEW IF NOT EXISTS %s.%s_view %s TO %s.%s
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
AS SELECT
toDateTime64(parseDateTimeBestEffort(timestamp), 9) AS timestamp,
content AS content,
_source_ AS source,
_container_id_ AS container_id,
pid as pid,
container.name AS container_name,
host.ip AS host_ip,
host.name AS host_name,
k8s.namespace.name AS k8s_namespace_name,
k8s.pod.name AS k8s_pod_name%s
FROM %s.%s_null
WHERE 1 = 1;
`

	dropViewSQL = `DROP TABLE IF EXISTS %s.%s_view %s;`
)

var _ SQLFactory = (*ViewTableFactory)(nil)

type ViewTableFactory struct {
}

func (v *ViewTableFactory) CreateTableSQL(params *request.LogTableRequest) string {
	var logFields string
	var viewFields string
	tablename := params.TableName
	for _, field := range params.Fields {
		logFields += fmt.Sprintf(",\n%s Nullable(%s)", field.Name, field.Type)
		viewFields += fmt.Sprintf(",\ntoNullable(to%s(replaceAll(JSONExtractRaw(content, '%s'), '\"', ''))) AS %s", field.Type, field.Name, field.Name)
	}
	cluster := params.ClusterString()
	if cluster != "" {
		tablename += "_local"
	}
	return fmt.Sprintf(viewSQL, params.DataBase, params.TableName, cluster, params.DataBase, tablename,
		logFields, viewFields, params.DataBase, params.TableName)
}

func (v *ViewTableFactory) DropTableSQL(params *request.LogTableRequest) string {
	return fmt.Sprintf(dropViewSQL, params.DataBase, params.TableName, params.ClusterString())
}
