// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"go.uber.org/zap"
)

type Handle func(ctx core.Context, record *model.WorkflowRecord) error

func HandleRecords(logger *zap.Logger, records <-chan *WorkflowRecordWithCtx, handlers ...Handle) {
	for record := range records {
		for _, handler := range handlers {
			err := handler(record.ctx, record.WorkflowRecord)
			if err != nil {
				logger.Error("handle workflow records failed", zap.Error(err))
			}
		}
	}
}
