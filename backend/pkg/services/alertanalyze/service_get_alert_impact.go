package alertanalyze

import (
	"errors"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

// AlertImpact 分析告警事件的影响面
// 1. 根据告警时间类型找到关联的Service,会检查Event中是否有满足要求的Label,如果没有会尝试所有预设的label组合
// 2. 通过ServiceTopology查询service的关联入口
func (s *service) AlertImpact(req *request.AlertImpactRequest) ([]clickhouse.EntryRelationship, []response.ImpactAlertEvent, map[model.EndpointKey]int, *model.Pagination, error) {
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)

	if req.PageParam == nil {
		// 按0,999分页
		req.PageParam = &request.PageParam{
			CurrentPage: 1,
			PageSize:    999,
		}
	}

	// 从Clickhouse中获取到所有的告警
	if req.EventID != "" {
		entrys, err := s.alertImpactTargetEvent(req.EventID, startTime, endTime)
		return entrys, nil, nil, nil, err
	}

	// eventId为空,获取所有的告警
	events, count, err := s.chRepo.GetAlertEvents(startTime, endTime, request.AlertFilter{
		Status: "firing", WithMutation: true,
	}, nil, nil, clickhouse.OrderAlertByReceivedTime)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if count == 0 {
		return []clickhouse.EntryRelationship{}, []response.ImpactAlertEvent{}, map[model.EndpointKey]int{}, &model.Pagination{
			Total:       0,
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
		}, nil
	}

	endpointsMap := EndpointsMap{
		EndpointSet:  map[clickhouse.AlertService]struct{}{},
		InferNodeSet: map[string][]clickhouse.AlertService{},
		PodSet:       map[podKey][]clickhouse.AlertService{},
		NetSrcSet:    map[netSrcProcessKey][]clickhouse.AlertService{},
		DatabaseSet:  map[databaseKey][]clickhouse.AlertService{},
	}

	var impactEvents = make([]response.ImpactAlertEvent, len(events))
	for i := 0; i < len(events); i++ {
		if req.AnalyzeSize == 0 && i > 50000 || req.AnalyzeSize != 0 && i > req.AnalyzeSize {
			impactEvents[i] = response.ImpactAlertEvent{
				AlertEvent:      &events[i].AlertEvent,
				IsAnalyzed:      false,
				ImpactEndpoints: []clickhouse.AlertService{},
			}
			continue
		}
		endpoints, associated := endpointsMap.SearchAndAddEndpointsForAlert(s.promRepo, &events[i].AlertEvent, startTime, endTime)
		// 记录关联状态
		impactEvents[i] = response.ImpactAlertEvent{
			AlertEvent:      &events[i].AlertEvent,
			IsAnalyzed:      associated,
			ImpactEndpoints: endpoints,
			ImpactEntries:   []model.EndpointKey{},
		}
	}

	// 所有告警涉及到的服务
	endpoints := endpointsMap.EndpointsList()
	// 查询入口和下游服务节点的关联
	entryRelation, err := s.chRepo.SearchEntryEndpointsByAlertService(endpoints, startTime.UnixMicro(), endTime.UnixMicro())
	if err != nil {
		pagation, eventList, _, _ := addRelatedEntryAndPagation(req, count, impactEvents, nil)
		return entryRelation, eventList, nil, pagation, err
	}

	// 关联告警并分页
	pagation, eventList, relatedEventCounts, err := addRelatedEntryAndPagation(req, count, impactEvents, entryRelation)
	return entryRelation, eventList, relatedEventCounts, pagation, err
}

// addRelatedEntryAndPagation 在告警事件中填充入口信息
// 同时每个入口关联的告警事件的数量
func addRelatedEntryAndPagation(req *request.AlertImpactRequest, count int, impactEvents []response.ImpactAlertEvent, entryRelation []clickhouse.EntryRelationship) (*model.Pagination, []response.ImpactAlertEvent, map[model.EndpointKey]int, error) {
	var pagation *model.Pagination = &model.Pagination{
		Total:       int64(count),
		CurrentPage: req.CurrentPage,
		PageSize:    req.PageSize,
	}
	var from, to int
	from = (req.PageParam.CurrentPage - 1) * req.PageParam.PageSize
	to = req.PageParam.CurrentPage * req.PageParam.PageSize
	if from > len(impactEvents) {
		return pagation, nil, nil, fmt.Errorf("page out of range")
	}
	if to > len(impactEvents) {
		to = len(impactEvents)
	}

	var relatedEventCount = make(map[model.EndpointKey]int)
	for i := 0; i < count; i++ {
		// 查询和事件关联的入口
		impactEntrys := searchEntryRelations(entryRelation, impactEvents[i].ImpactEndpoints)
		// 统计入口关联的告警数量
		for _, impactEntry := range impactEntrys {
			relatedEventCount[impactEntry]++
		}
		if from <= i && i < to {
			// 在当前页的告警事件中填充入口信息
			impactEvents[i].ImpactEntries = impactEntrys
		}
	}
	return pagation, impactEvents[from:to], relatedEventCount, nil
}

