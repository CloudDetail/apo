// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package metric

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

var queryDict *QueryDict

func init() {
	dirPath := "static/predefined-metrics"
	queryDict = &QueryDict{
		queryMap: make(map[string][]Query),
	}
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".json") {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var metrics PreDefinedMetrics
			if err := json.Unmarshal(data, &metrics); err != nil {
				log.Printf("failed to load predefined metrics: %s, err: %v", path, err)
				return nil
			}

			queryDict.AddPreDefinedMetrics(&metrics)
		}

		return nil
	})
	if err != nil {
		log.Println("failed to init predefined metrics", err)
	}
}

func (s *service) ListPreDefinedMetrics() []QueryInfo {
	return queryDict.ListMetrics()
}

func (s *service) ListQuerys() []Query {
	return queryDict.ListQuerys()
}

func (s *service) QueryMetrics(req *QueryMetricsRequest) *QueryMetricsResult {
	// auto params

	interval := prometheus.VecFromDuration(time.Duration(req.Step) * time.Microsecond)
	req.Params["__rate_interval"] = interval
	req.Params["__interval"] = interval

	var querys []Query

	if len(req.MetricName) > 0 {
		querys = queryDict.GetQuerysByNames([]string{req.MetricName})
		if len(querys) > 0 {
			return &QueryMetricsResult{Result: s.executeQuery(querys[0], req)}
		}
	} else if len(req.MetricIds) > 0 {
		querys = queryDict.GetQuerysByIds(req.MetricIds)
	} else {
		querys = queryDict.GetQuerysByNames(req.MetricNames)
	}

	if len(querys) == 0 {
		return &QueryMetricsResult{
			Msg: "metrics not found",
		}
	}

	var resp []QueryResult
	for _, query := range querys {
		queryRes := s.executeQuery(query, req)
		resp = append(resp, *queryRes)
	}

	return &QueryMetricsResult{
		Results: resp,
	}
}

var legendFormatReg = regexp.MustCompile("{{([^{}]*)}}")

func (s *service) executeQuery(
	query Query,
	req *QueryMetricsRequest,
) *QueryResult {
	var res = []Timeseries{}

	for _, target := range query.Targets {
		value, _, err := s.executeTargets(query.GroupID, &target, req)
		if err != nil {
			log.Println(err)
			continue
		}
		legendFormat := legendFormatReg.FindAllStringSubmatch(target.LegendFormat, -1)

		matrix := value.(model.Matrix)
		for i := 0; i < matrix.Len(); i++ {
			legend := target.LegendFormat

			for _, labels := range legendFormat {
				if len(labels) > 1 {
					label, find := matrix[i].Metric[model.LabelName(strings.Trim(labels[1], " "))]
					if find {
						legend = strings.ReplaceAll(legend, fmt.Sprintf("{{%s}}", labels[1]), string(label))
					}
				}
			}

			var chart = response.TempChartObject{
				ChartData: map[int64]float64{},
			}
			for _, item := range matrix[i].Values {
				timestamp := item.Timestamp
				value := item.Value

				tsMicro := timestamp.Unix() * 1000000
				if !math.IsInf(float64(value), 0) { // does not assign value when it is infinity
					chart.ChartData[tsMicro] = float64(value)
				}
			}

			res = append(res, Timeseries{
				Legend:       legend,
				LegendFormat: target.LegendFormat,
				Labels:       matrix[i].Metric,
				// Values: matrix[i].Values,
				Chart: chart,
			})
		}
	}

	return &QueryResult{
		// Query: query,

		Title:      query.Title,
		Unit:       query.Unit,
		Timeseries: res,
	}
}

func (s *service) executeTargets(groupId int, target *Target, req *QueryMetricsRequest) (model.Value, v1.Warnings, error) {
	var varMap = make(map[string]string)

	var varSpecs []Variable
	for _, variable := range target.Variables {
		if v, find := req.Params[variable]; find {
			varMap[variable] = v
			continue
		}

		varSpec, find := queryDict.GetVarSpec(groupId, variable)
		if !find {
			varMap[variable] = ""
			continue
		}
		varSpecs = append(varSpecs, *varSpec)
	}

	var retry = 10
	for {
		if retry <= 0 {
			break
		}

		var unknownSpec = make([]Variable, 0)
		for i := 0; i < len(varSpecs); i++ {
			variable, find, dep := s.queryVar(&varSpecs[i], req.StartTime, req.EndTime)
			if find {
				varMap[varSpecs[i].Name] = variable
				continue
			}

			for _, v := range dep {
				if value, find := varMap[v]; find {
					varSpecs[i].Query.Query = strings.ReplaceAll(varSpecs[i].Query.Query, "$"+v, prometheus.EscapeRegexp(value))
				} else if value, find := req.Params[v]; find {
					varSpecs[i].Query.Query = strings.ReplaceAll(varSpecs[i].Query.Query, "$"+v, prometheus.EscapeRegexp(value))
				}
			}
			unknownSpec = append(unknownSpec, varSpecs[i])
		}

		if len(unknownSpec) == 0 {
			break
		}
		varSpecs = unknownSpec

		retry--
	}

	var expr string = target.Expr
	for k, v := range varMap {
		expr = strings.ReplaceAll(expr, "$"+k, v)
	}

	return s.promRepo.GetApi().QueryRange(context.Background(), expr, v1.Range{
		Start: time.UnixMicro(req.StartTime),
		End:   time.UnixMicro(req.EndTime),
		Step:  time.Microsecond * time.Duration(req.Step),
	})
}

var labelValuesQry = regexp.MustCompile(`label_values\(([^,]+),([^)]+)\)`)
var variableQry = regexp.MustCompile(`\$([a-zA-Z0-9_]+)`)

func (s *service) queryVar(
	varSpec *Variable,
	startTime, endTime int64,
) (res string, find bool, dep []string) {
	if varSpec.Type == "custom" {
		return strings.Join(varSpec.Current.Value, "|"), true, nil
	}

	if varSpec.Type != "query" {
		return "", true, []string{}
	}

	matches := variableQry.FindAllStringSubmatch(varSpec.Query.Query, -1)
	if len(matches) > 0 {
		var deps = make([]string, 0)
		for _, match := range matches {
			deps = append(deps, match[1])
		}
		return "", false, deps
	}

	switch varSpec.Query.QryType {
	case 1: // label_values
		matches := labelValuesQry.FindStringSubmatch(varSpec.Query.Query)
		if len(matches) != 3 {
			log.Println("unexpected var query spec:", varSpec.Query.Query)
			return "", true, nil
		}

		labelValues, err := s.promRepo.LabelValues(matches[1], matches[2], startTime, endTime)
		if err != nil {
			log.Println("query result err:", err)
			return "", true, nil
		}

		var labels []string
		for _, v := range labelValues {
			labels = append(labels, prometheus.EscapeRegexp(string(v)))
		}
		return strings.Join(labels, "|"), true, nil
	case 3: // query_result
		expr, _ := strings.CutPrefix(varSpec.Query.Query, "query_result(")
		expr, _ = strings.CutSuffix(expr, ")")
		labels, err := s.promRepo.QueryResult(expr, varSpec.Regex, startTime, endTime)
		if err != nil {
			log.Println("query result err:", err)
			return "", true, nil
		}
		for i := 0; i < len(labels); i++ {
			labels[i] = prometheus.EscapeRegexp(labels[i])
		}

		return strings.Join(labels, "|"), true, nil
	default:
		return "", false, nil
	}
}
