// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/hashicorp/golang-lru/v2/expirable"
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
			record = w.createExpiredRecord(c, event, endTime)
		} else {
			runner.Add(1)
			record = w.doAlertCheck(c, event, endTime)
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

var cache = expirable.NewLRU[string, model.AlertEventClassify](10, nil, time.Hour)

func (w *worker) getAlertClassify(c *DifyClient, event *alert.AlertEvent) model.AlertEventClassify {
	inputs, _ := json.Marshal(map[string]interface{}{
		"alertGroup": event.Group,
		"alertName":  event.Name,
	})
	r, ok := cache.Get(event.Group + event.Name)
	if ok {
		return r
	}

	request := &WorkflowRequest{
		Inputs:       inputs,
		ResponseMode: "blocking",
		User:         "apo-backend",
	}

	difyconf := config.Get().Dify
	resp, err := c.WorkflowsRun(request, "Bearer "+difyconf.APIKeys.AlertClassify)

	classify := model.AlertEventClassify{
		WorkflowId:     w.FlowId,
		WorkflowApiKey: w.AnalyzeAuth,
	}
	if err != nil {
		w.logger.Error("failed to alert event classify", zap.Error(err))
		return classify
	}
	completResp, ok := resp.(*CompletionResponse)
	if !ok {
		w.logger.Error("failed to alert event classify", zap.Error(err))
		return classify
	}

	var res map[string]string
	err = json.Unmarshal(completResp.Data.Outputs, &res)
	if err != nil {
		w.logger.Error("failed to get alert event classify api", zap.Error(err))
		return classify
	}
	classify.WorkflowApiKey = res["workflowApiKey"]
	classify.WorkflowId = res["workflowId"]
	cache.Add(event.Group+event.Name, classify)
	return classify
}

func (w *worker) createExpiredRecord(c *DifyClient, event *alert.AlertEvent, endTime int64) model.WorkflowRecord {
	w.logger.Error("alert event is expired, skip alert check",
		zap.String("event_id", event.EventID),
		zap.Int64("expired_ts", w.expiredTS),
		zap.Int64("event_ts", endTime),
	)

	tw := time.Duration(w.CacheMinutes) * time.Minute
	roundedTime := event.UpdateTime.Truncate(tw).Add(tw)
	classify := w.getAlertClassify(c, event)

	return model.WorkflowRecord{
		WorkflowRunID: "",
		WorkflowID:    classify.WorkflowId,
		WorkflowName:  w.FlowName,
		Ref:           event.AlertID,
		Input:         event.EventID,
		Output:        "failed: alert check too late, could be too many event to check or last check cost too much time, skipped",
		CreatedAt:     roundedTime.UnixMicro(),
		RoundedTime:   roundedTime.UnixMicro(),

		InputRef: event,
	}
}

func (w *worker) doAlertCheck(c *DifyClient, event *alert.AlertEvent, endTime int64) model.WorkflowRecord {
	startTime := event.UpdateTime.Add(-15 * time.Minute).UnixMicro()
	inputs, _ := json.Marshal(map[string]interface{}{
		"alert":     event.Name,
		"params":    event.TagsInStr(),
		"startTime": startTime,
		"endTime":   endTime,
		"edition":   "ce",
	})
	classify := w.getAlertClassify(c, event)
	resp, err := c.alertCheck(&WorkflowRequest{Inputs: inputs}, w.Authorization, w.User)
	if err != nil || resp == nil {
		w.logger.Error("failed to to alert check", zap.Error(err))
		tw := time.Duration(w.CacheMinutes) * time.Minute
		roundedTime := event.UpdateTime.Truncate(tw).Add(tw)
		return model.WorkflowRecord{
			WorkflowRunID: "",
			WorkflowID:    classify.WorkflowId,
			WorkflowName:  w.FlowName,
			Ref:           event.AlertID,
			Input:         event.EventID,
			Output:        "failed: workflow execution failed due to API call failure",
			CreatedAt:     roundedTime.UnixMicro(),
			RoundedTime:   roundedTime.UnixMicro(),

			InputRef: event,
		}
	}

	tw := time.Duration(w.CacheMinutes) * time.Minute
	roundedTime := event.UpdateTime.Truncate(tw).Add(tw)

	var record model.WorkflowRecord
	output := resp.getOutput("failed: not find expected output")
	record = model.WorkflowRecord{
		WorkflowRunID: resp.WorkflowRunID(),
		WorkflowID:    classify.WorkflowId,
		WorkflowName:  w.FlowName,
		Ref:           event.AlertID,
		Input:         event.EventID,
		Output:        output,
		CreatedAt:     resp.CreatedAt(),
		RoundedTime:   roundedTime.UnixMicro(),

		InputRef: event,
	}
	difyconf := config.Get().Dify
	if difyconf.AutoAnalyze && output == "false" {
		param := w.getWorkflowParams(event)
		if param == nil {
			// unexpected err
			record.AnalyzeErr = "failed to get analyze workflow params"
			record.AlertDirection = ""
		}
		inputStr, err := json.Marshal(param)
		if err != nil {
			w.logger.Info("failed to marshal workflow params", zap.Error(err))
			record.AnalyzeErr = err.Error()
			record.AlertDirection = ""
		} else {
			resp, err := c.alertAnalyze(&WorkflowRequest{Inputs: inputStr}, "Bearer "+classify.WorkflowApiKey, w.User)
			if err != nil {
				record.AnalyzeErr = err.Error()
				record.AlertDirection = ""
			} else {
				record.AnalyzeRunID = resp.WorkflowRunID()
				record.AlertDirection = resp.getOutput("")
			}
		}
	}
	return record
}

func (w *worker) getWorkflowParams(event *alert.AlertEvent) *alert.WorkflowParams {
	var startTime, endTime time.Time
	if event.Status == alert.StatusResolved {
		startTime = event.EndTime.Add(-15 * time.Minute)
		endTime = event.EndTime
	} else {
		startTime = event.UpdateTime.Add(-15 * time.Minute)
		endTime = event.UpdateTime
	}

	alertServices, _ := tryGetAlertService(core.EmptyCtx(), w.Prom, event, startTime, endTime)

	res := alert.WorkflowParams{
		StartTime: startTime.UnixMicro(),
		EndTime:   endTime.UnixMicro(),
		NodeName:  event.GetInfraNodeTag(),
		Edition:   "ce",
	}

	var services, endpoints []string
	for _, alertService := range alertServices {
		services = append(services, alertService.Service)
		if len(alertService.Endpoint) == 0 {
			endpoints = append(endpoints, ".*")
		} else {
			endpoints = append(endpoints, alertService.Endpoint)
		}
	}

	rawTags := event.Tags.Clone()
	rawTags["alertEventId"] = event.EventID
	parmas := alert.AlertAnalyzeWorkflowParams{
		AlertName:    event.Name,
		Node:         event.GetInfraNodeTag(),
		Namespace:    event.GetK8sNamespaceTag(),
		Pod:          event.GetK8sPodTag(),
		Pid:          event.GetPidTag(),
		Detail:       event.Detail,
		ContainerID:  event.GetContainerIDTag(),
		Tags:         event.EnrichTags,
		RawTags:      rawTags,
		AlertEventId: event.EventID,
	}

	if len(services) == 1 {
		parmas.Service = services[0]
		parmas.Endpoint = endpoints[0]
	}

	jsonStr, err := json.Marshal(parmas)
	if err != nil {
		res.Params = "{}"
	} else {
		res.Params = string(jsonStr)
	}

	return &res
}