func (s *service) alertImpactTargetEvent(eventid string, startTime, endTime time.Time) ([]clickhouse.EntryRelationship, error) {
	event, err := s.chRepo.GetAlertEventById(eventid, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert eventId[%s]: %v", eventid, err)
	}

	// 根据告警事件类型查询影响入口
	var endpoints []clickhouse.AlertService
	switch clickhouse.AlertGroup(event.Group) {
	case clickhouse.APP_GROUP:
		endpoints, err = tryGetAlertServiceByService(s.promRepo, nil, event, startTime, endTime)
	case clickhouse.CONTAINER_GROUP:
		// 先尝试K8sPod匹配
		endpoints, err = tryGetAlertServiceByK8sPod(s.promRepo, nil, event, startTime, endTime)
	case clickhouse.NETWORK_GROUP:
		// 先尝试K8sPod匹配
		endpoints, err = tryGetAlertServiceByK8sPod(s.promRepo, nil, event, startTime, endTime)
		if err != nil && errors.As(err, &model.ErrAlertImpactMissingTag{}) {
			// 再尝试VMProcess匹配
			endpoints, err = tryGetAlertServiceByVMProcess(s.promRepo, nil, event, startTime, endTime)
		}
	case clickhouse.INFRA_GROUP:
		endpoints, err = tryGetAlertServiceByInfraNode(s.promRepo, nil, event, startTime, endTime)
	}

	if err != nil && errors.As(err, &model.ErrAlertImpactMissingTag{}) {
		// 预期的Label不存在,尝试所有预设的label组合
		endpoints, err = tryGetAlertService(s.promRepo, nil, event, startTime, endTime)
	}

	if err != nil {
		return nil, err
	}

	// 通过ServiceTopology关联查询入口
	return s.chRepo.SearchEntryEndpointsByAlertService(endpoints, startTime.UnixMicro(), endTime.UnixMicro())
}

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

