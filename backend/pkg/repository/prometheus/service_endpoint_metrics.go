package prometheus

type ServiceEndpointMetrics struct {
	EndpointKey

	REDMetrics REDMetrics
	LatencyData   []Points
	ErrorRateData []Points
	TPMData       []Points
}

func (e *ServiceEndpointMetrics) InitEmptyGroup(key ConvertFromLabels) MetricGroup {
	return &ServiceEndpointMetrics {
		EndpointKey: key.(EndpointKey),
	}
}

func (e *ServiceEndpointMetrics) AppendGroupIfNotExist(_ MGroupName, metricName MName) bool {
	return metricName == LATENCY
}

func (e *ServiceEndpointMetrics) SetValue(metricGroup MGroupName, metricName MName, value float64) {
	e.REDMetrics.SetValue(metricGroup, metricName, value)
}


func (e *ServiceEndpointMetrics) SetValues(_ MGroupName, metricName MName, points []Points) {
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
