// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) GetSingleTraceID(ctx core.Context, req *request.GetSingleTraceInfoRequest) (string, error) {
	result, err := s.jaegerRepo.GetSingleTrace(req.TraceID)
	if err != nil {
		return "", err
	}
	return result, nil
}
