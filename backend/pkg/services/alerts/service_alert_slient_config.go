// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"errors"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) GetSlienceConfigByAlertID(ctx core.Context, alertID string) (*slienceconfig.AlertSlienceConfig, error) {
	if !s.enableInnerReceiver {
		return nil, errors.New("inner alert is not open")
	}
	return s.receivers.GetSlienceConfigByAlertID(ctx, alertID)
}

func (s *service) ListSlienceConfig(ctx core.Context) ([]slienceconfig.AlertSlienceConfig, error) {
	if !s.enableInnerReceiver {
		return nil, errors.New("inner alert is not open")
	}
	return s.receivers.ListSlienceConfig(ctx)
}

func (s *service) SetSlienceConfigByAlertID(ctx core.Context, req *request.SetAlertSlienceConfigRequest) error {
	if !s.enableInnerReceiver {
		return errors.New("inner alert is not open")
	}
	return s.receivers.SetSlienceConfigByAlertID(ctx, req.AlertID, req.ForDuration)
}

func (s *service) RemoveSlienceConfigByAlertID(ctx core.Context, alertID string) error {
	if !s.enableInnerReceiver {
		return errors.New("inner alert is not open")
	}
	return s.receivers.RemoveSlienceConfigByAlertID(ctx, alertID)
}
