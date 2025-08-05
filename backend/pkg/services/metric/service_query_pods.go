// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package metric

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) QueryPods(ctx core.Context, req *request.QueryPodsRequest) (*response.QueryPodsResponse, error) {
	pods, err := s.promRepo.GetPodList(ctx, req.StartTime, req.EndTime, req.NodeName, req.Namespace, req.PodName)
	if err != nil {
		return nil, err
	}
	return &response.QueryPodsResponse{
		Pods: pods,
	}, nil
}
