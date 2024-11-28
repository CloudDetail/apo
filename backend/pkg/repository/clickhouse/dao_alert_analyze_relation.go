// File: dao_alert_analyzer_realtion.go
// Package: clickhouse
// Description: 包含了一些告警分析中使用到的结构和方法
package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

const (
	// 查询指定节点的入口信息,如果链路上存在异步请求,则返回异步入口的信息
	SQL_GET_ENTRY_RELATIONSHIP = `WITH
	  direct_relation AS (
	    SELECT trace_id,path,service,url,entry_service,entry_url
	    FROM %s.service_relationship
	    WHERE %s
	  ),
	  async_path AS (
	    SELECT path,service,url,trace_id
	    FROM service_relationship
	    GLOBAL JOIN (select DISTINCT trace_id from direct_relation) t ON service_relationship.trace_id = t.trace_id
	    WHERE flags['is_async'] = TRUE
	  )
	SELECT service,url,entry_service,entry_url,async_path.service as async_entry_service,async_path.url as async_entry_endpoint FROM direct_relation
	LEFT JOIN async_path ON direct_relation.trace_id = async_path.trace_id
	WHERE startsWith(path,async_path.path)
	GROUP BY service,url,entry_service,entry_url,async_entry_service,async_entry_endpoint`
)

// EntryRelationship 表示下游节点和其入口的直接对应关系
type EntryRelationship struct {
	Service  string `ch:"service"`
	Endpoint string `ch:"url"`

	EntryService  string `ch:"entry_service"`
	EntryEndpoint string `ch:"entry_url"`

	AsyncEntryService  string `ch:"async_entry_service"`
	AsyncEntryEndpoint string `ch:"async_entry_endpoint"`
}

// 和ServiceNode结构完全相同
// 作为查询条件时,允许Endpoint为空
// 表示对应Service下的所有Endpoint
type AlertService = ServiceNode

// SearchEntryEndpointsByAlertService 根据下游节点查询入口节点
// 链路中存在异步节点时,不保留真正的入口节点
func (ch *chRepo) SearchEntryEndpointsByAlertService(alertServices []AlertService, startTime, endTime int64) ([]EntryRelationship, error) {
	// microseconds -> seconds
	startTime = startTime / 1000000
	endTime = endTime / 1000000

	// services中可能包含两类数据,contentKey为空时,表示忽略contentKey
	var endpoints = ValueInGroups{
		Keys: []string{"service", "url"},
	}
	var services = ValueInGroups{
		Keys: []string{"service"},
	}

	for _, endpoint := range alertServices {
		if len(endpoint.Endpoint) > 0 {
			endpoints.ValueGroups = append(endpoints.ValueGroups, clickhouse.GroupSet{
				Value: []any{endpoint.Service, endpoint.Endpoint},
			})
		} else {
			services.ValueGroups = append(services.ValueGroups, clickhouse.GroupSet{
				Value: []any{endpoint.Service},
			})
		}
	}

	queryBuilder := NewQueryBuilder().
		Between("timestamp", startTime, endTime).
		Equals("miss_top", false).
		And(MergeWheres(OrSep, InGroup(endpoints), InGroup(services)))

	parentTopologys := []EntryRelationship{}
	sql := fmt.Sprintf(SQL_GET_ENTRY_RELATIONSHIP,
		ch.database,
		queryBuilder.String(),
	)
	if err := ch.conn.Select(context.Background(), &parentTopologys, sql, queryBuilder.values...); err != nil {
		return nil, err
	}

	return parentTopologys, nil
}
