// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (ch *chRepo) AddWorkflowRecord(ctx core.Context, record *model.WorkflowRecord) error {
	batch, err := ch.GetContextDB(ctx).PrepareBatch(ctx.GetContext(), `
		INSERT INTO workflow_records (workflow_run_id, workflow_id, workflow_name, ref, input, output, created_at, rounded_time)
		VALUES
	`)
	if err != nil {
		return err
	}
	if err := batch.Append(
		record.WorkflowRunID,
		record.WorkflowID,
		record.WorkflowName,
		record.Ref,
		record.Input,
		record.Output,
		time.UnixMicro(record.CreatedAt),
		time.UnixMicro(record.RoundedTime),
	); err != nil {
		return err
	}
	if err := batch.Send(); err != nil {
		return err
	}
	return nil
}

func (ch *chRepo) AddWorkflowRecords(ctx core.Context, records []model.WorkflowRecord) error {
	batch, err := ch.GetContextDB(ctx).PrepareBatch(ctx.GetContext(), `
		INSERT INTO workflow_records (workflow_run_id, workflow_id, workflow_name, ref, input, output, created_at, rounded_time)
		VALUES
	`)
	if err != nil {
		return err
	}
	for _, record := range records {
		if err := batch.Append(
			record.WorkflowRunID,
			record.WorkflowID,
			record.WorkflowName,
			record.Ref,
			record.Input,
			record.Output,
			time.UnixMicro(record.CreatedAt),
			time.UnixMicro(record.RoundedTime),
		); err != nil {
			continue
		}
	}

	if err := batch.Send(); err != nil {
		return err
	}
	return nil
}
