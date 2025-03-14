package request

type GetAlertEventsSvcReq struct {
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime"`

	*PageParam // 告警分页参数

	// TODO Analyzed Filter
	// 告警事件来自于Clickhouse, 查询告警事件关联的服务需要先预处理, 完成到服务的关联
	// 针对关联的服务/入口进行过滤,
	// 这意味着分页场景下, 为了得到正确的数量, 应从头分析每一个事件,直到填充完目标页

}

type AlertImpactRequest struct {
	// 告警事件编号
	StartTime int64 `form:"startTime" json:"startTime" binding:"required"`               // 查询开始时间
	EndTime   int64 `form:"endTime" json:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间

	AlertImpactFilter
	AnalyzeSize int `form:"analyzeSize" json:"analyzeSize"` // 最大分析数量
}

type AlertImpactFilter struct {
	Services   []string `json:"services"`
	Endpoints  []string `json:"endpoints"`
	Namespaces []string `json:"namespaces"`
}

type GetDefectDetectExecListRequest struct {
	*PageParam
}

type GetDefectDetectRuleListRequest struct {
	*PageParam
}
