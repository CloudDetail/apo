package serviceoverview

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = (*service)(nil)

type Service interface {
	GetServiceMoreUrl(startTime time.Time, endTime time.Time, step time.Duration, serviceNames string, sortRule SortType) (res []response.ServiceDetail, err error)
	GetThreshold(level string, serviceName string, endPoint string) (res response.GetThresholdResponse, err error)
	SetThreshold(level string, serviceName string, endPoint string, latency float64, errorRate float64, tps float64, log float64) (res response.SetThresholdResponse, err error)
	GetServicesAlert(startTime time.Time, endTime time.Time, step time.Duration, serviceNames []string, returnData []string) (res []response.ServiceAlertRes, err error)
	GetServicesEndPointData(startTime time.Time, endTime time.Time, step time.Duration, filter EndpointsFilter, sortRule SortType) (res []response.ServiceEndPointsRes, err error)

	GetServicesRYGLightStatus(startTime time.Time, endTime time.Time, filter EndpointsFilter) (response.ServiceRYGLightRes, error)
	GetMonitorStatus(startTime time.Time, endTime time.Time) (response.GetMonitorStatusResponse, error)
}

type service struct {
	dbRepo   database.Repo
	promRepo prometheus.Repo
	chRepo   clickhouse.Repo
}

func New(chRepo clickhouse.Repo, dbRepo database.Repo, promRepo prometheus.Repo) Service {
	return &service{
		dbRepo:   dbRepo,
		promRepo: promRepo,
		chRepo:   chRepo,
	}
}

type EndpointsFilter struct {
	ContainsSvcName      string   // SvcName,包含匹配
	ContainsEndpointName string   // EndpointName,包含匹配
	Namespace            string   // Namespace, 完全匹配
	ServiceName          string   // 指定服务名, 完全匹配
	MultiService         []string // 多个服务名，完全匹配
	MultiNamespace       []string // 多个namespace，完全匹配
}
