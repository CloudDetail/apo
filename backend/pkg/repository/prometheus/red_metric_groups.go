package prometheus

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
			micros := value / 1e3
			m.Realtime.Latency = &micros
		case ERROR_RATE:
			errorRatePercent := value * 100
			m.Realtime.ErrorRate = &errorRatePercent
		}
	case AVG:
		switch metricName {
		case LATENCY:
			micros := value / 1e3
			m.Avg.Latency = &micros
		case ERROR_RATE:
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
		case ERROR_RATE:
			m.DOD.ErrorRate = &radio
		case THROUGHPUT:
			m.DOD.TPM = &radio
		}
	case WOW:
		radio := (value - 1) * 100
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

type REDMetric struct {
	Latency   *float64
	ErrorRate *float64
	TPM       *float64
}
