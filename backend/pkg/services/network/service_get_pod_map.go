// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"encoding/json"
	"fmt"
	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"io"
	"net/http"
	"net/url"
)

type podMap struct {
	OptStatus   string                   `json:"OPT_STATUS"`
	Description string                   `json:"DESCRIPTION"`
	Result      *response.PodMapResponse `json:"result"`
	Debug       string                   `json:"debug"`
}

const (
	sqlTemplate = "SELECT %s FROM vtap_app_edge_port %s %s"
)

func (s *service) GetPodMap(req *request.PodMapRequest) (*response.PodMapResponse, error) {
	deepflowServer := config.Get().DeepFlow.ServerAddress
	queryFields := "PerSecond(Avg(`request`)) AS `PerSecond(Avg(request))`, Avg(`server_error_ratio`) AS `Avg(server_error_ratio)`, Avg(`rrt`) AS `Avg(rrt)`, node_type(pod_0) AS `client_node_type`, icon_id(pod_0) AS `client_icon_id`, pod_id_0, pod_0, node_type(pod_1) AS `server_node_type`, icon_id(pod_1) AS `server_icon_id`, pod_id_1, pod_1, Enum(tap_side), tap_side"
	queryWheres := fmt.Sprintf("WHERE time >= %d AND time <= %d", req.StartTime/1e6, req.EndTime/1e6)
	if req.Namespace != "" {
		queryWheres += fmt.Sprintf(" AND pod_ns_1 = '%s'", req.Namespace)
	}
	if req.Workload != "" {
		queryWheres += fmt.Sprintf(" AND pod_group_1 = '%s'", req.Workload)
	}

	queryBys := clickhouse.NewByLimitBuilder().
		GroupBy("pod_0, pod_1, tap_side, pod_id_0, pod_id_1, `client_node_type`, `server_node_type`").
		Limit(500)
	sql := fmt.Sprintf(sqlTemplate, queryFields, queryWheres, queryBys.String())
	db := "flow_metrics"
	dataPrecision := "1m"

	// 构建请求体参数
	formData := url.Values{
		"db":             {db},
		"data_precision": {dataPrecision},
		"sql":            {sql},
	}
	// 发起 POST 请求
	resp, err := http.PostForm(deepflowServer+"/v1/query", formData)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to deepflow server: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应体到 podMapRequest 结构体
	podMapResp := &podMap{
		Result: new(response.PodMapResponse),
	}
	err = json.Unmarshal(body, &podMapResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return podMapResp.Result, nil
}

func getProtocolNum(protocol string) int {
	switch protocol {
	case "HTTP":
		return 20
	case "HTTPS":
		return 21
	}
	return 0
}
