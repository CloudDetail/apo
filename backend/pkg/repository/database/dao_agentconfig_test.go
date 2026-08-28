package database

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	integrationmodel "github.com/CloudDetail/apo/backend/pkg/model/integration"
	profilemodel "github.com/CloudDetail/apo/backend/pkg/model/profile"
	"github.com/CloudDetail/apo/backend/pkg/repository/database/driver"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetNamespaceInjectionUsesStoredClusterIntegration(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:agent-namespace-injection?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&integrationmodel.TraceIntegration{}, &ac.ClusterInjectionConfig{}); err != nil {
		t.Fatal(err)
	}
	trace := integrationmodel.TraceIntegration{
		ClusterID: "cluster-a",
		SelfCollectConfig: integrationmodel.JSONField[integrationmodel.TraceSelfCollectConfig]{Obj: integrationmodel.TraceSelfCollectConfig{
			InstrumentAll:        true,
			InstrumentNS:         []string{"orders"},
			InstrumentDisabledNS: []string{"monitoring"},
		}},
	}
	if err := db.Create(&trace).Error; err != nil {
		t.Fatal(err)
	}
	config, err := (&daoRepo{DB: &driver.DB{DB: db}}).GetNamespaceInjection(nil, "cluster-a")
	if err != nil {
		t.Fatal(err)
	}
	if !config.InstrumentAll || len(config.InstrumentNS) != 1 || config.InstrumentNS[0] != "orders" || len(config.InstrumentDisabledNS) != 1 || config.InstrumentDisabledNS[0] != "monitoring" {
		t.Fatalf("namespace injection config=%+v", config)
	}

	data, err := json.Marshal(ac.NamespaceInjection{InstrumentNS: []string{"payments"}})
	if err != nil {
		t.Fatal(err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}
	if err := repo.SaveClusterInjectionConfig(nil, &ac.ClusterInjectionConfig{ClusterID: "cluster-a", Configured: true, Config: ac.JSON(data)}, nil); err != nil {
		t.Fatal(err)
	}
	config, err = repo.GetNamespaceInjection(nil, "cluster-a")
	if err != nil || config == nil || len(config.InstrumentNS) != 1 || config.InstrumentNS[0] != "payments" {
		t.Fatalf("dedicated namespace injection config=%+v err=%v", config, err)
	}
	if err := repo.SaveClusterInjectionConfig(nil, &ac.ClusterInjectionConfig{ClusterID: "cluster-a", Configured: false, Config: ac.JSON(`{}`)}, nil); err != nil {
		t.Fatal(err)
	}
	config, err = repo.GetNamespaceInjection(nil, "cluster-a")
	if err != nil || config == nil || config.InstrumentAll || len(config.InstrumentNS) != 0 {
		t.Fatalf("empty remote namespace injection config=%+v err=%v", config, err)
	}
}

func TestGetOdigletReadinessAggregatesFreshClusterNodes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:agent-odiglet-readiness?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&ac.Node{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	ready, notReady := true, false
	nodes := []ac.Node{
		{ID: "node-a", ClusterID: "cluster-a", OdigletReady: &notReady, LastSeenAt: &now},
		{ID: "node-b", ClusterID: "cluster-a", OdigletReady: &ready, LastSeenAt: &now},
		{ID: "node-c", ClusterID: "cluster-b", OdigletReady: &ready, LastSeenAt: &now},
	}
	if err := db.Create(&nodes).Error; err != nil {
		t.Fatal(err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}
	result, err := repo.GetOdigletReadiness(nil, "cluster-missing", now.Add(-time.Minute))
	if err != nil || result != nil {
		t.Fatalf("missing-cluster readiness=%v err=%v", result, err)
	}

	result, err = repo.GetOdigletReadiness(nil, "cluster-a", now.Add(-time.Minute))
	if err != nil || result == nil || !*result {
		t.Fatalf("readiness=%v err=%v", result, err)
	}

	if err := db.Model(&ac.Node{}).Where("cluster_id = ?", "cluster-a").Update("odiglet_ready", false).Error; err != nil {
		t.Fatal(err)
	}
	result, err = repo.GetOdigletReadiness(nil, "cluster-a", now.Add(-time.Minute))
	if err != nil || result == nil || *result {
		t.Fatalf("all-down readiness=%v err=%v", result, err)
	}

	stale := now.Add(-2 * time.Minute)
	if err := db.Model(&ac.Node{}).Where("id = ?", "node-b").Updates(map[string]any{"odiglet_ready": true, "last_seen_at": stale}).Error; err != nil {
		t.Fatal(err)
	}
	result, err = repo.GetOdigletReadiness(nil, "cluster-a", now.Add(-time.Minute))
	if err != nil || result == nil || *result {
		t.Fatalf("stale ready node affected readiness=%v err=%v", result, err)
	}

	if err := db.Model(&ac.Node{}).Where("id = ?", "node-a").Update("odiglet_ready", nil).Error; err != nil {
		t.Fatal(err)
	}
	result, err = repo.GetOdigletReadiness(nil, "cluster-a", now.Add(-time.Minute))
	if err != nil || result != nil {
		t.Fatalf("mixed-version readiness=%v err=%v", result, err)
	}
}

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
	}, nil)
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

