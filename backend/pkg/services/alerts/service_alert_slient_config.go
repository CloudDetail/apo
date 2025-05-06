package alerts

import (
	"errors"

	"github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) GetSlienceConfigByAlertID(alertID string) (*slienceconfig.AlertSlienceConfig, error) {
	if !s.enableInnerReceiver {
		return nil, errors.New("inner alert is not open")
	}
	return s.receivers.GetSlienceConfigByAlertID(alertID)
}

func (s *service) ListSlienceConfig() ([]slienceconfig.AlertSlienceConfig, error) {
	if !s.enableInnerReceiver {
		return nil, errors.New("inner alert is not open")
	}
	return s.receivers.ListSlienceConfig()
}

func (s *service) SetSlienceConfigByAlertID(req *request.SetAlertSlienceConfigRequest) error {
	if !s.enableInnerReceiver {
		return errors.New("inner alert is not open")
	}
	return s.receivers.SetSlienceConfigByAlertID(req.AlertID, req.ForDuration)
}

func (s *service) RemoveSlienceConfigByAlertID(alertID string) error {
	if !s.enableInnerReceiver {
		return errors.New("inner alert is not open")
	}
	return s.receivers.RemoveSlienceConfigByAlertID(alertID)
}
