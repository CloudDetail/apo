// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"context"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"go.uber.org/zap"
)

type Handle func(ctx context.Context, record *model.WorkflowRecord) error

func HandleRecords(ctx context.Context, logger *zap.Logger, records <-chan *model.WorkflowRecord, handlers ...Handle) {
	for record := range records {
		for _, handler := range handlers {
			err := handleRecord(ctx, handler, record)
			if err != nil {
				logger.Error("handle workflow records failed", zap.Error(err))
			}
		}
	}
}

func handleRecord(ctx context.Context, handler Handle, record *model.WorkflowRecord) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	return handler(ctx, record)
}