func TestSaveAgentBindingsPersistsDeliveryAndTargetsAtomically(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:agent-config-deliveries?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&ac.Node{}, &ac.Binding{}, &ac.Release{}, &ac.Delivery{}, &ac.DeliveryTarget{}); err != nil {
		t.Fatal(err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}
	if err := db.Create(&ac.Node{ID: "node-1", ClusterID: "cluster-a"}).Error; err != nil {
		t.Fatal(err)
	}
	release := &ac.Release{ID: uuid.New(), NodeID: "node-1", ProbeType: "probe-a", Generation: 1, Variables: ac.JSON("{}"), Files: ac.JSON("null"), LogSources: ac.JSON("null")}
	binding := &ac.Binding{NodeID: "node-1", ProbeType: "probe-a", DesiredGeneration: 1, Variables: ac.JSON("{}")}
	delivery := &ac.Delivery{ID: uuid.New(), Type: "NODE_CONFIG", Scope: "node-1", TargetNodeCount: 1, Targets: []ac.DeliveryTarget{{NodeID: "node-1", ClusterID: "cluster-a", Generations: ac.JSON(`{"probe-a":1}`)}}}
	if err := repo.SaveAgentBindings(nil, []*ac.Binding{binding}, []*ac.Release{release}, map[string]int64{agentReleaseKey("node-1", "probe-a"): 0}, delivery); err != nil {
		t.Fatal(err)
	}
	deliveries, err := repo.ListAgentDeliveries(nil, 20)
	if err != nil || len(deliveries) != 1 || len(deliveries[0].Targets) != 1 || deliveries[0].Targets[0].NodeID != "node-1" {
		t.Fatalf("deliveries=%+v err=%v", deliveries, err)
	}

	invalidRelease := &ac.Release{ID: uuid.New(), NodeID: "node-1", ProbeType: "probe-a", Generation: 2, Variables: ac.JSON("{}"), Files: ac.JSON("null"), LogSources: ac.JSON("null")}
	invalidBinding := &ac.Binding{NodeID: "node-1", ProbeType: "probe-a", DesiredGeneration: 2, Variables: ac.JSON("{}")}
	invalidDelivery := &ac.Delivery{ID: uuid.New(), Type: "NODE_CONFIG", Scope: "node-1", TargetNodeCount: 1}
	if err := repo.SaveAgentBindings(nil, []*ac.Binding{invalidBinding}, []*ac.Release{invalidRelease}, map[string]int64{agentReleaseKey("node-1", "probe-a"): 1}, invalidDelivery); err == nil {
		t.Fatal("delivery without targets was accepted")
	}
	var releaseCount int64
	if err := db.Model(&ac.Release{}).Count(&releaseCount).Error; err != nil || releaseCount != 1 {
		t.Fatalf("invalid delivery left a release behind: count=%d err=%v", releaseCount, err)
	}
}

