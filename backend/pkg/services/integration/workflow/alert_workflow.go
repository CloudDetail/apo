// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

const MAX_CACHE_SIZE = 100

// 记录5min内产生的不同告警的AlertId
// 5min结束时,查询截止时间点,每个AlertId最新的Event 跑flow
// 此外还有一个立刻触发的API,主动清空现有的Map,运行workflow并等待结果
type AlertWorkflow struct {
	// AlertId -> AlertEvent
	PrepareToRun map[string]alert.AlertEvent
	AddMutex     sync.Mutex

	chRepo clickhouse.Repo
	*DifyClient

	difyAPIKey    string
	authorization string // f'Bearer {difyAPIKey}'
	user          string

	EventAnalyzeFlowId string
	CheckId            string

	logger *zap.Logger
}

func New(chRepo clickhouse.Repo, client *DifyClient, apiKey string, user string, logger *zap.Logger) *AlertWorkflow {
	return &AlertWorkflow{
		chRepo: chRepo,

		PrepareToRun:  make(map[string]alert.AlertEvent),
		DifyClient:    client,
		difyAPIKey:    apiKey,
		authorization: fmt.Sprintf("Bearer %s", apiKey),
		user:          user,
		logger:        logger,
	}
}

func (c *AlertWorkflow) Run(ctx context.Context) error {
	ctab := cron.New()
	_, err := ctab.AddFunc("*/5 * * * *", func() {
		err := c.Submit()
		if err != nil {
			// TODO deal with err
			panic(err)
		}
	})
	if err != nil {
		return err
	}
	ctab.Start()
	return nil
}

func (c *AlertWorkflow) AddAlertEvent(event *alert.AlertEvent) {
	if event.Status == alert.StatusResolved {
		return
	}

	c.AddMutex.Lock()
	defer c.AddMutex.Unlock()

	if len(c.PrepareToRun) > MAX_CACHE_SIZE {
		return
	}
	c.PrepareToRun[event.AlertID] = *event
}

func (c *AlertWorkflow) AddAlertEvents(events []alert.AlertEvent) {
	c.AddMutex.Lock()
	defer c.AddMutex.Unlock()

	if len(c.PrepareToRun) > MAX_CACHE_SIZE {
		return
	}

	for _, event := range events {
		if event.Status == alert.StatusResolved {
			continue
		}
		c.PrepareToRun[event.AlertID] = event
	}
}

type eventCheckResult struct {
	event  *alert.AlertEvent
	record model.WorkflowRecord
}

func (c *AlertWorkflow) worker(
	events <-chan alert.AlertEvent,
	results chan<- model.WorkflowRecord,
	startTime int64, endTime int64,
	wg *sync.WaitGroup) {
	defer wg.Done()
	for event := range events { // 从通道获取任务
		inputs, _ := json.Marshal(map[string]interface{}{
			"alert":     event.Name,
			"params":    event.TagsInStr(),
			"startTime": startTime,
			"endTime":   endTime,
		})

		resp, err := c.alertCheck(&DifyRequest{Inputs: inputs}, c.authorization, c.user)

		if err != nil {
			// TODO deal with error
			c.logger.Error("failed to to alert check", zap.Error(err))
			// panic(err)
			continue
		}

		record := model.WorkflowRecord{
			WorkflowRunID: resp.WorkflowRunID(),
			// WorkflowID:    "",
			WorkflowName: "告警事件分析",
			Ref:          event.AlertID,
			Input:        "",
			Output:       resp.IsValidOrDefault("false"), // 'false' means valid alert
			CreatedAt:    resp.CreatedAt(),
		}

		results <- record
	}
}

func (c *AlertWorkflow) Submit() error {
	c.AddMutex.Lock()
	submitEvents := c.PrepareToRun
	c.PrepareToRun = make(map[string]alert.AlertEvent)
	c.AddMutex.Unlock()

	submitEventList := make([]alert.AlertEvent, 0, len(submitEvents))
	for _, event := range submitEvents {
		submitEventList = append(submitEventList, event)
	}

	c.logger.Info("start to submit alert need to analyze", zap.Int("size", len(submitEventList)))

	if len(submitEventList) == 0 {
		return nil
	}

	var records = make([]model.WorkflowRecord, 0, len(submitEvents))

	var wg sync.WaitGroup
	var results = make(chan model.WorkflowRecord)

	// 每次提交最多跑4min, 其他数据暂不计算
	endTime := time.Now()
	startTime := endTime.Add(-15 * time.Minute).UnixMicro()
	var events = make(chan alert.AlertEvent)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go c.worker(events, results, startTime, endTime.UnixMicro(), &wg)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	go func() {
		for i := 0; i < len(submitEventList); i++ {
			select {
			case <-ctx.Done():
				close(events)
				return
			case events <- submitEventList[i]:
			}
		}
		close(events)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		records = append(records, result)
	}

	if len(records) == 0 {
		return nil
	}

	executeTime := time.Now().Unix() - endTime.Unix()
	c.logger.Info("alert check finished", zap.Int("record size", len(records)), zap.Int("cost(s)", int(executeTime)))

	return c.chRepo.AddWorkflowRecords(context.Background(), records)
}

type AlertCheckRespose struct {
	resp *CompletionResponse
}

func (r *AlertCheckRespose) WorkflowRunID() string {
	return r.resp.WorkflowRunID
}

// UnixMicro Timestamp
func (r *AlertCheckRespose) CreatedAt() int64 {
	return r.resp.Data.CreatedAt * 1e6
}

func (r *AlertCheckRespose) IsValidOrDefault(defaultV string) string {
	var res map[string]string
	err := json.Unmarshal(r.resp.Data.Outputs, &res)
	if err != nil {
		return defaultV
	}

	text, find := res["text"]
	if !find {
		return defaultV
	}

	return text
}
