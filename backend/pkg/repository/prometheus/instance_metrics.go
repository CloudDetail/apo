package prometheus

// InstanceMetrics instance粒度的指标结果
type InstanceMetrics struct {
	InstanceKey

	IsLatencyExceeded   bool
	IsErrorRateExceeded bool
	IsTPSExceeded       bool

	REDMetrics      REDMetrics
	LogDayOverDay   *float64
	LogWeekOverWeek *float64

	LogAVGData    *float64
	LatencyData   []Points
	ErrorRateData []Points
	TPMData       []Points
	LogData       []Points
}

func (e *InstanceMetrics) InitEmptyGroup(key ConvertFromLabels) MetricGroup {
	return &InstanceMetrics{
		InstanceKey: key.(InstanceKey),
	}
}

func (e *InstanceMetrics) AppendGroupIfNotExist(_ MGroupName, metricName MName) bool {
	return metricName == LATENCY
}

func (e *InstanceMetrics) SetValue(metricGroup MGroupName, metricName MName, value float64) {
	e.REDMetrics.SetValue(metricGroup, metricName, value)
}
