package storage

import (
	"context"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"go.uber.org/zap"
)

func SaveRecords(ctx context.Context, ch clickhouse.Repo, logger *zap.Logger, records <-chan model.WorkflowRecord) {
	for record := range records {
		err := ch.AddWorkflowRecords(ctx, []model.WorkflowRecord{record})
		if err != nil {
			logger.Error("store workflow records failed", zap.Error(err))
		}
	}
}
