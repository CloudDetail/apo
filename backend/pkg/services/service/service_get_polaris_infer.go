// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

// GetPolarisInfer implements Service.
func (s *service) GetPolarisInfer(req *request.GetPolarisInferRequest) (*response.GetPolarisInferResponse, error) {
	return s.polRepo.QueryPolarisInfer(req)
}
