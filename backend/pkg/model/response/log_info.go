package response

type LogTableInfoResponse struct {
	Parses    []Parse    `json:"parses"`
	LogTables []LogTable `json:"logTables"`
	Err       string     `json:"error"`
}

type Parse struct {
	DataBase   string      `json:"dataBase"`
	ParseInfos []ParseInfo `json:"parseInfos"`
}
type ParseInfo struct {
	TableName string `json:"tableName"`
	ParseName string `json:"parseName"`
}

type LogTable struct {
	DataBase string         `json:"dataBase"`
	Cluster  string         `json:"cluster"`
	Tables   []LogTableInfo `json:"tableInfos"`
}

type LogTableInfo struct {
	TableName string `json:"tableName"`
	TimeField string `json:"timeField"`
	LogField  string `json:"logField"`
}
