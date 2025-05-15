// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"fmt"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	startTimeLayout = "2006-01-02 15:04:05 -0700 MST"

	// SQL _GET_K8S_EVENTS get K8s event alarm
	SQL_GET_K8S_EVENTS = `WITH grouped_events AS (
			SELECT Timestamp,SeverityText,Body,ResourceAttributes,LogAttributes,
				ROW_NUMBER() OVER (PARTITION BY ResourceAttributes['k8s.object.kind'] ORDER BY SeverityNumber) AS rn
			FROM k8s_events
			%s
		)
		SELECT Timestamp,SeverityText,Body,ResourceAttributes,LogAttributes
		FROM grouped_events
		WHERE rn <= %d`
)

// K8sAlert query K8S alarm
func (ch *chRepo) GetK8sAlertEventsSample(ctx core.Context, startTime time.Time, endTime time.Time, instances []*model.ServiceInstance) ([]K8sEvents, error) {
	relatedObj := make([]string, 0)
	for _, instance := range instances {
		if instance == nil {
			continue
		}

		if len(instance.PodName) > 0 {
			relatedObj = append(relatedObj, instance.PodName)
		}
		if len(instance.NodeName) > 0 {
			relatedObj = append(relatedObj, instance.NodeName)
		}
	}

	builder := NewQueryBuilder().
		Between("Timestamp", startTime.Unix(), endTime.Unix()).
		InStrings("ResourceAttributes['k8s.object.name']", relatedObj).
		InStrings("ResourceAttributes['k8s.object.kind']", []string{"Pod", "Node"}).
		GreaterThan("SeverityNumber", 9)

	// Take one event per ObjectKind
	query := fmt.Sprintf(SQL_GET_K8S_EVENTS, builder.String(), 1)
	// Execute query
	var res []K8sEvents
	err := ch.GetContextDB(ctx).Select(context.Background(), &res, query, builder.values...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

type K8sEvents struct {
	Timestamp          time.Time         `ch:"Timestamp" json:"timestamp"`
	SeverityText       string            `ch:"SeverityText" json:"SeverityText"`
	Body               string            `ch:"Body" json:"body"`
	ResourceAttributes map[string]string `ch:"ResourceAttributes" json:"resourceAttributes"`
	LogAttributes      map[string]string `ch:"LogAttributes" json:"logAttributes"`
}

func (e *K8sEvents) GetObjKind() string {
	if e != nil && e.ResourceAttributes != nil {
		return e.ResourceAttributes["k8s.object.kind"]
	}
	return "unknown"
}

func (e *K8sEvents) GetObjName() string {
	if e != nil && e.ResourceAttributes != nil {
		return e.ResourceAttributes["k8s.object.name"]
	}
	return "unknown"
}

func (e *K8sEvents) GetReason() string {
	if e != nil && e.ResourceAttributes != nil {
		return e.LogAttributes["k8s.event.reason"]
	}
	return "unknown"
}
