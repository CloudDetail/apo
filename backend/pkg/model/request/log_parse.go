package request

type LogParseRequest struct {
	DataBase  string `json:"dataBase"`
	TableName string `json:"tableName"`
}
