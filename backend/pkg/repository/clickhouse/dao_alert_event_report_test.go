// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"strings"
	"testing"
)

func TestQuerytagsConditionEscapesTagLiteralsAndSeparatesConditions(t *testing.T) {
	got := querytagsCondition(map[string]string{
		"alertEventId": "event-1\\' OR 1=1 --",
		"service":      "checkout",
	})

	want := "AND tags['alertEventId'] = 'event-1\\\\'' OR 1=1 --' AND tags['service'] = 'checkout'"
	if got != want {
		t.Fatalf("querytagsCondition() = %q, want %q", got, want)
	}
}

func TestLatestAlertEventByAlertIDQueryUsesPerAlertLimit(t *testing.T) {
	sql := fmt.Sprintf(SQL_GET_LATEST_ALERT_EVENT_BY_ALERTID, "WHERE alert_id = ?")

	if strings.Contains(sql, "SELECT *") {
		t.Fatalf("latest alert event query = %q, want explicit columns for stable ScanStruct mapping", sql)
	}
	if !strings.Contains(sql, "ORDER BY received_time DESC LIMIT 1 BY alert_id") {
		t.Fatalf("latest alert event query = %q, want per-alert latest ordering", sql)
	}
}
