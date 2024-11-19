package alertanalyze

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	ck "github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

const (
	alertEvent = "alert"
	errorEvent = "error"

	mutationEventPrefix = "mutation-"
)

// SearchAnormalDeltaByEntry 基于入口查询异常事件
func (s *service) SearchAnormalDeltaByEntry(req *request.GetDescendantAnormalDeltaEventRequest) (*response.GetDescendantDeltaAnormalEventResponse, error) {
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)

	descendants, err := s.chRepo.ListDescendantNodes(&request.GetDescendantMetricsRequest{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Service:   req.Service,
		Endpoint:  req.Endpoint,
	})
	if err != nil {
		return nil, err
	}

	descendants = append(descendants, ck.TopologyNode{
		Service:  req.Service,
		Endpoint: req.Endpoint,
	})

	var selectedEvents []string
	if len(req.AnormalTypes) > 0 {
		selectedEvents = strings.Split(req.AnormalTypes, ",")
	}

	// 便于后续查询受告警影响的服务
	var instances []*model.ServiceInstance
	var endpoints []model.EndpointKey
	instanceMap := newInstanceMap()
	// TODO 优化,减少查询次数

	for _, descendant := range descendants {
		// 获取每个endpoint下的所有实例
		instanceList, err := s.promRepo.GetActiveRequestInstanceList(req.StartTime, req.EndTime, descendant.Service, descendant.Endpoint)
		if err != nil {
			continue
		}

		// 构建好子孙节点的Node/Service -> descendant 映射
		endpoint := model.EndpointKey{
			ServiceName: descendant.Service,
			Endpoint:    descendant.Endpoint,
		}
		instancesForDescendant := instanceList.GetInstances()
		instanceMap.AddInstances(endpoint, instancesForDescendant)

		instances = append(instances, instancesForDescendant...)
		endpoints = append(endpoints, endpoint)
	}

	// 返回结果列表
	var anormalEventList []model.AnormalEvent

	// 获取匹配的error
	if len(selectedEvents) == 0 || isSelect(selectedEvents, errorEvent) {
		propagations, err := s.chRepo.ListErrorByEntryService(req.StartTime, req.EndTime, req.Service, req.Endpoint, endpoints)
		if err == nil {
			errorEvents := s.parseErrorEvent(propagations, instanceMap, req.Step)
			// 存放error事件
			anormalEventList = append(anormalEventList, errorEvents...)
		} else {
			return nil, err
		}
	}

	if len(selectedEvents) == 0 || isSelect(selectedEvents, alertEvent) {
		// 获取匹配的alertEvents
		alertEvents, err := s.chRepo.GetAlertEventsWithKeyByInstanceAndEndpoints(
			startTime, endTime,
			// 同时获取Firing和Resolved故障
			request.AlertFilter{WithMutation: true},
			instances, endpoints,
		)

		if err == nil {
			anormalAlerts := s.parseAlertEvents(alertEvents, instanceMap, selectedEvents, req.StartTime)
			// 存放告警事件
			anormalEventList = append(anormalEventList, anormalAlerts...)
		} else {
			return nil, err
		}
	}

	originAnormalCounts := map[model.EndpointKey]map[model.AnormalType]int64{}
	alertTriggeredCounts := map[model.EndpointKey]map[model.AnormalType]int64{}
	finalAnormalCounts := map[model.EndpointKey]map[model.AnormalType]int64{}

	originAnormalEvents := []response.DescendantAnormalEventRecord{}
	deltaAnormalEvents := []response.DescendantAnormalEventRecord{}
	finalAnrmalEvents := []response.DescendantAnormalEventRecord{}

	// anormalEventTimeGroup := make(map[int64][]model.AnormalEvent)
	for _, event := range anormalEventList {
		// status Before Firing
		statusBefore := model.StatusResolved
		lastEventTSBefore := int64(-1)
		lastEventTSAfter := int64(-1)
		isFiringDurning := model.StatusResolved
		statusAfter := model.StatusResolved
		deltaEventTS := []model.AnormalUpdateTS{}

		// 填充Chart
		for _, updateTS := range event.UpdateTSs {
			// 统计用户时间片前的状态
			if updateTS.Timestamp < req.DeltaStartTime {
				statusBefore = updateTS.AnormalStatus
				lastEventTSBefore = updateTS.Timestamp
			}
			if updateTS.Timestamp >= req.DeltaStartTime &&
				updateTS.Timestamp <= req.DeltaEndTime &&
				updateTS.AnormalStatus == model.StatusFiring {
				isFiringDurning = model.StatusFiring
			}
			if updateTS.Timestamp < req.DeltaEndTime {
				statusAfter = updateTS.AnormalStatus
				lastEventTSAfter = updateTS.Timestamp
				if updateTS.Timestamp > req.DeltaStartTime {
					deltaEventTS = append(deltaEventTS, updateTS)
				}
			}
		}

		if statusBefore == model.StatusFiring {
			for _, impactEndpoints := range event.ImpactEndpoints {
				alertCounts, find := originAnormalCounts[impactEndpoints.EndpointKey]
				if !find {
					alertCounts = map[model.AnormalType]int64{}
					originAnormalCounts[impactEndpoints.EndpointKey] = alertCounts
				}
				alertCounts[event.AnormalType]++
			}

			// 记录已经发生的告警事件和当时的状态
			for _, impactEndpoint := range event.ImpactEndpoints {
				originAnormalEvents = append(originAnormalEvents, response.DescendantAnormalEventRecord{
					EndpointKey:   impactEndpoint.EndpointKey,
					AlertKey:      impactEndpoint.GetEventKey(),
					AnormalType:   event.AnormalType,
					Timestamp:     lastEventTSBefore,
					AnormalObject: impactEndpoint.AlertObject,
					AnormalReason: impactEndpoint.AlertReason,
					AnormalMsg:    impactEndpoint.AlertMessage[lastEventTSBefore],
					AnormalStatus: "updatedFiring",
				})
			}
		}

		if statusBefore == model.StatusFiring || isFiringDurning == model.StatusFiring {
			for _, impactEndpoints := range event.ImpactEndpoints {
				alertCounts, find := alertTriggeredCounts[impactEndpoints.EndpointKey]
				if !find {
					alertCounts = map[model.AnormalType]int64{}
					alertTriggeredCounts[impactEndpoints.EndpointKey] = alertCounts
				}
				alertCounts[event.AnormalType]++
			}
		}

		if statusAfter == model.StatusFiring {
			for _, impactEndpoints := range event.ImpactEndpoints {
				alertCounts, find := finalAnormalCounts[impactEndpoints.EndpointKey]
				if !find {
					alertCounts = map[model.AnormalType]int64{}
					finalAnormalCounts[impactEndpoints.EndpointKey] = alertCounts
				}
				alertCounts[event.AnormalType]++
			}

			// 记录已经发生的告警事件和当时的状态
			for _, impactEndpoint := range event.ImpactEndpoints {
				finalAnrmalEvents = append(finalAnrmalEvents, response.DescendantAnormalEventRecord{
					EndpointKey:   impactEndpoint.EndpointKey,
					AlertKey:      impactEndpoint.GetEventKey(),
					AnormalType:   event.AnormalType,
					Timestamp:     lastEventTSAfter,
					AnormalObject: impactEndpoint.AlertObject,
					AnormalReason: impactEndpoint.AlertReason,
					AnormalMsg:    impactEndpoint.AlertMessage[lastEventTSAfter],
					AnormalStatus: "updatedFiring",
				})
			}
		}

		if len(deltaEventTS) > 0 {
			var hasStarted bool = false
			if statusBefore == model.StatusFiring {
				hasStarted = true
			}
			// 记录发生变化的事件 (create / update / resolved)
			for _, ts := range deltaEventTS {
				var anormalStatus string = "updatedFiring"
				if hasStarted && ts.AnormalStatus == model.StatusFiring {
					anormalStatus = "updatedFiring"
				} else if !hasStarted && ts.AnormalStatus == model.StatusFiring {
					anormalStatus = "startFiring"
					hasStarted = true
				} else if ts.AnormalStatus == model.StatusResolved {
					anormalStatus = "resolved"
					hasStarted = false
				}

				for _, impactEndpoint := range event.ImpactEndpoints {
					deltaAnormalEvents = append(deltaAnormalEvents, response.DescendantAnormalEventRecord{
						EndpointKey:   impactEndpoint.EndpointKey,
						AlertKey:      impactEndpoint.GetEventKey(),
						AnormalType:   event.AnormalType,
						Timestamp:     ts.Timestamp,
						AnormalObject: impactEndpoint.AlertObject,
						AnormalReason: impactEndpoint.AlertReason,
						AnormalMsg:    impactEndpoint.AlertMessage[ts.Timestamp],
						AnormalStatus: anormalStatus,
					})
				}
			}
		}
	}

	var originAnormalCountsList = []response.DescendantAnormalCounts{}
	for endpoint, anormalCounts := range originAnormalCounts {
		originAnormalCountsList = append(originAnormalCountsList, response.DescendantAnormalCounts{
			EndpointKey:      endpoint,
			AnormalCountsMap: anormalCounts,
		})
	}

	var alertTriggeredCountsList = []response.DescendantAnormalCounts{}
	for endpoint, anormalCounts := range alertTriggeredCounts {
		alertTriggeredCountsList = append(alertTriggeredCountsList, response.DescendantAnormalCounts{
			EndpointKey:      endpoint,
			AnormalCountsMap: anormalCounts,
		})
	}

	var finalAnormalCountsList = []response.DescendantAnormalCounts{}
	for endpoint, anormalCounts := range finalAnormalCounts {
		finalAnormalCountsList = append(finalAnormalCountsList, response.DescendantAnormalCounts{
			EndpointKey:      endpoint,
			AnormalCountsMap: anormalCounts,
		})
	}

	return &response.GetDescendantDeltaAnormalEventResponse{
		OriginAnormalCounts:  originAnormalCountsList,
		AlertTriggeredCounts: alertTriggeredCountsList,
		FinalAnormalCounts:   finalAnormalCountsList,

		OriginAnormalEvents: originAnormalEvents,
		DeltaAnormalEvents:  deltaAnormalEvents,
		FinalAnormalEvents:  finalAnrmalEvents,
	}, nil
}

