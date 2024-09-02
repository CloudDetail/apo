package prometheus

import (
	"context"
	"fmt"
	"strconv"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prometheus_model "github.com/prometheus/common/model"

	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	queryProcessStartPromQL   = "max by (node_name, pid) (originx_process_start_time{%s})"
	queryContainerStartPromQL = "max by (node_name, container_id) (originx_process_start_time{%s})"
)

// QueryProcessStartTime 查询进程最近启动时间
// 返回结果中包含了查询pid在多个节点上的进程，因此需要根据node_name进行筛选;
// 注意同一个容器中可能存在多个进程，这里假设这些进程启动时间接近，取最近的进程启动时间作为容器启动时间
func (repo *promRepo) QueryProcessStartTime(startTime time.Time, endTime time.Time, step time.Duration, pids []string, containerIds []string) (map[model.ServiceInstance]int64, error) {
	// pid 有可能在不同主机上是重复的，无法在查询时筛选出对应node和pid的数据，因此查询后要针对node_name进行筛选
	tRange := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  step,
	}
	queryPids := fmt.Sprintf(queryProcessStartPromQL, fmt.Sprintf("pid=~'%s'", RegexMultipleValue(pids...)))
	queryContainerIds := fmt.Sprintf(queryContainerStartPromQL, fmt.Sprintf("container_id=~'%s'", RegexMultipleValue(containerIds...)))
	var queryPidsOrContainerIds = fmt.Sprintf("%s or %s", queryPids, queryContainerIds)
	res, _, err := repo.GetApi().QueryRange(context.Background(), queryPidsOrContainerIds, tRange)
	if err != nil {
		return nil, err
	}

	result := make(map[model.ServiceInstance]int64)
	matrix, ok := res.(prometheus_model.Matrix)
	if !ok {
		return result, nil
	}
	for _, sample := range matrix {
		// 虚拟机采用Pid查询，因此针对ContainerId为空的单独存储
		// 容器只能用ContainerId查询，因此针对ContainerId不为空的单独存储
		pid, _ := strconv.ParseInt(string(sample.Metric["pid"]), 10, 64)
		containerId := string(sample.Metric["container_id"])
		nodeName := string(sample.Metric["node_name"])
		value := sample.Values[0].Value
		if containerId == "" {
			instance := model.ServiceInstance{
				Pid:      pid,
				NodeName: nodeName,
			}
			result[instance] = int64(value)
		} else {
			instance := model.ServiceInstance{
				ContainerId: containerId,
				NodeName:    nodeName,
			}
			result[instance] = int64(value)
		}
	}
	return result, nil
}
