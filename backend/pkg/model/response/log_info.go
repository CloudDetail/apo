package response

type LogTableInfoResponse struct {
	LogTables map[string][]LogTable `json:"logTables"`
	Err       string                `json:"error"`
}

type LogTable struct {
	TableName string `json:"tableName"`
	Cluster   string `json:"cluster"`
}
