package prometheus

import "math"

// metricGroup Name
type MGroupName string

// metricName
type MName string

const (
	// metricGroup
	REALTIME MGroupName = "realtime" // endpoint时刻瞬时值
	AVG      MGroupName = "avg"      // start~endpoint之间的平均值
	DOD      MGroupName = "dod"      // start~endpoint时段和昨日日同比
	WOW      MGroupName = "wow"      // start~endpoint时段和上周周同比

	// metricName
	LATENCY    MName = "latency"
	ERROR      MName = "error"
	THROUGHPUT MName = "throughput"
)

type MetricGroupMap[K interface {
	comparable
	ConvertFromLabels
}, V MetricGroup] struct {
	// 用于返回列表
	MetricGroupList []V
	// 用于通过Key快速查询对应的MetricGroup
	MetricGroupMap map[K]V
}

func (m *MetricGroupMap[K, V]) MergeMetricResults(metricGroup MGroupName, metricName MName, metricResults []MetricResult) {
	for _, metric := range metricResults {
		if len(metric.Values) <= 0 {
			continue
		}
		var kType K
		key, ok := kType.ConvertFromLabels(metric.Metric).(K)
		if !ok {
			continue
		}
		mg, find := m.MetricGroupMap[key]
		if !find && mg.AppendGroupIfNotExist(metricGroup, metricName) {
			mg, ok = mg.initEmptyGroup(key).(V)
			if !ok {
				// 通常不会发生,这意味着initEmptyGroup返回的结构不是它本身
				continue
			}
			m.MetricGroupList = append(m.MetricGroupList, mg)
			m.MetricGroupMap[key] = mg
		} else if !find {
			continue
		}
		// 所有合并值均只包含最新时间点的结果,直接取metricResult.Values[0]
		value := metric.Values[0].Value
		if math.IsInf(value, 0) {
			continue
		}
		mg.SetValue(metricGroup, metricName, value)
	}
}

type ConvertFromLabels interface {
	ConvertFromLabels(labels Labels) ConvertFromLabels
}

type MetricGroup interface {
	initEmptyGroup(key ConvertFromLabels) MetricGroup
	AppendGroupIfNotExist(metricGroup MGroupName, metricName MName) bool
	SetValue(metricGroup MGroupName, metricName MName, value float64)
}

var _ MetricGroup = &REDMetrics{}

type REDMetrics struct {
	Realtime REDMetric
	Avg      REDMetric
	DOD      REDMetric
	WOW      REDMetric
}

func (m *REDMetrics) initEmptyGroup(_ ConvertFromLabels) MetricGroup {
	return &REDMetrics{}
}

func (m *REDMetrics) AppendGroupIfNotExist(_ MGroupName, metricName MName) bool {
	return metricName == LATENCY
}

func (m *REDMetrics) SetValue(metricGroup MGroupName, metricName MName, value float64) {
	switch metricGroup {
	case REALTIME:
		switch metricName {
		case LATENCY:
			micros := value / 1e3
			m.Realtime.Latency = &micros
		case ERROR:
			errorRatePercent := value * 100
			m.Realtime.ErrorRate = &errorRatePercent
		}
	case AVG:
		switch metricName {
		case LATENCY:
			micros := value / 1e3
			m.Avg.Latency = &micros
		case ERROR:
			errorRatePercent := value * 100
			m.Avg.ErrorRate = &errorRatePercent
		case THROUGHPUT:
			tpm := value * 60
			m.Avg.TPM = &tpm
		}
	case DOD:
		radio := (value - 1) * 100
		switch metricName {
		case LATENCY:
			m.DOD.Latency = &radio
		case ERROR:
			m.DOD.ErrorRate = &radio
		case THROUGHPUT:
			m.DOD.TPM = &radio
		}
	case WOW:
		radio := (value - 1) * 100
		switch metricName {
		case LATENCY:
			m.WOW.Latency = &radio
		case ERROR:
			m.WOW.ErrorRate = &radio
		case THROUGHPUT:
			m.WOW.TPM = &radio
		}
	}
}

type REDMetric struct {
	Latency   *float64
	ErrorRate *float64
	TPM       *float64
}
