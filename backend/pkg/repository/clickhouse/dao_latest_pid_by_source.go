// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"strings"

	core "github.com/CloudDetail/apo/backend/pkg/core"
)

type NodePID struct {
	Node string
	PID  int64
}

func (ch *chRepo) GetLatestSourceByNodePIDs(ctx core.Context, pairs []NodePID) (map[NodePID]string, error) {
	result := make(map[NodePID]string)
	if len(pairs) == 0 {
		return result, nil
	}

	seen := make(map[NodePID]struct{}, len(pairs))
	var where []string
	args := make([]any, 0, len(pairs)*2)
	for _, pair := range pairs {
		if pair.Node == "" || pair.PID <= 0 {
			continue
		}
		if _, ok := seen[pair]; ok {
			continue
		}
		seen[pair] = struct{}{}
		where = append(where, "(host_name = ? AND latest_pid = ?)")
		args = append(args, pair.Node, pair.PID)
	}
	if len(where) == 0 {
		return result, nil
	}

	sql := fmt.Sprintf(`
SELECT
    source,
    host_name,
    toInt64(argMaxMerge(pid_state)) AS latest_pid
FROM %s.latest_pid_by_source_node
GROUP BY
    source,
    host_name
HAVING %s
`, ch.database, strings.Join(where, " OR "))

	rows, err := ch.GetContextDB(ctx).Query(ctx.GetContext(), sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var source string
		var node string
		var pid int64
		if err := rows.Scan(&source, &node, &pid); err != nil {
			return nil, err
		}
		result[NodePID{Node: node, PID: pid}] = source
	}
	return result, rows.Err()
}
