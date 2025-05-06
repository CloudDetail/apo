package alerts

import (
	"errors"

	"github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) GetSlienceConfig(alertID string) (*slienceconfig.AlertSlienceConfig, error) {
	if !s.enableInnerReceiver {
		return nil, errors.New("inner alert is not open")
	}
	return s.receivers.GetSlienceConfig(alertID)
}

func (s *service) ListSlienceConfig() ([]slienceconfig.AlertSlienceConfig, error) {
	if !s.enableInnerReceiver {
		return nil, errors.New("inner alert is not open")
	}
	return s.receivers.ListSlienceConfig()
}

func (s *service) SetSlienceConfig(req *request.SetAlertSlienceConfigRequest) error {
	if !s.enableInnerReceiver {
		return errors.New("inner alert is not open")
	}
	return s.receivers.SetSlienceConfig(req.AlertID, req.ForDuration)
}

func (s *service) RemoveSlienceConfig(alertID string) error {
	if !s.enableInnerReceiver {
		return errors.New("inner alert is not open")
	}
	return s.receivers.RemoveSlienceConfig(alertID)
}
