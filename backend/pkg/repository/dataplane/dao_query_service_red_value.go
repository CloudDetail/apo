// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	ServiceRedValueAPI = "/datasource/queryServiceRedValue"
)

type queryServiceRedValueRequest struct {
	ProviderId  int    `json:"providerId"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	ServiceId   string `json:"serviceId"`
	ServiceName string `json:"serviceName"`
}

type queryServiceValueResponse struct {
	Success  bool                  `json:"success"`
	Data     *model.RedMetricValue `json:"data"`
	ErrorMsg string                `json:"errorMsg"`
}

func (repo *dataplaneRepo) QueryServiceRedValue(providerId int, startTime int64, endTime int64, service *model.Service) (*model.RedMetricValue, error) {
	queryParams := &queryServiceRedValueRequest{
		ProviderId:  providerId,
		StartTime:   startTime,
		EndTime:     endTime,
		ServiceId:   service.Id,
		ServiceName: service.Name,
	}
	requestBody, err := json.Marshal(queryParams)
	if err != nil {
		return nil, fmt.Errorf("query param is invalid, %v", err)
	}

	resp, err := repo.client.Post(
		fmt.Sprintf("%s%s", repo.address, ServiceRedValueAPI),
		"application/json",
		bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var response queryServiceValueResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if !response.Success {
		return nil, errors.New(response.ErrorMsg)
	}
	return response.Data, nil
}
