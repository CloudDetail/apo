// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package workflow

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"go.uber.org/zap"
)

type AlertWorkflow struct {
	chRepo clickhouse.Repo
	client *DifyClient
	logger *zap.Logger

	AnalyzeFlowId string
	AlertCheck    *AlertCheckCfg

	check alertCheck
}

type Option func(f *AlertWorkflow)

func New(chRepo clickhouse.Repo, client *DifyClient, logger *zap.Logger, opts ...Option) *AlertWorkflow {
	workflow := &AlertWorkflow{
		chRepo: chRepo,
		client: client,
		logger: logger,
	}

	for _, opt := range opts {
		opt(workflow)
	}
	return workflow
}

func WithAlertCheckFlow(cfg *AlertCheckCfg) Option {
	return func(f *AlertWorkflow) {
		f.AlertCheck = cfg

		if cfg.MaxConcurrency <= 0 {
			cfg.MaxConcurrency = 1
		}

		cfg.CacheMinutes = maxFactorOf60LessThanN(cfg.CacheMinutes)
		cfg.FlowName = "AlertCheck"

		if !cfg.HasValidAPIKey() {
			f.logger.Info("can not find a valid workflow APIKey")
			return
		}

		check := newAlertCheck(cfg, f.logger)
		records, err := check.Run(context.Background(), f.client)
		if err != nil {
			f.logger.Error("failed to init alertWorkflow", zap.Error(err))
			return
		}
		f.logger.Info("start to process alert check",
			zap.Int("cacheMinutes", cfg.CacheMinutes),
			zap.Int("maxConcurrency", cfg.MaxConcurrency),
		)
		go f.SaveRecords(context.Background(), records)
		f.check = check
	}
}

func WithAlertAnalyzeFlow(analyzeFlowId string) Option {
	return func(f *AlertWorkflow) {
		f.AnalyzeFlowId = analyzeFlowId
	}
}

func (c *AlertWorkflow) SaveRecords(ctx context.Context, records <-chan model.WorkflowRecord) {
	for record := range records {
		err := c.chRepo.AddWorkflowRecords(ctx, []model.WorkflowRecord{record})
		if err != nil {
			c.logger.Error("store workflow records failed", zap.Error(err))
		}
	}
}

func (c *AlertWorkflow) AddAlertEvents(events []alert.AlertEvent) {
	if c.check == nil {
		return // AlertCheck flow is not start correctly
	}
	c.check.AddEvents(events)
}

func maxFactorOf60LessThanN(n int) int {
	if n <= 0 {
		return 5
	}

	factors := []int{5, 10, 15, 20, 30}
	maxFactor := 0
	for _, f := range factors {
		if f <= n {
			maxFactor = f
		} else {
			break
		}
	}
	return maxFactor
}
