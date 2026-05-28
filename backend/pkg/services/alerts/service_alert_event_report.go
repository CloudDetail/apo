// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"fmt"
	"strings"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	alertmodel "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

const (
	alertEventReportWindowPaddingMicros = int64(time.Hour / time.Microsecond)
	alertEventReportCandidateLimit      = 50
	minLegacyReportMatchScore           = 2
)

type alertEventReportViewSnapshot struct {
	report        *model.AlertEventReport
	event         *model.AlertEvent
	startTime     time.Time
	endTime       time.Time
	viewStartTime time.Time
	viewEndTime   time.Time
}

func (s *service) GetAlertEventReport(ctx core.Context, req *request.GetAlertEventReportRequest) (*response.GetAlertEventReportResponse, error) {

	startTime, endTime := alertEventReportQueryWindow(req)
	report, err := s.chRepo.GetAlertEventReport(ctx, startTime, endTime, req.Tags)
	if err != nil {
		return nil, err
	}
	resp := response.ToAlertReportResponse(*report)
	return &resp, nil
}

func (s *service) GetAlertEventReportView(ctx core.Context, req *request.GetAlertEventReportRequest) (*response.GetAlertEventReportViewResponse, error) {
	startTime, endTime := alertEventReportQueryWindow(req)
	snapshot, err := s.getAlertEventReportViewSnapshot(ctx, req, startTime, endTime)
	if err != nil {
		return nil, err
	}
	viewStartTime, viewEndTime := alertEventReportViewWindow(snapshot)
	viewReq := alertEventReportRequestWithWindow(req, viewStartTime, viewEndTime)
	resp := response.ToAlertReportViewResponse(*snapshot.report, viewReq)
	applyAlertEventFactsToReportView(&resp.View, snapshot.event)
	return &resp, nil
}

func alertEventReportQueryWindow(req *request.GetAlertEventReportRequest) (time.Time, time.Time) {
	return time.UnixMicro(req.StartTime - alertEventReportWindowPaddingMicros),
		time.UnixMicro(req.EndTime + alertEventReportWindowPaddingMicros)
}

func (s *service) getAlertEventReportViewSnapshot(ctx core.Context, req *request.GetAlertEventReportRequest, startTime time.Time, endTime time.Time) (*alertEventReportViewSnapshot, error) {
	report, err := s.chRepo.GetAlertEventReport(ctx, startTime, endTime, req.Tags)
	if err == nil {
		snapshot := &alertEventReportViewSnapshot{
			report:        report,
			startTime:     startTime,
			endTime:       endTime,
			viewStartTime: time.UnixMicro(req.StartTime),
			viewEndTime:   time.UnixMicro(req.EndTime),
		}
		alertEventID := strings.TrimSpace(req.Tags["alertEventId"])
		if alertEventID != "" {
			event, eventErr := s.chRepo.GetAlertEventByEventID(ctx, alertEventID, startTime, endTime)
			if eventErr == nil {
				snapshot.event = s.getLatestAlertEventForReportRuntime(ctx, event, req)
			}
		}
		return snapshot, nil
	}

	alertEventID := strings.TrimSpace(req.Tags["alertEventId"])
	if alertEventID == "" {
		return nil, err
	}

	legacyReport, legacyErr := s.findLegacyAlertEventReport(ctx, req, alertEventID, startTime, endTime)
	if legacyErr != nil {
		eventTimeReport, eventTimeErr := s.findAlertEventReportAroundEvent(ctx, req, alertEventID)
		if eventTimeErr == nil {
			return eventTimeReport, nil
		}
		return nil, err
	}
	return legacyReport, nil
}

func alertEventReportRequestWithWindow(req *request.GetAlertEventReportRequest, startTime time.Time, endTime time.Time) *request.GetAlertEventReportRequest {
	if req == nil || startTime.IsZero() || endTime.IsZero() {
		return req
	}
	copied := *req
	copied.StartTime = startTime.UnixMicro()
	copied.EndTime = endTime.UnixMicro()
	return &copied
}

func alertEventReportViewWindow(snapshot *alertEventReportViewSnapshot) (time.Time, time.Time) {
	if snapshot == nil {
		return time.Time{}, time.Time{}
	}
	if !snapshot.viewStartTime.IsZero() && !snapshot.viewEndTime.IsZero() {
		return snapshot.viewStartTime, snapshot.viewEndTime
	}
	return snapshot.startTime, snapshot.endTime
}

