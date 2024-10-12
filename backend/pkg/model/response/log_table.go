package response

type LogTableResponse struct {
	Sqls []string `json:"sqls"`
	Err  string   `json:"error"`
}
