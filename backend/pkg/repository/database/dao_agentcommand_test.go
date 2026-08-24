package database

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestOpenCircuitCommandPersistenceClaimAndTransitions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&ac.Command{}); err != nil {
		t.Fatal(err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}
	epoch := uint64(12)
	command := &ac.Command{
		ID:                   uuid.New(),
		NodeID:               "node-e2e",
		ProbeType:            "ebpf-agent",
		Type:                 "OPEN_CIRCUIT",
		ExpectedCircuitEpoch: &epoch,
		Reason:               "memory_limit",
		Status:               "PENDING",
	}
	if err := repo.CreateAgentCommand(nil, command); err != nil {
		t.Fatal(err)
	}

	claimed, err := repo.ClaimAgentCommands(nil, "node-e2e", "ebpf-agent", "daemon-e2e", time.Unix(100, 0), 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(claimed) != 1 {
		t.Fatalf("claimed %d commands, want 1", len(claimed))
	}
	got := claimed[0]
	if got.Status != "CLAIMED" || got.ExpectedCircuitEpoch == nil || *got.ExpectedCircuitEpoch != epoch || got.Reason != "memory_limit" {
		t.Fatalf("claimed command lost circuit fields: %+v", got)
	}
	payload, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), `"expectedCircuitEpoch":12`) || !strings.Contains(string(payload), `"reason":"memory_limit"`) {
		t.Fatalf("command JSON lost circuit fields: %s", payload)
	}

	started, err := repo.TransitionAgentCommand(nil, got.ID, got.NodeID, got.ClaimantID, got.LeaseToken, "STARTED", "", time.Unix(101, 0))
	if err != nil {
		t.Fatal(err)
	}
	if started.Status != "STARTED" || started.StartedAt == nil {
		t.Fatalf("unexpected STARTED command: %+v", started)
	}
	succeeded, err := repo.TransitionAgentCommand(nil, got.ID, got.NodeID, got.ClaimantID, got.LeaseToken, "SUCCEEDED", `{"circuit":{"state":"FUSED","epoch":13}}`, time.Unix(102, 0))
	if err != nil {
		t.Fatal(err)
	}
	if succeeded.Status != "SUCCEEDED" || succeeded.FinishedAt == nil || !strings.Contains(succeeded.Result, `"state":"FUSED"`) {
		t.Fatalf("unexpected SUCCEEDED command: %+v", succeeded)
	}
}
