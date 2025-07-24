// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"fmt"
	"strings"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

// GetFaultLogContent implements Service.
func (s *service) GetFaultLogContent(ctx core.Context, req *request.GetFaultLogContentRequest) (*response.GetFaultLogContentResponse, error) {
	logContest, sources, err := s.chRepo.QueryApplicationLogs(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(logContest.Contents) == 0 {
		result := s.trySearchInFullLog(req, ctx)
		if result != nil {
			return result, nil
		}
	}

	return &response.GetFaultLogContentResponse{
		Sources:     sources,
		LogContents: logContest,
	}, nil
}

func (s *service) trySearchInFullLog(req *request.GetFaultLogContentRequest, ctx core.Context) *response.GetFaultLogContentResponse {
	var query []string
	if len(req.ContainerId) > 0 {
		query = append(query, fmt.Sprintf("container_id = '%s'", req.ContainerId))
	} else if len(req.NodeName) > 0 {
		query = append(query, fmt.Sprintf("host_name = '%s'", req.NodeName))
		query = append(query, fmt.Sprintf("pid = '%d'", req.Pid))
	}

	if len(req.SourceFrom) > 0 {
		query = append(query, fmt.Sprintf("source = '%s'", req.SourceFrom))
	}

	if len(query) == 0 {
		query = append(query, "(1='1')")
	}

	logs, _, err := s.chRepo.QueryAllLogsInOrder(ctx, &request.LogQueryRequest{
		StartTime:  int64(req.StartTime),
		EndTime:    int64(req.EndTime),
		TableName:  "raw_logs",
		DataBase:   "apo",
		PageNum:    1,
		PageSize:   2000,
		Query:      strings.Join(query, " AND "),
		TimeField:  "timestamp",
		LogField:   "content",
		IsExternal: false,
	})

	if len(logs) == 0 || err != nil {
		return nil
	}

	sourceFrom := map[string]struct{}{}
	logContents := &clickhouse.Logs{
		Source:   req.SourceFrom,
		Contents: []clickhouse.LogContent{},
	}
	for _, row := range logs {
		if contentPtr, find := row["content"]; find {
			content := contentPtr.(string)
			timestamp := row["timestamp"].(time.Time)

			logContents.Contents = append(logContents.Contents, clickhouse.LogContent{
				Timestamp: timestamp.UnixMicro(),
				Body:      content,
			})

			if source, ok := row["source"].(string); ok {
				sourceFrom[source] = struct{}{}
			}
		}
	}

	if len(logContents.Contents) == 0 {
		return nil
	}

	var sources = make([]string, 0)
	for source := range sourceFrom {
		sources = append(sources, source)
	}
	if len(sources) == 0 {
		sources = append(sources, "stdout")
	}

	return &response.GetFaultLogContentResponse{
		Sources:     sources,
		LogContents: logContents,
	}
}
