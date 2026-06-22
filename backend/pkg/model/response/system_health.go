// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

type SystemHealthOverviewResp struct {
	Summary       SystemHealthOverviewSummary `json:"summary"`
	Systems       []SystemHealthSystem        `json:"systems"`
	PendingIssues []SystemHealthPendingIssue  `json:"pendingIssues"`
	Window        string                      `json:"window"`
	GeneratedAt   string                      `json:"generatedAt"`
	Source        string                      `json:"source"`
}

type SystemHealthSystemDetailResp struct {
	System      SystemHealthSystem `json:"system"`
	Window      string             `json:"window"`
	GeneratedAt string             `json:"generatedAt"`
	Source      string             `json:"source"`
}

type SystemHealthOverviewSummary struct {
	TotalSystems       int `json:"totalSystems"`
	AbnormalSystems    int `json:"abnormalSystems"`
	LowScoreSystems    int `json:"lowScoreSystems"`
	AbnormalMetrics    int `json:"abnormalMetrics"`
	SevereMetrics      int `json:"severeMetrics"`
	NewAbnormalMetrics int `json:"newAbnormalMetrics"`
	NoDataSystems      int `json:"noDataSystems"`
	Services           int `json:"services"`
	AbnormalServices   int `json:"abnormalServices"`
}

type SystemHealthSystem struct {
	ID            int64                               `json:"id"`
	GroupID       int64                               `json:"groupId"`
	Name          string                              `json:"name"`
	Score         int                                 `json:"score"`
	Status        string                              `json:"status"`
	Metrics       map[string]string                   `json:"metrics"`
	MetricDetails map[string]SystemHealthMetricDetail `json:"metricDetails"`
	Services      []SystemHealthService               `json:"services"`
	Source        string                              `json:"source"`
}

type SystemHealthService struct {
	Name          string                              `json:"name"`
	Desc          string                              `json:"desc"`
	Score         int                                 `json:"score"`
	Status        string                              `json:"status"`
	Metrics       map[string]string                   `json:"metrics"`
	MetricDetails map[string]SystemHealthMetricDetail `json:"metricDetails"`
}

type SystemHealthMetricDetail struct {
	Status     string                        `json:"status"`
	Value      string                        `json:"value"`
	Note       string                        `json:"note"`
	Unit       string                        `json:"unit,omitempty"`
	Source     string                        `json:"source,omitempty"`
	Series     []SystemHealthMetricPoint     `json:"series,omitempty"`
	Thresholds []SystemHealthMetricThreshold `json:"thresholds,omitempty"`
}

type SystemHealthMetricPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type SystemHealthMetricThreshold struct {
	Label  string  `json:"label"`
	Value  float64 `json:"value"`
	Status string  `json:"status"`
}

type SystemHealthPendingIssue struct {
	ID           string `json:"id"`
	GroupID      int64  `json:"groupId"`
	SystemName   string `json:"systemName"`
	ServiceName  string `json:"serviceName"`
	MetricKey    string `json:"metricKey"`
	MetricLabel  string `json:"metricLabel"`
	Status       string `json:"status"`
	Value        string `json:"value"`
	Note         string `json:"note"`
	ServiceScore int    `json:"serviceScore"`
	SystemScore  int    `json:"systemScore"`
}

type SystemHealthPendingIssueAnalyzeResp struct {
	AIStatus      string                       `json:"aiStatus"`
	Source        string                       `json:"source"`
	Endpoint      string                       `json:"endpoint,omitempty"`
	WorkflowRunID string                       `json:"workflowRunId,omitempty"`
	Error         string                       `json:"error,omitempty"`
	Issue         SystemHealthPendingIssue     `json:"issue"`
	Analysis      string                       `json:"analysis"`
	Impact        string                       `json:"impact"`
	ActionPlan    []SystemHealthActionPlanStep `json:"actionPlan"`
	Validation    string                       `json:"validation"`
	Risk          string                       `json:"risk"`
	DataBoundary  string                       `json:"dataBoundary"`
}

type SystemHealthReportResp struct {
	Report       SystemHealthReport        `json:"report"`
	ScoreFactors []SystemHealthScoreFactor `json:"scoreFactors"`
	Issues       []SystemHealthIssue       `json:"issues"`
}

type SystemHealthReportRecordListResp struct {
	Records  []SystemHealthReportRecordSummary `json:"records"`
	Total    int64                             `json:"total"`
	Page     int                               `json:"page"`
	PageSize int                               `json:"pageSize"`
}

