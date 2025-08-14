// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	SQL_LIST_APP_INFO_LABEL_KEYS = `SELECT DISTINCT arrayJoin(mapKeys(labels)) AS key
FROM originx_app_info
%s`
	SQL_LIST_APP_INFO_LABEL_VALUES = `SELECT DISTINCT %s
FROM originx_app_info
%s`
	SQL_GET_APP_INFO_LABELS = `SELECT 
		labels,
		labels['pod_name'] as pod_name,
		labels['service_name'] as service_name, 
		labels['node_name'] as node_name,
		container_id,
		host_pid
	FROM originx_app_info %s ORDER BY timestamp DESC`
)

func (repo *chRepo) ListAppInfoLabelKeys(ctx core.Context, startTime, endTime int64) ([]string, error) {
	qb := NewQueryBuilder().
		Between("timestamp", startTime/1e6, endTime/1e6)

	sql := fmt.Sprintf(SQL_LIST_APP_INFO_LABEL_KEYS, qb.String())
	rows, err := repo.GetConn(ctx).Query(ctx.GetContext(), sql, qb.values...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var result []string
	for rows.Next() {
		var labelsKey string
		if err := rows.Scan(&labelsKey); err != nil {
			return nil, err
		}
		result = append(result, labelsKey)
	}
	return result, nil
}

var safeKeyRe = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

func escapeAppInfoKeys(key string) (string, error) {
	if !safeKeyRe.MatchString(key) {
		return "", fmt.Errorf("invalid label key")
	}
	escapedKey := strings.ReplaceAll(key, `'`, `''`)
	return "labels['" + escapedKey + "']", nil
}

func (repo *chRepo) ListAppInfoLabelValues(ctx core.Context, startTime, endTime int64, key string) ([]string, error) {
	labelKey, err := escapeAppInfoKeys(key)
	if err != nil {
		return nil, err
	}
	fb := NewFieldBuilder().
		Alias(labelKey, "value")

	qb := NewQueryBuilder().
		Between("timestamp", startTime/1e6, endTime/1e6)

	sql := fmt.Sprintf(SQL_LIST_APP_INFO_LABEL_VALUES, fb.String(), qb.String())
	rows, err := repo.GetConn(ctx).Query(ctx.GetContext(), sql, qb.values...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var result []string
	for rows.Next() {
		var labelsValue string
		if err := rows.Scan(&labelsValue); err != nil {
			return nil, err
		}
		result = append(result, labelsValue)
	}
	return result, nil
}

func (repo *chRepo) GetInstancesLabels(ctx core.Context, startTime, endTime int64, instances []*model.ServiceInstance) (map[string]map[string]string, error) {
	if len(instances) == 0 {
		return make(map[string]map[string]string), nil
	}

	queryBuilder := NewQueryBuilder().
		Between("toUnixTimestamp(timestamp)", time.UnixMicro(startTime), time.UnixMicro(endTime))

	podNamesSet := make(map[string]bool)
	serviceNamesSet := make(map[string]bool)
	nodeNamesSet := make(map[string]bool)
	containerIdsSet := make(map[string]bool)
	pidSet := make(map[uint32]bool)

	for _, instance := range instances {
		podNamesSet[instance.PodName] = true
		serviceNamesSet[instance.ServiceName] = true
		nodeNamesSet[instance.NodeName] = true
		containerIdsSet[instance.ContainerId] = true
		pidSet[uint32(instance.Pid)] = true

	}

	if len(podNamesSet) > 0 {
		podNames := make([]string, 0, len(podNamesSet))
		for podName := range podNamesSet {
			podNames = append(podNames, podName)
		}
		queryBuilder.In("labels['pod_name']", podNames)
	}

	if len(serviceNamesSet) > 0 {
		serviceNames := make([]string, 0, len(serviceNamesSet))
		for serviceName := range serviceNamesSet {
			serviceNames = append(serviceNames, serviceName)
		}
		queryBuilder.In("labels['service_name']", serviceNames)
	}

	if len(nodeNamesSet) > 0 {
		nodeNames := make([]string, 0, len(nodeNamesSet))
		for nodeName := range nodeNamesSet {
			nodeNames = append(nodeNames, nodeName)
		}
		queryBuilder.In("labels['node_name']", nodeNames)
	}

	if len(containerIdsSet) > 0 {
		containerIds := make([]string, 0, len(containerIdsSet))
		for containerId := range containerIdsSet {
			containerIds = append(containerIds, containerId)
		}
		queryBuilder.In("container_id", containerIds)
	}

	if len(pidSet) > 0 {
		pids := make([]uint32, 0, len(containerIdsSet))
		for pid := range pidSet {
			pids = append(pids, pid)
		}
		queryBuilder.In("host_pid", pids)
	}

	type QueryResult struct {
		Labels      map[string]string `ch:"labels"`
		PodName     string            `ch:"pod_name"`
		ServiceName string            `ch:"service_name"`
		NodeName    string            `ch:"node_name"`
		ContainerId string            `ch:"container_id"`
		HostPid     uint32            `ch:"host_pid"`
	}

	res := []QueryResult{}
	sql := fmt.Sprintf(SQL_GET_APP_INFO_LABELS, queryBuilder.String())

	err := repo.GetContextDB(ctx).Select(ctx.GetContext(), &res, sql, queryBuilder.values...)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]string)

	instanceMap := make(map[string]*model.ServiceInstance, len(instances))
	for _, inst := range instances {
		key := fmt.Sprintf("%s-%s-%s-%s-%d", inst.ServiceName, inst.PodName, inst.NodeName, inst.ContainerId, inst.Pid)
		instanceMap[key] = inst
	}

	for _, row := range res {
		key := fmt.Sprintf("%s-%s-%s-%s-%d", row.ServiceName, row.PodName, row.NodeName, row.ContainerId, row.HostPid)
		if _, ok := instanceMap[key]; ok {
			result[key] = row.Labels
		}
	}

	return result, nil
}
