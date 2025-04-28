// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/CloudDetail/apo/backend/pkg/util"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

type podMap struct {
	OptStatus   string                   `json:"OPT_STATUS"`
	Description string                   `json:"DESCRIPTION"`
	Result      *response.PodMapResponse `json:"result"`
	Debug       string                   `json:"debug"`
}

const (
	sqlTemplate = "SELECT %s FROM vtap_app_edge_port %s %s"
	// Predefined Query Fields
	queryFields = "PerSecond(Avg(`request`)) AS `PerSecond(Avg(request))`, " +
		"Avg(`server_error_ratio`) AS `Avg(server_error_ratio)`, " +
		"Avg(`rrt`) AS `Avg(rrt)`, " +
		"node_type(pod_0) AS `client_node_type`, " +
		"icon_id(pod_0) AS `client_icon_id`, " +
		"pod_id_0, pod_0, " +
		"node_type(pod_1) AS `server_node_type`, " +
		"icon_id(pod_1) AS `server_icon_id`, " +
		"pod_id_1, pod_1, " +
		"Enum(tap_side), tap_side"
)

var byteUnmarshallingValidator = util.NewByteValidator(10*1024*1024, []string{}, []string{"$func", "$eval", "constructor", "prototype"}, 10)

// Validate input parameters
func validatePodMapRequest(req *request.PodMapRequest) error {
	// Validate time range
	if req.StartTime <= 0 || req.EndTime <= 0 || req.StartTime > req.EndTime {
		return fmt.Errorf("invalid time range")
	}

	// Validate namespace (only letters, numbers, underscores and hyphens allowed)
	if req.Namespace != "" {
		if !util.IsValidIdentifier(req.Namespace) {
			return fmt.Errorf("invalid namespace format")
		}
	}

	// Validate workload (only letters, numbers, underscores and hyphens allowed)
	if req.Workload != "" {
		if !util.IsValidIdentifier(req.Workload) {
			return fmt.Errorf("invalid workload format")
		}
	}

	return nil
}

func buildPodMapQuery(req *request.PodMapRequest) (string, error) {
	// Validate input parameters
	if err := validatePodMapRequest(req); err != nil {
		return "", err
	}

	// Build base query conditions
	queryWheres := fmt.Sprintf("WHERE time >= %d AND time <= %d", req.StartTime/1e6, req.EndTime/1e6)

	// Add optional conditions using parameterized approach
	if req.Namespace != "" {
		queryWheres += fmt.Sprintf(" AND pod_ns_1 = '%s'", util.EscapeSQLString(req.Namespace))
	}
	if req.Workload != "" {
		queryWheres += fmt.Sprintf(" AND pod_group_1 = '%s'", util.EscapeSQLString(req.Workload))
	}

	// Use predefined query builder
	queryBys := clickhouse.NewByLimitBuilder().
		GroupBy("pod_0, pod_1, tap_side, pod_id_0, pod_id_1, `client_node_type`, `server_node_type`").
		Limit(500)

	// Use predefined query fields
	sql := fmt.Sprintf(sqlTemplate, queryFields, queryWheres, queryBys.String())
	return sql, nil
}

func (s *service) GetPodMap(req *request.PodMapRequest) (*response.PodMapResponse, error) {
	deepflowServer := config.Get().DeepFlow.ServerAddress

	// Build secure SQL query
	sql, err := buildPodMapQuery(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	db := "flow_metrics"
	dataPrecision := "1m"

	// Build request body parameters
	formData := url.Values{
		"db":             {db},
		"data_precision": {dataPrecision},
		"sql":            {sql},
	}
	// Initiate a POST request
	resp, err := http.PostForm(deepflowServer+"/v1/query", formData)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to deepflow server: %w", err)
	}
	defer resp.Body.Close()

	// Read Response Body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response body to podMapRequest structure
	podMapResp := &podMap{
		Result: new(response.PodMapResponse),
	}
	err = byteUnmarshallingValidator.ValidateAndUnmarshalJSON(body, &podMapResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return podMapResp.Result, nil
}
