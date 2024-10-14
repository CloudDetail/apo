package response

type LogTableInfoResponse struct {
	Parses    []Parse    `json:"parses"`
	Instances []Instance `json:"instances"`
	Err       string     `json:"error"`
}

type Parse struct {
	DataBase  string `json:"dataBase"`
	TableName string `json:"tableName"`
	ParseName string `json:"parseName"`
	ParseInfo string `json:"parseInfo"`
}

type Instance struct {
	InstanceName string   `json:"instanceName"`
	DataBases    []DBInfo `json:"dataBases"`
}

type DBInfo struct {
	DataBase string         `json:"dataBase"`
	Tables   []LogTableInfo `json:"tables"`
}

type LogTableInfo struct {
	Cluster   string `json:"cluster"`
	TableName string `json:"tableName"`
	TimeField string `json:"timeField"`
	LogField  string `json:"logField"`
}