func (*service) parseErrorEvent(propagations []ck.ErrorPropation, instanceMap *instanceMap, step int64) []model.AnormalEvent {
	var anormalEventList []model.AnormalEvent
	for _, propagation := range propagations {
		errorEvent := model.AnormalEvent{
			UpdateTSs: []model.AnormalUpdateTS{
				{
					AnormalStatus: model.StatusFiring,
					Timestamp:     propagation.Timestamp.UnixMicro(),
				},
				{
					// 自定义一个一分钟结束的状态
					AnormalStatus: model.StatusResolved,
					Timestamp:     propagation.Timestamp.Add(time.Duration(step * 1000)).UnixMicro(),
				},
			},
			AnormalType:     model.AnormalTypeError,
			ImpactEndpoints: []model.AnormalEventDetail{},
		}
		for idx, service := range propagation.NodesService {
			if !propagation.NodesIsError[idx] {
				continue
			}
			// 检查数据存在,防止数组越界
			if len(propagation.NodesUrl) <= idx ||
				len(propagation.NodesInstance) <= idx ||
				len(propagation.NodesErrorTypes) <= idx ||
				len(propagation.NodesErrorMsgs) <= idx {
				break
			}
			endpointKey := model.EndpointKey{
				ServiceName: service,
				Endpoint:    propagation.NodesUrl[idx],
			}
			// 检查ContentKey是否在endpoints中
			if !instanceMap.IsEndpointKeyExist(endpointKey) {
				continue
			}
			errorEvent.ImpactEndpoints = append(errorEvent.ImpactEndpoints, model.AnormalEventDetail{
				EndpointKey: endpointKey,
				AlertKey:    model.GenUUID().String(),
				AlertObject: propagation.NodesInstance[idx],
				AlertReason: strings.Join(propagation.NodesErrorTypes[idx], ";"),
				AlertMessage: map[int64]string{
					propagation.Timestamp.UnixMicro(): strings.Join(propagation.NodesErrorMsgs[idx], ";"),
				},
			})
		}
		anormalEventList = append(anormalEventList, errorEvent)
	}
	return anormalEventList
}

