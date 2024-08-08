package serviceoverview

import (
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) SetThreshold(level string, serviceName string, endPoint string, latency float64, errorRate float64, tps float64, log float64) (res response.SetThresholdResponse, err error) {
	threshold := &database.Threshold{
		Latency:   latency,
		Tps:       tps,
		ErrorRate: errorRate,
		Log:       log,
	}
	if level == database.GLOBAL {
		threshold.Level = database.GLOBAL
	}
	err = s.dbRepo.CreateOrUpdateThreshold(threshold)
	return res, err

}
