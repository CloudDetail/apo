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

const queryProcessStartPromQL = "originx_process_start_time{%s}"

// QueryProcessStartTime 查询进程启动时间
// 返回结果中包含了查询pid在多个节点上的进程，因此需要根据node_name进行筛选
func (repo *promRepo) QueryProcessStartTime(startTime time.Time, endTime time.Time, step time.Duration, pids []string) (map[model.ServiceInstance]int64, error) {
	// pid 有可能在不同主机上是重复的，无法在查询时筛选出对应node和pid的数据，因此查询后要针对node_name进行筛选
	var queryCondition = fmt.Sprintf("pid=~'%s'", MultipleValue(pids...))
	tRange := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  step,
	}
	query := fmt.Sprintf(queryProcessStartPromQL, queryCondition)
	res, _, err := repo.GetApi().QueryRange(context.Background(), query, tRange)
	if err != nil {
		return nil, err
	}

	result := make(map[model.ServiceInstance]int64)
	matrix, ok := res.(prometheus_model.Matrix)
	if !ok {
		return result, nil
	}
	for _, sample := range matrix {
		pid, _ := strconv.ParseInt(string(sample.Metric["pid"]), 10, 64)
		instance := model.ServiceInstance{
			Pid:      pid,
			NodeName: string(sample.Metric["node_name"]),
		}
		value := sample.Values[0].Value
		result[instance] = int64(value)
	}

	return result, nil
}