type SystemHealthReportRecordSummary struct {
	RecordID         string `json:"recordId"`
	GroupID          int64  `json:"groupId"`
	Status           string `json:"status"`
	Score            int    `json:"score"`
	Summary          string `json:"summary"`
	Window           string `json:"window"`
	StartTime        int64  `json:"startTime"`
	EndTime          int64  `json:"endTime"`
	GeneratedAt      string `json:"generatedAt"`
	IssueCount       int    `json:"issueCount"`
	ActiveAlertCount int    `json:"activeAlertCount"`
	AIStatus         string `json:"aiStatus"`
	CreatedAt        string `json:"createdAt"`
}

type SystemHealthReportRecordResp struct {
	Record SystemHealthReportRecordSummary `json:"record"`
	Data   SystemHealthReportResp          `json:"data"`
}

type SystemHealthReport struct {
	Status           string                 `json:"status"`
	Score            int                    `json:"score"`
	AvailableSources []string               `json:"availableSources"`
	ResourceUsage    int                    `json:"resourceUsage"`
	InfraAlerts      int                    `json:"infraAlerts"`
	ServiceRiskCount int                    `json:"serviceRiskCount"`
	ActiveAlertCount int                    `json:"activeAlertCount"`
	LogRiskCount     int                    `json:"logRiskCount"`
	Summary          string                 `json:"summary"`
	Window           string                 `json:"window"`
	GeneratedAt      string                 `json:"generatedAt"`
	AIAnalysis       SystemHealthAIAnalysis `json:"aiAnalysis"`
}

type SystemHealthAIAnalysis struct {
	Status        string `json:"status"`
	Source        string `json:"source"`
	Endpoint      string `json:"endpoint,omitempty"`
	WorkflowRunID string `json:"workflowRunId,omitempty"`
	Error         string `json:"error,omitempty"`
}

type SystemHealthScoreFactor struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Value  int    `json:"value"`
	Tone   string `json:"tone"`
	Reason string `json:"reason"`
}

type SystemHealthIssue struct {
	Key                   string                       `json:"key"`
	AggregationKey        string                       `json:"aggregationKey"`
	Priority              string                       `json:"priority"`
	Title                 string                       `json:"title"`
	Status                string                       `json:"status"`
	Scope                 string                       `json:"scope"`
	AffectedObjects       []SystemHealthAffectedObject `json:"affectedObjects"`
	ActiveAlerts          int                          `json:"activeAlerts"`
	DeduplicatedAlerts    int                          `json:"deduplicatedAlerts"`
	RawEventCount         int                          `json:"rawEventCount"`
	Owner                 string                       `json:"owner"`
	SLA                   string                       `json:"sla"`
	WorkflowStatus        string                       `json:"workflowStatus"`
	Progress              string                       `json:"progress"`
	ScoreImpact           int                          `json:"scoreImpact"`
	RiskBreakdown         SystemHealthRiskBreakdown    `json:"riskBreakdown"`
	UpdatedAt             string                       `json:"updatedAt"`
	StartedAt             string                       `json:"startedAt"`
	MergeBasis            string                       `json:"mergeBasis"`
	MergeReasons          []string                     `json:"mergeReasons"`
	DataQuality           []string                     `json:"dataQuality"`
	MissingData           []string                     `json:"missingData"`
	EBPFFinding           SystemHealthEBPFFinding      `json:"ebpfFinding"`
	SystemProblem         string                       `json:"systemProblem"`
	AIConclusion          string                       `json:"aiConclusion"`
	RootCause             string                       `json:"rootCause,omitempty"`
	MetricAnalysis        string                       `json:"metricAnalysis,omitempty"`
	EBPFRole              string                       `json:"ebpfRole"`
	Impact                string                       `json:"impact"`
	DataSources           []string                     `json:"dataSources"`
	Actions               []string                     `json:"actions"`
	ActionPlan            []SystemHealthActionPlanStep `json:"actionPlan"`
	ExistingAlertAnalyses []SystemHealthAlertAnalysis  `json:"existingAlertAnalyses,omitempty"`
	AlertTrend            *SystemHealthAlertTrend      `json:"alertTrend,omitempty"`
	Evidence              []SystemHealthEvidence       `json:"evidence"`
}

type SystemHealthAffectedObject struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type SystemHealthRiskBreakdown struct {
	EntryImpactScore    int `json:"entryImpactScore"`
	AlertSeverityScore  int `json:"alertSeverityScore"`
	RedDegradationScore int `json:"redDegradationScore"`
	PropagationScore    int `json:"propagationScore"`
	PathRisk            int `json:"pathRisk"`
}

