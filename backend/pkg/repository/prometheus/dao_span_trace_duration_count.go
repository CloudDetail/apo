// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prometheus_model "github.com/prometheus/common/model"

	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	TEMPLATE_GET_SERVICES                = `sum by(svc_name) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))`
	TEMPLATE_GET_ENDPOINTS               = `sum by(content_key) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))`
	TEMPLATE_GET_SERVICE_INSTANCE        = `sum by(svc_name, pod, pid, container_id, node_name, namespace, node_ip) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))`
	TEMPLATE_GET_ACTIVE_SERVICE_INSTANCE = `sum by(svc_name, pod, pid, container_id, node_name, namespace) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))`
	TEMPLATE_ERROR_RATE_INSTANCE         = "100*(" +
		"(sum by(%s)(increase(kindling_span_trace_duration_nanoseconds_count{%s, is_error='true'}[%s])) or 0)" + // or 0补充缺失数据场景
		"/sum by(%s)(increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))" +
		") or (sum by(%s)(increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s])) * 0)" // or * 0补充中间缺失数据的场景
	TEMPLATE_GET_NAMESPACES = `sum(kindling_span_trace_duration_nanoseconds_count{namespace=~".+"}[%s]) by (namespace)`
)

// GetServiceList 查询服务名列表
func (repo *promRepo) GetServiceList(startTime int64, endTime int64, namespace []string) ([]string, error) {
	var namespaceFilter string
	if len(namespace) > 0 {
		namespaceFilter = fmt.Sprintf(`%s"%s"`, NamespaceRegexPQLFilter, strings.Join(namespace, "|"))
	}
	query := fmt.Sprintf(TEMPLATE_GET_SERVICES, namespaceFilter, VecFromS2E(startTime, endTime))
	value, _, err := repo.GetApi().Query(context.Background(), query, time.UnixMicro(endTime))

	if err != nil {
		return nil, err
	}
	result := make([]string, 0)
	vector, ok := value.(prometheus_model.Vector)
	if !ok {
		return result, nil
	}
	for _, sample := range vector {
		result = append(result, string(sample.Metric["svc_name"]))
	}
	return result, nil
}

// GetServiceEndPointList 查询服务Endpoint列表, 服务名允许为空
// TODO not allowed empty?
func (repo *promRepo) GetServiceEndPointList(startTime int64, endTime int64, serviceName string) ([]string, error) {
	queryCondition := ""
	if serviceName != "" {
		queryCondition = fmt.Sprintf("svc_name='%s'", serviceName)
	}
	query := fmt.Sprintf(TEMPLATE_GET_ENDPOINTS, queryCondition, VecFromS2E(startTime, endTime))
	value, _, err := repo.GetApi().Query(context.Background(), query, time.UnixMicro(endTime))

	if err != nil {
		return nil, err
	}
	result := make([]string, 0)
	vector, ok := value.(prometheus_model.Vector)
	if !ok {
		return result, nil
	}
	for _, sample := range vector {
		result = append(result, string(sample.Metric["content_key"]))
	}
	return result, nil
}

// 查询活跃实例列表
func (repo *promRepo) GetActiveInstanceList(startTime int64, endTime int64, serviceName string) (*model.ServiceInstances, error) {
	queryCondition := fmt.Sprintf("svc_name='%s'", serviceName)
	query := fmt.Sprintf(TEMPLATE_GET_ACTIVE_SERVICE_INSTANCE, queryCondition, VecFromS2E(startTime, endTime))
	res, _, err := repo.GetApi().Query(context.Background(), query, time.UnixMicro(endTime))
	if err != nil {
		return nil, err
	}
	result := model.NewServiceInstances()
	vector, ok := res.(prometheus_model.Vector)
	if !ok {
		return result, nil
	}
	instances := make([]*model.ServiceInstance, 0)
	for _, sample := range vector {
		if float64(sample.Value) > 0 {
			pidStr := sample.Metric["pid"]
			pid, _ := strconv.ParseInt(string(pidStr), 10, 64)

			instances = append(instances, &model.ServiceInstance{
				ServiceName: string(sample.Metric["svc_name"]),
				ContainerId: string(sample.Metric["container_id"]),
				PodName:     string(sample.Metric["pod"]),
				Namespace:   string(sample.Metric["namespace"]),
				NodeName:    string(sample.Metric["node_name"]),
				Pid:         pid,
			})
		}
	}
	result.AddInstances(instances)
	return result, nil
}

