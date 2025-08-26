// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"errors"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/services/integration/alert/provider"
)

func (s *service) SetupProviderWebhook(ctx core.Context, req *request.SetupAlertProviderWebhookRequest) error {
	sourceInfo, err := s.dbRepo.GetAlertSource(ctx, req.SourceID)
	if err != nil {
		return err
	}

	pType, find := provider.ProviderRegistry[sourceInfo.SourceType]
	if !find || !pType.SupportWebhookInstall {
		return errors.New("alert source not support setup webhook now")
	}

	provider := pType.New(sourceInfo.SourceFrom, sourceInfo.Params.Obj)
	if provider == nil {
		return errors.New("alert source not support setup webhook now")
	}
	return provider.SetupWebhook(ctx, req.URL, sourceInfo.Params.Obj)
}
