package clickhouse

import (
	"context"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"time"
)

func (ch *chRepo) GetDetectExecList(req *request.GetDefectDetectExecListRequest) ([]model.DetectMutation, int64, error) {
	sql := `SELECT 
    		name,
    		mutation_check_pql,
    		for_duration,
    		step,
    		toUnixTimestamp64Micro(start_time) AS start_time,
    		toUnixTimestamp64Micro(end_time) AS end_time,
    		synchronize_to_alert_rules,
    		group,
    		toUnixTimestamp64Micro(timestamp) AS timestamp
			FROM detect_list 
			ORDER BY timestamp DESC 
			LIMIT ? OFFSET ?`
	resp := []model.DetectMutation{}
	countSql := `SELECT count(*) as total FROM detect_list`
	var count []QueryCount
	err := ch.conn.Select(context.Background(), &count, countSql, nil)
	if err != nil {
		return nil, 0, err
	}
	err = ch.conn.Select(context.Background(), &resp, sql, req.PageSize, (req.CurrentPage-1)*req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	return resp, int64(count[0].Total), nil
}

func (ch *chRepo) AddDetectMutation(mutation model.DetectMutation) error {
	sql := `
		INSERT INTO 
		detect_list(
		            name, 
		            mutation_check_pql, 
		            for_duration, step, 
		            start_time, end_time, 
		            timestamp,
		            group)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	err := ch.conn.Exec(context.Background(), sql,
		mutation.DetectName,
		mutation.MutationCheck,
		mutation.For, mutation.Step,
		time.UnixMicro(mutation.StartTime),
		time.UnixMicro(mutation.EndTime),
		time.UnixMicro(mutation.Timestamp),
		mutation.Group,
	)

	return err
}

func (ch *chRepo) GetDistinctDetectExecList(req *request.GetDefectDetectRuleListRequest) ([]model.DetectMutation, int64, error) {
	sql := `SELECT 
    	name,
    	mutation_check_pql,
    	for_duration,
    	step,
    	group,
		FROM detect_list
		GROUP BY 
    	name, 
    	mutation_check_pql, 
    	for_duration, 
    	step, 
    	group
		ORDER BY 
    	MAX(timestamp) DESC
		LIMIT ? OFFSET ?`

	var mutations []model.DetectMutation
	err := ch.conn.Select(context.Background(), &mutations, sql, req.PageSize, (req.CurrentPage-1)*req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	countSql := `SELECT COUNT(DISTINCT name, mutation_check_pql, for_duration, step, group) as total FROM detect_list`
	var count []QueryCount
	err = ch.conn.Select(context.Background(), &count, countSql, nil)
	if err != nil {
		return nil, 0, err
	}

	return mutations, int64(count[0].Total), nil
}
