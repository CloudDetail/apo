package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	startTimeLayout = "2006-01-02 15:04:05 -0700 MST"

	// SQL_GET_K8S_EVENTS 获取K8s事件告警
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

// K8sAlert   查询K8S告警
func (ch *chRepo) GetK8sAlertEventsSample(startTime time.Time, endTime time.Time, instances []*model.ServiceInstance) ([]K8sEvents, error) {
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

	// 每个ObjectKind取一个事件
	query := fmt.Sprintf(SQL_GET_K8S_EVENTS, builder.String(), 1)
	// 执行查询
	var res []K8sEvents
	err := ch.conn.Select(context.Background(), &res, query, builder.values...)
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