type SystemHealthEBPFFinding struct {
	Observed string `json:"observed"`
	Excluded string `json:"excluded"`
	Narrowed string `json:"narrowed"`
}

type SystemHealthActionPlanStep struct {
	Phase      string `json:"phase"`
	Target     string `json:"target"`
	Action     string `json:"action"`
	Validation string `json:"validation"`
	Risk       string `json:"risk"`
	Source     string `json:"source,omitempty"`
}

type SystemHealthAlertAnalysis struct {
	EventID        string `json:"eventId,omitempty"`
	AlertID        string `json:"alertId,omitempty"`
	AlertName      string `json:"alertName,omitempty"`
	AlertDirection string `json:"alertDirection,omitempty"`
	AnalyzeRunID   string `json:"analyzeRunId,omitempty"`
	AnalyzeError   string `json:"analyzeError,omitempty"`
	CheckOutput    string `json:"checkOutput,omitempty"`
	Validity       string `json:"validity,omitempty"`
}

type SystemHealthAlertTrend struct {
	Window                 string                          `json:"window,omitempty"`
	WindowText             string                          `json:"windowText,omitempty"`
	StartTime              string                          `json:"startTime,omitempty"`
	EndTime                string                          `json:"endTime,omitempty"`
	OccurrenceCount        int                             `json:"occurrenceCount"`
	EventCount             int                             `json:"eventCount"`
	TotalCount             int                             `json:"totalCount,omitempty"`
	RawEventCount          int                             `json:"rawEventCount,omitempty"`
	LongestDurationSeconds float64                         `json:"longestDurationSeconds,omitempty"`
	MaxDurationSeconds     float64                         `json:"maxDurationSeconds,omitempty"`
	TotalDurationSeconds   float64                         `json:"totalDurationSeconds,omitempty"`
	Summary                string                          `json:"summary,omitempty"`
	ActivityText           string                          `json:"activityText,omitempty"`
	Problem                string                          `json:"problem,omitempty"`
	AlertTypes             []SystemHealthAlertTrendType    `json:"alertTypes,omitempty"`
	Segments               []SystemHealthAlertTrendSegment `json:"segments,omitempty"`
}

type SystemHealthAlertTrendType struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type SystemHealthAlertTrendSegment struct {
	AlertID         string  `json:"alertId,omitempty"`
	AlertName       string  `json:"alertName,omitempty"`
	RuleName        string  `json:"ruleName,omitempty"`
	Title           string  `json:"title,omitempty"`
	Summary         string  `json:"summary,omitempty"`
	Type            string  `json:"type,omitempty"`
	Severity        string  `json:"severity,omitempty"`
	Status          string  `json:"status,omitempty"`
	StartTime       string  `json:"startTime,omitempty"`
	StartedAt       string  `json:"startedAt,omitempty"`
	EndTime         string  `json:"endTime,omitempty"`
	EndedAt         string  `json:"endedAt,omitempty"`
	DurationSeconds float64 `json:"durationSeconds,omitempty"`
	Count           int     `json:"count"`
	EventCount      int     `json:"eventCount"`
	AlertCount      int     `json:"alertCount,omitempty"`
	IsActive        bool    `json:"isActive,omitempty"`
}

type SystemHealthEvidence struct {
	Type             string                                `json:"type"`
	Source           string                                `json:"source,omitempty"`
	Severity         string                                `json:"severity,omitempty"`
	Query            string                                `json:"query"`
	Result           string                                `json:"result"`
	AlertEvents      []SystemHealthAlertEventEvidence      `json:"alertEvents,omitempty"`
	Metric           *SystemHealthMetricEvidence           `json:"metric,omitempty"`
	PolarisBreakdown *SystemHealthPolarisBreakdownEvidence `json:"polarisBreakdown,omitempty"`
	ObjectStatuses   []SystemHealthObjectStatusEvidence    `json:"objectStatuses,omitempty"`
}

