package factory

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

var (
	nullSQL = `
CREATE TABLE IF NOT EXISTS %s.%s_null %s
(
"timestamp"          String,
"content"            String CODEC(ZSTD(1)),
"_source_"           String CODEC(ZSTD(1)),
"_container_id_"     String CODEC(ZSTD(1)),
"pid"                String CODEC(ZSTD(1)),
"container.name" 	   LowCardinality(String) CODEC(ZSTD(1)),
"host.ip"            LowCardinality(String) CODEC(ZSTD(1)),
"host.name"          LowCardinality(String) CODEC(ZSTD(1)),
"k8s.namespace.name" LowCardinality(String) CODEC(ZSTD(1)),
"k8s.pod.name"       LowCardinality(String) CODEC(ZSTD(1))
)
ENGINE = Null;
`

	dropNullSQL = `DROP TABLE IF EXISTS %s.%s_null %s;`
)

var _ SQLFactory = (*NullTableFactory)(nil)

type NullTableFactory struct {
}

func (n *NullTableFactory) CreateTableSQL(params *request.LogTableRequest) string {
	return fmt.Sprintf(nullSQL, params.DataBase, params.TableName, params.ClusterString())
}

func (n *NullTableFactory) DropTableSQL(params *request.LogTableRequest) string {
	return fmt.Sprintf(dropNullSQL, params.DataBase, params.TableName, params.ClusterString())
}
