// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"os"
	"slices"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetIntegratedCapabilities(ctx core.Context) (*response.GetIntegratedCapabilitiesResponse, error) {
	dps, err := s.dbRepo.ListDataPlane(ctx)
	if err != nil {
		return nil, err
	}

	var res response.GetIntegratedCapabilitiesResponse
	capMap := make(map[integration.TelemetryType]*integration.TelemetryCapability)
	for _, dp := range dps {
		for _, t := range integration.AllTelemetryTypes {
			cap, ok := capMap[t]
			if !ok {
				cap = &integration.TelemetryCapability{
					Type:    t,
					Enabled: false,
					Sources: []integration.DataplaneWithName{},
				}
				capMap[t] = cap
			}
			if slices.Contains(dp.DataPlaneWithType.DataPlane.Capability.Obj, string(t)) {
				dpType := dp.DataPlaneWithType.DataPlane.Typ
				dpWithName := integration.DataplaneWithName{Type: dpType, Name: dp.DataPlaneWithType.DataPlane.Name}
				if !slices.Contains(cap.Sources, dpWithName) && (dpType != "apo" || os.Getenv(showAPODataPlaneTypeENV) == "true") {
					cap.Enabled = true
					cap.Sources = append(cap.Sources, dpWithName)
				}
			}
		}
	}
	for _, cap := range capMap {
		res.TotalSupported = append(res.TotalSupported, *cap)
	}
	return &res, nil
}
