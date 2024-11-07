package network

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

type SegmentLatency struct {
	// 客户端进程发出到收到回复的信息
	clientProcess Duration
	// 客户端容器网卡发出到收到回复总时长
	clientNic uint64
	// 客户端主机网卡发出到收到回复总时长
	clientK8sNodeNic uint64
	// 服务端主机网卡发出到收到回复总时长
	serverK8sNodeNic uint64
	// 服务端容器网卡发出到收到回复总时长
	serverNic uint64
	// 客户端主机ID
	PodNodeId0 uint32
	// 服务端主机ID
	PodNodeId1 uint32
}

type Duration struct {
	// 处理网络包时间戳，单位微秒
	startTime uint64
	//
	endTime          uint64
	responseDuration uint64
}

func (s *service) GetSpanSegmentsMetrics(req *request.SpanSegmentMetricsRequest) (*response.SpanSegmentMetricsResponse, error) {
	if len(req.SpanId) == 0 {

	}
	return nil, nil
}