// GetInstanceList 查询服务实例列表, URL允许为空
func (repo *promRepo) GetInstanceList(startTime int64, endTime int64, serviceName string, url string) (*model.ServiceInstances, error) {
	var queryCondition string
	if url == "" {
		queryCondition = fmt.Sprintf("svc_name='%s'", serviceName)
	} else {
		queryCondition = fmt.Sprintf("svc_name='%s',content_key='%s'", serviceName, url)
	}
	query := fmt.Sprintf(TEMPLATE_GET_SERVICE_INSTANCE, queryCondition, VecFromS2E(startTime, endTime))
	res, _, err := repo.GetApi().Query(context.Background(), query, time.UnixMicro(endTime))
	if err != nil {
		return nil, err
	}

	result := model.NewServiceInstances()
	vector, ok := res.(prometheus_model.Vector)
	if !ok {
		return result, nil
	}

	instances := make([]*model.ServiceInstance, 0)
	for _, sample := range vector {
		pidStr := sample.Metric["pid"]
		pid, _ := strconv.ParseInt(string(pidStr), 10, 64)

		instances = append(instances, &model.ServiceInstance{
			ServiceName: string(sample.Metric["svc_name"]),
			ContainerId: string(sample.Metric["container_id"]),
			PodName:     string(sample.Metric["pod"]),
			Namespace:   string(sample.Metric["namespace"]),
			NodeName:    string(sample.Metric["node_name"]),
			Pid:         pid,
			NodeIP:      string(sample.Metric["node_ip"]),
		})
	}
	result.AddInstances(instances)
	return result, nil
}

func (repo *promRepo) GetMultiServicesInstanceList(startTime int64, endTime int64, services []string) (map[string]*model.ServiceInstances, error) {
	var queryCondition = fmt.Sprintf("svc_name=~'%s'", RegexMultipleValue(services...))
	query := fmt.Sprintf(TEMPLATE_GET_SERVICE_INSTANCE, queryCondition, VecFromS2E(startTime, endTime))
	res, _, err := repo.GetApi().Query(context.Background(), query, time.UnixMicro(endTime))
	if err != nil {
		return nil, err
	}

	result := make(map[string]*model.ServiceInstances)
	vector, ok := res.(prometheus_model.Vector)
	if !ok {
		return result, nil
	}
	serviceMapList := make(map[string][]*model.ServiceInstance)
	for _, sample := range vector {
		pidStr := sample.Metric["pid"]
		pid, _ := strconv.ParseInt(string(pidStr), 10, 64)

		instance := &model.ServiceInstance{
			ServiceName: string(sample.Metric["svc_name"]),
			ContainerId: string(sample.Metric["container_id"]),
			PodName:     string(sample.Metric["pod"]),
			Namespace:   string(sample.Metric["namespace"]),
			NodeName:    string(sample.Metric["node_name"]),
			Pid:         pid,
		}
		if list, ok := serviceMapList[instance.ServiceName]; ok {
			serviceMapList[instance.ServiceName] = append(list, instance)
		} else {
			serviceMapList[instance.ServiceName] = []*model.ServiceInstance{instance}
		}
	}
	for k, v := range serviceMapList {
		result[k] = model.NewServiceInstances()
		result[k].AddInstances(v)
	}
	return result, nil
}

func (repo *promRepo) QueryInstanceErrorRate(startTime int64, endTime int64, step int64, endpoint string, instance *model.ServiceInstance) (map[int64]float64, error) {
	tRange := v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(step * 1000),
	}
	var queryCondition string
	var queryGroup string
	if instance.PodName != "" {
		queryGroup = "pod"
		queryCondition = fmt.Sprintf("svc_name='%s', content_key='%s', pod='%s'",
			instance.ServiceName,
			endpoint,
			instance.PodName,
		)
	} else if instance.ContainerId != "" {
		queryGroup = "node_name, container_id"
		queryCondition = fmt.Sprintf("svc_name='%s', content_key='%s', node_name='%s', container_id='%s'",
			instance.ServiceName,
			endpoint,
			instance.NodeName,
			instance.ContainerId,
		)
	} else {
		// VM场景
		queryGroup = "node_name, pid"
		queryCondition = fmt.Sprintf("svc_name='%s', content_key='%s', node_name='%s', pid='%d'",
			instance.ServiceName,
			endpoint,
			instance.NodeName,
			instance.Pid,
		)
	}
	queryStep := getDurationFromStep(tRange.Step)
	query := fmt.Sprintf(TEMPLATE_ERROR_RATE_INSTANCE,
		queryGroup, queryCondition, queryStep,
		queryGroup, queryCondition, queryStep,
		queryGroup, queryCondition, queryStep,
	)
	res, _, err := repo.GetApi().QueryRange(context.Background(), query, tRange)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]float64)
	values, ok := res.(prometheus_model.Matrix)
	if !ok {
		return result, nil
	}
	if len(values) == 1 {
		val := values[0]
		for _, pair := range val.Values {
			result[int64(pair.Timestamp)*1000] = float64(pair.Value)
		}
	}
	return result, nil
}
