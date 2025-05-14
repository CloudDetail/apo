// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"errors"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) GetSlienceConfigByAlertID(ctx_core core.Context, alertID string) (*slienceconfig.AlertSlienceConfig, error) {
	if !s.enableInnerReceiver {
		return nil, errors.New("inner alert is not open")
	}
	return s.receivers.GetSlienceConfigByAlertID(ctx_core, alertID)
}

func (s *service) ListSlienceConfig(ctx_core core.Context) ([]slienceconfig.AlertSlienceConfig, error) {
	if !s.enableInnerReceiver {
		return nil, errors.New("inner alert is not open")
	}
	return s.receivers.ListSlienceConfig(ctx_core)
}

func (s *service) SetSlienceConfigByAlertID(ctx_core core.Context, req *request.SetAlertSlienceConfigRequest) error {
	if !s.enableInnerReceiver {
		return errors.New("inner alert is not open")
	}
	return s.receivers.SetSlienceConfigByAlertID(ctx_core, req.AlertID, req.ForDuration)
}

func (s *service) RemoveSlienceConfigByAlertID(ctx_core core.Context, alertID string) error {
	if !s.enableInnerReceiver {
		return errors.New("inner alert is not open")
	}
	return s.receivers.RemoveSlienceConfigByAlertID(ctx_core, alertID)
}
