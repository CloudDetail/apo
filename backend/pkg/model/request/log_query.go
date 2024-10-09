package request

type LogQueryRequest struct {
	StartTime int64  `json:"startTime" binding:"min=0"`
	EndTime   int64  `json:"endTime" binding:"required,gtfield=StartTime"`
	TableName string `json:"tableName"`
	DataBase  string `json:"dataBase"`
	Query     string `json:"query"`
	PageNum   int    `json:"pageNum"`
	PageSize  int    `json:"pageSize"`
}
