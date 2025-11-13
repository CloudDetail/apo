package alertanalyze

import (
	"errors"
	"strconv"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

type EndpointsMap struct {
	EndpointSet map[clickhouse.AlertService]struct{}

	// Node
	InferNodeSet map[string][]clickhouse.AlertService
	// Pod Namespace
	PodSet map[podKey][]clickhouse.AlertService
	// Node + Pid
	NetSrcSet map[netSrcProcessKey][]clickhouse.AlertService
	// Database
	DatabaseSet map[databaseKey][]clickhouse.AlertService

	EntryRelationsSet map[clickhouse.AlertService][]clickhouse.EntryRelationship
}

func newEndpointsMap() *EndpointsMap {
	return &EndpointsMap{
		EndpointSet:  map[clickhouse.AlertService]struct{}{},
		InferNodeSet: map[string][]clickhouse.AlertService{},
		PodSet:       map[podKey][]clickhouse.AlertService{},
		NetSrcSet:    map[netSrcProcessKey][]clickhouse.AlertService{},
		DatabaseSet:  map[databaseKey][]clickhouse.AlertService{},
	}
}

type podKey struct {
	Namespace string
	PodName   string
}

type netSrcProcessKey struct {
	Node string
	Pid  string
}

type databaseKey struct {
	URL  string
	IP   string
	Port string
}

func (m *EndpointsMap) SearchAndAddEndpointsForAlert(ctx core.Context, chRepo clickhouse.Repo, repo prometheus.Repo, event *alert.Alert, startTime, endTime time.Time) ([]clickhouse.AlertService, bool) {
	// 尝试下面的方法去获取event的关联入口
	endpoints, err := tryGetAlertService(ctx, repo, m, event, startTime, endTime)

	if len(endpoints) == 0 {
		var pid = 0
		pidStr := event.GetPidTag()
		if v, e1 := strconv.Atoi(pidStr); e1 == nil {
			pid = v
		}
		endpoints, _ = chRepo.GetDataplaneServiceByInstance(ctx,
			startTime.UnixMicro(),
			endTime.UnixMicro(),
			"",
			event.GetNetSrcPodTag(),
			event.GetK8sNamespaceTag(),
			event.GetInfraNodeTag(),
			pid,
		)
	}

	if err != nil && len(endpoints) == 0 {
		var vErr alert.ErrAlertImpactNoMatchedService
		if errors.As(err, &vErr) {
			// 分析过,但找不到
			return []clickhouse.AlertService{}, true
		}
		// 无法分析
		return []clickhouse.AlertService{}, false
	}

	if len(endpoints) == 0 {
		return []clickhouse.AlertService{}, true
	}
	return endpoints, true
}

func (m *EndpointsMap) EndpointsList() []clickhouse.AlertService {
	var endpointList []clickhouse.AlertService
	for endpoint := range m.EndpointSet {
		endpointList = append(endpointList, endpoint)
	}
	return endpointList
}

func tryGetAlertService(ctx core.Context, repo prometheus.Repo, endpointsMap *EndpointsMap, event *alert.Alert, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {

	var tryMethods = []func(core.Context, prometheus.Repo, *EndpointsMap, *alert.Alert, time.Time, time.Time) ([]clickhouse.AlertService, error){
		tryGetAlertServiceByService,
		tryGetAlertServiceByDB,
		tryGetAlertServiceByK8sPod,
		tryGetAlertServiceByVMProcess,
		tryGetAlertServiceByInfraNode,
	}
	var endpoints []clickhouse.AlertService
	checkedError := alert.ErrAlertImpactMissingTag{
		TagGroups: []alert.TagGroup{},
		Event:     event,
	}
	for _, tryGetService := range tryMethods {
		var err error
		endpoints, err = tryGetService(ctx, repo, endpointsMap, event, startTime, endTime)
		if err == nil {
			return endpoints, nil
		}
		// 如果是Tag不足,继续尝试别的Tag
		var vErr alert.ErrAlertImpactMissingTag
		if errors.As(err, &vErr) {
			checkedError.AddCheckedGroup(vErr)
			continue
		}
		var vErr2 alert.ErrAlertImpactNotFit
		if errors.As(err, &vErr2) {
			// 跳过不适合的搜寻方式
			continue
		}
		// 其他错误,直接返回
		return nil, err
	}

	return endpoints, checkedError
}

func tryGetAlertServiceByService(ctx core.Context, _ prometheus.Repo, endpointsMap *EndpointsMap, event *alert.Alert, _ time.Time, _ time.Time) ([]clickhouse.AlertService, error) {
	serviceName := event.GetServiceNameTag()
	if len(serviceName) == 0 {
		return nil, alert.ErrAlertImpactMissingTag{
			TagGroups: []alert.TagGroup{[]string{"serviceName"}},
			Event:     event,
		}
	}

	alertServices := []clickhouse.AlertService{
		{
			Service:  serviceName,
			Endpoint: event.GetEndpointTag(),
		},
	}

	if endpointsMap != nil {
		for _, alertService := range alertServices {
			endpointsMap.EndpointSet[alertService] = struct{}{}
		}
	}

	return alertServices, nil
}

func tryGetAlertServiceByK8sPod(ctx core.Context, repo prometheus.Repo, endpointsMap *EndpointsMap, event *alert.Alert, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	podName := event.GetNetSrcPodTag()
	namespace := event.GetK8sNamespaceTag()
	if len(podName) == 0 || len(namespace) == 0 {
		return nil, alert.ErrAlertImpactMissingTag{
			TagGroups: []alert.TagGroup{[]string{"pod", "namespace"}},
			Event:     event,
		}
	}

	podKey := podKey{
		Namespace: namespace,
		PodName:   podName,
	}
	if endpoints, find := endpointsMap.PodSet[podKey]; find {
		return endpoints, nil
	}

	// 通常也只会有一个Service
	services, err := repo.GetServiceListByFilter(
		ctx,
		startTime, endTime,
		prometheus.NamespacePQLFilter, namespace,
		prometheus.PodPQLFilter, podName,
	)
	if err != nil {
		return nil, err
	}
	var endpoints []clickhouse.AlertService
	// 通常只有一个service
	for _, service := range services {
		// 不关系ContentKey
		endpoints = append(endpoints, clickhouse.AlertService{
			Service: service,
		})
	}

	if endpointsMap != nil {
		endpointsMap.PodSet[podKey] = endpoints
		for _, endpoint := range endpoints {
			endpointsMap.EndpointSet[endpoint] = struct{}{}
		}
	}

	if len(endpoints) == 0 {
		return nil, alert.ErrAlertImpactNoMatchedService{
			TagGroup:  []string{"pod", "namespace"},
			TagValues: []string{podName, namespace},
		}
	}

	return endpoints, nil
}

func tryGetAlertServiceByVMProcess(ctx core.Context, repo prometheus.Repo, endpointsMap *EndpointsMap, event *alert.Alert, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	nodeName := event.GetNetSrcNodeTag()
	pid := event.GetNetSrcPidTag()
	if len(nodeName) == 0 || len(pid) == 0 {
		return nil, alert.ErrAlertImpactMissingTag{
			TagGroups: []alert.TagGroup{[]string{"node", "pid"}},
			Event:     event,
		}
	}

	processKey := netSrcProcessKey{
		Node: nodeName,
		Pid:  pid,
	}
	if endpointsMap != nil {
		if endpoints, find := endpointsMap.NetSrcSet[processKey]; find {
			return endpoints, nil
		}
	}

	services, err := repo.GetServiceListByFilter(
		ctx,
		startTime, endTime,
		prometheus.NodeNamePQLFilter, nodeName,
		prometheus.PidPQLFilter, pid,
	)

	if err != nil {
		return nil, err
	}

	var endpoints []clickhouse.AlertService
	for _, service := range services {
		endpoints = append(endpoints, clickhouse.AlertService{
			Service: service,
		})
	}

	if endpointsMap != nil {
		endpointsMap.NetSrcSet[processKey] = endpoints
		for _, endpoint := range endpoints {
			endpointsMap.EndpointSet[endpoint] = struct{}{}
		}
	}

	if len(endpoints) == 0 {
		return nil, alert.ErrAlertImpactNoMatchedService{
			TagGroup:  []string{"node_name", "pid"},
			TagValues: []string{nodeName, pid},
		}
	}

	return endpoints, nil
}

func tryGetAlertServiceByInfraNode(ctx core.Context, repo prometheus.Repo, endpointsMap *EndpointsMap, event *alert.Alert, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	if event.Group != string(clickhouse.INFRA_GROUP) {
		return nil, alert.ErrAlertImpactNotFit{
			TagGroup: []string{"node"},
			Group:    event.Group,
		}
	}

	nodeName := event.GetInfraNodeTag()
	if len(nodeName) == 0 {
		return nil, alert.ErrAlertImpactMissingTag{
			TagGroups: []alert.TagGroup{[]string{"node"}},
			Event:     event,
		}
	}

	if endpointsMap != nil {
		if endpoints, find := endpointsMap.InferNodeSet[nodeName]; find {
			return endpoints, nil
		}
	}

	services, err := repo.GetServiceListByFilter(
		ctx,
		startTime, endTime,
		prometheus.NodeNamePQLFilter, nodeName,
	)

	if err != nil {
		return nil, err
	}

	var endpoints []clickhouse.AlertService
	for _, service := range services {
		endpoints = append(endpoints, clickhouse.AlertService{
			Service: service,
		})
	}

	if endpointsMap != nil {
		endpointsMap.InferNodeSet[nodeName] = endpoints
		for _, endpoint := range endpoints {
			endpointsMap.EndpointSet[endpoint] = struct{}{}
		}
	}

	if len(endpoints) == 0 {
		return nil, alert.ErrAlertImpactNoMatchedService{
			TagGroup:  []string{"instance_name"},
			TagValues: []string{nodeName},
		}
	}

	return endpoints, nil
}

func tryGetAlertServiceByDB(ctx core.Context, repo prometheus.Repo, endpointsMap *EndpointsMap, event *alert.Alert, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	// 尝试获取数据库URL
	dbURL := event.GetDatabaseURL()
	dbIP := event.GetDatabaseIP()
	dbPort := event.GetDatabasePort()
	if len(dbURL) == 0 && len(dbIP) == 0 {
		return nil, alert.ErrAlertImpactMissingTag{
			TagGroups: []alert.TagGroup{
				[]string{"db_url"}, []string{"db_ip", "db_port"},
			},
			Event: event,
		}
	}
	dbKey := databaseKey{
		URL:  dbURL,
		IP:   dbIP,
		Port: dbPort,
	}
	if endpoints, find := endpointsMap.DatabaseSet[dbKey]; find {
		return endpoints, nil
	}

	// 查询受此数据库影响的服务
	services, err := repo.GetServiceListByDatabase(
		ctx,
		startTime, endTime, dbURL, dbIP, dbPort)

	if err != nil {
		return nil, err
	}
	var endpoints []clickhouse.AlertService
	endpoints = append(endpoints, clickhouse.AlertService{
		DatabaseURL:  dbURL,
		DatabaseIP:   dbIP,
		DatabasePort: dbPort,
	})
	for _, service := range services {
		endpoints = append(endpoints, clickhouse.AlertService{
			Service: service,
		})
	}
	if endpointsMap != nil {
		endpointsMap.DatabaseSet[dbKey] = endpoints
		for _, endpoint := range endpoints {
			endpointsMap.EndpointSet[endpoint] = struct{}{}
		}
	}

	if len(endpoints) == 0 {
		return nil, alert.ErrAlertImpactNoMatchedService{
			TagGroup:  []string{"db_url", "db_ip", "db_port"},
			TagValues: []string{dbURL, dbIP, dbPort},
		}
	}

	return endpoints, nil
}

func searchEntryByServices(relations []clickhouse.EntryRelationship, alertServices []clickhouse.AlertService) []model.EndpointKey {
	var entrySet = make(map[model.EndpointKey]struct{}, 0)
	for _, relation := range relations {
		for _, endpoint := range alertServices {
			if endpoint.Service != relation.Service {
				continue
			}

			if len(endpoint.Endpoint) > 0 && endpoint.Endpoint != relation.Endpoint {
				continue
			}

			var entryKey = model.EndpointKey{
				ServiceName: relation.GetEntryService(),
				Endpoint:    relation.GetEntryEndpoint(),
			}
			entrySet[entryKey] = struct{}{}

		}
	}

	var entryList = make([]model.EndpointKey, 0)
	for k, _ := range entrySet {
		entryList = append(entryList, k)
	}
	return entryList
}
