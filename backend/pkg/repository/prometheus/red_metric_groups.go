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

func (m *REDMetrics) SetValue(metricGroup MGroupName, metricName MName, value float64) {

	switch metricGroup {
	case REALTIME:
		switch metricName {
		case LATENCY:
			if math.IsInf(value, 1) {
				value = RES_MAX_VALUE
			} else {
				micros := value / 1e3
				m.Realtime.Latency = &micros
			}
		case ERROR_RATE:
			if math.IsInf(value, 1) {
				value = RES_MAX_VALUE
			} else {
				errorRatePercent := value * 100
				m.Realtime.ErrorRate = &errorRatePercent
			}
		}
	case AVG:
		switch metricName {
		case LATENCY:
			if math.IsInf(value, 1) {
				value = RES_MAX_VALUE
			} else {
				micros := value / 1e3
				m.Avg.Latency = &micros
			}
		case ERROR_RATE:
			if math.IsInf(value, 1) {
				value = RES_MAX_VALUE
			} else {
				errorRatePercent := value * 100
				m.Avg.ErrorRate = &errorRatePercent
			}
		case THROUGHPUT:
			if math.IsInf(value, 1) {
				value = RES_MAX_VALUE
			} else {
				tpm := value * 60
				m.Avg.TPM = &tpm
			}
		}
	case DOD:
		var radio float64
		if math.IsInf(value, 1) {
			radio = RES_MAX_VALUE
		} else {
			radio = (value - 1) * 100
		}
		switch metricName {
		case LATENCY:
			m.DOD.Latency = &radio
		case ERROR_RATE:
			m.DOD.ErrorRate = &radio
		case THROUGHPUT:
			m.DOD.TPM = &radio
		}
	case WOW:
		var radio float64
		if math.IsInf(value, 1) {
			radio = RES_MAX_VALUE
		} else {
			radio = (value - 1) * 100
		}
		switch metricName {
		case LATENCY:
			m.WOW.Latency = &radio
		case ERROR_RATE:
			m.WOW.ErrorRate = &radio
		case THROUGHPUT:
			m.WOW.TPM = &radio
		}
	}
}

func (m *REDMetrics) CleanUPNullValue() {
	if m.Avg.Latency != nil && *m.Avg.Latency > 0 {
		if m.DOD.Latency == nil {
			value := RES_MAX_VALUE
			m.DOD.Latency = &value
		}
		if m.WOW.Latency == nil {
			value := RES_MAX_VALUE
			m.WOW.Latency = &value
		}
	}

	if m.Avg.ErrorRate != nil && *m.Avg.ErrorRate > 0 {
		if m.DOD.ErrorRate == nil {
			value := RES_MAX_VALUE
			m.DOD.ErrorRate = &value
		}
		if m.WOW.ErrorRate == nil {
			value := RES_MAX_VALUE
			m.WOW.ErrorRate = &value
		}
	}

	if m.Avg.TPM != nil && *m.Avg.TPM > 0 {
		if m.DOD.TPM == nil {
			value := RES_MAX_VALUE
			m.DOD.TPM = &value
		}
		if m.WOW.TPM == nil {
			value := RES_MAX_VALUE
			m.WOW.TPM = &value
		}
	}
}

type REDMetric struct {
	Latency   *float64
	ErrorRate *float64
	TPM       *float64
}
