package response

type LogQueryResponse struct {
	Limited       int                      `json:"limited"`
	HiddenFields  []string                 `json:"hiddenFields"`
	DefaultFields []string                 `json:"defaultFields"`
	Logs          []map[string]interface{} `json:"logs"`
	Query         string                   `json:"query"`
	Cost          int64                    `json:"cost"`
}
