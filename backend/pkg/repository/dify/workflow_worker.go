// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
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
			output := resp.getOutput("failed: not find expected output")
			record = model.WorkflowRecord{
				WorkflowRunID: resp.WorkflowRunID(),
				WorkflowID:    w.FlowId,
				WorkflowName:  w.FlowName,
				Ref:           event.AlertID,
				Input:         event.ID.String(), // TODO record input param
				Output:        output,            // 'false' means valid alert
				CreatedAt:     resp.CreatedAt(),
				RoundedTime:   roundedTime.UnixMicro(),

				InputRef: event,
			}

			if output == "false" {
				param := w.getWorkflowParams(event)
				if param == nil {
					// unexpected err
					record.AnalyzeErr = "failed to get analyze workflow params"
					record.AlertDirection = "生成告警分析参数失败"
				}
				inputStr, err := json.Marshal(param)
				if err != nil {
					w.logger.Info("failed to marshal workflow params", zap.Error(err))
					record.AnalyzeErr = err.Error()
					record.AlertDirection = "序列化告警分析参数失败"
				} else {
					resp, err := c.alertAnalyze(&WorkflowRequest{Inputs: inputStr}, w.AnalyzeAuth, w.User)
					if err != nil {
						record.AnalyzeRunID = resp.WorkflowRunID()
						record.AnalyzeErr = err.Error()
						record.AlertDirection = "执行告警分析工作流失败"
					} else {
						record.AnalyzeRunID = resp.WorkflowRunID()
						record.AlertDirection = resp.getOutput("failed: not find expected output: alertDirection")
					}
				}
			}
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

	parmas := alert.AlertAnalyzeWorkflowParams{
		AlertName:   event.Name,
		Node:        event.GetInfraNodeTag(),
		Namespace:   event.GetK8sNamespaceTag(),
		Pod:         event.GetK8sPodTag(),
		Pid:         event.GetPidTag(),
		Detail:      event.Detail,
		ContainerID: event.GetContainerIDTag(),
		Tags:        event.EnrichTags,
		RawTags:     event.Tags,
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
