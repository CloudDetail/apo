// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package k8s

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s service) GetNamespaceList() (*response.GetNamespaceListResponse, error) {
	list, err := s.k8sRepo.GetNamespaceList()
	if err != nil {
		return nil, err
	}
	return &response.GetNamespaceListResponse{
		NamespaceList: list,
	}, nil
}

func (s service) GetNamespaceInfo(req *request.GetNamespaceInfoRequest) (*response.GetNamespaceInfoResponse, error) {
	info, err := s.k8sRepo.GetNamespaceInfo(req.Namespace)
	if err != nil {
		return nil, err
	}
	return &response.GetNamespaceInfoResponse{
		Namespace: info,
	}, nil
}
