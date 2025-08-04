// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"go.uber.org/zap"
)

func (r *difyRepo) PrepareAsyncAlertCheckWorkflow(prom prometheus.Repo, logger *zap.Logger) (records <-chan *WorkflowRecordWithCtx, err error) {
	r.AlertCheckCFG.Prom = prom
	r.asyncAlertCheck = newAsyncAlertCheck(r.AlertCheckCFG, logger)
	records, err = r.asyncAlertCheck.Run(context.Background(), r.cli)
	if err != nil {
		r.asyncAlertCheck = nil
	}
	return records, err
}

func (r *difyRepo) SubmitAlertEvents(ctx core.Context, events []alert.AlertEvent) error {
	if r.asyncAlertCheck == nil {
		return nil
	}
	r.asyncAlertCheck.AddEvents(ctx, events)
	return nil
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
