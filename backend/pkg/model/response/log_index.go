package response

type LogIndexResponse struct {
	Indexs []IndexItem `json:"indexs"`
	Err    string      `json:"error"`
}

type IndexItem struct {
	IndexName string  `json:"indexName"`
	Count     uint64  `json:"count"`
	Percent   float64 `json:"percent"`
}
