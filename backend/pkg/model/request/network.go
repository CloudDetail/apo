package request

type PodMapRequest struct {
	StartTime int64  `form:"statTime" json:"startTime" binding:"required,min=0"`          // 查询开始时间
	EndTime   int64  `form:"endTime" json:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Namespace string `form:"namespace"`
	Workload  string `form:"workload"`
	Protocol  string `form:"protocol"`
}

type SpanSegmentMetricsRequest struct {
	TraceId string `uri:"traceId" binding:"required"`
	SpanId  string `uri:"spanId"`
}