func (*service) parseAlertEvents(alertEvents []ck.AlertEventWithKey, instanceMap *instanceMap, selectedEvents []string, searchStartTime int64) []model.AnormalEvent {
	var anormalEventList []model.AnormalEvent

	for i := 0; i < len(alertEvents); i++ {
		alertEvent := alertEvents[i]
		if len(selectedEvents) > 0 &&
			!isSelect(selectedEvents, alertEvent.Group) {
			// 跳过未选择的事件
			continue
		}
		if alertEvent.Status == model.StatusResolved {
			// 跳过第一个状态就是Resolved的事件
			continue
		}

		// 初始化每个告警的首个事件
		var updateTSs []model.AnormalUpdateTS
		var alertMessage = map[int64]string{}
		if alertEvent.CreateTime.UnixMicro() < searchStartTime {
			updateTSs = append(updateTSs, model.AnormalUpdateTS{
				Timestamp:     alertEvent.CreateTime.UnixMicro(),
				AnormalStatus: model.StatusFiring,
			})
			alertMessage = map[int64]string{
				alertEvent.CreateTime.UnixMicro(): "在查询时间开始前已告警,初次告警在" + alertEvent.CreateTime.Format("2006-01-02 15:04:05"),
			}
		}
		updateTSs = append(updateTSs, model.AnormalUpdateTS{
			Timestamp:     alertEvent.ReceivedTime.UnixMicro(),
			AnormalStatus: alertEvent.Status,
		})
		alertMessage[alertEvent.ReceivedTime.UnixMicro()] = alertEvent.Detail
		var anormalEvent model.AnormalEvent = model.AnormalEvent{
			UpdateTSs:       updateTSs,
			ImpactEndpoints: []model.AnormalEventDetail{},
		}

		// 完善告警信息
		switch ck.AlertGroup(alertEvent.Group) {
		case ck.APP_GROUP, ck.MUTATION_APP_GROUP:
			anormalEvent.AnormalType = model.AnormalTypeAlertApp
			anormalEvent.ImpactEndpoints = append(anormalEvent.ImpactEndpoints, model.AnormalEventDetail{
				EndpointKey: model.EndpointKey{
					ServiceName: alertEvent.GetServiceNameTag(),
					Endpoint:    alertEvent.GetContentKeyTag(),
				},
				AlertKey:     alertEvent.AlertKey,
				AlertObject:  alertEvent.GetTargetObj(),
				AlertReason:  alertEvent.Name,
				AlertMessage: alertMessage,
			})
		case ck.CONTAINER_GROUP, ck.MUTATION_CONTAINER_GROUP:
			anormalEvent.AnormalType = model.AnormalTypeAlertContainer
			instance, endpoints := instanceMap.GetEndpointsByK8sPodNS(alertEvent.GetK8sPodTag(), alertEvent.GetK8sNamespaceTag())
			if instance == nil {
				continue
			}
			for _, endpoint := range endpoints {
				anormalEvent.ImpactEndpoints = append(anormalEvent.ImpactEndpoints, model.AnormalEventDetail{
					EndpointKey:  endpoint,
					AlertKey:     alertEvent.AlertKey,
					AlertObject:  alertEvent.GetTargetObj(),
					AlertReason:  alertEvent.Name,
					AlertMessage: alertMessage,
				})
			}
		case ck.NETWORK_GROUP, ck.MUTATION_NETWORK_GROUP:
			anormalEvent.AnormalType = model.AnormalTypeAlertNet
			var endpoints []model.EndpointKey
			var instance *model.ServiceInstance
			pod := alertEvent.GetK8sPodTag()
			if len(pod) > 0 {
				instance, endpoints = instanceMap.GetEndpointsByK8sPodNS(alertEvent.GetK8sPodTag(), alertEvent.GetK8sNamespaceTag())
				if instance == nil {
					continue
				}
			} else {
				instance, endpoints = instanceMap.GetEndpointsByNodePid(alertEvent.GetNetSrcNodeTag(), alertEvent.GetNetSrcPidTag())
				if instance == nil {
					continue
				}
			}
			for _, endpoint := range endpoints {
				anormalEvent.ImpactEndpoints = append(anormalEvent.ImpactEndpoints, model.AnormalEventDetail{
					EndpointKey:  endpoint,
					AlertKey:     alertEvent.AlertKey,
					AlertObject:  alertEvent.GetTargetObj(),
					AlertReason:  alertEvent.Name,
					AlertMessage: alertMessage,
				})
			}
		case ck.INFRA_GROUP, ck.MUTATION_INFRA_GROUP:
			anormalEvent.AnormalType = model.AnormalTypeAlertInfra
			endpointsMaps := instanceMap.GetEndpointsByNode(alertEvent.GetInfraNodeTag())
			for instance, endpoints := range endpointsMaps {
				var instanceName string
				if len(instance.PodName) > 0 {
					instanceName = fmt.Sprintf("%s/%s", instance.Namespace, instance.PodName)
				} else {
					instanceName = fmt.Sprintf("(pid:%d)", instance.Pid)
				}
				for _, endpoint := range endpoints {
					anormalEvent.ImpactEndpoints = append(anormalEvent.ImpactEndpoints, model.AnormalEventDetail{
						EndpointKey:  endpoint,
						AlertKey:     alertEvent.AlertKey,
						AlertObject:  instanceName + " at " + alertEvent.GetTargetObj(),
						AlertReason:  alertEvent.Name,
						AlertMessage: alertMessage,
					})
				}
			}
		}

		var tryMergeIndex = i + 1
		var lastStatus model.Status = model.StatusFiring
		for tryMergeIndex < len(alertEvents) {
			nextEvent := alertEvents[tryMergeIndex]
			if nextEvent.AlertKey != alertEvent.AlertKey {
				break
			}
			// 跳过重复的resolved事件
			if nextEvent.Status != model.StatusResolved || lastStatus != model.StatusResolved {
				anormalEvent.UpdateTSs = append(anormalEvent.UpdateTSs, model.AnormalUpdateTS{
					Timestamp:     nextEvent.ReceivedTime.UnixMicro(),
					AnormalStatus: nextEvent.Status,
				})
				for _, endpoint := range anormalEvent.ImpactEndpoints {
					endpoint.AlertMessage[nextEvent.ReceivedTime.UnixMicro()] = nextEvent.Detail
				}
				lastStatus = nextEvent.Status
			}
			// 后续处理时跳过已经完成处理的事件
			i = tryMergeIndex
			// 继续查询处理下一个事件
			tryMergeIndex++
		}
		anormalEventList = append(anormalEventList, anormalEvent)
	}
	return anormalEventList
}

