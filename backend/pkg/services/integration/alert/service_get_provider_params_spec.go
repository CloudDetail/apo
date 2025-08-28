// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/provider"
)

func (s *service) GetProviderParamsSpec(ctx core.Context, sourceType string) *response.GetAlertProviderParamsSpecResponse {
	if len(sourceType) == 0 {
		var res = make(map[string]provider.ProviderType)
		for typ, item := range provider.ProviderRegistry {
			res[typ] = item
		}

		return &response.GetAlertProviderParamsSpecResponse{
			ProviderTypes: res,
		}
	}

	if item, ok := provider.ProviderRegistry[sourceType]; ok {
		return &response.GetAlertProviderParamsSpecResponse{
			ProviderType:  &item,
			ProviderTypes: map[string]provider.ProviderType{sourceType: item},
		}
	}

	item := provider.BasicEncoder(sourceType)
	return &response.GetAlertProviderParamsSpecResponse{
		ProviderType:  &item,
		ProviderTypes: map[string]provider.ProviderType{sourceType: item},
	}
}
