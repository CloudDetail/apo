// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

type EndpointMetrics struct {
	EndpointKey

	/*
		DelaySource shows the main reason for time-cost
			[0,0.5): delay mainly from their self;
			[0.5,1]: delay mainly from their downstream
			'nil': no network metric found, unable to analyze major delay causes
	*/
	DelaySource   *float64
	AlertCount    int
	NamespaceList []string // Namespace containing the endpoint

	IsLatencyExceeded   bool
	IsErrorRateExceeded bool
	IsTPSExceeded       bool

	Avg1MinLatencyMutationRate float64 // delayed mutation rate
	Avg1MinErrorMutationRate   float64 // error rate mutation rate

	REDMetrics REDMetrics

	LatencyData   []Points // Data of delay time period
	ErrorRateData []Points // Data for the error rate time period
	TPMData       []Points // Data for TPM time period
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

func (e *EndpointMetrics) SetValues(_ MGroupName, metricName MName, points []Points) {
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
