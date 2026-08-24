package database

import (
	"strings"
	"testing"
	"time"

	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSaveAgentBindingsChecksEverySubmittedGeneration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:agent-config-generations?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&ac.Node{}, &ac.Binding{}, &ac.Release{}); err != nil {
		t.Fatal(err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}
	if err := db.Create(&ac.Node{ID: "node-1"}).Error; err != nil {
		t.Fatal(err)
	}
	for _, release := range []ac.Release{
		{ID: uuid.New(), NodeID: "node-1", ProbeType: "probe-a", Generation: 1, Variables: ac.JSON("{}"), Files: ac.JSON("null"), LogSources: ac.JSON("null")},
		{ID: uuid.New(), NodeID: "node-1", ProbeType: "probe-b", Generation: 2, Variables: ac.JSON("{}"), Files: ac.JSON("null"), LogSources: ac.JSON("null")},
	} {
		if err := db.Create(&release).Error; err != nil {
			t.Fatal(err)
		}
	}
	next := &ac.Release{ID: uuid.New(), NodeID: "node-1", ProbeType: "probe-a", Generation: 2, Variables: ac.JSON("{}"), Files: ac.JSON("null"), LogSources: ac.JSON("null")}
	binding := &ac.Binding{NodeID: "node-1", ProbeType: "probe-a", DesiredGeneration: 2, Variables: ac.JSON("{}")}
	err = repo.SaveAgentBindings(nil, []*ac.Binding{binding}, []*ac.Release{next}, map[string]int64{
		agentReleaseKey("node-1", "probe-a"): 1,
		agentReleaseKey("node-1", "probe-b"): 1,
	})
	if err == nil || !strings.Contains(err.Error(), "node-1/probe-b") {
		t.Fatalf("stale unchanged probe generation was not rejected: %v", err)
	}
	var count int64
	if err := db.Model(&ac.Release{}).Where("node_id=? AND probe_type=?", "node-1", "probe-a").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("conflicted update wrote a release: count=%d", count)
	}
}

func TestUpsertAgentNodePreservesCreatedAtAndRefreshesHeartbeat(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:agent-node-heartbeat?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&ac.Node{}); err != nil {
		t.Fatal(err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}
	firstSeen := time.Now().Add(-time.Minute).Round(time.Millisecond)
	first := &ac.Node{ID: "node-69", Name: "first", ClusterID: "cluster-a", LastSeenAt: &firstSeen}
	if err := repo.UpsertAgentNode(nil, first); err != nil {
		t.Fatal(err)
	}
	createdAt := first.CreatedAt

	secondSeen := time.Now().Round(time.Millisecond)
	second := &ac.Node{ID: "node-69", Name: "second", ClusterID: "cluster-b", LastSeenAt: &secondSeen}
	if err := repo.UpsertAgentNode(nil, second); err != nil {
		t.Fatal(err)
	}
	var stored ac.Node
	if err := db.First(&stored, "id=?", "node-69").Error; err != nil {
		t.Fatal(err)
	}
	if !stored.CreatedAt.Equal(createdAt) {
		t.Fatalf("createdAt changed during heartbeat: first=%s stored=%s", createdAt, stored.CreatedAt)
	}
	if stored.LastSeenAt == nil || !stored.LastSeenAt.Equal(secondSeen) || stored.Name != "second" || stored.ClusterID != "cluster-b" {
		t.Fatalf("registration fields were not refreshed: %+v", stored)
	}
}

func TestReportAgentStateRefreshesHeartbeatAndRuntime(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:agent-state-heartbeat?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&ac.ApplyState{}); err != nil {
		t.Fatal(err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}
	first := &ac.ApplyState{
		NodeID:    "node-69",
		ProbeType: "ilogtail",
		Status:    "RUNNING",
		Runtime:   ac.JSON(`{"childPid":8}`),
		Circuit:   ac.JSON(`{"state":"NORMAL"}`),
	}
	if err := repo.ReportAgentState(nil, first); err != nil {
		t.Fatal(err)
	}
	firstReportedAt := *first.ReportedAt
	time.Sleep(10 * time.Millisecond)

	second := &ac.ApplyState{
		NodeID:    "node-69",
		ProbeType: "ilogtail",
		Status:    "FUSED",
		Reason:    "auto_resource_limit",
		Runtime:   ac.JSON(`{"childRunning":false}`),
		Circuit:   ac.JSON(`{"state":"FUSED","epoch":1}`),
	}
	if err := repo.ReportAgentState(nil, second); err != nil {
		t.Fatal(err)
	}

	var stored ac.ApplyState
	if err := db.First(&stored, "node_id=? AND probe_type=?", "node-69", "ilogtail").Error; err != nil {
		t.Fatal(err)
	}
	if stored.ReportedAt == nil || !stored.ReportedAt.After(firstReportedAt) {
		t.Fatalf("heartbeat did not advance: first=%s stored=%v", firstReportedAt, stored.ReportedAt)
	}
	if stored.Status != "FUSED" || stored.Reason != "auto_resource_limit" || string(stored.Runtime) != `{"childRunning":false}` {
		t.Fatalf("runtime state was not replaced: %+v", stored)
	}
}
