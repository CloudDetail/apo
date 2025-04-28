// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package serviceoverview

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetThreshold(ctx core.Context, level string, serviceName string, endPoint string) (res response.GetThresholdResponse, err error) {
	threshold, err := s.dbRepo.GetOrCreateThreshold(ctx, serviceName, endPoint, level)
	if err != nil {
		return res, err
	}
	res.Log = threshold.Log
	res.ErrorRate = threshold.ErrorRate
	res.Tps = threshold.Tps
	res.Latency = threshold.Latency
	return res, nil
}
