package request

type AlertImpactRequest struct {
	// 告警事件编号
	EventID   string `form:"eventId" json:"eventId"`                                      // 需要分析的告警的编号
	StartTime int64  `form:"startTime" json:"startTime" binding:"required"`               // 查询开始时间
	EndTime   int64  `form:"endTime" json:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Step      int64  `form:"step" json:"step" binding:"required"`                         // 查询步长(us)

	AnalyzeSize int `form:"analyzeSize" json:"analyzeSize"` // 分析数量

	*PageParam // 告警分页参数
}

type GetDefectDetectExecListRequest struct {
	*PageParam
}

type GetDefectDetectRuleListRequest struct {
	*PageParam
}
