package alerts

import (
	"errors"
	"testing"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	alertmodel "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

type fakeReportViewCHRepo struct {
	clickhouse.Repo

	exactErr          error
	exactReport       *model.AlertEventReport
	event             *model.AlertEvent
	latestEvent       *alertmodel.AlertEvent
	latestWindowEvent *alertmodel.AlertEvent
	candidates        []model.AlertEventReport

	eventLookupCount        int
	latestLookupCount       int
	latestWindowLookupCount int
	candidateLookupCount    int
	latestWindowStartTime   time.Time
	latestWindowEndTime     time.Time
}

func (f *fakeReportViewCHRepo) GetAlertEventReport(ctx core.Context, startTime time.Time, endTime time.Time, tags map[string]string) (*model.AlertEventReport, error) {
	if f.exactReport != nil {
		return f.exactReport, nil
	}
	return nil, f.exactErr
}

func (f *fakeReportViewCHRepo) GetAlertEventByEventID(ctx core.Context, eventID string, startTime, endTime time.Time) (*model.AlertEvent, error) {
	f.eventLookupCount++
	if f.event == nil {
		return nil, errors.New("event not found")
	}
	return f.event, nil
}

func (f *fakeReportViewCHRepo) GetLatestAlertEventByAlertID(ctx core.Context, alertID string) (*alertmodel.AlertEvent, error) {
	f.latestLookupCount++
	if f.latestEvent == nil {
		return nil, errors.New("latest event not found")
	}
	return f.latestEvent, nil
}

func (f *fakeReportViewCHRepo) GetLatestAlertEventByAlertIDInWindow(ctx core.Context, alertID string, startTime, endTime time.Time) (*alertmodel.AlertEvent, error) {
	f.latestWindowLookupCount++
	f.latestWindowStartTime = startTime
	f.latestWindowEndTime = endTime
	if f.latestWindowEvent == nil {
		return nil, errors.New("latest event not found")
	}
	return f.latestWindowEvent, nil
}

func (f *fakeReportViewCHRepo) ListAlertEventReports(ctx core.Context, startTime time.Time, endTime time.Time, limit int) ([]model.AlertEventReport, error) {
	f.candidateLookupCount++
	return f.candidates, nil
}

func TestGetAlertEventReportViewReturnsSnapshotAndRuntimeStatusOnly(t *testing.T) {
	req := &request.GetAlertEventReportRequest{
		StartTime: 1779256980000000,
		EndTime:   1779258240000000,
		Tags:      map[string]string{"alertEventId": "event-001"},
	}
	chRepo := &fakeReportViewCHRepo{
		exactReport: &model.AlertEventReport{
			ReportType:        "infra",
			OverviewTimestamp: time.UnixMicro(1779258240000000),
			OverviewDetail:    "错误日志数升高",
			OverviewReason:    "CPU 与错误日志同步升高",
			OverviewTags:      map[string]string{"service": "checkout"},
			RootCause:         "CPU 方向异常",
		},
		event: &model.AlertEvent{
			Status:     model.StatusFiring,
			UpdateTime: time.UnixMicro(1779258240000000),
		},
	}

	resp, err := (&service{chRepo: chRepo}).GetAlertEventReportView(core.EmptyCtx(), req)
	if err != nil {
		t.Fatalf("GetAlertEventReportView() error = %v", err)
	}

	if resp.Snapshot.Type != "infra" {
		t.Fatalf("snapshot type = %q, want infra", resp.Snapshot.Type)
	}
	if resp.View.RuntimeStatus == nil || resp.View.RuntimeStatus.State != "active" {
		t.Fatalf("runtime status = %#v, want active", resp.View.RuntimeStatus)
	}
	if resp.View.BusinessImpact != nil {
		t.Fatalf("business impact = %#v, want nil", resp.View.BusinessImpact)
	}
	if len(resp.View.MetricDeltas) != 0 {
		t.Fatalf("metric deltas = %#v, want empty", resp.View.MetricDeltas)
	}
	if resp.View.FaultLocation != nil {
		t.Fatalf("fault location = %#v, want nil", resp.View.FaultLocation)
	}
	if len(resp.View.Timeline) != 0 {
		t.Fatalf("timeline = %#v, want empty", resp.View.Timeline)
	}
}

func TestGetAlertEventReportViewUsesWindowLatestResolvedStatus(t *testing.T) {
	req := &request.GetAlertEventReportRequest{
		StartTime: 1779256980000000,
		EndTime:   1779258240000000,
		Tags:      map[string]string{"alertEventId": "event-001"},
	}
	chRepo := &fakeReportViewCHRepo{
		exactReport: &model.AlertEventReport{
			ReportType:        "infra",
			OverviewTimestamp: time.UnixMicro(1779258240000000),
			OverviewDetail:    "错误日志数升高",
			OverviewReason:    "CPU 与错误日志同步升高",
			OverviewTags:      map[string]string{"service": "checkout"},
			RootCause:         "CPU 方向异常",
		},
		event: &model.AlertEvent{
			AlertID:    "alert-001",
			Status:     model.StatusFiring,
			UpdateTime: time.UnixMicro(1779258240000000),
		},
		latestWindowEvent: &alertmodel.AlertEvent{
			Status:   alertmodel.StatusResolved,
			EndTime:  time.UnixMicro(1779258240000000),
			Alert:    alertmodel.Alert{AlertID: "alert-001"},
			EventID:  "event-002",
			Severity: alertmodel.SeverityWarnLevel,
		},
	}

	resp, err := (&service{chRepo: chRepo}).GetAlertEventReportView(core.EmptyCtx(), req)
	if err != nil {
		t.Fatalf("GetAlertEventReportView() error = %v", err)
	}

	if resp.View.RuntimeStatus == nil || resp.View.RuntimeStatus.State != "resolved" {
		t.Fatalf("runtime status = %#v, want resolved", resp.View.RuntimeStatus)
	}
	if chRepo.latestWindowLookupCount != 1 {
		t.Fatalf("latest event window lookup count = %d, want 1", chRepo.latestWindowLookupCount)
	}
	if chRepo.latestWindowStartTime.UnixMicro() != req.StartTime || chRepo.latestWindowEndTime.UnixMicro() != req.EndTime {
		t.Fatalf("latest event window = %d-%d, want %d-%d", chRepo.latestWindowStartTime.UnixMicro(), chRepo.latestWindowEndTime.UnixMicro(), req.StartTime, req.EndTime)
	}
	if chRepo.latestLookupCount != 0 {
		t.Fatalf("latest event lookup count = %d, want 0", chRepo.latestLookupCount)
	}
}

func TestGetAlertEventReportViewKeepsRequestedResolvedEventStatus(t *testing.T) {
	req := &request.GetAlertEventReportRequest{
		StartTime: 1779256980000000,
		EndTime:   1779258240000000,
		Tags:      map[string]string{"alertEventId": "event-001"},
	}
	chRepo := &fakeReportViewCHRepo{
		exactReport: &model.AlertEventReport{
			ReportType:        "infra",
			OverviewTimestamp: time.UnixMicro(1779258240000000),
			OverviewDetail:    "RTT 偏高",
			OverviewReason:    "网络 RTT 在告警窗口升高。",
			OverviewTags:      map[string]string{"node": "worker-23"},
			RootCause:         "网络方向异常",
		},
		event: &model.AlertEvent{
			AlertID:    "alert-001",
			Status:     model.StatusResolved,
			UpdateTime: time.UnixMicro(1779258240000000),
			EndTime:    time.UnixMicro(1779258240000000),
		},
	}

	resp, err := (&service{chRepo: chRepo}).GetAlertEventReportView(core.EmptyCtx(), req)
	if err != nil {
		t.Fatalf("GetAlertEventReportView() error = %v", err)
	}

	if resp.View.RuntimeStatus == nil || resp.View.RuntimeStatus.State != "resolved" {
		t.Fatalf("runtime status = %#v, want resolved", resp.View.RuntimeStatus)
	}
}

func TestGetAlertEventReportViewStillFallsBackToLegacyReportTags(t *testing.T) {
	req := &request.GetAlertEventReportRequest{
		StartTime: 1779256980000000,
		EndTime:   1779258240000000,
		Tags:      map[string]string{"alertEventId": "event-001"},
	}
	chRepo := &fakeReportViewCHRepo{
		exactErr: errors.New("no report found"),
		event: &model.AlertEvent{
			Tags:    map[string]string{"pod": "checkout-7d9c6f", "node": "node-1", "pid": "12345"},
			RawTags: map[string]string{"container_id": "container-abc"},
		},
		candidates: []model.AlertEventReport{
			{
				ReportType:        "infra",
				OverviewTimestamp: time.UnixMicro(1779258240000000),
				OverviewDetail:    "checkout pod CPU 异常",
				Tags: map[string]string{
					"pod":         "checkout-7d9c6f",
					"nodeName":    "node-1",
					"pid":         "12345",
					"containerId": "container-abc",
				},
				RootCause: "node-1 上 checkout pod CPU 异常",
			},
		},
	}

	resp, err := (&service{chRepo: chRepo}).GetAlertEventReportView(core.EmptyCtx(), req)
	if err != nil {
		t.Fatalf("GetAlertEventReportView() error = %v", err)
	}

	if resp.Snapshot.Type != "infra" {
		t.Fatalf("snapshot type = %q, want infra", resp.Snapshot.Type)
	}
	if chRepo.eventLookupCount != 1 {
		t.Fatalf("event lookup count = %d, want 1", chRepo.eventLookupCount)
	}
	if chRepo.candidateLookupCount != 1 {
		t.Fatalf("candidate lookup count = %d, want 1", chRepo.candidateLookupCount)
	}
}
