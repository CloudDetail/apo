// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

func (ch *chRepo) GetToResolveApps(ctx core.Context) ([]*model.AppInfo, error) {
	today := time.Now()
	startTime := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()).Unix()

	query := "SELECT * FROM originx_app_info where toUnixTimestamp(timestamp) >= ?"
	// Query list data
	apps := []model.AppInfo{}
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &apps, query, startTime)
	if err != nil {
		return nil, err
	}

	relatedApp := make(map[string]*model.AppInfo)
	for _, app := range apps {
		if app.Labels["service_name"] != "" {
			relatedApp[fmt.Sprintf("%s-%s-%s-%d-%d", app.Labels["node_ip"], app.Labels["node_name"], app.Labels["cluster_id"], app.StartTime, app.HostPid)] = &app
		}
	}

	result := make([]*model.AppInfo, 0)
	for _, app := range apps {
		if _, ok := relatedApp[fmt.Sprintf("%s-%s-%s-%d-%d", app.Labels["node_ip"], app.Labels["node_name"], app.Labels["cluster_id"], app.StartTime, app.HostPid)]; !ok {
			result = append(result, &app)
		}
	}

	return result, nil
}
