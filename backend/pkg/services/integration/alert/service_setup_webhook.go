// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"errors"
	"fmt"

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

	var webhookURL string
	if len(req.URL) == 0 {
		// only first create
		host := ctx.GetHeader("Origin")
		webhookURL = fmt.Sprintf("%s/api/alertinput/event/source?sourceId=%s", host, req.SourceID)
	} else {
		webhookURL = req.URL
	}

	return provider.SetupWebhook(ctx, webhookURL)
}
