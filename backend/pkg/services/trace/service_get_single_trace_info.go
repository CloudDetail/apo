// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) GetSingleTraceID(ctx_core core.Context, req *request.GetSingleTraceInfoRequest) (string, error) {
	result, err := s.jaegerRepo.GetSingleTrace(req.TraceID)
	if err != nil {
		return "", err
	}
	return result, nil
}
