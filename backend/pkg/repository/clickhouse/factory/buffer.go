package factory

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

var (
	bufferSQL = `
CREATE TABLE IF NOT EXISTS %s.%s_buffer %s
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
ENGINE = Buffer('%s', '%s_null', %d, %d, %d, %d, %d, %d, %d);
`
	dropBufferSQL = `DROP TABLE IF EXISTS %s.%s_buffer %s;`
)

var _ SQLFactory = (*BufferTableFactory)(nil)

type BufferTableFactory struct {
}

func (b *BufferTableFactory) CreateTableSQL(params *request.LogTableRequest) string {
	return fmt.Sprintf(bufferSQL,
		params.DataBase, params.TableName, params.ClusterString(),
		params.DataBase, params.TableName,
		params.Buffer.NumLayers,
		params.Buffer.MinTime, params.Buffer.MaxTime,
		params.Buffer.MinRows, params.Buffer.MaxRows,
		params.Buffer.MinBytes, params.Buffer.MaxBytes,
	)
}

func (b *BufferTableFactory) DropTableSQL(params *request.LogTableRequest) string {
	return fmt.Sprintf(dropBufferSQL, params.DataBase, params.TableName, params.ClusterString())
}
