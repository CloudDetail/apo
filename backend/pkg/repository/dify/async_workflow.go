// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"go.uber.org/zap"
)

func (r *difyRepo) PrepareAsyncAlertCheckWorkflow(cfg *AlertCheckConfig, logger *zap.Logger) (records <-chan *model.WorkflowRecord, err error) {
	if len(cfg.Sampling) == 0 {
		cfg.Sampling = "first"
	}
	if len(cfg.FlowName) == 0 {
		cfg.FlowName = "AlertCheck"
	}
	if cfg.CacheMinutes <= 0 {
		cfg.CacheMinutes = 20
	} else {
		cfg.CacheMinutes = maxFactorOf60LessThanN(cfg.CacheMinutes)
	}
	if cfg.MaxConcurrency <= 0 {
		cfg.MaxConcurrency = 1
	}
	r.AlertCheckCFG = *cfg
	r.asyncAlertCheck = newAsyncAlertCheck(&r.AlertCheckCFG, logger)
	records, err = r.asyncAlertCheck.Run(context.Background(), r.cli)
	if err != nil {
		r.asyncAlertCheck = nil
	}
	return records, err
}

func (r *difyRepo) SubmitAlertEvents(events []alert.AlertEvent) {
	if r.asyncAlertCheck == nil {
		return
	}
	r.asyncAlertCheck.AddEvents(events)
}

func DefaultAlertCheckConfig() AlertCheckConfig {
	return AlertCheckConfig{
		Sampling:       "first",
		CacheMinutes:   20,
		MaxConcurrency: 1,
	}
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