type SystemHealthAlertEventEvidence struct {
	EventID   string            `json:"eventId,omitempty"`
	AlertID   string            `json:"alertId,omitempty"`
	Name      string            `json:"name"`
	Group     string            `json:"group,omitempty"`
	Source    string            `json:"source,omitempty"`
	Status    string            `json:"status,omitempty"`
	Severity  string            `json:"severity,omitempty"`
	StartedAt string            `json:"startedAt,omitempty"`
	UpdatedAt string            `json:"updatedAt,omitempty"`
	EndedAt   string            `json:"endedAt,omitempty"`
	Summary   string            `json:"summary,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

type SystemHealthMetricEvidence struct {
	RuleGroup       string                             `json:"ruleGroup,omitempty"`
	RuleName        string                             `json:"ruleName,omitempty"`
	MetricName      string                             `json:"metricName,omitempty"`
	Expression      string                             `json:"expression,omitempty"`
	Step            string                             `json:"step,omitempty"`
	StartTime       int64                              `json:"startTime,omitempty"`
	EndTime         int64                              `json:"endTime,omitempty"`
	SeriesReturned  int                                `json:"seriesReturned"`
	SeriesMatched   int                                `json:"seriesMatched"`
	PointCount      int                                `json:"pointCount"`
	FirstValue      float64                            `json:"firstValue"`
	LastValue       float64                            `json:"lastValue"`
	AverageValue    float64                            `json:"averageValue"`
	MinValue        float64                            `json:"minValue"`
	MaxValue        float64                            `json:"maxValue"`
	MaxDeviation    float64                            `json:"maxDeviation"`
	Delta           float64                            `json:"delta"`
	ChangePercent   *float64                           `json:"changePercent,omitempty"`
	MatchedLabels   map[string]string                  `json:"matchedLabels,omitempty"`
	MatchExpression string                             `json:"matchExpression,omitempty"`
	EventScoped     bool                               `json:"eventScoped"`
	Message         string                             `json:"message,omitempty"`
	Series          []SystemHealthMetricSeriesEvidence `json:"series,omitempty"`
	PointLimit      int                                `json:"pointLimit,omitempty"`
	PointsTruncated bool                               `json:"pointsTruncated,omitempty"`
}

type SystemHealthPolarisBreakdownEvidence struct {
	ServiceName string                                     `json:"serviceName,omitempty"`
	ContentKey  string                                     `json:"contentKey,omitempty"`
	Target      string                                     `json:"target,omitempty"`
	StartTime   int64                                      `json:"startTime,omitempty"`
	EndTime     int64                                      `json:"endTime,omitempty"`
	Latency     *SystemHealthPolarisBreakdownItemEvidence  `json:"latency,omitempty"`
	Resources   []SystemHealthPolarisBreakdownItemEvidence `json:"resources,omitempty"`
	TPS         *SystemHealthPolarisBreakdownItemEvidence  `json:"tps,omitempty"`
}

type SystemHealthPolarisBreakdownItemEvidence struct {
	Key             string                        `json:"key"`
	Name            string                        `json:"name"`
	ResourceKind    string                        `json:"resourceKind,omitempty"`
	Unit            string                        `json:"unit,omitempty"`
	PointCount      int                           `json:"pointCount"`
	FirstValue      float64                       `json:"firstValue"`
	LastValue       float64                       `json:"lastValue"`
	AverageValue    float64                       `json:"averageValue"`
	MinValue        float64                       `json:"minValue"`
	MaxValue        float64                       `json:"maxValue"`
	Delta           float64                       `json:"delta"`
	ChangePercent   *float64                      `json:"changePercent,omitempty"`
	Abnormal        bool                          `json:"abnormal,omitempty"`
	Reason          string                        `json:"reason,omitempty"`
	Points          []SystemHealthMetricDataPoint `json:"points,omitempty"`
	PointLimit      int                           `json:"pointLimit,omitempty"`
	PointsTruncated bool                          `json:"pointsTruncated,omitempty"`
}

type SystemHealthMetricSeriesEvidence struct {
	Labels map[string]string             `json:"labels,omitempty"`
	Points []SystemHealthMetricDataPoint `json:"points,omitempty"`
}

type SystemHealthMetricDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type SystemHealthObjectStatusEvidence struct {
	Type       string                                `json:"type"`
	Name       string                                `json:"name"`
	Namespace  string                                `json:"namespace,omitempty"`
	Status     string                                `json:"status,omitempty"`
	Ready      *bool                                 `json:"ready,omitempty"`
	Restarts   int                                   `json:"restarts,omitempty"`
	NodeName   string                                `json:"nodeName,omitempty"`
	AgeSeconds int64                                 `json:"ageSeconds,omitempty"`
	Message    string                                `json:"message,omitempty"`
	Labels     map[string]string                     `json:"labels,omitempty"`
	Conditions []SystemHealthObjectConditionEvidence `json:"conditions,omitempty"`
	Containers []SystemHealthContainerStatusEvidence `json:"containers,omitempty"`
}

type SystemHealthObjectConditionEvidence struct {
	Type    string `json:"type,omitempty"`
	Status  string `json:"status,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

type SystemHealthContainerStatusEvidence struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int32  `json:"restartCount"`
	State        string `json:"state,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Image        string `json:"image,omitempty"`
}
