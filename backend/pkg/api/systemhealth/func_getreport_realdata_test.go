// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	alertmodel "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"go.uber.org/zap"
)

func TestRealDataBackendCanQueryAlertEvidenceChain(t *testing.T) {
	if os.Getenv("APO_SYSTEM_HEALTH_REALDATA") != "1" {
		t.Skip("set APO_SYSTEM_HEALTH_REALDATA=1 to query real Alertmanager, vmalert, and VM data")
	}

	groupID, err := strconv.ParseInt(envOrDefault("APO_SYSTEM_HEALTH_REALDATA_GROUP_ID", "692489354112"), 10, 64)
	if err != nil {
		t.Fatalf("parse group id: %v", err)
	}
	alertManagerURL := envOrDefault("APO_SYSTEM_HEALTH_REALDATA_ALERTMANAGER", "http://192.168.1.6:30093")
	vmAlertURL := envOrDefault("APO_SYSTEM_HEALTH_REALDATA_VMALERT", "http://192.168.1.6:30888")
	vmURL := envOrDefault("APO_SYSTEM_HEALTH_REALDATA_VM", "http://192.168.1.6:8428")
	targetAlertName := os.Getenv("APO_SYSTEM_HEALTH_REALDATA_ALERT")

	ctx := core.EmptyCtx()
	promRepo, err := prometheus.New(zap.NewNop(), vmURL, "vm")
	if err != nil {
		t.Fatalf("new prometheus repo: %v", err)
	}
	h := &handler{
		promRepo:        promRepo,
		alertManagerURL: alertManagerURL,
		vmAlertURL:      vmAlertURL,
	}

	events, err := h.queryAlertManagerActiveEvents(ctx, groupID)
	if err != nil {
		t.Fatalf("query alertmanager active events: %v", err)
	}
	if len(events) == 0 {
		t.Fatalf("expected active alert events for group %d", groupID)
	}
	event := pickRealDataEvent(events, targetAlertName)
	if event.Name == "" {
		t.Fatalf("no matching alert event found for target alert %q", targetAlertName)
	}
	if event.Status != alertmodel.StatusFiring || !isValidAlert(event) {
		t.Fatalf("expected valid firing event, got status=%s validity=%s isValid=%s", event.Status, event.Validity, event.IsValid)
	}

	rule, found := h.findAlertRuleForEvent(ctx, event, groupID)
	if !found || rule == nil || rule.Expr == "" {
		t.Fatalf("expected alert rule for event %q, found=%v rule=%+v", event.Name, found, rule)
	}

	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)
	input := reportInput{
		events:          []alertmodel.AEventWithWRecord{event},
		alertNames:      []string{event.Name},
		alertSeverities: []string{event.Severity},
	}
	evidence := h.alertRuleMetricTrendEvidence(ctx, input, groupID, startTime, endTime)
	if evidence.Metric == nil {
		t.Fatalf("expected metric evidence, result=%s", evidence.Result)
	}
	if evidence.Metric.SeriesReturned == 0 || evidence.Metric.PointCount == 0 {
		t.Fatalf("expected non-empty metric series, metric=%+v result=%s", *evidence.Metric, evidence.Result)
	}
	if !evidence.Metric.EventScoped || evidence.Metric.SeriesMatched == 0 {
		t.Fatalf("expected event-scoped metric match, metric=%+v result=%s", *evidence.Metric, evidence.Result)
	}
	if len(evidence.Metric.Series) == 0 || len(evidence.Metric.Series[0].Points) == 0 {
		t.Fatalf("expected chartable metric points, metric=%+v result=%s", *evidence.Metric, evidence.Result)
	}

	t.Logf(
		"backend realdata ok: group=%d activeEvents=%d alert=%q rule=%q series=%d matched=%d points=%d chartPoints=%d match=%s",
		groupID,
		len(events),
		event.Name,
		rule.Expr,
		evidence.Metric.SeriesReturned,
		evidence.Metric.SeriesMatched,
		evidence.Metric.PointCount,
		len(evidence.Metric.Series[0].Points),
		evidence.Metric.MatchExpression,
	)
}

func pickRealDataEvent(events []alertmodel.AEventWithWRecord, targetAlertName string) alertmodel.AEventWithWRecord {
	for _, event := range events {
		if targetAlertName != "" && event.Name != targetAlertName {
			continue
		}
		if event.GetStringTagWithRaw("namespace") != "" && event.GetStringTagWithRaw("pod") != "" {
			return event
		}
	}
	for _, event := range events {
		if targetAlertName == "" || event.Name == targetAlertName {
			return event
		}
	}
	return alertmodel.AEventWithWRecord{}
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
