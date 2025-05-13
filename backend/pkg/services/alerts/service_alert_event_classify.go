// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/dify"
	"github.com/hashicorp/golang-lru/v2/expirable"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

var cache = expirable.NewLRU[string, string](10, nil, time.Hour)

func (s *service) AlertEventClassify(ctx_core core.Context, req *request.AlertEventClassifyRequest) (*response.AlertEventClassifyResponse, error) {
	inputs, _ := json.Marshal(map[string]interface{}{
		"alertGroup":	req.AlertGroup,
		"alertName":	req.AlertName,
	})
	r, ok := cache.Get(req.AlertGroup + req.AlertName)
	if ok {
		return &response.AlertEventClassifyResponse{
			WorkflowId: r,
		}, nil
	}

	request := &dify.WorkflowRequest{
		Inputs:		inputs,
		ResponseMode:	"blocking",
		User:		"apo-backend",
	}

	difyconf := config.Get().Dify
	resp, err := s.difyRepo.WorkflowsRun(request, "Bearer "+difyconf.APIKeys.AlertClassify)
	if err != nil {
		return nil, err
	}
	if resp.Data.Status != "succeeded" {
		return nil, fmt.Errorf("workflow run failed")
	}
	var res map[string]string
	err = json.Unmarshal(resp.Data.Outputs, &res)
	if err != nil {
		return nil, err
	}

	cache.Add(req.AlertGroup+req.AlertName, res["workflowId"])
	return &response.AlertEventClassifyResponse{
		WorkflowId: res["workflowId"],
	}, nil
}
