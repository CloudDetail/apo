// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

// InstanceMetrics instance granularity metric results
type InstanceMetrics struct {
	InstanceKey

	REDMetrics REDMetrics

	LogDayOverDay   *float64
	LogWeekOverWeek *float64
	LogAVGData      *float64

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
