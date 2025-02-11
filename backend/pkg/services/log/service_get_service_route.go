// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceRoute(req *request.GetServiceRouteRequest) (*response.GetServiceRouteResponse, error) {
	serviceNames := []string{}
	now := time.Now()
	currentTimestamp := now.UnixMicro()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	sevenDaysAgoTimestamp := sevenDaysAgo.UnixMicro()
	for _, service := range req.Service {
		instances, err := s.promRepo.GetActiveInstanceList(sevenDaysAgoTimestamp, currentTimestamp, service, nil)
		if err != nil {
			return nil, err
		}
		for instanceName := range instances.GetInstanceIdMap() {
			parts := strings.Split(instanceName, "-")
			if len(parts) >= 3 {
				serviceNames = append(serviceNames, strings.Join(parts[:len(parts)-2], "-"))
				break
			}
		}
	}

	return &response.GetServiceRouteResponse{
		RouteRule: map[string]string{"k8s.pod.name": strings.Join(serviceNames, ",")},
	}, nil
}