func TestListAgentDeliveriesResolvesOperatorUsername(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:agent-config-delivery-users?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&ac.Delivery{}, &ac.DeliveryTarget{}, &profilemodel.User{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&profilemodel.User{UserID: 42, Username: "operator"}).Error; err != nil {
		t.Fatal(err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}
	delivery := &ac.Delivery{ID: uuid.New(), Type: "NODE_CIRCUIT", Scope: "node-1", RequestedBy: 42, Configuration: ac.JSON(`{"enabled":false}`)}
	if err := saveAgentDelivery(db, delivery); err != nil {
		t.Fatal(err)
	}
	deliveries, err := repo.ListAgentDeliveries(nil, 20)
	if err != nil || len(deliveries) != 1 || deliveries[0].RequestedByName != "operator" {
		t.Fatalf("deliveries=%+v err=%v", deliveries, err)
	}
}

func TestSaveAgentInitialConfigRefreshesOnlyBeforeFirstRelease(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:agent-initial-refresh?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&ac.InitialConfig{}, &ac.Release{}); err != nil {
		t.Fatal(err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}
	first := &ac.InitialConfig{NodeID: "node-69", ProbeType: "ilogtail", Files: ac.JSON(`[{"path":"config/local/old.yaml"}]`), ContentSHA256: "old", CapturedAt: time.Now().Add(-time.Minute)}
	if saved, err := repo.SaveAgentInitialConfigIfUnassigned(nil, first); err != nil || !saved {
		t.Fatalf("save first baseline: saved=%v err=%v", saved, err)
	}
	updated := &ac.InitialConfig{NodeID: "node-69", ProbeType: "ilogtail", Files: ac.JSON(`[{"path":"conf/plugin_logger.xml"}]`), ContentSHA256: "new", CapturedAt: time.Now()}
	if saved, err := repo.SaveAgentInitialConfigIfUnassigned(nil, updated); err != nil || !saved {
		t.Fatalf("refresh unassigned baseline: saved=%v err=%v", saved, err)
	}
	stored, err := repo.GetAgentInitialConfig(nil, "node-69", "ilogtail")
	if err != nil || stored.ContentSHA256 != "new" || string(stored.Files) != string(updated.Files) {
		t.Fatalf("baseline was not refreshed: stored=%+v err=%v", stored, err)
	}

	release := ac.Release{ID: uuid.New(), NodeID: "node-69", ProbeType: "ilogtail", Generation: 1, Variables: ac.JSON("{}"), Files: ac.JSON("null"), LogSources: ac.JSON("null")}
	if err := db.Create(&release).Error; err != nil {
		t.Fatal(err)
	}
	blocked := &ac.InitialConfig{NodeID: "node-69", ProbeType: "ilogtail", Files: ac.JSON(`[{"path":"wrong.yaml"}]`), ContentSHA256: "wrong", CapturedAt: time.Now().Add(time.Minute)}
	if saved, err := repo.SaveAgentInitialConfigIfUnassigned(nil, blocked); err != nil || saved {
		t.Fatalf("released baseline changed: saved=%v err=%v", saved, err)
	}
	stored, err = repo.GetAgentInitialConfig(nil, "node-69", "ilogtail")
	if err != nil || stored.ContentSHA256 != "new" {
		t.Fatalf("released baseline was overwritten: stored=%+v err=%v", stored, err)
	}
}

func TestSaveNodeLogConfigsReplacesEachNodeAtomically(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:agent-log-configs?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&ac.NodeLogConfig{}); err != nil {
		t.Fatal(err)
	}
	repo := &daoRepo{DB: &driver.DB{DB: db}}
	first := []*ac.NodeLogConfig{
		{NodeID: "node-a", ProbeType: "ilogtail", Sources: ac.JSON(`[{"type":"file","path":"/var/log/a"}]`), Version: 1},
		{NodeID: "node-b", ProbeType: "ilogtail", Sources: ac.JSON(`[{"type":"file","path":"/var/log/b"}]`), Version: 1},
	}
	if err := repo.SaveNodeLogConfigs(nil, first, nil); err != nil {
		t.Fatal(err)
	}
	second := []*ac.NodeLogConfig{
		{NodeID: "node-a", ProbeType: "ilogtail", Sources: ac.JSON(`[]`), Version: 1},
		{NodeID: "node-b", ProbeType: "ilogtail", Sources: ac.JSON(`[]`), Version: 1},
	}
	if err := repo.SaveNodeLogConfigs(nil, second, nil); err != nil {
		t.Fatal(err)
	}
	for _, node := range []string{"node-a", "node-b"} {
		stored, err := repo.GetNodeLogConfig(nil, node, "ilogtail")
		if err != nil || stored.Version != 2 || string(stored.Sources) != "[]" {
			t.Fatalf("%s stored=%+v err=%v", node, stored, err)
		}
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
		NodeID:          "node-69",
		ProbeType:       "ilogtail",
		Status:          "RUNNING",
		Runtime:         ac.JSON(`{"childPid":8}`),
		EffectiveConfig: ac.JSON(`{"configSource":"DEFAULT","configGeneration":0,"effectiveConfigFiles":[{"path":"config.yaml","content":"initial"}]}`),
		Circuit:         ac.JSON(`{"state":"NORMAL"}`),
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
	if string(stored.EffectiveConfig) != string(first.EffectiveConfig) {
		t.Fatalf("ordinary heartbeat replaced effective config: %s", stored.EffectiveConfig)
	}
}
