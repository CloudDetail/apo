// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

type ServiceMetrics struct {
	ServiceKey

	REDMetrics    REDMetrics
	LatencyData   []Points
	ErrorRateData []Points
	TPMData       []Points
}

func (e *ServiceMetrics) InitEmptyGroup(key ConvertFromLabels) MetricGroup {
	return &ServiceMetrics{
		ServiceKey: key.(ServiceKey),
	}
}

func (e *ServiceMetrics) AppendGroupIfNotExist(_ MGroupName, metricName MName) bool {
	return metricName == LATENCY
}

func (e *ServiceMetrics) SetValue(metricGroup MGroupName, metricName MName, value float64) {
	e.REDMetrics.SetValue(metricGroup, metricName, value)
}

func (e *ServiceMetrics) SetValues(_ MGroupName, metricName MName, points []Points) {
	var data = make([]Points, len(points))
	for idx, point := range points {
		data[idx].TimeStamp = point.TimeStamp
		data[idx].Value = AdjustREDValue(AVG, metricName, point.Value)
	}

	switch metricName {
	case LATENCY:
		e.LatencyData = data
	case ERROR_RATE:
		e.ErrorRateData = data
	case THROUGHPUT:
		e.TPMData = data
	}
}
