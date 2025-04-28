package dify

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"go.uber.org/zap"
)

type Handle func(ctx context.Context, record *model.WorkflowRecord) error

func HandleRecords(ctx context.Context, logger *zap.Logger, records <-chan *model.WorkflowRecord, handlers ...Handle) {
	for record := range records {
		// TODO Async Handle With Timeout
		for _, handler := range handlers {
			err := handler(ctx, record)
			if err != nil {
				logger.Error("handle workflow records failed", zap.Error(err))
			}
		}
	}
}