type instanceMap struct {
	Pod2InstanceMap     map[K8sPodNSKey]model.ServiceInstance
	NodePid2InstanceMap map[NodePidKey]model.ServiceInstance
	Node2InstancesMap   map[string]map[model.ServiceInstance]struct{}

	InstanceMap map[model.ServiceInstance]map[model.EndpointKey]struct{}

	EndpointMap map[model.EndpointKey]struct{}
}

type NodePidKey struct {
	Node string
	Pid  int
}

type K8sPodNSKey struct {
	Namespace string
	Pod       string
}

func newInstanceMap() *instanceMap {
	return &instanceMap{
		Pod2InstanceMap:     map[K8sPodNSKey]model.ServiceInstance{},
		NodePid2InstanceMap: map[NodePidKey]model.ServiceInstance{},
		Node2InstancesMap:   map[string]map[model.ServiceInstance]struct{}{},
		InstanceMap:         map[model.ServiceInstance]map[model.EndpointKey]struct{}{},
		EndpointMap:         map[model.EndpointKey]struct{}{},
	}
}

func (m *instanceMap) AddInstances(endpointKey model.EndpointKey, instances []*model.ServiceInstance) {
	m.EndpointMap[endpointKey] = struct{}{}

	for _, instance := range instances {
		endpointKeys, find := m.InstanceMap[*instance]
		if !find {
			endpointKeys = make(map[model.EndpointKey]struct{})
		}
		endpointKeys[endpointKey] = struct{}{}
		m.InstanceMap[*instance] = endpointKeys

		if len(instance.PodName) > 0 {
			m.Pod2InstanceMap[K8sPodNSKey{instance.Namespace, instance.PodName}] = *instance
		}

		if instance.Pid > 0 {
			m.NodePid2InstanceMap[NodePidKey{instance.NodeName, int(instance.Pid)}] = *instance
		}

		if len(instance.NodeName) > 0 {
			instancesOnNode, find := m.Node2InstancesMap[instance.NodeName]
			if !find {
				instancesOnNode = make(map[model.ServiceInstance]struct{})
			}
			instancesOnNode[*instance] = struct{}{}
			m.Node2InstancesMap[instance.NodeName] = instancesOnNode
		}
	}
}

