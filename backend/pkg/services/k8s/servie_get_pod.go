// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s service) GetPodList(req *request.GetPodListRequest) (*response.GetPodListResponse, error) {
	list, err := s.k8sRepo.GetPodList(req.Namespace)
	if err != nil {
		return nil, err
	}
	return &response.GetPodListResponse{
		PodList: list,
	}, nil
}

func (s service) GetPodInfo(req *request.GetPodInfoRequest) (*response.GetPodInfoResponse, error) {
	info, err := s.k8sRepo.GetPodInfo(req.Namespace, req.Pod)
	if err != nil {
		return nil, err
	}
	return &response.GetPodInfoResponse{
		Pod: info,
	}, nil
}
