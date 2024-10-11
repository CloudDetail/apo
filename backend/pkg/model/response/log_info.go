package response

type LogTableInfoResponse struct {
	LogTables map[string][]LogTable `json:"logTables"`
}

type LogTable struct {
	TableName string `json:"tableName"`
	Cluster   string `json:"cluster"`
}
