// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"strconv"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

const (
	SQL_GET_APP_LOG_SOURCE = `SELECT LogAttributes['_source_'] as LogSource
		FROM ilogtail_logs
		%s %s`
	SQL_GET_APP_LOG = `SELECT toUnixTimestamp64Micro(Timestamp) as ts,Body
		FROM ilogtail_logs
		%s %s`
)

func (ch *chRepo) QueryApplicationLogs(ctx core.Context, req *request.GetFaultLogContentRequest) (*Logs, []string, error) {

	builder := NewQueryBuilder().
		Between("Timestamp", int64(req.StartTime/1000000), int64(req.EndTime/1000000))

	if len(req.ContainerId) > 0 {
		builder.Equals("LogAttributes['_container_id_']", req.ContainerId)
	} else if req.Pid > 0 {
		builder.Equals("LogAttributes['pid']", strconv.FormatUint(uint64(req.Pid), 10))
		builder.Equals("LogAttributes['k8s.node.name']", req.NodeName)
	} else {
		builder.Equals("LogAttributes['k8s.pod.name']", req.PodName)
	}

	var sources []string
	sources, err := ch.queryApplicationLogsSource(ctx, builder)
	if err != nil {
		return nil, nil, err
	}
	if len(sources) == 0 {
		return &Logs{}, []string{}, nil
	}
	if len(req.SourceFrom) == 0 {
		req.SourceFrom = sources[0]
	}

	builder.Equals("LogAttributes['_source_']", req.SourceFrom)

	byBuilder := NewByLimitBuilder().
		Limit(2000).
		OrderBy("ts", true).
		OrderBy("LogAttributes['log_seq']", true)

	sql := fmt.Sprintf(SQL_GET_APP_LOG, builder.String(), byBuilder.String())
	var logRaws []LogContent
	err = ch.GetContextDB(ctx).Select(ctx.GetContext(), &logRaws, sql, builder.values...)
	return &Logs{req.SourceFrom, logRaws}, sources, err
}

func (ch *chRepo) QueryApplicationLogsAvailableSource(ctx core.Context, faultLog FaultLogResult) ([]string, error) {
	builder := NewQueryBuilder().
		Between("Timestamp", int64(faultLog.StartTime), int64(faultLog.EndTime))

	if len(faultLog.ContainerId) > 0 {
		builder.Equals("LogAttributes['_container_id_']", faultLog.ContainerId)
	}
	if len(faultLog.PodName) > 0 {
		builder.Equals("LogAttributes['k8s.pod.name']", faultLog.PodName)
	}
	if len(faultLog.NodeName) > 0 {
		builder.Equals("LogAttributes['k8s.node.name']", faultLog.NodeName)
	}
	if faultLog.Pid > 0 {
		builder.Equals("LogAttributes['pid']", strconv.FormatUint(uint64(faultLog.Pid), 10))
	}

	return ch.queryApplicationLogsSource(ctx, builder)
}

func (ch *chRepo) queryApplicationLogsSource(ctx core.Context, builder *QueryBuilder) ([]string, error) {
	byBuilder := NewByLimitBuilder().
		GroupBy("LogSource")

	sql := fmt.Sprintf(SQL_GET_APP_LOG_SOURCE, builder.String(), byBuilder.String())
	var sources []Source
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &sources, sql, builder.values...)
	if err != nil {
		return nil, err
	}

	var res []string
	for _, source := range sources {
		res = append(res, source.LogSource)
	}
	return res, err
}

type Logs struct {
	Source   string       `json:"source"`
	Contents []LogContent `json:"contents"`
}

type LogContent struct {
	Timestamp int64  `ch:"ts" json:"timestamp"`
	Body      string `ch:"Body" json:"body"`
}

type Source struct {
	LogSource string `ch:"LogSource"`
}
