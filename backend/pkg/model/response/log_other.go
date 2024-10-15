package response

type OtherTableResponse struct {
	OtherTables []OtherDB `json:"otherTables"`
	Err         string    `json:"error"`
}

type OtherDB struct {
	DataBase string       `json:"dataBase"`
	Tables   []OtherTable `json:"tables"`
}

type OtherTable struct {
	TableName string `json:"tableName"`
}

type OtherTableInfoResponse struct {
	Columns []Column `json:"columns"`
	Err     string   `json:"error"`
}

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type AddOtherTableResponse struct {
	Err string `json:"error"`
}

type DeleteOtherTableResponse struct {
	Err string `json:"error"`
}
