// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

type EndpointMetrics struct {
	EndpointKey

	DelaySource   *float64 // main source of delay
	AlertCount    int
	NamespaceList []string // Namespace containing the endpoint

	// TODO DelaySource values of nil and 0 are two scenarios.
	// nil indicates that no data has been queried (this metric may not be available) and the display is unknown; 0 indicates that there is no network percentage and the display itself
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
