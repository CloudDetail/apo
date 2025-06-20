// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prometheus_model "github.com/prometheus/common/model"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	TEMPLATE_GET_SERVICES_BY_FILTER      = `group by (svc_name) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s])> 0)`
	TEMPLATE_GET_SERVICES                = `sum by(svc_name) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))`
	TEMPLATE_GET_SERVICES_WITH_NAMESPACE = `sum by(svc_name, namespace) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))`
	TEMPLATE_GET_ENDPOINTS               = `sum by(content_key) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))`
	TEMPLATE_GET_SERVICE_INSTANCE        = `sum by(svc_name, pod, pid, container_id, node_name, namespace, node_ip) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))`
	TEMPLATE_GET_ACTIVE_INSTANCE         = `sum by(svc_name, pod, pid, container_id, node_name, namespace) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s])) > 0`
	TEMPLATE_GET_ACTIVE_SERVICE_INSTANCE = `sum by(svc_name, pod, pid, container_id, node_name, namespace) (increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))`
	TEMPLATE_ERROR_RATE_INSTANCE         = "100*(" +
		"(sum by(%s)(increase(kindling_span_trace_duration_nanoseconds_count{%s, is_error='true'}[%s])) or 0)" + // or 0 Supplements missing data scenarios
		"/sum by(%s)(increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s]))" +
		") or (sum by(%s)(increase(kindling_span_trace_duration_nanoseconds_count{%s}[%s])) * 0)" // or * 0补充中间缺失数据的场景
	TEMPLATE_GET_NAMESPACES            = `sum(kindling_span_trace_duration_nanoseconds_count{namespace=~".+"}[%s]) by (namespace)`
	TEMPLATE_GET_NAMESPACES_BY_SERVICE = `sum(kindling_span_trace_duration_nanoseconds_count{%s}[%s]) by (namespace)`

	TEMPLATE_GET_SERVICE_BY_DB = `group by (svc_name)
(
	last_over_time(kindling_db_duration_nanoseconds_count{%s}[%s])
	or
	(apo_network_middleware_connect{%s} * on (node_name,pid,container_id)
		group_left (svc_name) (group by(node_name,pid,container_id,svc_name)
		(last_over_time(kindling_span_trace_duration_nanoseconds_count[%s]))))
)`
)

