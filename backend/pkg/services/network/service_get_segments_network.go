package network

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetSpanSegmentsMetrics(req *request.SpanSegmentMetricsRequest) (response.SpanSegmentMetricsResponse, error) {
	netSegments, err := s.chRepo.GetNetworkSpanSegments(req.TraceId, req.SpanId)
	if err != nil {
		return nil, err
	}
	result := make(response.SpanSegmentMetricsResponse)
	for _, segment := range netSegments {
		if segment.SpanId == "" {
			continue
		}
		segmentLatency, ok := result[segment.SpanId]
		if !ok {
			segmentLatency = &response.SegmentLatency{}
			result[segment.SpanId] = segmentLatency
		}

		duration := response.Duration{
			StartTime:        segment.StartTime.UnixMicro(),
			EndTime:          segment.EndTime.UnixMicro(),
			ResponseDuration: segment.ResponseDuration,
		}
		switch segment.ObservationPoint {
		case "c-p":
			segmentLatency.ClientProcess = duration
		case "c":
			segmentLatency.ClientNic = duration
		case "c-nd":
			segmentLatency.ClientK8sNodeNic = duration
		case "s-p":
			segmentLatency.ServerProcess = duration
		case "s":
			segmentLatency.ServerNic = duration
		case "s-nd":
			segmentLatency.ServerK8sNodeNic = duration
		}
	}
	return result, nil
}
