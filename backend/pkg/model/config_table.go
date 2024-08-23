package model

// type TableTTLMap struct {
// 	Name          string
// 	TTLExpression string
// 	OriginalDays  int
// }

type ModifyTableTTLMap struct {
	Name          string `json:"name"`
	TTLExpression string `json:"TTLExpression"`
	OriginalDays  *int   `json:"originalDays"`
}

type TablesQuery struct {
	Name             string `ch:"name" json:"name"`
	CreateTableQuery string `ch:"create_table_query" json:"createTableQuery"`
}
