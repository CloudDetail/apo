package dataplane

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	ServiceInstanceAPI = "/datasource/queryServiceInstances"
)

type queryApmServiceInstancesRequest struct {
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	Cluster     string `json:"clusterId"`
	Source      string `json:"dataSource"`
	ServiceId   string `json:"serviceId"`
	ServiceName string `json:"serviceName"`
}

type queryApmServiceInstancesResponse struct {
	Success  bool                        `json:"success"`
	Data     []*model.ApmServiceInstance `json:"data"`
	ErrorMsg string                      `json:"errorMsg"`
}

func (repo *dataplaneRepo) QueryApmServiceInstances(startTime int64, endTime int64, service *model.Service) ([]*model.ApmServiceInstance, error) {
	queryParams := &queryApmServiceInstancesRequest{
		StartTime:   startTime,
		EndTime:     endTime,
		Cluster:     service.ClusterId,
		Source:      service.Source,
		ServiceId:   service.Id,
		ServiceName: service.Name,
	}
	requestBody, err := json.Marshal(queryParams)
	if err != nil {
		return nil, fmt.Errorf("query param is invalid, %v", err)
	}

	resp, err := repo.client.Post(
		fmt.Sprintf("%s%s", repo.address, ServiceInstanceAPI),
		"application/json",
		bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var response queryApmServiceInstancesResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if !response.Success {
		return nil, errors.New(response.ErrorMsg)
	}
	return response.Data, nil
}
