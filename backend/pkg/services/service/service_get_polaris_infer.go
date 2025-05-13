// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

// GetPolarisInfer implements Service.
func (s *service) GetPolarisInfer(ctx_core core.Context, req *request.GetPolarisInferRequest) (*response.GetPolarisInferResponse, error) {
	return s.polRepo.QueryPolarisInfer(req)
}
