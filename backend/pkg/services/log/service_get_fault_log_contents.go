// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

// GetFaultLogContent implements Service.
func (s *service) GetFaultLogContent(ctx_core core.Context, req *request.GetFaultLogContentRequest) (*response.GetFaultLogContentResponse, error) {
	logContest, sources, err := s.chRepo.QueryApplicationLogs(ctx_core, req)
	if err != nil {
		return nil, err
	}
	return &response.GetFaultLogContentResponse{
		Sources:	sources,
		LogContents:	logContest,
	}, nil
}
