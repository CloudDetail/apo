// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"strings"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prometheus_model "github.com/prometheus/common/model"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	TEMPLATE_GET_PODS_BY_FILTER = `group by(node_name, namespace, pod, pid) (increase(originx_thread_polaris_nanoseconds_count{%s}[%s])> 0)`
)

func (repo *promRepo) GetPodList(ctx core.Context, startTime int64, endTime int64, nodeName string, namespace string, podName string) ([]*model.Pod, error) {
	filters := []string{
		`type="cpu"`,
	}
	if nodeName != "" {
		filters = append(filters, fmt.Sprintf(`node_name="%s"`, nodeName))
	}
	if namespace != "" {
		filters = append(filters, fmt.Sprintf(`namespace="%s"`, namespace))
	}
	if podName != "" {
		filters = append(filters, fmt.Sprintf(`pod="%s"`, podName))
	}
	query := fmt.Sprintf(
		TEMPLATE_GET_PODS_BY_FILTER,
		strings.Join(filters, ","),
		VecFromS2E(startTime, endTime),
	)
	value, _, err := repo.GetApi().QueryRange(ctx.GetContext(), query, v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(endTime-startTime) * time.Microsecond,
	})
	if err != nil {
		return nil, err
	}

	vector, ok := value.(prometheus_model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected model.Vector", value)
	}

	resultMap := make(map[_podKey]*model.Pod, 0)
	for _, sample := range vector {
		nodeName := string(sample.Metric["node_name"])
		namespace := string(sample.Metric["namespace"])
		pod := string(sample.Metric["pod"])

		podKey := _podKey{
			NodeName:  nodeName,
			Namespace: namespace,
			Pod:       pod,
		}

		if pod, find := resultMap[podKey]; find {
			pod.Pids = append(pod.Pids, string(sample.Metric["pid"]))
			continue
		}

		resultMap[podKey] = &model.Pod{
			NodeName:  nodeName,
			Namespace: namespace,
			Pod:       pod,
			Pids:      []string{string(sample.Metric["pid"])},
		}
	}

	result := make([]*model.Pod, 0)
	for _, pod := range resultMap {
		result = append(result, pod)
	}

	return result, nil
}

type _podKey struct {
	NodeName  string
	Namespace string
	Pod       string
}
