// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"testing"
	"time"
)

func TestSpanTraceP9x(t *testing.T) {
	svcs := []string{"ts-station-service", "ts-travel2-service"}
	endpoints := []string{
		"POST /api/v1/stationservice/stations/idlist",
		"POST /api/v1/travel2service/trips/left",
	}
	got := getSpanTraceP9xSql("vmrange", 5*time.Minute, svcs, endpoints)
	expect := "union(" +
		"histogram_quantile(0.9, sum by(vmrange, content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_bucket{svc_name='ts-station-service', content_key='POST /api/v1/stationservice/stations/idlist'}[5m])))," +
		"histogram_quantile(0.9, sum by(vmrange, content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_bucket{svc_name='ts-travel2-service', content_key='POST /api/v1/travel2service/trips/left'}[5m])))" +
		")"
	if expect != got.ToString() {
		t.Errorf("want=%s, got=%s", expect, got.ToString())
	}
}

func TestSplitSpanTraceP9x(t *testing.T) {
	svcs := []string{"ts-station-service", "ts-travel2-service", "ts-travel-service"}
	endpoints := []string{
		"POST /api/v1/stationservice/stations/idlist",
		"POST /api/v1/travel2service/trips/left",
		"POST /api/v1/travel/service/trips/left",
	}

	got := getSpanTraceP9xSql("vmrange", 5*time.Minute, svcs, endpoints)

	res1 := got.buildQueryWithCondRange(0, 1)
	res2 := got.buildQueryWithCondRange(0, 2)
	res3 := got.buildQueryWithCondRange(1, 3)

	e1 := "histogram_quantile(0.9, sum by(vmrange, content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_bucket{svc_name='ts-station-service', content_key='POST /api/v1/stationservice/stations/idlist'}[5m])))"
	e2 := "union(" +
		"histogram_quantile(0.9, sum by(vmrange, content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_bucket{svc_name='ts-station-service', content_key='POST /api/v1/stationservice/stations/idlist'}[5m])))," +
		"histogram_quantile(0.9, sum by(vmrange, content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_bucket{svc_name='ts-travel2-service', content_key='POST /api/v1/travel2service/trips/left'}[5m])))" +
		")"
	e3 := "union(" +
		"histogram_quantile(0.9, sum by(vmrange, content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_bucket{svc_name='ts-travel2-service', content_key='POST /api/v1/travel2service/trips/left'}[5m])))," +
		"histogram_quantile(0.9, sum by(vmrange, content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_bucket{svc_name='ts-travel-service', content_key='POST /api/v1/travel/service/trips/left'}[5m])))" +
		")"

	if e1 != res1 {
		t.Errorf("want=%s, got=%s", e1, res1)
	}
	if e2 != res2 {
		t.Errorf("want=%s, got=%s", e2, res2)
	}
	if e3 != res3 {
		t.Errorf("want=%s, got=%s", e3, res3)
	}
}

func TestSpanTraceInstanceP9x(t *testing.T) {
	endpoint := "POST /api/v1/stationservice/stations/idlist"
	extraCondition := fmt.Sprintf("pod='%s'", "ts-station-service-8b76754bc-gbst8")
	got := getSpanTraceInstanceP9xSql("vmrange", 5*time.Minute, endpoint, extraCondition)
	expect := "histogram_quantile(0.9, sum by(vmrange) (increase(kindling_span_trace_duration_nanoseconds_bucket{content_key='POST /api/v1/stationservice/stations/idlist', pod='ts-station-service-8b76754bc-gbst8'}[5m])))"
	if expect != got.ToString() {
		t.Errorf("want=%s, got=%s", expect, got.ToString())
	}
}

func TestExternalP9x(t *testing.T) {
	svcs := []string{"ts-basic-service:15680", "ts-order-service:12031"}
	endpoints := []string{"POST /api", "GET /api"}
	systems := []string{"http", "http"}
	got := getExternalP9xSql("vmrange", 5*time.Minute, svcs, endpoints, systems)
	expect := "union(" +
		"histogram_quantile(0.9, sum by(vmrange, address, name) (increase(kindling_external_duration_nanoseconds_bucket{address='ts-basic-service:15680', name='POST /api', system='http'}[5m])))," +
		"histogram_quantile(0.9, sum by(vmrange, address, name) (increase(kindling_external_duration_nanoseconds_bucket{address='ts-order-service:12031', name='GET /api', system='http'}[5m])))" +
		")"
	if expect != got.ToString() {
		t.Errorf("want=%s, got=%s", expect, got.ToString())
	}
}

func testDbP9x(t *testing.T) {
	svcs := []string{"train-ticket-mysql:3306", "train-ticket-mysql:3306"}
	endpoints := []string{"SELECT ts.train_type", "SELECT ts.trip"}
	systems := []string{"mysql", "mysql"}
	got := getDbP9xSql("vmrange", 5*time.Minute, svcs, endpoints, systems)
	expect := "union(" +
		"histogram_quantile(0.9, sum by(vmrange, db_url, name) (increase(kindling_db_duration_nanoseconds_bucket{db_url='train-ticket-mysql:3306', name='SELECT ts.train_type', db_system='mysql'}[5m])))," +
		"histogram_quantile(0.9, sum by(vmrange, db_url, name) (increase(kindling_db_duration_nanoseconds_bucket{db_url='train-ticket-mysql:3306', name='SELECT ts.trip', db_system='mysql'}[5m])))" +
		")"
	if expect != got.ToString() {
		t.Errorf("want=%s, got=%s", expect, got.ToString())
	}
}

func TestMqP9x(t *testing.T) {
	svcs := []string{"", ""}
	endpoints := []string{"topicA", "topicB"}
	systems := []string{"kafka", "rabbitmq"}
	got := getMqP9xSql("vmrange", 5*time.Minute, svcs, endpoints, systems)
	expect := "union(" +
		"histogram_quantile(0.9, sum by(vmrange, address, name) (increase(kindling_mq_duration_nanoseconds_bucket{address='', name='topicA', system='kafka', role!='consumer'}[5m])))," +
		"histogram_quantile(0.9, sum by(vmrange, address, name) (increase(kindling_mq_duration_nanoseconds_bucket{address='', name='topicB', system='rabbitmq', role!='consumer'}[5m])))" +
		")"
	if expect != got.ToString() {
		t.Errorf("want=%s, got=%s", expect, got.ToString())
	}
}
