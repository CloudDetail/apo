package prometheus

import "math"

// RES_MAX_VALUE 返回前端的最大值，同比为该值时表示最大值
const RES_MAX_VALUE float64 = 9999999

var _ MetricGroup = &REDMetrics{}

type REDMetrics struct {
	Realtime REDMetric
	Avg      REDMetric
	DOD      REDMetric
	WOW      REDMetric
}

func (m *REDMetrics) InitEmptyGroup(_ ConvertFromLabels) MetricGroup {
	return &REDMetrics{}
}

func (m *REDMetrics) AppendGroupIfNotExist(_ MGroupName, metricName MName) bool {
	return metricName == LATENCY
}

func AdjustREDValue(metricGroup MGroupName, metricName MName, value float64) float64 {
	switch metricGroup {
	case REALTIME, AVG:
		if math.IsInf(value, 1) {
			return RES_MAX_VALUE
		} else if math.IsInf(value, -1) {
			return -RES_MAX_VALUE
		}
		switch metricName {
		case LATENCY:
			micros := value / 1e3
			return micros
		case ERROR_RATE:
			errorRatePercent := value * 100
			return errorRatePercent
		case THROUGHPUT:
			tpm := value * 60
			return tpm
		}
	case DOD, WOW:
		var radio float64
		if math.IsInf(value, 1) {
			radio = RES_MAX_VALUE
		} else if math.IsInf(value, -1) {
			radio = -RES_MAX_VALUE
		} else {
			radio = (value - 1) * 100
		}
		return radio
	}
	return value
}

func (m *REDMetrics) SetValue(metricGroup MGroupName, metricName MName, value float64) {
	adjustedValue := AdjustREDValue(metricGroup, metricName, value)

	var mg *REDMetric
	switch metricGroup {
	case REALTIME:
		mg = &m.Realtime
	case AVG:
		mg = &m.Avg
	case DOD:
		mg = &m.DOD
	case WOW:
		mg = &m.WOW
	default:
		return
	}

	switch metricName {
	case LATENCY:
		mg.Latency = &adjustedValue
	case ERROR_RATE:
		mg.ErrorRate = &adjustedValue
	case THROUGHPUT:
		mg.TPM = &adjustedValue
	}
}

type REDMetric struct {
	Latency   *float64
	ErrorRate *float64
	TPM       *float64
}
