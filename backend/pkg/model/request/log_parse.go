package request

type QueryLogParseRequest struct {
	DataBase  string `form:"dataBase" json:"dataBase"`
	TableName string `form:"tableName" json:"tableName"`
}

type UpdateLogParseRequest struct {
	DataBase  string            `json:"dataBase"`
	TableName string            `json:"tableName"`
	ParseName string            `json:"parseName"`
	RouteRule map[string]string `json:"routeRule"`
	ParseRule string            `json:"parseRule"`
}

type AddLogParseRequest struct {
	ParseName string            `json:"parseName"`
	ParseInfo string            `json:"parseInfo"`
	RouteRule map[string]string `json:"routeRule"`
	ParseRule string            `json:"parseRule"`
	TableName string            `json:"tableName"`
	LogTable  LogTable          `json:"logTable"`
}

type LogTable struct {
	TTL    uint               `json:"ttl"`
	Fields []Field            `json:"fields"`
	Buffer BufferEngineConfig `json:"buffer"`
}

type DeleteLogParseRequest struct {
	DataBase  string `json:"dataBase"`
	TableName string `json:"tableName"`
	ParseName string `json:"parseName"`
}
