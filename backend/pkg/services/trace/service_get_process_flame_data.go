// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"encoding/json"
	"math"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"go.uber.org/zap"
)

func (s *service) GetProcessFlameGraphData(ctx core.Context, req *request.GetProcessFlameGraphRequest) (response.GetProcessFlameGraphResponse, error) {
	data, err := s.chRepo.GetFlameGraphData(ctx, req.StartTime, req.EndTime, req.NodeName,
		req.PID, -1, req.SampleType, "", "")
	if err != nil {
		return response.GetProcessFlameGraphResponse{}, err
	}

	if len(*data) == 0 {
		return response.GetProcessFlameGraphResponse{}, nil
	}

	var startTime int64 = math.MaxInt64
	var endTime int64 = 0
	var labels map[string]string
	var sampleRate uint32
	raw := make([]model.FlameBearer, len(*data))
	for i, d := range *data {
		err := json.Unmarshal([]byte(d.FlameBearer), &raw[i])
		if err != nil {
			s.logger.Warn("unmarshal to flamebaerer failed", zap.String("flamebearer", d.FlameBearer), zap.Error(err))
			continue
		}
		startTime = min(startTime, d.StartTime)
		endTime = max(endTime, d.EndTime)
		if d.TID != 0 {
			s.logger.Warn("process level flame graph's thread id doesn't equal to 0", zap.Uint32("pid", d.PID), zap.Uint32("tid", d.TID))
		}
		if labels == nil {
			labels = d.Labels
		} else if !checkLabelsEqual(labels, d.Labels) {
			s.logger.Warn("labels doesn't equal in the same process", zap.Any("src", labels), zap.Any("dst", labels))
		}

		if sampleRate == 0 {
			sampleRate = d.SampleRate
		}
	}
	t := &model.Tree{}
	for _, graph := range raw {
		t.MergeFlameGraph(&graph)
	}
	mergedFlame := model.NewFlameGraph(t, req.MaxNodes)
	bearerStr, err := json.Marshal(&mergedFlame)
	if err != nil {
		return response.GetProcessFlameGraphResponse{}, err
	}
	return response.GetProcessFlameGraphResponse{
		StartTime:   startTime,
		EndTime:     endTime,
		Labels:      labels,
		SampleRate:  sampleRate,
		SampleType:  req.SampleType,
		PID:         uint32(req.PID),
		TID:         0,
		FlameBearer: string(bearerStr),
	}, nil
}

func checkLabelsEqual(src, dst map[string]string) bool {
	if len(src) != len(dst) {
		return false
	}

	for k, v := range src {
		val, ok := dst[k]
		if !ok {
			return false
		}
		if v != val {
			return false
		}
	}
	return true
}
