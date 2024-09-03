package prometheus

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	pmodel "github.com/prometheus/common/model"
)

const queryProcessStartPromQL = "max by (node_name,pid) (last_over_time(originx_process_start_time{%s}[%s]))"

func (repo *promRepo) QueryProcessStartTime(startTime time.Time, endTime time.Time, instances []*model.ServiceInstance) (map[model.ServiceInstance]int64, error) {
	vector := VecFromS2E(startTime.UnixMicro(), endTime.UnixMicro())

	var filters []string
	var pids []string
	var nodeNames []string
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		pids = append(pids, strconv.FormatInt(instance.Pid, 10))
		nodeNames = append(nodeNames, EscapeRegexp(instance.NodeName))
	}

	if len(pids) > 0 {
		filters = append(filters, fmt.Sprintf("pid=~\"%s\"", strings.Join(pids, "|")))
	}
	if len(nodeNames) > 0 {
		filters = append(filters, fmt.Sprintf("node_name=~\"%s\"", strings.Join(nodeNames, "|")))
	}

	query := fmt.Sprintf(
		queryProcessStartPromQL,
		strings.Join(filters, ","),
		vector,
	)

	res, _, err := repo.GetApi().Query(context.Background(), query, endTime)
	if err != nil {
		return nil, err
	}

	samples, ok := res.(pmodel.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected query result type %T, expected model.Vector", res)
	}

	var startTSmap = make(map[model.ServiceInstance]int64)
	for _, sample := range samples {
		if math.IsNaN(float64(sample.Value)) || math.IsInf(float64(sample.Value), 0) {
			continue
		}

		nodeName := string(sample.Metric["node_name"])
		pid := string(sample.Metric["pid"])
		pidI64, err := strconv.ParseInt(pid, 10, 64)
		if err != nil {
			continue
		}
		for _, instance := range instances {
			if instance == nil {
				continue
			}
			if instance.NodeName == nodeName && instance.Pid == pidI64 {
				startTSmap[*instance] = int64(sample.Value)
				break
			}
		}
	}

	return startTSmap, nil
}
