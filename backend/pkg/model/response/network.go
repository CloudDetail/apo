package response

type PodMapResponse struct {
	Columns []string `json:"columns"`
	Schemas []struct {
		LabelType string `json:"label_type"`
		PreAs     string `json:"pre_as"`
		Type      int    `json:"type"`
		Unit      string `json:"unit"`
		ValueType string `json:"value_type"`
	} `json:"schemas"`
	Values [][]interface{} `json:"values"`
}

type SpanSegmentMetricsResponse struct {
}
