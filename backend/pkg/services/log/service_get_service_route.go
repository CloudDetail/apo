package log

import (
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceRoute(req *request.GetServiceRouteRequest) (*response.GetServiceRouteResponse, error) {
	now := time.Now()
	currentTimestamp := now.UnixMicro()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	sevenDaysAgoTimestamp := sevenDaysAgo.UnixMicro()

	instances, err := s.promRepo.GetActiveInstanceList(sevenDaysAgoTimestamp, currentTimestamp, req.Service)
	if err != nil {
		return nil, err
	}
	var deployName string
	for instanceName, _ := range instances.GetInstanceIdMap() {
		parts := strings.Split(instanceName, "-")
		if len(parts) >= 3 {
			deployName = strings.Join(parts[:len(parts)-2], "-")
			break
		}
	}

	return &response.GetServiceRouteResponse{
		RouteRule: map[string]string{"k8s.pod.name": deployName},
	}, nil

}
