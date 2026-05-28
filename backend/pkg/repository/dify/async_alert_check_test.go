// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dify

import (
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"go.uber.org/zap"
)

func TestSampleWithFirstRecordWaitsForQueueSpaceBeforeDropping(t *testing.T) {
	checker := newAsyncAlertCheck(&AlertCheckConfig{
		QueueSize:      1,
		EnqueueTimeout: 100 * time.Millisecond,
	}, zap.NewNop())
	sampler, ok := checker.(*sampleWithFirstRecord)
	if !ok {
		t.Fatalf("newAsyncAlertCheck() = %T, want *sampleWithFirstRecord", checker)
	}

	sampler.AddEvents([]alert.AlertEvent{newAlertCheckTestEvent("alert-1")})

	done := make(chan struct{})
	go func() {
		defer close(done)
		sampler.AddEvents([]alert.AlertEvent{newAlertCheckTestEvent("alert-2")})
	}()

	select {
	case <-done:
		t.Fatal("AddEvents returned before queue space was available")
	case <-time.After(10 * time.Millisecond):
	}

	<-sampler.eventInput.Ch

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("AddEvents did not enqueue after queue space became available")
	}

	got := <-sampler.eventInput.Ch
	if got.AlertID != "alert-2" {
		t.Fatalf("queued alertId = %q, want alert-2", got.AlertID)
	}
}

func TestSampleWithFirstRecordDoesNotMarkDroppedAlertAsChecked(t *testing.T) {
	checker := newAsyncAlertCheck(&AlertCheckConfig{
		QueueSize:      1,
		EnqueueTimeout: time.Millisecond,
	}, zap.NewNop())
	sampler, ok := checker.(*sampleWithFirstRecord)
	if !ok {
		t.Fatalf("newAsyncAlertCheck() = %T, want *sampleWithFirstRecord", checker)
	}

	sampler.AddEvents([]alert.AlertEvent{newAlertCheckTestEvent("alert-1")})
	sampler.AddEvents([]alert.AlertEvent{newAlertCheckTestEvent("alert-2")})

	if _, exists := sampler.checkedAlert["alert-2"]; exists {
		t.Fatal("dropped alert was marked as checked")
	}

	<-sampler.eventInput.Ch
	sampler.AddEvents([]alert.AlertEvent{newAlertCheckTestEvent("alert-2")})

	got := <-sampler.eventInput.Ch
	if got.AlertID != "alert-2" {
		t.Fatalf("queued alertId = %q, want alert-2", got.AlertID)
	}
}

func newAlertCheckTestEvent(alertID string) alert.AlertEvent {
	return alert.AlertEvent{
		Alert: alert.Alert{
			AlertID: alertID,
			Name:    "test alert",
		},
		EventID:    alertID + "-event",
		Status:     alert.StatusFiring,
		UpdateTime: time.Now(),
	}
}
