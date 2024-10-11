package request

type QueryLogParseRequest struct {
	DataBase  string `json:"dataBase"`
	TableName string `json:"tableName"`
}

type UpdateLogParseRequest struct {
	DataBase  string `json:"dataBase"`
	TableName string `json:"tableName"`
	ParseName string `json:"parseName"`
	RouteRule string `json:"routeRule"`
	ParseRule string `json:"parseRule"`
}
