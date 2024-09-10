package prometheus

type EndpointMetrics struct {
	EndpointKey

	DelaySource   *float64 //延时主要来源
	AlertCount    int
	NamespaceList []string // 包含该端点的Namespace

	// TODO DelaySource值为nil和值为0是两种场景。
	//  nil表示没有查询到数据（可能没有这个指标），显示未知；0表示无网络占比，显示自身
	IsLatencyExceeded   bool
	IsErrorRateExceeded bool
	IsTPSExceeded       bool

	Avg1MinLatencyMutationRate float64 //延时突变率
	Avg1MinErrorMutationRate   float64 //错误率突变率

	REDMetrics REDMetrics

	LatencyData   []Points // 延时时间段的数据
	ErrorRateData []Points // 错误率时间段的数据
	TPMData       []Points // TPM 时间段的数据
}

func (e *EndpointMetrics) InitEmptyGroup(key ConvertFromLabels) MetricGroup {
	return &EndpointMetrics{
		EndpointKey: key.(EndpointKey),
	}
}

func (e *EndpointMetrics) AppendGroupIfNotExist(_ MGroupName, metricName MName) bool {
	return metricName == LATENCY
}

func (e *EndpointMetrics) SetValue(metricGroup MGroupName, metricName MName, value float64) {
	e.REDMetrics.SetValue(metricGroup, metricName, value)
}
