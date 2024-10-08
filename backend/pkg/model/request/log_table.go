package request

import "fmt"

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type BufferEngineConfig struct {
	NumLayers uint `json:"numLayers"`
	MinTime   uint `json:"minTime"`
	MaxTime   uint `json:"maxTime"`
	MinRows   uint `json:"minRows"`
	MaxRows   uint `json:"maxRows"`
	MinBytes  uint `json:"minBytes"`
	MaxBytes  uint `json:"maxBytes"`
}

type LogTableRequest struct {
	DataBase  string             `json:"database"`
	TableName string             `json:"tableName"`
	Cluster   string             `json:"cluster"`
	TTL       uint               `json:"ttl"`
	Fields    []Field            `json:"fields"`
	Buffer    BufferEngineConfig `json:"buffer"`
}

func (q *LogTableRequest) ClusterString() string {
	if q.Cluster == "" {
		return ""
	}
	return fmt.Sprintf("ON CLUSTER %s", q.Cluster)
}

func (q *LogTableRequest) FillerValue() {
	if q.TTL == 0 {
		q.TTL = 7
	}
	if q.TableName == "" {
		q.TableName = "apo_logs"
	}
	if q.DataBase == "" {
		q.DataBase = "default"
	}
	if q.Buffer.NumLayers == 0 {
		q.Buffer.NumLayers = 16
	}
	if q.Buffer.MinTime == 0 {
		q.Buffer.MinTime = 10
	}
	if q.Buffer.MaxTime == 0 {
		q.Buffer.MaxTime = 100
	}
	if q.Buffer.MinRows == 0 {
		q.Buffer.MinRows = 1000000
	}
	if q.Buffer.MaxRows == 0 {
		q.Buffer.MaxRows = 10000000
	}
	if q.Buffer.MinBytes == 0 {
		q.Buffer.MinBytes = 10000000
	}
	if q.Buffer.MaxBytes == 0 {
		q.Buffer.MaxBytes = 100000000
	}
}