func (s *service) findLegacyAlertEventReport(ctx core.Context, req *request.GetAlertEventReportRequest, alertEventID string, startTime time.Time, endTime time.Time) (*alertEventReportViewSnapshot, error) {
	event, err := s.chRepo.GetAlertEventByEventID(ctx, alertEventID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	candidates, err := s.chRepo.ListAlertEventReports(ctx, startTime, endTime, alertEventReportCandidateLimit)
	if err != nil {
		return nil, err
	}

	report, ok := selectLegacyAlertEventReport(event, candidates)
	if !ok {
		return nil, fmt.Errorf("no legacy alert event report matched event %s", alertEventID)
	}
	return &alertEventReportViewSnapshot{
		report:        report,
		event:         s.getLatestAlertEventForReportRuntime(ctx, event, req),
		startTime:     startTime,
		endTime:       endTime,
		viewStartTime: time.UnixMicro(req.StartTime),
		viewEndTime:   time.UnixMicro(req.EndTime),
	}, nil
}

func (s *service) findAlertEventReportAroundEvent(ctx core.Context, req *request.GetAlertEventReportRequest, alertEventID string) (*alertEventReportViewSnapshot, error) {
	event, err := s.chRepo.GetAlertEventByEventID(ctx, alertEventID, time.Time{}, time.Time{})
	if err != nil {
		return nil, err
	}

	startTime, endTime, ok := alertEventReportWindowFromEvent(event)
	if !ok {
		return nil, fmt.Errorf("alert event %s has no usable report window", alertEventID)
	}

	report, err := s.chRepo.GetAlertEventReport(ctx, startTime, endTime, req.Tags)
	if err == nil {
		return &alertEventReportViewSnapshot{report: report, event: s.getLatestAlertEventForReportRuntime(ctx, event, req), startTime: startTime, endTime: endTime, viewStartTime: startTime, viewEndTime: endTime}, nil
	}

	candidates, err := s.chRepo.ListAlertEventReports(ctx, startTime, endTime, alertEventReportCandidateLimit)
	if err != nil {
		return nil, err
	}

	report, ok = selectLegacyAlertEventReport(event, candidates)
	if !ok {
		return nil, fmt.Errorf("no alert event report around event %s", alertEventID)
	}
	return &alertEventReportViewSnapshot{report: report, event: s.getLatestAlertEventForReportRuntime(ctx, event, req), startTime: startTime, endTime: endTime, viewStartTime: startTime, viewEndTime: endTime}, nil
}

func (s *service) getLatestAlertEventForReportRuntime(ctx core.Context, event *model.AlertEvent, req *request.GetAlertEventReportRequest) *model.AlertEvent {
	alertID := alertEventAlertID(event)
	if alertID == "" || req == nil || req.StartTime == 0 || req.EndTime == 0 {
		return event
	}

	latest, err := s.chRepo.GetLatestAlertEventByAlertIDInWindow(ctx, alertID, time.UnixMicro(req.StartTime), time.UnixMicro(req.EndTime))
	if err != nil || latest == nil {
		return event
	}

	return alertEventRuntimeFromLatestEvent(event, latest)
}

func alertEventAlertID(event *model.AlertEvent) string {
	if event == nil {
		return ""
	}
	if alertID := strings.TrimSpace(event.AlertID); alertID != "" {
		return alertID
	}
	return strings.TrimSpace(event.RawTags["alert_id"])
}

func alertEventRuntimeFromLatestEvent(event *model.AlertEvent, latest *alertmodel.AlertEvent) *model.AlertEvent {
	if latest == nil {
		return event
	}
	runtimeEvent := &model.AlertEvent{}
	if event != nil {
		copied := *event
		runtimeEvent = &copied
	}
	runtimeEvent.AlertID = latest.AlertID
	runtimeEvent.CreateTime = latest.CreateTime
	runtimeEvent.UpdateTime = latest.UpdateTime
	runtimeEvent.EndTime = latest.EndTime
	runtimeEvent.ReceivedTime = latest.ReceivedTime
	runtimeEvent.Status = alertEventRuntimeStatusFromString(latest.Status, runtimeEvent.Status)
	return runtimeEvent
}

func alertEventRuntimeStatusFromString(status string, fallback model.Status) model.Status {
	switch status {
	case alertmodel.StatusResolved:
		return model.StatusResolved
	case alertmodel.StatusFiring:
		return model.StatusFiring
	default:
		return fallback
	}
}

func alertEventReportWindowFromEvent(event *model.AlertEvent) (time.Time, time.Time, bool) {
	if event == nil {
		return time.Time{}, time.Time{}, false
	}

	times := []time.Time{
		event.CreateTime,
		event.UpdateTime,
		event.EndTime,
		event.ReceivedTime,
	}
	var startTime time.Time
	var endTime time.Time
	for _, current := range times {
		if current.IsZero() {
			continue
		}
		if startTime.IsZero() || current.Before(startTime) {
			startTime = current
		}
		if endTime.IsZero() || current.After(endTime) {
			endTime = current
		}
	}
	if startTime.IsZero() || endTime.IsZero() {
		return time.Time{}, time.Time{}, false
	}
	return startTime.Add(-time.Hour), endTime.Add(time.Hour), true
}

func applyAlertEventFactsToReportView(view *response.ReportView, event *model.AlertEvent) {
	if view == nil || event == nil {
		return
	}

	applyAlertEventRuntimeStatus(view, event)
}

func applyAlertEventRuntimeStatus(view *response.ReportView, event *model.AlertEvent) {
	status := response.ReportRuntimeStatus{UpdatedAt: alertEventRuntimeUpdatedAt(event)}
	switch event.Status {
	case model.StatusFiring:
		status.State = "active"
		status.Label = "仍异常"
		status.Summary = "告警事件当前仍处于触发状态。"
	case model.StatusResolved:
		status.State = "resolved"
		status.Label = "已恢复"
		status.Summary = "告警事件当前状态为已恢复。"
	default:
		status.State = "unknown"
		status.Label = "未知"
		status.Summary = "告警事件当前状态未知。"
	}

	view.RuntimeStatus = &status
	view.DataSources.RuntimeStatus = "alert_event"
}

func alertEventRuntimeUpdatedAt(event *model.AlertEvent) int64 {
	if event.Status == model.StatusResolved && !event.EndTime.IsZero() {
		return event.EndTime.UnixMicro()
	}
	if !event.UpdateTime.IsZero() {
		return event.UpdateTime.UnixMicro()
	}
	if !event.CreateTime.IsZero() {
		return event.CreateTime.UnixMicro()
	}
	if !event.ReceivedTime.IsZero() {
		return event.ReceivedTime.UnixMicro()
	}
	return 0
}

func selectLegacyAlertEventReport(event *model.AlertEvent, candidates []model.AlertEventReport) (*model.AlertEventReport, bool) {
	eventTags := buildAlertEventReportIdentityTags(event)
	bestScore := 0
	bestIndices := []int{}

	for index := range candidates {
		score := alertEventReportMatchScore(reportIdentityTags(candidates[index]), eventTags)
		if score > bestScore {
			bestScore = score
			bestIndices = []int{index}
		} else if score == bestScore && score > 0 {
			bestIndices = append(bestIndices, index)
		}
	}

	if len(bestIndices) == 0 {
		return nil, false
	}
	if bestScore < minLegacyReportMatchScore && len(candidates) > 1 {
		return nil, false
	}
	if len(bestIndices) > 1 && !legacyReportsShareSameIdentity(candidates, bestIndices) {
		return nil, false
	}

	return &candidates[bestIndices[0]], true
}

func buildAlertEventReportIdentityTags(event *model.AlertEvent) map[string]string {
	tags := map[string]string{}
	if event == nil {
		return tags
	}

	for key, value := range event.Tags {
		addReportIdentityTag(tags, key, value)
	}
	for key, value := range event.RawTags {
		addReportIdentityTag(tags, key, value)
	}

	addReportIdentityTag(tags, "alertName", event.Name)
	addReportIdentityTag(tags, "group", event.Group)
	addReportIdentityTag(tags, "service", firstReportIdentityValue(event.Tags, "service", "serviceName", "svc_name"))
	addReportIdentityTag(tags, "serviceName", firstReportIdentityValue(event.Tags, "serviceName", "service", "svc_name"))
	addReportIdentityTag(tags, "endpoint", firstReportIdentityValue(event.Tags, "endpoint", "content_key"))
	addReportIdentityTag(tags, "pod", firstReportIdentityValue(event.Tags, "pod", "pod_name", "src_pod", "podName"))
	addReportIdentityTag(tags, "podName", firstReportIdentityValue(event.Tags, "podName", "pod_name", "pod", "src_pod"))
	addReportIdentityTag(tags, "namespace", firstReportIdentityValue(event.Tags, "namespace", "src_namespace"))
	addReportIdentityTag(tags, "nodeName", firstReportIdentityValue(event.Tags, "nodeName", "node", "node_name", "instance_name"))
	addReportIdentityTag(tags, "node", firstReportIdentityValue(event.Tags, "node", "nodeName", "node_name", "instance_name"))
	addReportIdentityTag(tags, "pid", firstReportIdentityValue(event.Tags, "pid"))
	addReportIdentityTag(tags, "containerId", firstReportIdentityValue(event.Tags, "containerId", "container_id", "containerID"))
	addReportIdentityTag(tags, "containerId", firstReportIdentityValue(event.RawTags, "containerId", "container_id", "containerID"))
	addReportIdentityTag(tags, "container_id", firstReportIdentityValue(event.Tags, "container_id", "containerId", "containerID"))
	addReportIdentityTag(tags, "container_id", firstReportIdentityValue(event.RawTags, "container_id", "containerId", "containerID"))

	return tags
}

func reportIdentityTags(report model.AlertEventReport) map[string]string {
	tags := map[string]string{}
	for key, value := range report.Tags {
		addReportIdentityTag(tags, key, value)
	}
	for key, value := range report.OverviewTags {
		addReportIdentityTag(tags, key, value)
	}
	return tags
}

func alertEventReportMatchScore(reportTags map[string]string, eventTags map[string]string) int {
	score := 0
	for key, value := range reportTags {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if alertEventReportTagMatches(key, value, eventTags) {
			score++
		}
	}
	return score
}

func alertEventReportTagMatches(key string, value string, eventTags map[string]string) bool {
	for _, eventKey := range alertEventReportTagAliases(key) {
		if strings.TrimSpace(eventTags[eventKey]) == value {
			return true
		}
	}
	return false
}

func alertEventReportTagAliases(key string) []string {
	switch key {
	case "service", "serviceName", "svc_name":
		return []string{key, "service", "serviceName", "svc_name"}
	case "endpoint", "content_key":
		return []string{key, "endpoint", "content_key"}
	case "pod", "podName", "pod_name", "src_pod":
		return []string{key, "pod", "podName", "pod_name", "src_pod"}
	case "namespace", "src_namespace":
		return []string{key, "namespace", "src_namespace"}
	case "nodeName", "node", "node_name", "instance_name":
		return []string{key, "nodeName", "node", "node_name", "instance_name"}
	case "containerId", "container_id", "containerID":
		return []string{key, "containerId", "container_id", "containerID"}
	default:
		return []string{key}
	}
}

func legacyReportsShareSameIdentity(candidates []model.AlertEventReport, indices []int) bool {
	if len(indices) == 0 {
		return false
	}

	signature := legacyReportIdentitySignature(reportIdentityTags(candidates[indices[0]]))
	if signature == "" {
		return false
	}
	for _, index := range indices[1:] {
		if legacyReportIdentitySignature(reportIdentityTags(candidates[index])) != signature {
			return false
		}
	}
	return true
}

func legacyReportIdentitySignature(tags map[string]string) string {
	parts := []string{}
	appendPart := func(canonicalKey string, aliases ...string) {
		value := firstReportIdentityValue(tags, aliases...)
		if value != "" {
			parts = append(parts, canonicalKey+"="+value)
		}
	}

	appendPart("service", "service", "serviceName", "svc_name")
	appendPart("endpoint", "endpoint", "content_key")
	appendPart("pod", "pod", "podName", "pod_name", "src_pod")
	appendPart("namespace", "namespace", "src_namespace")
	appendPart("nodeName", "nodeName", "node", "node_name", "instance_name")
	appendPart("pid", "pid")
	appendPart("containerId", "containerId", "container_id", "containerID")
	appendPart("group", "group")
	appendPart("alertName", "alertName")

	return strings.Join(parts, "|")
}

func firstReportIdentityValue(tags map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(tags[key]); value != "" {
			return value
		}
	}
	return ""
}

func addReportIdentityTag(tags map[string]string, key string, value string) {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" || value == "" {
		return
	}
	if tags[key] == "" {
		tags[key] = value
	}
}
