package response

type LogQueryResponse struct {
	Limited       int       `json:"limited"`
	HiddenFields  []string  `json:"hiddenFields"`
	DefaultFields []string  `json:"defaultFields"`
	Logs          []LogItem `json:"logs"`
	Query         string    `json:"query"`
	Cost          int64     `json:"cost"`
	Err           string    `json:"error"`
}

type LogItem struct {
	Content interface{}            `json:"content"`
	Tags    map[string]interface{} `json:"tags"`
}
