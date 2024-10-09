package response

type LogChartResponse struct {
	Histograms []*LogHistogram `json:"histograms"`
	Count      uint64          `json:"count"`
	Progress   string          `json:"progress"`
}

type LogHistogram struct {
	Count    uint64 `json:"count"`
	Progress string `json:"progress"`
	From     int64  `json:"from"`
	To       int64  `json:"to"`
}