func (m *EndpointsMap) SearchAndAddEndpointsForAlert(repo prometheus.Repo, event *model.AlertEvent, startTime, endTime time.Time) ([]clickhouse.AlertService, bool) {
	// 尝试下面的方法去获取event的关联入口
	endpoints, err := tryGetAlertService(repo, m, event, startTime, endTime)
	if err != nil && len(endpoints) == 0 {
		var vErr model.ErrAlertImpactNoMatchedService
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

// searchEntryRelations 查询每个alertServices关联的入口
func searchEntryRelations(relations []clickhouse.EntryRelationship, alertServices []clickhouse.AlertService) []model.EndpointKey {
	var tmpEntrySet = make(map[model.EndpointKey]struct{}, 0)
	var entryList = make([]model.EndpointKey, 0)
	for _, relation := range relations {
		for _, endpoint := range alertServices {
			if endpoint.Service != relation.Service {
				continue
			}
			// endpoint.ContentKey 为空表示所有关联业务匹配所有接口
			if len(endpoint.Endpoint) > 0 && endpoint.Endpoint != relation.Endpoint {
				continue
			}

			var entryKey = model.EndpointKey{
				ServiceName: relation.GetEntryService(),
				Endpoint:    relation.GetEntryEndpoint(),
			}

			if _, find := tmpEntrySet[entryKey]; find {
				continue
			}
			tmpEntrySet[entryKey] = struct{}{}
			entryList = append(entryList, entryKey)
		}
	}
	return entryList
}

func tryGetAlertService(repo prometheus.Repo, endpointsMap *EndpointsMap, event *model.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	var tryMethods = []func(prometheus.Repo, *EndpointsMap, *model.AlertEvent, time.Time, time.Time) ([]clickhouse.AlertService, error){
		tryGetAlertServiceByService,
		tryGetAlertServiceByDB,
		tryGetAlertServiceByK8sPod,
		tryGetAlertServiceByVMProcess,
		tryGetAlertServiceByInfraNode,
	}
	var endpoints []clickhouse.AlertService
	checkedError := model.ErrAlertImpactMissingTag{
		TagGroups: []model.TagGroup{},
		Event:     event,
	}
	for _, tryGetService := range tryMethods {
		var err error
		endpoints, err = tryGetService(repo, endpointsMap, event, startTime, endTime)
		if err == nil {
			return endpoints, nil
		}
		// 如果是Tag不足,继续尝试别的Tag
		var vErr model.ErrAlertImpactMissingTag
		if errors.As(err, &vErr) {
			checkedError.AddCheckedGroup(vErr)
			continue
		}
		var vErr2 model.ErrAlertImpactNotFit
		if errors.As(err, &vErr2) {
			// 跳过不适合的搜寻方式
			continue
		}
		// 其他错误,直接返回
		return nil, err
	}

	return endpoints, checkedError
}

func tryGetAlertServiceByService(_ prometheus.Repo, endpointsMap *EndpointsMap, event *model.AlertEvent, _ time.Time, _ time.Time) ([]clickhouse.AlertService, error) {
	serviceName := event.GetServiceNameTag()
	if len(serviceName) == 0 {
		return nil, model.ErrAlertImpactMissingTag{
			TagGroups: []model.TagGroup{[]string{"serviceName"}},
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

func tryGetAlertServiceByK8sPod(repo prometheus.Repo, endpointsMap *EndpointsMap, event *model.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	podName := event.GetNetSrcPodTag()
	namespace := event.GetK8sNamespaceTag()
	if len(podName) == 0 || len(namespace) == 0 {
		return nil, model.ErrAlertImpactMissingTag{
			TagGroups: []model.TagGroup{[]string{"pod", "namespace"}},
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
		return nil, model.ErrAlertImpactNoMatchedService{
			TagGroup:  []string{"pod", "namespace"},
			TagValues: []string{podName, namespace},
		}
	}

	return endpoints, nil
}

func tryGetAlertServiceByVMProcess(repo prometheus.Repo, endpointsMap *EndpointsMap, event *model.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	nodeName := event.GetNetSrcNodeTag()
	pid := event.GetNetSrcPidTag()
	if len(nodeName) == 0 || len(pid) == 0 {
		return nil, model.ErrAlertImpactMissingTag{
			TagGroups: []model.TagGroup{[]string{"node", "pid"}},
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
		return nil, model.ErrAlertImpactNoMatchedService{
			TagGroup:  []string{"node_name", "pid"},
			TagValues: []string{nodeName, pid},
		}
	}

	return endpoints, nil
}

func tryGetAlertServiceByInfraNode(repo prometheus.Repo, endpointsMap *EndpointsMap, event *model.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	if event.Group != string(clickhouse.INFRA_GROUP) {
		return nil, model.ErrAlertImpactNotFit{
			TagGroup: []string{"node"},
			Group:    event.Group,
		}
	}

	nodeName := event.GetInfraNodeTag()
	if len(nodeName) == 0 {
		return nil, model.ErrAlertImpactMissingTag{
			TagGroups: []model.TagGroup{[]string{"node"}},
			Event:     event,
		}
	}

	if endpointsMap != nil {
		if endpoints, find := endpointsMap.InferNodeSet[nodeName]; find {
			return endpoints, nil
		}
	}

	services, err := repo.GetServiceListByFilter(
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
		return nil, model.ErrAlertImpactNoMatchedService{
			TagGroup:  []string{"instance_name"},
			TagValues: []string{nodeName},
		}
	}

	return endpoints, nil
}

// func tryGetAlertServiceByContainer(repo prometheus.Repo, endpointsMap *EndpointsMap, event *model.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
// 	podName := event.GetK8sPodTag()
// 	namespace := event.GetK8sNamespaceTag()
// 	if len(podName) == 0 || len(namespace) == 0 {
// 		return nil, model.ErrAlertImpactMissingTag{
// 			TagGroups: []model.TagGroup{[]string{"pod/pod_name", "namespace"}},
// 			Event:     event,
// 		}
// 	}

// 	podKey := podKey{
// 		Namespace: namespace,
// 		PodName:   podName,
// 	}
// 	if endpoints, find := endpointsMap.PodSet[podKey]; find {
// 		return endpoints, nil
// 	}

// 	// 通常也只会有一个Service
// 	services, err := repo.GetServiceListByFilter(
// 		startTime, endTime,
// 		prometheus.NamespacePQLFilter, namespace,
// 		prometheus.PodPQLFilter, podName,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var endpoints []clickhouse.AlertService
// 	// 通常只有一个service
// 	for _, service := range services {
// 		// 不关系ContentKey
// 		endpoints = append(endpoints, clickhouse.AlertService{
// 			Service: service,
// 		})
// 	}

// 	if endpointsMap != nil {
// 		endpointsMap.PodSet[podKey] = endpoints
// 		for _, endpoint := range endpoints {
// 			endpointsMap.EndpointSet[endpoint] = struct{}{}
// 		}
// 	}

// 	if len(endpoints) == 0 {
// 		return nil, model.ErrAlertImpactNoMatchedService{
// 			TagGroup:  []string{"pod/pod_name", "namespace"},
// 			TagValues: []string{podName, namespace},
// 		}
// 	}

// 	return endpoints, nil
// }

func tryGetAlertServiceByDB(repo prometheus.Repo, endpointsMap *EndpointsMap, event *model.AlertEvent, startTime time.Time, endTime time.Time) ([]clickhouse.AlertService, error) {
	// 尝试获取数据库URL
	dbURL := event.GetDatabaseURL()
	dbIP := event.GetDatabaseIP()
	dbPort := event.GetDatabasePort()
	if len(dbURL) == 0 && len(dbIP) == 0 {
		return nil, model.ErrAlertImpactMissingTag{
			TagGroups: []model.TagGroup{
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
		return nil, model.ErrAlertImpactNoMatchedService{
			TagGroup:  []string{"db_url", "db_ip", "db_port"},
			TagValues: []string{dbURL, dbIP, dbPort},
		}
	}

	return endpoints, nil
}
