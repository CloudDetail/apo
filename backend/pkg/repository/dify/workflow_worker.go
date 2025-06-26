// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"go.uber.org/zap"
)

const MAX_CACHE_SIZE = 100

type inputChan struct {
	Ch         chan *alert.AlertEvent
	IsShutDown bool
}

func newInputChan() *inputChan {
	return &inputChan{
		Ch:         make(chan *alert.AlertEvent, MAX_CACHE_SIZE+10),
		IsShutDown: false,
	}
}

type worker struct {
	logger *zap.Logger
	*AlertCheckConfig

	expiredTS int64
}

func (w *worker) run(c *DifyClient, eventInput <-chan *alert.AlertEvent, results chan<- *model.WorkflowRecord, wg *sync.WaitGroup) {
	defer wg.Done()
	timeout := time.NewTimer(3 * time.Second)
	defer timeout.Stop()
	for event := range eventInput {
		endTime := event.UpdateTime.UnixMicro()
		var record model.WorkflowRecord
		if w.expiredTS > 0 && endTime < w.expiredTS {
			record = w.createExpiredRecord(event)
		} else {
			runner.Add(1)
			record = w.doAlertCheck(event, endTime, c)
			runner.Add(-1)
		}
		if !timeout.Stop() {
			select {
			case <-timeout.C:
			default:
			}
		}
		timeout.Reset(3 * time.Second)
		select {
		case <-timeout.C:
			w.logger.Error("too many record need to handler, drop")
			continue
		case results <- &record:
		}
	}
}

func (w *worker) createExpiredRecord(event *alert.AlertEvent) model.WorkflowRecord {
	w.logger.Debug("alert event is expired, skip alert check", zap.String("event_id", event.ID.String()))
	tw := time.Duration(w.CacheMinutes) * time.Minute
	roundedTime := event.UpdateTime.Truncate(tw).Add(tw)

	return model.WorkflowRecord{
		WorkflowRunID: "",
		WorkflowID:    w.FlowId,
		WorkflowName:  w.FlowName,
		Ref:           event.AlertID,
		Input:         event.ID.String(),
		Output:        "failed: alert check too late, could be too many event too check or last check cost too much time, skipped",
		CreatedAt:     roundedTime.UnixMicro(),
		RoundedTime:   roundedTime.UnixMicro(),

		InputRef: event,
	}
}

func (w *worker) doAlertCheck(event *alert.AlertEvent, endTime int64, c *DifyClient) model.WorkflowRecord {
	startTime := event.UpdateTime.Add(-15 * time.Minute).UnixMicro()
	inputs, _ := json.Marshal(map[string]interface{}{
		"alert":     event.Name,
		"params":    event.TagsInStr(),
		"startTime": startTime,
		"endTime":   endTime,
	})
	resp, err := c.alertCheck(&WorkflowRequest{Inputs: inputs}, w.Authorization, w.User)
	if err != nil {
		w.logger.Error("failed to to alert check", zap.Error(err))
	}

	tw := time.Duration(w.CacheMinutes) * time.Minute
	roundedTime := event.UpdateTime.Truncate(tw).Add(tw)

	if resp == nil {
		return model.WorkflowRecord{
			WorkflowRunID: "",
			WorkflowID:    w.FlowId,
			WorkflowName:  w.FlowName,
			Ref:           event.AlertID,
			Input:         event.ID.String(),
			Output:        fmt.Sprintf("failed: workflow execution failed due to API call failure: %s", err.Error()),
			CreatedAt:     roundedTime.UnixMicro(),
			RoundedTime:   roundedTime.UnixMicro(),
			InputRef:      event,
		}
	}

	return model.WorkflowRecord{
		WorkflowRunID: resp.WorkflowRunID(),
		WorkflowID:    w.FlowId,
		WorkflowName:  w.FlowName,
		Ref:           event.AlertID,
		Input:         event.ID.String(),
		Output:        resp.getOutput("failed: not find expected output"), // 'false' means valid alert
		CreatedAt:     resp.CreatedAt(),
		RoundedTime:   roundedTime.UnixMicro(),

		InputRef: event,
	}
}
