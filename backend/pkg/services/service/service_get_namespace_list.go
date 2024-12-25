// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceNamespaceList(req *request.GetServiceNamespaceListRequest) (response.GetServiceNamespaceListResponse, error) {
	list, err := s.promRepo.GetNamespaceList(req.StartTime, req.EndTime)
	var resp response.GetServiceNamespaceListResponse
	if err != nil {
		return resp, err
	}

	resp.NamespaceList = list
	return resp, nil
}