func (m *instanceMap) GetEndpointsByK8sPodNS(pod, namespace string) (*model.ServiceInstance, []model.EndpointKey) {
	instance, find := m.Pod2InstanceMap[K8sPodNSKey{namespace, pod}]
	if !find {
		return nil, nil
	}

	endpointsMap, find := m.InstanceMap[instance]
	if !find {
		return nil, nil
	}
	var endpoints []model.EndpointKey
	for endpoint := range endpointsMap {
		endpoints = append(endpoints, endpoint)
	}
	return &instance, endpoints
}

func (m *instanceMap) GetEndpointsByNodePid(node string, pid string) (*model.ServiceInstance, []model.EndpointKey) {
	if len(pid) == 0 {

	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return nil, nil
	}

	instance, find := m.NodePid2InstanceMap[NodePidKey{node, pidInt}]
	if !find {
		return nil, nil
	}

	endpointsMap, find := m.InstanceMap[instance]
	if !find {
		return nil, nil
	}
	var endpoints []model.EndpointKey
	for endpoint := range endpointsMap {
		endpoints = append(endpoints, endpoint)
	}
	return &instance, endpoints
}

func (m *instanceMap) GetEndpointsByNode(node string) map[model.ServiceInstance][]model.EndpointKey {
	instances, find := m.Node2InstancesMap[node]
	if !find || len(instances) == 0 {
		return nil
	}

	var res = make(map[model.ServiceInstance][]model.EndpointKey)
	for instance := range instances {
		endpointsMap, find := m.InstanceMap[instance]
		if !find {
			continue
		}
		var endpoints []model.EndpointKey
		for endpoint := range endpointsMap {
			endpoints = append(endpoints, endpoint)
		}

		res[instance] = endpoints
	}

	return res
}

func (m *instanceMap) IsEndpointKeyExist(endpointKey model.EndpointKey) bool {
	_, find := m.EndpointMap[endpointKey]
	return find
}

func isSelect(selectedEvents []string, eventGroup string) bool {
	// 是否叠加alert告警
	if eventGroup == alertEvent {
		for _, eventType := range selectedEvents {
			if eventType == string(ck.APP_GROUP) ||
				eventType == string(ck.CONTAINER_GROUP) ||
				eventType == string(ck.INFRA_GROUP) ||
				eventType == string(ck.NETWORK_GROUP) {
				return true
			}
		}
		return false
	}

	// 是否叠加alert-xxx类型告警事件
	// app: app/mutation-app
	// container: container/mutation-container
	// infra: infra/mutation-infra
	// network: network/mutation-network
	// error: error
	for _, eventType := range selectedEvents {
		mutationType := mutationEventPrefix + eventType
		if eventType == eventGroup ||
			mutationType == eventGroup {
			return true
		}
	}
	return false
}
