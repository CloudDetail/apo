package request

type QueryLogParseRequest struct {
	DataBase  string `form:"dataBase" json:"dataBase"`
	TableName string `form:"tableName" json:"tableName"`
}

type UpdateLogParseRequest struct {
	DataBase  string            `json:"dataBase"`
	TableName string            `json:"tableName"`
	ParseInfo string            `json:"parseInfo"`
	ParseName string            `json:"parseName"`
	RouteRule map[string]string `json:"routeRule"`
	ParseRule string            `json:"parseRule"`
}

type AddLogParseRequest struct {
	ParseName string            `json:"parseName"`
	Service   string            `json:"serviceName"`
	ParseInfo string            `json:"parseInfo"`
	RouteRule map[string]string `json:"routeRule"`
	ParseRule string            `json:"parseRule"`
	LogTable  LogTable          `json:"logTable"`
}

type GetServiceRouteRequest struct {
	Service []string `form:"serviceNames"`
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