// GetServiceList to query the service name list
func (repo *promRepo) GetServiceList(ctx core.Context, startTime int64, endTime int64, namespace []string) ([]string, error) {
	var namespaceFilter string
	if len(namespace) > 0 {
		namespaceFilter = fmt.Sprintf(`%s"%s"`, NamespaceRegexPQLFilter, RegexMultipleValue(namespace...))
	}
	query := fmt.Sprintf(TEMPLATE_GET_SERVICES, namespaceFilter, VecFromS2E(startTime, endTime))
	value, _, err := repo.GetApi().Query(ctx.GetContext(), query, time.UnixMicro(endTime))

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

// GetServiceWithNamespace Get service list and namespace that belong to.
func (repo *promRepo) GetServiceWithNamespace(ctx core.Context, startTime, endTime int64, namespaces []string) (map[string][]string, error) {
	var namespaceFilter string
	if len(namespaces) > 0 {
		namespaceFilter = fmt.Sprintf(`%s"%s"`, NamespaceRegexPQLFilter, RegexMultipleValue(namespaces...))
	}
	query := fmt.Sprintf(TEMPLATE_GET_SERVICES_WITH_NAMESPACE, namespaceFilter, VecFromS2E(startTime, endTime))
	value, _, err := repo.GetApi().Query(ctx.GetContext(), query, time.UnixMicro(endTime))

	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	vector, ok := value.(prometheus_model.Vector)
	if !ok {
		return result, nil
	}

	for _, sample := range vector {
		service := string(sample.Metric["svc_name"])
		namespace := string(sample.Metric["namespace"])
		if len(service) == 0 {
			continue
		}
		if _, exists := result[service]; !exists {
			result[service] = []string{}
		}
		result[service] = append(result[service], namespace)
	}

	return result, nil
}

func (repo *promRepo) GetServiceNamespace(ctx core.Context, startTime, endTime int64, service string) ([]string, error) {
	var serviceFilter string
	if len(service) > 0 {
		serviceFilter = fmt.Sprintf(`%s"%s"`, ServicePQLFilter, service)
	}
	query := fmt.Sprintf(TEMPLATE_GET_NAMESPACES_BY_SERVICE, serviceFilter, VecFromS2E(startTime, endTime))
	value, _, err := repo.GetApi().Query(ctx.GetContext(), query, time.UnixMicro(endTime))

	if err != nil {
		return nil, err
	}

	result := make([]string, 0)
	vector, ok := value.(prometheus_model.Vector)
	if !ok {
		return result, nil
	}

	for _, sample := range vector {
		namespace := string(sample.Metric["namespace"])
		result = append(result, namespace)
	}

	return result, nil
}

// GetNamespaceWithService Get namespace list and service that under it.
func (repo *promRepo) GetNamespaceWithService(ctx core.Context, startTime, endTime int64) (map[string][]string, error) {
	var namespaceFilter string

	query := fmt.Sprintf(TEMPLATE_GET_SERVICES_WITH_NAMESPACE, namespaceFilter, VecFromS2E(startTime, endTime))
	value, _, err := repo.GetApi().Query(ctx.GetContext(), query, time.UnixMicro(endTime))

	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	vector, ok := value.(prometheus_model.Vector)
	if !ok {
		return result, nil
	}

	for _, sample := range vector {
		service := string(sample.Metric["svc_name"])
		namespace := string(sample.Metric["namespace"])
		if len(service) == 0 && len(namespace) == 0 {
			continue
		}
		if _, exists := result[namespace]; !exists {
			result[namespace] = []string{}
		}
		result[namespace] = append(result[namespace], service)
	}

	return result, nil
}

var svcDomainRegex = regexp.MustCompile(`(.*)\.svc\..*$`)

func cutServiceSuffixFromURL(url string) string {
	// <svc>.<ns>.svc.<cluster_domain> / <svc>.<ns>.svc.cluster.local
	parts := svcDomainRegex.FindStringSubmatch(url)
	if len(parts) > 1 {
		url = parts[1]
	}
	// <svc>.<ns>
	lastDot := strings.LastIndexByte(url, '.')
	if lastDot != -1 {
		return url[:lastDot]
	}
	// <svc>
	return url
}

// GetServiceListByDBInfo 基于数据库信息获取关联的服务列表
func (repo *promRepo) GetServiceListByDatabase(ctx core.Context,
	startTime, endTime time.Time,
	dbURL, dbIP, dbPort string,
) ([]string, error) {
	// 清理URL后面的Namespace信息,会导致无法查询namespace
	dbURL = cutServiceSuffixFromURL(dbURL)
	vector := VecFromS2E(startTime.UnixMicro(), endTime.UnixMicro())
	pql := fmt.Sprintf(TEMPLATE_GET_SERVICE_BY_DB,
		fmt.Sprintf(`db_url=~"%s.*",db_url!=""`, EscapeRegexp(dbURL)),
		vector,
		fmt.Sprintf(`db_ip="%s",db_port="%s",db_ip!="",db_port!=""`, dbIP, dbPort),
		vector,
	)
	ress, err := repo.QueryData(ctx, endTime, pql)
	if err != nil {
		return nil, err
	}
	var services []string
	for _, res := range ress {
		services = append(services, res.Metric.SvcName)
	}
	return services, nil
}

func (repo *promRepo) GetServiceListByFilter(ctx core.Context, startTime time.Time, endTime time.Time, filterKVs ...string) ([]string, error) {
	if len(filterKVs)%2 != 0 {
		return nil, fmt.Errorf("size of filterKVs is not even: %d", len(filterKVs))
	}
	var filters []string
	for i := 0; i+1 < len(filterKVs); i += 2 {
		filters = append(filters, fmt.Sprintf("%s\"%s\"", filterKVs[i], filterKVs[i+1]))
	}

	// 如果时间低于1h就用1h
	var vectorStr = "1h"
	if endTime.Sub(startTime) > time.Hour {
		vectorStr = VecFromS2E(startTime.UnixMicro(), endTime.UnixMicro())
	}

	pql := fmt.Sprintf(
		TEMPLATE_GET_SERVICES_BY_FILTER,
		strings.Join(filters, ","),
		vectorStr,
	)
	ress, err := repo.QueryData(ctx, endTime, pql)
	if err != nil {
		return nil, err
	}
	var services []string
	for _, res := range ress {
		services = append(services, res.Metric.SvcName)
	}
	return services, nil
}

// GetServiceEndPointList to query the service Endpoint list. The service name can be empty.
func (repo *promRepo) GetServiceEndPointList(ctx core.Context, startTime int64, endTime int64, serviceName string) ([]string, error) {
	queryCondition := ""
	if serviceName != "" {
		queryCondition = fmt.Sprintf("svc_name='%s'", serviceName)
	}
	query := fmt.Sprintf(TEMPLATE_GET_ENDPOINTS, queryCondition, VecFromS2E(startTime, endTime))
	value, _, err := repo.GetApi().Query(ctx.GetContext(), query, time.UnixMicro(endTime))

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

// Query the list of active instances
func (repo *promRepo) GetActiveInstanceList(ctx core.Context, startTime int64, endTime int64, serviceNames []string) (*model.ServiceInstances, error) {
	queryCondition := fmt.Sprintf("%s'%s'", ServiceRegexPQLFilter, RegexMultipleValue(serviceNames...))

	query := fmt.Sprintf(TEMPLATE_GET_ACTIVE_SERVICE_INSTANCE, queryCondition, VecFromS2E(startTime, endTime))
	res, _, err := repo.GetApi().Query(ctx.GetContext(), query, time.UnixMicro(endTime))
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

// GetInstanceList to query the service instance list. The URL can be empty.
func (repo *promRepo) GetInstanceList(ctx core.Context, startTime int64, endTime int64, serviceName string, url string) (*model.ServiceInstances, error) {
	var queryCondition string
	if url == "" {
		queryCondition = fmt.Sprintf("svc_name='%s'", serviceName)
	} else {
		queryCondition = fmt.Sprintf("svc_name='%s',content_key='%s'", serviceName, url)
	}
	query := fmt.Sprintf(TEMPLATE_GET_SERVICE_INSTANCE, queryCondition, VecFromS2E(startTime, endTime))
	res, _, err := repo.GetApi().Query(ctx.GetContext(), query, time.UnixMicro(endTime))
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

func (repo *promRepo) GetMultiServicesInstanceList(ctx core.Context, startTime int64, endTime int64, services []string) (map[string]*model.ServiceInstances, error) {
	var queryCondition = fmt.Sprintf("svc_name=~'%s'", RegexMultipleValue(services...))
	query := fmt.Sprintf(TEMPLATE_GET_SERVICE_INSTANCE, queryCondition, VecFromS2E(startTime, endTime))
	res, _, err := repo.GetApi().Query(ctx.GetContext(), query, time.UnixMicro(endTime))
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

func (repo *promRepo) QueryInstanceErrorRate(ctx core.Context, startTime int64, endTime int64, step int64, endpoint string, instance *model.ServiceInstance) (map[int64]float64, error) {
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
		// VM scenario
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
	res, _, err := repo.GetApi().QueryRange(ctx.GetContext(), query, tRange)
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
