package agentconfig

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	integrationmodel "github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type configSemanticsRepo struct {
	*commandRepo
	nodes             map[string]ac.Node
	clusters          map[string]integrationmodel.Cluster
	baselines         map[string]ac.InitialConfig
	releases          map[string]ac.Release
	templates         map[uuid.UUID]ac.Template
	globalCircuit     *ac.GlobalCircuitConfig
	nodeCircuits      map[string]ac.NodeCircuitConfig
	nodeLogs          map[string]ac.NodeLogConfig
	namespaceConfigs  map[string]ac.NamespaceInjection
	workloadInventory map[string]ac.ClusterWorkloadInventory
	odigletReady      *bool
	states            []ac.ApplyState
	deliveries        []ac.Delivery
	batchSaveCalls    int
	lastExpected      map[string]int64
}

func newConfigSemanticsRepo() *configSemanticsRepo {
	return &configSemanticsRepo{
		commandRepo:       &commandRepo{},
		nodes:             map[string]ac.Node{},
		clusters:          map[string]integrationmodel.Cluster{},
		baselines:         map[string]ac.InitialConfig{},
		releases:          map[string]ac.Release{},
		templates:         map[uuid.UUID]ac.Template{},
		nodeCircuits:      map[string]ac.NodeCircuitConfig{},
		nodeLogs:          map[string]ac.NodeLogConfig{},
		namespaceConfigs:  map[string]ac.NamespaceInjection{},
		workloadInventory: map[string]ac.ClusterWorkloadInventory{},
	}
}

func (r *configSemanticsRepo) SaveClusterWorkloadInventory(_ core.Context, inventory *ac.ClusterWorkloadInventory) error {
	r.workloadInventory[inventory.ClusterID] = *inventory
	return nil
}

func (r *configSemanticsRepo) GetClusterWorkloadInventory(_ core.Context, clusterID string) (*ac.ClusterWorkloadInventory, error) {
	inventory, ok := r.workloadInventory[clusterID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &inventory, nil
}

func (r *configSemanticsRepo) GetNamespaceInjection(_ core.Context, clusterID string) (*ac.NamespaceInjection, error) {
	value, ok := r.namespaceConfigs[clusterID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &value, nil
}

func (r *configSemanticsRepo) SaveClusterInjectionConfig(_ core.Context, stored *ac.ClusterInjectionConfig) error {
	if !stored.Configured {
		delete(r.namespaceConfigs, stored.ClusterID)
		return nil
	}
	var config ac.NamespaceInjection
	if err := json.Unmarshal(stored.Config, &config); err != nil {
		return err
	}
	r.namespaceConfigs[stored.ClusterID] = config
	return nil
}

func (r *configSemanticsRepo) GetOdigletReadiness(core.Context, string, time.Time) (*bool, error) {
	return r.odigletReady, nil
}

func (r *configSemanticsRepo) GetAgentNode(_ core.Context, node string) (*ac.Node, error) {
	value, ok := r.nodes[node]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &value, nil
}
func (r *configSemanticsRepo) UpsertAgentNode(_ core.Context, node *ac.Node) error {
	r.nodes[node.ID] = *node
	return nil
}
func (r *configSemanticsRepo) GetCluster(_ core.Context, clusterID string) (integrationmodel.Cluster, error) {
	cluster, ok := r.clusters[clusterID]
	if !ok {
		return integrationmodel.Cluster{}, gorm.ErrRecordNotFound
	}
	return cluster, nil
}
func (r *configSemanticsRepo) ListAgentNodes(core.Context) ([]ac.Node, error) {
	result := make([]ac.Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		result = append(result, node)
	}
	return result, nil
}
func (r *configSemanticsRepo) GetAgentInitialConfig(_ core.Context, node, probe string) (*ac.InitialConfig, error) {
	value, ok := r.baselines[releaseKey(node, probe)]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &value, nil
}
func (r *configSemanticsRepo) ListAgentInitialConfigs(_ core.Context, node string) ([]ac.InitialConfig, error) {
	result := []ac.InitialConfig{}
	for _, baseline := range r.baselines {
		if baseline.NodeID == node {
			result = append(result, baseline)
		}
	}
	return result, nil
}
func (r *configSemanticsRepo) GetAgentRelease(_ core.Context, node, probe string) (*ac.Release, error) {
	value, ok := r.releases[releaseKey(node, probe)]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &value, nil
}
func (r *configSemanticsRepo) GetAgentReleaseGeneration(_ core.Context, node, probe string, generation int64) (*ac.Release, error) {
	value, ok := r.releases[releaseKey(node, probe)]
	if !ok || value.Generation != generation {
		return nil, gorm.ErrRecordNotFound
	}
	return &value, nil
}
func (r *configSemanticsRepo) ListAgentReleases(_ core.Context, node, probe string, limit int) ([]ac.Release, error) {
	value, ok := r.releases[releaseKey(node, probe)]
	if !ok || limit == 0 {
		return []ac.Release{}, nil
	}
	return []ac.Release{value}, nil
}
func (r *configSemanticsRepo) ListLatestAgentReleases(_ core.Context, node string) ([]ac.Release, error) {
	result := []ac.Release{}
	for _, release := range r.releases {
		if release.NodeID == node {
			result = append(result, release)
		}
	}
	return result, nil
}
func (r *configSemanticsRepo) GetAgentTemplate(_ core.Context, id uuid.UUID) (*ac.Template, error) {
	value, ok := r.templates[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &value, nil
}
func (r *configSemanticsRepo) CreateAgentTemplate(_ core.Context, template *ac.Template) error {
	r.templates[template.ID] = *template
	return nil
}
func (r *configSemanticsRepo) SaveAgentBindings(_ core.Context, bindings []*ac.Binding, releases []*ac.Release, expected map[string]int64, delivery *ac.Delivery) error {
	if len(bindings) != len(releases) {
		return errors.New("mismatched writes")
	}
	r.lastExpected = make(map[string]int64, len(expected))
	for key, want := range expected {
		r.lastExpected[key] = want
		actual := int64(0)
		if current, ok := r.releases[key]; ok {
			actual = current.Generation
		}
		if actual != want {
			return errors.New("generation conflict")
		}
	}
	r.batchSaveCalls++
	for _, release := range releases {
		r.releases[releaseKey(release.NodeID, release.ProbeType)] = *release
	}
	if delivery != nil {
		r.deliveries = append(r.deliveries, *delivery)
	}
	return nil
}
func (r *configSemanticsRepo) GetGlobalCircuitConfig(core.Context) (*ac.GlobalCircuitConfig, error) {
	if r.globalCircuit == nil {
		return nil, gorm.ErrRecordNotFound
	}
	copy := *r.globalCircuit
	return &copy, nil
}
func (r *configSemanticsRepo) SaveGlobalCircuitConfig(_ core.Context, config *ac.GlobalCircuitConfig) error {
	copy := *config
	r.globalCircuit = &copy
	return nil
}
func (r *configSemanticsRepo) GetNodeCircuitConfig(_ core.Context, node string) (*ac.NodeCircuitConfig, error) {
	value, ok := r.nodeCircuits[node]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &value, nil
}
func (r *configSemanticsRepo) SaveNodeCircuitConfigs(_ core.Context, configs []*ac.NodeCircuitConfig) error {
	for _, config := range configs {
		copy := *config
		if old, ok := r.nodeCircuits[config.NodeID]; ok {
			copy.Version = old.Version + 1
		}
		r.nodeCircuits[config.NodeID] = copy
	}
	return nil
}
func (r *configSemanticsRepo) GetNodeLogConfig(_ core.Context, node, probe string) (*ac.NodeLogConfig, error) {
	value, ok := r.nodeLogs[releaseKey(node, probe)]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &value, nil
}
func (r *configSemanticsRepo) SaveNodeLogConfigs(_ core.Context, configs []*ac.NodeLogConfig) error {
	for _, config := range configs {
		copy := *config
		key := releaseKey(config.NodeID, config.ProbeType)
		if old, ok := r.nodeLogs[key]; ok {
			copy.Version = old.Version + 1
		}
		r.nodeLogs[key] = copy
	}
	return nil
}
func (r *configSemanticsRepo) ListAgentStates(core.Context) ([]ac.ApplyState, error) {
	return append([]ac.ApplyState(nil), r.states...), nil
}
func (r *configSemanticsRepo) GetAgentState(_ core.Context, node, probe string) (*ac.ApplyState, error) {
	for index := range r.states {
		if r.states[index].NodeID == node && r.states[index].ProbeType == probe {
			copy := r.states[index]
			return &copy, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *configSemanticsRepo) ReportAgentState(_ core.Context, state *ac.ApplyState) error {
	for index := range r.states {
		if r.states[index].NodeID == state.NodeID && r.states[index].ProbeType == state.ProbeType {
			copy := *state
			if len(copy.EffectiveConfig) == 0 {
				copy.EffectiveConfig = r.states[index].EffectiveConfig
			}
			r.states[index] = copy
			return nil
		}
	}
	r.states = append(r.states, *state)
	return nil
}
func (r *configSemanticsRepo) ListAgentDeliveries(core.Context, int) ([]ac.Delivery, error) {
	return append([]ac.Delivery(nil), r.deliveries...), nil
}

func addBaseline(t *testing.T, repo *configSemanticsRepo, node, probe, content string) {
	t.Helper()
	files, err := json.Marshal([]ac.InitialConfigFile{{Path: "config.yaml", Content: content, ContentSHA256: ac.HashContent(content)}})
	if err != nil {
		t.Fatal(err)
	}
	repo.baselines[releaseKey(node, probe)] = ac.InitialConfig{NodeID: node, ProbeType: probe, Files: ac.JSON(files)}
}

func addBaselineFiles(t *testing.T, repo *configSemanticsRepo, node, probe string, input []ac.InitialConfigFile) {
	t.Helper()
	files := append([]ac.InitialConfigFile(nil), input...)
	for index := range files {
		files[index].ContentSHA256 = ac.HashContent(files[index].Content)
	}
	data, err := json.Marshal(files)
	if err != nil {
		t.Fatal(err)
	}
	repo.baselines[releaseKey(node, probe)] = ac.InitialConfig{NodeID: node, ProbeType: probe, Files: ac.JSON(data), ContentSHA256: ac.HashInitialConfigFiles(files)}
}

func addBundleTemplate(t *testing.T, repo *configSemanticsRepo, id uuid.UUID, configs []ac.TemplateProbeConfig, variables []ac.TemplateVariable) {
	t.Helper()
	data, err := json.Marshal(configs)
	if err != nil {
		t.Fatal(err)
	}
	repo.templates[id] = ac.Template{ID: id, Configs: ac.JSON(data), Version: 1, Variables: variables}
}

func TestInitialConfigIsHistoryVersionOneWithoutServerDelivery(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addBaseline(t, repo, "node-1", "node-agent", "sampling: 1\n")
	key := releaseKey("node-1", "node-agent")
	baseline := repo.baselines[key]
	baseline.CapturedAt = time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	repo.baselines[key] = baseline
	service := New(repo)

	current, err := service.CurrentNodeConfig(nil, "node-1", "node-agent")
	if err != nil {
		t.Fatal(err)
	}
	if current.Generation != 0 || current.Version != 1 || current.Source != "INITIAL" {
		t.Fatalf("initial current config=%+v", current)
	}
	history, err := service.ListNodeConfigHistory(nil, "node-1", "node-agent", 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Generation != 0 || history[0].Version != 1 || history[0].Source != "INITIAL" || !history[0].CreatedAt.Equal(baseline.CapturedAt) {
		t.Fatalf("initial history=%+v", history)
	}
	if len(repo.releases) != 0 || repo.batchSaveCalls != 0 {
		t.Fatalf("initial history created a server release: releases=%d saves=%d", len(repo.releases), repo.batchSaveCalls)
	}

	updated, err := service.UpdateNodeConfigs(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{
		ProbeType:          "node-agent",
		Content:            "sampling: 2\n",
		ExpectedGeneration: 0,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Configs) != 1 || updated.Configs[0].Generation != 0 || updated.Configs[0].Version != 1 || updated.Configs[0].Source != "INITIAL" {
		t.Fatalf("pending server update changed current config=%+v", updated.Configs)
	}
	if release := repo.releases[key]; release.Generation != 1 || release.RenderedContent != "sampling: 2\n" {
		t.Fatalf("first server release=%+v", release)
	}
	history, err = service.ListNodeConfigHistory(nil, "node-1", "node-agent", 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 || history[0].Version != 2 || history[0].Source != "RELEASE" || history[1].Version != 1 || history[1].Source != "INITIAL" {
		t.Fatalf("history after first server update=%+v", history)
	}
}

func TestReportPreservesInitialSnapshotAndAdvancesOnlyAppliedRelease(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addBaseline(t, repo, "node-1", "probe-a", "source: baseline\n")
	remoteFiles := []ac.InitialConfigFile{{Path: "config.yaml", Content: "source: remote\n", ContentSHA256: ac.HashContent("source: remote\n")}}
	remoteJSON, err := json.Marshal(remoteFiles)
	if err != nil {
		t.Fatal(err)
	}
	repo.releases[releaseKey("node-1", "probe-a")] = ac.Release{
		ID: uuid.New(), NodeID: "node-1", ProbeType: "probe-a", Generation: 1,
		RenderedContent: "source: remote\n", Files: ac.JSON(remoteJSON), ContentSHA256: ac.HashInitialConfigFiles(remoteFiles),
	}
	service := New(repo)

	initialRuntime := ac.JSON(`{"childPid":10,"configSource":"DEFAULT","configGeneration":0,"contentId":"","effectiveConfigFiles":[{"path":"config.yaml","content":"source: actual-default\n"}]}`)
	if err := service.Report(nil, &ac.ApplyState{NodeID: "node-1", ProbeType: "probe-a", AppliedGeneration: 0, Runtime: initialRuntime}); err != nil {
		t.Fatal(err)
	}
	if err := service.Report(nil, &ac.ApplyState{NodeID: "node-1", ProbeType: "probe-a", AppliedGeneration: 0, Runtime: ac.JSON(`{"childPid":11}`)}); err != nil {
		t.Fatal(err)
	}
	current, err := service.CurrentNodeConfig(nil, "node-1", "probe-a")
	if err != nil {
		t.Fatal(err)
	}
	if current.Generation != 0 || current.Source != "INITIAL" || current.Files[0].Content != "source: actual-default\n" {
		t.Fatalf("initial daemon snapshot was not preserved: %+v", current)
	}
	stored, err := repo.GetAgentState(nil, "node-1", "probe-a")
	if err != nil {
		t.Fatal(err)
	}
	storedFields, err := runtimeFields(stored.Runtime)
	if err != nil {
		t.Fatal(err)
	}
	var childPID int
	if err := json.Unmarshal(storedFields["childPid"], &childPID); err != nil || childPID != 11 {
		t.Fatalf("latest runtime fields were not saved: runtime=%s err=%v", stored.Runtime, err)
	}
	if _, ok := storedFields["effectiveConfigFiles"]; ok {
		t.Fatalf("effective config remained mixed into runtime: %s", stored.Runtime)
	}
	effectiveFields, err := runtimeFields(stored.EffectiveConfig)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := effectiveFields["effectiveConfigFiles"]; !ok {
		t.Fatalf("effective config snapshot was not stored separately: %s", stored.EffectiveConfig)
	}

	if err := service.Report(nil, &ac.ApplyState{NodeID: "node-1", ProbeType: "probe-a", DesiredGeneration: 1, AppliedGeneration: 1, Runtime: ac.JSON(`{"childPid":12}`)}); err != nil {
		t.Fatal(err)
	}
	current, err = service.CurrentNodeConfig(nil, "node-1", "probe-a")
	if err != nil {
		t.Fatal(err)
	}
	if current.Generation != 1 || current.Source != "RELEASE" || current.Files[0].Content != "source: remote\n" {
		t.Fatalf("applied release did not become current: %+v", current)
	}
	stored, err = repo.GetAgentState(nil, "node-1", "probe-a")
	if err != nil {
		t.Fatal(err)
	}
	storedFields, err = runtimeFields(stored.EffectiveConfig)
	if err != nil {
		t.Fatal(err)
	}
	var source string
	var generation int64
	_ = json.Unmarshal(storedFields["configSource"], &source)
	_ = json.Unmarshal(storedFields["configGeneration"], &generation)
	if source != "REMOTE" || generation != 1 {
		t.Fatalf("server did not replace the effective config snapshot: %s", stored.EffectiveConfig)
	}
}

func TestReportWithoutSnapshotBuildsGenerationZeroFromBaseline(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addBaseline(t, repo, "node-1", "probe-a", "source: baseline\n")
	service := New(repo)

	if err := service.Report(nil, &ac.ApplyState{NodeID: "node-1", ProbeType: "probe-a", Runtime: ac.JSON(`{"childRunning":true}`)}); err != nil {
		t.Fatal(err)
	}
	current, err := service.CurrentNodeConfig(nil, "node-1", "probe-a")
	if err != nil {
		t.Fatal(err)
	}
	if current.Generation != 0 || current.Files[0].Content != "source: baseline\n" {
		t.Fatalf("Generation 0 baseline was not maintained: %+v", current)
	}
}

func TestListStatesMarksExpiredReportsOffline(t *testing.T) {
	now := time.Date(2026, time.August, 26, 10, 0, 0, 0, time.UTC)
	fresh := now.Add(-agentStateFreshnessTimeout + time.Second)
	expired := now.Add(-agentStateFreshnessTimeout - time.Second)
	repo := newConfigSemanticsRepo()
	repo.states = []ac.ApplyState{
		{NodeID: "node-1", ProbeType: "fresh", Status: "RUNNING", ReportedAt: &fresh},
		{NodeID: "node-1", ProbeType: "expired", Status: "RUNNING", ReportedAt: &expired},
		{NodeID: "node-1", ProbeType: "never-reported", Status: "RUNNING"},
	}
	service := New(repo)
	service.now = func() time.Time { return now }

	states, err := service.ListStates(nil)
	if err != nil {
		t.Fatal(err)
	}
	if states[0].Status != "RUNNING" || states[0].Reason != "" {
		t.Fatalf("fresh state changed: %+v", states[0])
	}
	for _, state := range states[1:] {
		if state.Status != "OFFLINE" || state.Reason != "status report expired" {
			t.Fatalf("expired state remained healthy: %+v", state)
		}
	}
}

func TestNodeConfigUpdatePreservesCompleteMultiFileProbe(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addBaselineFiles(t, repo, "node-1", "ebpf-agent", []ac.InitialConfigFile{
		{Path: "config/camera-agent-config.yml", Content: "enabled: true\n"},
		{Path: "config/log4rs.yml", Content: "level: info\n"},
		{Path: "load-env.sh", Content: "export SNAPLEN=1000\n"},
	})
	service := New(repo)

	current, err := service.CurrentNodeConfigs(nil, "node-1")
	if err != nil || len(current.Configs) != 1 || len(current.Configs[0].Files) != 3 || current.Configs[0].Content != "" {
		t.Fatalf("unexpected multi-file baseline: configs=%+v err=%v", current.Configs, err)
	}
	files := append([]ac.InitialConfigFile(nil), current.Configs[0].Files...)
	files[1].Content = "enable: changed\n"
	updated, err := service.UpdateNodeConfigs(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{ProbeType: "ebpf-agent", Files: files, ExpectedGeneration: 0}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Configs[0].Files) != 3 || updated.Configs[0].Generation != 0 {
		t.Fatalf("pending multi-file release changed current config: %+v", updated.Configs[0])
	}
	release := repo.releases[releaseKey("node-1", "ebpf-agent")]
	var persisted []ac.InitialConfigFile
	if err := json.Unmarshal(release.Files, &persisted); err != nil || len(persisted) != 3 {
		t.Fatalf("persisted files=%+v err=%v", persisted, err)
	}
	if release.ContentSHA256 != ac.HashInitialConfigFiles(persisted) {
		t.Fatalf("bundle sha=%s want=%s", release.ContentSHA256, ac.HashInitialConfigFiles(persisted))
	}
	if err := service.Report(nil, &ac.ApplyState{NodeID: "node-1", ProbeType: "ebpf-agent", AppliedGeneration: 1, Runtime: ac.JSON(`{"childRunning":true}`)}); err != nil {
		t.Fatal(err)
	}

	_, err = service.UpdateNodeConfigs(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{ProbeType: "ebpf-agent", Files: persisted[:2], ExpectedGeneration: 1}}})
	if err == nil || !strings.Contains(err.Error(), "cannot be added, removed, or renamed") || repo.batchSaveCalls != 1 {
		t.Fatalf("changed registered file set was accepted: err=%v calls=%d", err, repo.batchSaveCalls)
	}
}

func TestImportedTemplateCannotChangeRegisteredFileSet(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addBaselineFiles(t, repo, "node-1", "ebpf-agent", []ac.InitialConfigFile{
		{Path: "config.yaml", Content: "old: true\n"},
		{Path: "logging.yaml", Content: "level: info\n"},
	})
	service := New(repo)
	template, err := service.ImportTemplate(nil, TemplateImportInput{YAML: `
name: compact-ebpf
description: imported template
variables:
  - name: LEVEL
    required: true
configs:
  - probeType: ebpf-agent
    files:
      - path: runtime.yaml
        template: |
          level: {% LEVEL %}
`}, 7)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Assign(nil, "node-1", "ebpf-agent", AssignmentInput{
		TemplateID: template.ID,
		Variables:  map[string]string{"LEVEL": "debug"},
	})
	if err == nil || !strings.Contains(err.Error(), "do not match registered configuration files") || repo.batchSaveCalls != 0 {
		t.Fatalf("imported template changed the registered file set: err=%v calls=%d", err, repo.batchSaveCalls)
	}
}

func TestV103ExampleTemplateIsImportable(t *testing.T) {
	data, err := os.ReadFile("../../../../docs/examples/agent-config-template-v1.0.3.yaml")
	if err != nil {
		t.Fatal(err)
	}
	template, err := New(newConfigSemanticsRepo()).ImportTemplate(nil, TemplateImportInput{YAML: string(data)}, 7)
	if err != nil {
		t.Fatal(err)
	}
	configs, err := templateProbeConfigs(template)
	if err != nil {
		t.Fatal(err)
	}
	fileCount := 0
	foundCollectorCluster := false
	foundVMAgentCluster := false
	foundOriginxEnv := false
	for _, config := range configs {
		fileCount += len(config.Files)
		for _, file := range config.Files {
			if config.ProbeType == "originx-gc-agent" && file.Path == "originx-gc-agent.env" &&
				strings.Contains(file.Content, "RECEIVER_HOST=") && strings.Contains(file.Content, "RECEIVER_PORT=") {
				foundOriginxEnv = true
			}
			if !strings.Contains(file.Content, "{% CLUSTER_ID %}") {
				continue
			}
			if config.ProbeType == "apo-otel-collector-agent" {
				foundCollectorCluster = true
			}
			if config.ProbeType == "vmagent" {
				foundVMAgentCluster = true
			}
		}
	}
	if len(configs) != 9 || fileCount != 11 || !foundCollectorCluster || !foundVMAgentCluster || !foundOriginxEnv {
		t.Fatalf("unexpected v1.0.3 example: configs=%d files=%d collectorCluster=%t vmagentCluster=%t originxEnv=%t", len(configs), fileCount, foundCollectorCluster, foundVMAgentCluster, foundOriginxEnv)
	}
	if len(template.Variables) != 1 || template.Variables[0].Name != "CLUSTER_ID" || !template.Variables[0].Required {
		t.Fatalf("unexpected template variables: %+v", template.Variables)
	}
	if template.ClusterType != templateClusterTypeKubernetes {
		t.Fatalf("example cluster type=%q", template.ClusterType)
	}
}

func TestVMV110ExampleTemplateIsImportable(t *testing.T) {
	data, err := os.ReadFile("../../../../docs/examples/agent-config-template-vm-v1.1.0.yaml")
	if err != nil {
		t.Fatal(err)
	}
	repo := newConfigSemanticsRepo()
	service := New(repo)
	template, err := service.ImportTemplate(nil, TemplateImportInput{YAML: string(data)}, 7)
	if err != nil {
		t.Fatal(err)
	}
	configs, err := templateProbeConfigs(template)
	if err != nil {
		t.Fatal(err)
	}
	wantFiles := map[string][]string{
		"alloy":          {"config.alloy"},
		"ebpf-agent":     {"config/camera-agent-config.yml", "config/log4rs.yml", "load-env.sh"},
		"loongcollector": {"config/local/all-log-config.yaml", "config/local/config.yaml"},
		"node-agent":     {"config.yaml"},
		"preload":        {"libapoinstrument.conf"},
		"vmagent":        {"config/vmagent/scrape.yml"},
	}
	if len(configs) != len(wantFiles) {
		t.Fatalf("unexpected VM example config count: got %d want %d", len(configs), len(wantFiles))
	}
	for _, config := range configs {
		paths, ok := wantFiles[config.ProbeType]
		if !ok {
			t.Fatalf("unexpected VM example probe: %s", config.ProbeType)
		}
		if len(config.Files) != len(paths) {
			t.Fatalf("unexpected VM example files for %s: %+v", config.ProbeType, config.Files)
		}
		for index, file := range config.Files {
			if file.Path != paths[index] {
				t.Fatalf("unexpected VM example file for %s: got %s want %s", config.ProbeType, file.Path, paths[index])
			}
		}
	}
	if template.ClusterType != templateClusterTypeVM {
		t.Fatalf("VM example cluster type=%q", template.ClusterType)
	}
	wantVariables := map[string]struct {
		required bool
		fallback string
	}{
		"APO_APP_RED_TRACE_ENABLED": {fallback: "true"},
		"APO_PROXY_IP":              {required: true},
		"CLUSTER_ID":                {required: true},
		"HOST_IP":                   {required: true},
	}
	if len(template.Variables) != len(wantVariables) {
		t.Fatalf("unexpected VM example variables: %+v", template.Variables)
	}
	for _, variable := range template.Variables {
		want, ok := wantVariables[variable.Name]
		if !ok || variable.Required != want.required || variable.DefaultValue != want.fallback {
			t.Fatalf("unexpected VM example variable: %+v", variable)
		}
	}
	repo.nodes["node-vm"] = ac.Node{
		ID:          "node-vm",
		ClusterID:   "cluster-vm",
		Environment: "binary",
		Labels:      ac.JSON(`{"HOST_IP":"192.168.1.43","APO_PROXY_IP":"192.168.1.69"}`),
	}
	addBaselineFiles(t, repo, "node-vm", "vmagent", []ac.InitialConfigFile{{Path: "config/vmagent/scrape.yml", Content: "old\n"}})
	release, err := service.Assign(nil, "node-vm", "vmagent", AssignmentInput{TemplateID: template.ID})
	if err != nil {
		t.Fatal(err)
	}
	var rendered []ac.InitialConfigFile
	if err := json.Unmarshal(release.Files, &rendered); err != nil {
		t.Fatal(err)
	}
	if len(rendered) != 1 || strings.Contains(rendered[0].Content, "{% CLUSTER_ID %}") ||
		!strings.Contains(rendered[0].Content, `cluster_id: "cluster-vm"`) {
		t.Fatalf("VM example was not rendered with registered node variables: %+v", rendered)
	}
}

func TestTemplateClusterTypeImportNormalization(t *testing.T) {
	tests := []struct {
		name        string
		clusterType string
		want        string
	}{
		{name: "legacy default", want: templateClusterTypeKubernetes},
		{name: "k8s alias", clusterType: "k8s", want: templateClusterTypeKubernetes},
		{name: "binary alias", clusterType: "binary", want: templateClusterTypeVM},
		{name: "vm", clusterType: "vm", want: templateClusterTypeVM},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			template, err := New(newConfigSemanticsRepo()).ImportTemplate(nil, TemplateImportInput{YAML: `
name: categorized
clusterType: ` + test.clusterType + `
configs:
  - probeType: probe-a
    files:
      - path: config.yaml
        template: |
          enabled: true
`}, 7)
			if err != nil {
				t.Fatal(err)
			}
			if template.ClusterType != test.want {
				t.Fatalf("cluster type=%q, want %q", template.ClusterType, test.want)
			}
		})
	}

	_, err := New(newConfigSemanticsRepo()).ImportTemplate(nil, TemplateImportInput{YAML: `
name: invalid
clusterType: bare-metal
configs:
  - probeType: probe-a
    files:
      - path: config.yaml
        content: enabled
`}, 7)
	if err == nil || !strings.Contains(err.Error(), "unsupported template cluster type") {
		t.Fatalf("unsupported cluster type was accepted: %v", err)
	}
}

func TestImportedMultiFileTemplateRendersAssignmentEndToEnd(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addBaselineFiles(t, repo, "node-1", "ebpf-agent", []ac.InitialConfigFile{
		{Path: "config/camera-agent-config.yml", Content: "legacy camera config\n"},
		{Path: "config/log4rs.yml", Content: "legacy logging config\n"},
		{Path: "load-env.sh", Content: "legacy environment\n"},
	})
	service := New(repo)

	template, err := service.ImportTemplate(nil, TemplateImportInput{YAML: `
name: ebpf-runtime
description: multi-file runtime configuration
variables:
  - name: LEVEL
    required: true
    validationRegex: "^(debug|info)$"
  - name: ENDPOINT
    defaultValue: http://collector:4317
configs:
  - probeType: ebpf-agent
    files:
      - path: config/camera-agent-config.yml
        template: |
          log_level: {% LEVEL %}
          endpoint: {% ENDPOINT %}
      - path: config/log4rs.yml
        template: |
          level: {% LEVEL %}
      - path: load-env.sh
        template: |
          export LD_LIBRARY_PATH=${LD_LIBRARY_PATH:-}
          export OTEL_ENDPOINT={% ENDPOINT %}
`}, 7)
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Assign(nil, "node-1", "ebpf-agent", AssignmentInput{TemplateID: template.ID})
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) || len(validationErr.MissingVariables) != 1 || validationErr.MissingVariables[0] != "LEVEL" {
		t.Fatalf("missing required variable was not rejected: %v", err)
	}

	release, err := service.Assign(nil, "node-1", "ebpf-agent", AssignmentInput{
		TemplateID: template.ID,
		Variables:  map[string]string{"LEVEL": "debug"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var files []ac.InitialConfigFile
	if err := json.Unmarshal(release.Files, &files); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"config/camera-agent-config.yml": "log_level: debug\nendpoint: http://collector:4317\n",
		"config/log4rs.yml":              "level: debug\n",
		"load-env.sh":                    "export LD_LIBRARY_PATH=${LD_LIBRARY_PATH:-}\nexport OTEL_ENDPOINT=http://collector:4317\n",
	}
	if len(files) != len(want) || release.Generation != 1 {
		t.Fatalf("unexpected release: generation=%d files=%#v", release.Generation, files)
	}
	for _, file := range files {
		if file.Content != want[file.Path] {
			t.Fatalf("unexpected rendered file %s: %q", file.Path, file.Content)
		}
	}
	if release.ContentSHA256 != ac.HashInitialConfigFiles(files) {
		t.Fatalf("bundle checksum=%s", release.ContentSHA256)
	}
	history, err := service.ListNodeConfigHistory(nil, "node-1", "ebpf-agent", 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 || history[0].Generation != 1 || history[1].Generation != 0 {
		t.Fatalf("unexpected version history: %#v", history)
	}
}

func TestNodeConfigUpdateSavesAllProbeChangesTogether(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addBaseline(t, repo, "node-1", "probe-a", "a: 1\n")
	addBaseline(t, repo, "node-1", "probe-b", "b: 1\n")
	service := New(repo)

	updated, err := service.UpdateNodeConfigs(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{
		{ProbeType: "probe-a", Content: "a: 2\n", ExpectedGeneration: 0},
		{ProbeType: "probe-b", Content: "b: 2\n", ExpectedGeneration: 0},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if repo.batchSaveCalls != 1 || len(updated.Configs) != 2 {
		t.Fatalf("node update was not one multi-probe save: calls=%d configs=%+v", repo.batchSaveCalls, updated.Configs)
	}
	for _, config := range updated.Configs {
		if config.Generation != 0 || config.Source != "INITIAL" {
			t.Fatalf("pending release changed current config: %+v", config)
		}
	}
	for _, probe := range []string{"probe-a", "probe-b"} {
		if err := service.Report(nil, &ac.ApplyState{NodeID: "node-1", ProbeType: probe, AppliedGeneration: 1, Runtime: ac.JSON(`{"childRunning":true}`)}); err != nil {
			t.Fatal(err)
		}
	}

	updated, err = service.UpdateNodeConfigs(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{
		{ProbeType: "probe-a", Content: "a: 3\n", ExpectedGeneration: 1},
		{ProbeType: "probe-b", Content: "b: 2\n", ExpectedGeneration: 1},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if repo.batchSaveCalls != 2 || len(repo.lastExpected) != 2 {
		t.Fatalf("node update did not recheck every submitted generation: calls=%d expected=%v", repo.batchSaveCalls, repo.lastExpected)
	}

	before := repo.batchSaveCalls
	_, err = service.UpdateNodeConfigs(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{
		{ProbeType: "probe-a", Content: "a: stale\n", ExpectedGeneration: 0},
		{ProbeType: "probe-b", Content: "b: 2\n", ExpectedGeneration: 1},
	}})
	if err == nil || repo.batchSaveCalls != before {
		t.Fatalf("stale generation was not rejected before saving: err=%v calls=%d", err, repo.batchSaveCalls)
	}
	_, err = service.UpdateNodeConfigs(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{ProbeType: "probe-a", Content: "a: 4\n", ExpectedGeneration: 2}}})
	if err == nil || repo.batchSaveCalls != before {
		t.Fatalf("partial node update was not rejected: err=%v calls=%d", err, repo.batchSaveCalls)
	}
}

func TestNodeConfigExpectedGenerationAdvancesBeforeApply(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addBaseline(t, repo, "node-1", "probe-a", "a: 1\n")
	service := New(repo)

	updated, err := service.UpdateNodeConfigGroup(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{
		ProbeType: "probe-a", Content: "a: 2\n", ExpectedGeneration: 0,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Configs) != 1 || updated.Configs[0].Generation != 0 || updated.Configs[0].ExpectedGeneration != 1 {
		t.Fatalf("first pending release returned invalid generations: %+v", updated.Configs)
	}

	updated, err = service.UpdateNodeConfigGroup(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{
		ProbeType: "probe-a", Content: "a: 3\n", ExpectedGeneration: 1,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if release := repo.releases[releaseKey("node-1", "probe-a")]; release.Generation != 2 {
		t.Fatalf("second release generation=%d, want 2", release.Generation)
	}
	if updated.Configs[0].Generation != 0 || updated.Configs[0].ExpectedGeneration != 2 {
		t.Fatalf("second pending release returned invalid generations: %+v", updated.Configs[0])
	}

	before := repo.batchSaveCalls
	if _, err = service.UpdateNodeConfigGroup(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{
		ProbeType: "probe-a", Content: "a: stale\n", ExpectedGeneration: 1,
	}}}); err == nil || repo.batchSaveCalls != before {
		t.Fatalf("stale expected generation was accepted: err=%v calls=%d", err, repo.batchSaveCalls)
	}
}

func TestInjectionControlIsSeparateFromConfigReleaseGroup(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addBaseline(t, repo, "node-1", "probe-a", "a: 1\n")
	addBaselineFiles(t, repo, "node-1", "preload", []ac.InitialConfigFile{{
		Path: "libapoinstrument.conf", Content: "[python]\nPYTHONPATH=/opt/apo\n",
	}})
	service := New(repo)

	updated, err := service.UpdateNodeConfigGroup(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{
		ProbeType:          "probe-a",
		Content:            "a: 2\n",
		ExpectedGeneration: 0,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if repo.batchSaveCalls != 1 || len(repo.lastExpected) != 1 {
		t.Fatalf("group update was not isolated: calls=%d expected=%v", repo.batchSaveCalls, repo.lastExpected)
	}
	for _, config := range updated.Configs {
		switch config.ProbeType {
		case "preload":
			if config.Generation != 0 || config.Files[0].Content != "[python]\nPYTHONPATH=/opt/apo\n" {
				t.Fatalf("ordinary config group changed injection control: %+v", config)
			}
		case "probe-a":
			if config.Generation != 0 || config.Files[0].Content != "a: 1\n" {
				t.Fatalf("pending ordinary config changed current config: %+v", config)
			}
		}
	}

	before := repo.batchSaveCalls
	_, err = service.UpdateNodeConfigGroup(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{
		ProbeType: "preload", Content: "; python disabled\n", ExpectedGeneration: 0,
	}}})
	if err == nil || repo.batchSaveCalls != before {
		t.Fatalf("ordinary config API accepted injection control: err=%v calls=%d", err, repo.batchSaveCalls)
	}
	_, err = service.UpdateNodeConfigs(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{
		{ProbeType: "probe-a", Content: "a: 3\n", ExpectedGeneration: 1},
		{ProbeType: "preload", Content: "; python disabled\n", ExpectedGeneration: 0},
	}})
	if err == nil || repo.batchSaveCalls != before {
		t.Fatalf("legacy config API accepted injection control: err=%v calls=%d", err, repo.batchSaveCalls)
	}

	control, err := service.GetNodeInjectionControl(nil, "node-1")
	if err != nil {
		t.Fatal(err)
	}
	if !control.Available || control.Generation != 0 || control.FilePath != "libapoinstrument.conf" {
		t.Fatalf("unexpected initial injection control: %+v", control)
	}
	saved, err := service.SaveNodeInjectionControl(nil, "node-1", NodeInjectionControlInput{
		Content: "; python disabled\n", ExpectedGeneration: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Generation != 1 || saved.Content != "; python disabled\n" || repo.batchSaveCalls != before+1 {
		t.Fatalf("injection control did not return its latest release: control=%+v calls=%d", saved, repo.batchSaveCalls)
	}
	saved, err = service.SaveNodeInjectionControl(nil, "node-1", NodeInjectionControlInput{
		Content: "; python disabled\n", ExpectedGeneration: 1,
	})
	if err != nil || saved.Generation != 1 || repo.batchSaveCalls != before+1 {
		t.Fatalf("unchanged injection control created another release: control=%+v err=%v calls=%d", saved, err, repo.batchSaveCalls)
	}
	saved, err = service.SaveNodeInjectionControl(nil, "node-1", NodeInjectionControlInput{
		Content: "[node]\nenabled=true\n", ExpectedGeneration: 1,
	})
	if err != nil || saved.Generation != 2 || saved.Content != "[node]\nenabled=true\n" || repo.batchSaveCalls != before+2 {
		t.Fatalf("second injection release did not advance before apply: control=%+v err=%v calls=%d", saved, err, repo.batchSaveCalls)
	}
	_, err = service.SaveNodeInjectionControl(nil, "node-1", NodeInjectionControlInput{
		Content: "stale\n", ExpectedGeneration: 1,
	})
	if err == nil || repo.batchSaveCalls != before+2 {
		t.Fatalf("stale injection control was not rejected: err=%v calls=%d", err, repo.batchSaveCalls)
	}
	if _, err = service.ListNodeConfigHistory(nil, "node-1", "preload", 20); err == nil {
		t.Fatal("injection control exposed configuration history")
	}
}

func TestBatchInjectionControlAtomicallyUpdatesAllTargetNodes(t *testing.T) {
	repo := newConfigSemanticsRepo()
	for _, node := range []ac.Node{
		{ID: "node-a", ClusterID: "cluster-a"},
		{ID: "node-b", ClusterID: "cluster-b"},
	} {
		repo.nodes[node.ID] = node
		addBaselineFiles(t, repo, node.ID, "preload", []ac.InitialConfigFile{{
			Path: "preload/.apo-config/libapoinstrument.conf", Content: "[java]\nenabled=true\n",
		}})
	}
	service := New(repo)

	result, err := service.BatchSaveNodeInjectionControl(nil, BatchNodeInjectionControlInput{
		ClusterIDs: []string{"cluster-a", "cluster-b"},
		Content:    "[python]\nenabled=true\n",
		ExpectedGenerations: map[string]int64{
			"node-a": 0,
			"node-b": 0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.TargetNodeCount != 2 || result.SucceededCount != 2 || repo.batchSaveCalls != 1 || len(repo.lastExpected) != 2 {
		t.Fatalf("unexpected atomic injection result: result=%+v calls=%d expected=%v", result, repo.batchSaveCalls, repo.lastExpected)
	}
	for _, node := range []string{"node-a", "node-b"} {
		release := repo.releases[releaseKey(node, "preload")]
		if release.Generation != 1 || release.RenderedContent != "[python]\nenabled=true\n" {
			t.Fatalf("%s injection control was not updated: %+v", node, release)
		}
	}
}

func TestBatchInjectionControlRejectsIncompleteOrStaleGenerationsBeforeWriting(t *testing.T) {
	for _, test := range []struct {
		name      string
		expected  map[string]int64
		wantError string
	}{
		{
			name:      "missing node generation",
			expected:  map[string]int64{"node-a": 0},
			wantError: "expected generation for node node-b is required",
		},
		{
			name:      "stale node generation",
			expected:  map[string]int64{"node-a": 0, "node-b": 1},
			wantError: "generation conflict for node-b/preload",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := newConfigSemanticsRepo()
			for _, node := range []ac.Node{
				{ID: "node-a", ClusterID: "cluster-a"},
				{ID: "node-b", ClusterID: "cluster-b"},
			} {
				repo.nodes[node.ID] = node
				addBaselineFiles(t, repo, node.ID, "preload", []ac.InitialConfigFile{{
					Path: "libapoinstrument.conf", Content: "[java]\nenabled=true\n",
				}})
			}
			service := New(repo)

			_, err := service.BatchSaveNodeInjectionControl(nil, BatchNodeInjectionControlInput{
				ClusterIDs:          []string{"cluster-a", "cluster-b"},
				Content:             "[python]\nenabled=true\n",
				ExpectedGenerations: test.expected,
			})
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.batchSaveCalls != 0 || len(repo.releases) != 0 {
				t.Fatalf("invalid batch wrote partial state: calls=%d releases=%v", repo.batchSaveCalls, repo.releases)
			}
		})
	}
}

func TestBatchInjectionControlRejectsUnregisteredTargetBeforeWriting(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-a"] = ac.Node{ID: "node-a", ClusterID: "cluster-a"}
	repo.nodes["node-b"] = ac.Node{ID: "node-b", ClusterID: "cluster-b"}
	addBaselineFiles(t, repo, "node-a", "preload", []ac.InitialConfigFile{{
		Path: "libapoinstrument.conf", Content: "[java]\nenabled=true\n",
	}})
	addBaseline(t, repo, "node-b", "probe-a", "enabled: true\n")
	service := New(repo)

	_, err := service.BatchSaveNodeInjectionControl(nil, BatchNodeInjectionControlInput{
		ClusterIDs: []string{"cluster-a", "cluster-b"},
		Content:    "[python]\nenabled=true\n",
		ExpectedGenerations: map[string]int64{
			"node-a": 0,
			"node-b": 0,
		},
	})
	if err == nil || !strings.Contains(err.Error(), "node node-b has no registered preload injection configuration") {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.batchSaveCalls != 0 || len(repo.releases) != 0 {
		t.Fatalf("unregistered target caused partial write: calls=%d releases=%v", repo.batchSaveCalls, repo.releases)
	}
}

func TestBatchAssignmentUsesOneCompleteTemplateAndAtomicallyOverwritesTargets(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-a"] = ac.Node{ID: "node-a", ClusterID: "cluster-a"}
	repo.nodes["node-b"] = ac.Node{ID: "node-b", ClusterID: "cluster-b"}
	repo.nodes["node-c"] = ac.Node{ID: "node-c", ClusterID: "cluster-c"}
	for index, node := range []string{"node-a", "node-b", "node-c"} {
		addBaseline(t, repo, node, "probe-a", "old-a\n")
		addBaseline(t, repo, node, "probe-b", "old-b\n")
		repo.releases[releaseKey(node, "probe-a")] = ac.Release{NodeID: node, ProbeType: "probe-a", Generation: int64(index + 3), RenderedContent: "node-specific\n", LogSources: ac.JSON("null")}
		repo.releases[releaseKey(node, "probe-b")] = ac.Release{NodeID: node, ProbeType: "probe-b", Generation: int64(index + 6), RenderedContent: "node-specific-b\n", LogSources: ac.JSON("null")}
	}
	templateID := uuid.New()
	addBundleTemplate(t, repo, templateID, []ac.TemplateProbeConfig{
		{ProbeType: "probe-a", Files: []ac.InitialConfigFile{{Path: "config.yaml", Content: "value: {% VALUE %}\n"}}},
		{ProbeType: "probe-b", Files: []ac.InitialConfigFile{{Path: "config.yaml", Content: "enabled: {% ENABLED %}\n"}}},
	}, []ac.TemplateVariable{{Name: "VALUE", Required: true}, {Name: "ENABLED", Required: true}})
	service := New(repo)

	result, err := service.BatchAssign(nil, BatchAssignmentInput{
		ClusterIDs: []string{"cluster-a", "cluster-b"},
		TemplateID: templateID,
		Variables:  map[string]string{"VALUE": "shared", "ENABLED": "true"},
	}, 7)
	if err != nil {
		t.Fatal(err)
	}
	if result.TargetNodeCount != 2 || result.SucceededCount != 2 || repo.batchSaveCalls != 1 || len(repo.lastExpected) != 4 {
		t.Fatalf("unexpected atomic batch result: result=%+v calls=%d expected=%v", result, repo.batchSaveCalls, repo.lastExpected)
	}
	for _, node := range []string{"node-a", "node-b"} {
		if got := repo.releases[releaseKey(node, "probe-a")].RenderedContent; got != "value: shared\n" {
			t.Fatalf("%s retained node-specific content: %q", node, got)
		}
		if got := repo.releases[releaseKey(node, "probe-b")].RenderedContent; got != "enabled: true\n" {
			t.Fatalf("%s retained node-specific probe-b content: %q", node, got)
		}
	}
	if got := repo.releases[releaseKey("node-c", "probe-a")].RenderedContent; got != "node-specific\n" {
		t.Fatalf("untargeted node was changed: %q", got)
	}
}

func TestBatchAssignmentRendersClusterIDFromTargetNode(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-a"] = ac.Node{ID: "node-a", ClusterID: "cluster-a"}
	addBaseline(t, repo, "node-a", "vmagent", "cluster_id: old\n")
	templateID := uuid.New()
	addBundleTemplate(t, repo, templateID, []ac.TemplateProbeConfig{{
		ProbeType: "vmagent",
		Files: []ac.InitialConfigFile{{
			Path:    "config.yaml",
			Content: "cluster_id: {% CLUSTER_ID %}\n",
		}},
	}}, []ac.TemplateVariable{{Name: "CLUSTER_ID", Required: true}})
	service := New(repo)

	_, err := service.BatchAssign(nil, BatchAssignmentInput{
		ClusterIDs: []string{"cluster-a"},
		TemplateID: templateID,
		Variables:  map[string]string{"CLUSTER_ID": "forged-cluster"},
	}, 7)
	if err != nil {
		t.Fatal(err)
	}
	if got := repo.releases[releaseKey("node-a", "vmagent")].RenderedContent; got != "cluster_id: cluster-a\n" {
		t.Fatalf("cluster ID was not rendered from the target node: %q", got)
	}
}

func TestBatchAssignmentOverwritesCompleteMultiFileProbeFromTemplate(t *testing.T) {
	repo := newConfigSemanticsRepo()
	paths := []string{
		"config/camera-agent-config.yml",
		"config/log4rs.yml",
		"load-env.sh",
	}
	for _, node := range []string{"node-a", "node-b"} {
		repo.nodes[node] = ac.Node{ID: node, ClusterID: "cluster-a"}
		addBaselineFiles(t, repo, node, "ebpf-agent", []ac.InitialConfigFile{
			{Path: paths[0], Content: "old logger\n"},
			{Path: paths[1], Content: "old all\n"},
			{Path: paths[2], Content: "old file\n"},
		})
	}
	templateFiles := []ac.InitialConfigFile{
		{Path: paths[0], Content: "level: {% LEVEL %}\n"},
		{Path: paths[1], Content: "all: {% LEVEL %}\n"},
		{Path: paths[2], Content: "file: {% LEVEL %}\n"},
	}
	for index := range templateFiles {
		templateFiles[index].ContentSHA256 = ac.HashContent(templateFiles[index].Content)
	}
	templateID := uuid.New()
	addBundleTemplate(t, repo, templateID, []ac.TemplateProbeConfig{{ProbeType: "ebpf-agent", Files: templateFiles}}, []ac.TemplateVariable{{Name: "LEVEL", Required: true}})
	service := New(repo)

	result, err := service.BatchAssign(nil, BatchAssignmentInput{
		ClusterIDs: []string{"cluster-a"},
		TemplateID: templateID,
		Variables:  map[string]string{"LEVEL": "info"},
	}, 7)
	if err != nil {
		t.Fatal(err)
	}
	if result.TargetNodeCount != 2 || result.SucceededCount != 2 || repo.batchSaveCalls != 1 {
		t.Fatalf("unexpected batch result: result=%+v calls=%d", result, repo.batchSaveCalls)
	}
	for _, node := range []string{"node-a", "node-b"} {
		release := repo.releases[releaseKey(node, "ebpf-agent")]
		var files []ac.InitialConfigFile
		if err := json.Unmarshal(release.Files, &files); err != nil {
			t.Fatalf("decode %s release: %v", node, err)
		}
		if len(files) != 3 || files[0].Content == "old logger\n" {
			t.Fatalf("%s did not receive the complete template: %+v", node, files)
		}
		for _, file := range files {
			if !strings.Contains(file.Content, "info") {
				t.Fatalf("%s file %s was not rendered: %q", node, file.Path, file.Content)
			}
		}
		if release.ContentSHA256 != ac.HashInitialConfigFiles(files) {
			t.Fatalf("%s bundle checksum=%s", node, release.ContentSHA256)
		}
	}
}

func TestBatchAssignmentRejectsChangedRegisteredFileSet(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-a"] = ac.Node{ID: "node-a", ClusterID: "cluster-a"}
	addBaselineFiles(t, repo, "node-a", "ebpf-agent", []ac.InitialConfigFile{
		{Path: "a.yaml", Content: "a: old\n"},
		{Path: "b.yaml", Content: "b: old\n"},
	})
	templateID := uuid.New()
	addBundleTemplate(t, repo, templateID, []ac.TemplateProbeConfig{{ProbeType: "ebpf-agent", Files: []ac.InitialConfigFile{{Path: "a.yaml", Content: "a: new\n"}}}}, nil)
	service := New(repo)

	_, err := service.BatchAssign(nil, BatchAssignmentInput{ClusterIDs: []string{"cluster-a"}, TemplateID: templateID}, 7)
	if err == nil || !strings.Contains(err.Error(), "do not match registered configuration files") || repo.batchSaveCalls != 0 {
		t.Fatalf("batch assignment changed the registered file set: err=%v calls=%d", err, repo.batchSaveCalls)
	}
}

func TestTemplateAssignmentRejectsDifferentClusterTypeBeforeWriting(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-k8s"] = ac.Node{ID: "node-k8s", ClusterID: "cluster-k8s", Environment: "kubernetes"}
	repo.nodes["node-vm"] = ac.Node{ID: "node-vm", ClusterID: "cluster-vm", Environment: "binary"}
	for _, node := range []string{"node-k8s", "node-vm"} {
		addBaseline(t, repo, node, "probe-a", "old: true\n")
	}
	templateID := uuid.New()
	addBundleTemplate(t, repo, templateID, []ac.TemplateProbeConfig{{
		ProbeType: "probe-a",
		Files:     []ac.InitialConfigFile{{Path: "config.yaml", Content: "new: true\n"}},
	}}, nil)
	template := repo.templates[templateID]
	template.ClusterType = templateClusterTypeVM
	repo.templates[templateID] = template
	service := New(repo)

	_, err := service.Assign(nil, "node-k8s", "probe-a", AssignmentInput{TemplateID: templateID})
	if err == nil || !strings.Contains(err.Error(), "does not match node node-k8s") || repo.batchSaveCalls != 0 {
		t.Fatalf("cross-category assignment was not rejected before writing: err=%v calls=%d", err, repo.batchSaveCalls)
	}

	_, err = service.BatchAssign(nil, BatchAssignmentInput{
		ClusterIDs: []string{"cluster-k8s", "cluster-vm"},
		TemplateID: templateID,
	}, 7)
	if err == nil || !strings.Contains(err.Error(), "does not match node node-k8s") || repo.batchSaveCalls != 0 {
		t.Fatalf("mixed-category batch was not rejected before writing: err=%v calls=%d", err, repo.batchSaveCalls)
	}
}

func TestTemplateUpdateCannotChangeRegisteredFileSet(t *testing.T) {
	repo := newConfigSemanticsRepo()
	templateID := uuid.New()
	addBundleTemplate(t, repo, templateID, []ac.TemplateProbeConfig{{
		ProbeType: "ebpf-agent",
		Files: []ac.InitialConfigFile{
			{Path: "config.yaml", Content: "level: info\n"},
			{Path: "logging.yaml", Content: "output: stdout\n"},
		},
	}}, nil)
	service := New(repo)

	_, err := service.UpdateTemplate(nil, templateID, TemplateInput{
		Name: "updated",
		Configs: []ac.TemplateProbeConfig{{
			ProbeType: "ebpf-agent",
			Files: []ac.InitialConfigFile{
				{Path: "config.yaml", Content: "level: debug\n"},
				{Path: "logging.yaml", Content: "output: file\n"},
			},
		}},
	}, 7)
	if err != nil {
		t.Fatalf("template content update was rejected: %v", err)
	}

	_, err = service.UpdateTemplate(nil, templateID, TemplateInput{
		Name: "invalid",
		Configs: []ac.TemplateProbeConfig{{
			ProbeType: "ebpf-agent",
			Files:     []ac.InitialConfigFile{{Path: "config.yaml", Content: "level: debug\n"}},
		}},
	}, 7)
	if err == nil || !strings.Contains(err.Error(), "template registration cannot be changed") {
		t.Fatalf("template file removal was accepted: %v", err)
	}

	_, err = service.UpdateTemplate(nil, templateID, TemplateInput{
		Name:        "invalid-category",
		ClusterType: templateClusterTypeVM,
		Configs: []ac.TemplateProbeConfig{{
			ProbeType: "ebpf-agent",
			Files: []ac.InitialConfigFile{
				{Path: "config.yaml", Content: "level: debug\n"},
				{Path: "logging.yaml", Content: "output: file\n"},
			},
		}},
	}, 7)
	if err == nil || !strings.Contains(err.Error(), "cluster type cannot be changed") {
		t.Fatalf("template cluster type change was accepted: %v", err)
	}
}

func TestBatchAssignmentRejectsTemplateWithoutEveryNodeProbe(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-a"] = ac.Node{ID: "node-a", ClusterID: "cluster-a"}
	addBaseline(t, repo, "node-a", "probe-a", "a: old\n")
	addBaseline(t, repo, "node-a", "probe-b", "b: old\n")
	templateA := uuid.New()
	addBundleTemplate(t, repo, templateA, []ac.TemplateProbeConfig{{ProbeType: "probe-a", Files: []ac.InitialConfigFile{{Path: "config.yaml", Content: "a: new\n"}}}}, nil)
	service := New(repo)

	_, err := service.BatchAssign(nil, BatchAssignmentInput{
		ClusterIDs: []string{"cluster-a"},
		TemplateID: templateA,
	}, 7)
	if err == nil || !strings.Contains(err.Error(), "contains 1 probes but node node-a has 2") {
		t.Fatalf("incomplete template set was accepted: %v", err)
	}
	if repo.batchSaveCalls != 0 {
		t.Fatalf("incomplete batch wrote releases: calls=%d", repo.batchSaveCalls)
	}
}

func TestCircuitPolicyIsNodeLevelAndIndependentFromConfigReleases(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addBaseline(t, repo, "node-1", "probe-a", "a: 1\n")
	service := New(repo)
	global := ac.CircuitBreakerSettings{Enabled: true, MaxCPUCores: 2, MaxMemoryBytes: 256 * 1024 * 1024, DurationSeconds: 20}
	if _, err := service.SaveGlobalCircuit(nil, global, 1); err != nil {
		t.Fatal(err)
	}
	settings, err := service.SettingsForNode(nil, "node-1")
	if err != nil || settings.CircuitBreaker.MaxCPUCores != 2 {
		t.Fatalf("node did not use global default: settings=%+v err=%v", settings, err)
	}
	nodePolicy := ac.CircuitBreakerSettings{
		Enabled:         true,
		MaxCPUCores:     1.5,
		MaxMemoryBytes:  128 * 1024 * 1024,
		DurationSeconds: 15,
		ComponentPolicies: map[string]ac.CircuitBreakerSettings{
			"probe-a": {Enabled: true, MaxCPUCores: 0.5, MaxMemoryBytes: 96 * 1024 * 1024, DurationSeconds: 10, AutoRecover: true, RecoveryInitialDelaySeconds: 30, RecoveryMaxDelaySeconds: 3600},
		},
	}
	if _, err := service.SaveNodeCircuit(nil, "node-1", nodePolicy, 2); err != nil {
		t.Fatal(err)
	}
	probeSettings, err := service.SettingsForProbe(nil, "node-1", "probe-a")
	if err != nil || probeSettings.CircuitBreaker.MaxCPUCores != 1.5 || probeSettings.CircuitBreaker.ComponentPolicies["probe-a"].MaxCPUCores != 0.5 {
		t.Fatalf("node did not receive its complete policy: settings=%+v err=%v", probeSettings, err)
	}
	defaultSettings, err := service.SettingsForProbe(nil, "node-1", "probe-b")
	if err != nil || defaultSettings.CircuitBreaker.MaxCPUCores != 1.5 {
		t.Fatalf("component settings did not retain the node total: settings=%+v err=%v", defaultSettings, err)
	}
	if _, err := service.UpdateNodeConfigs(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{ProbeType: "probe-a", Content: "a: 2\n", ExpectedGeneration: 0}}}); err != nil {
		t.Fatal(err)
	}
	after, source, err := service.GetNodeCircuit(nil, "node-1")
	if err != nil || source != "NODE" || after.MaxCPUCores != 1.5 {
		t.Fatalf("config release changed node circuit policy: policy=%+v source=%s err=%v", after, source, err)
	}
}

func TestCircuitPolicyPersistsNodeSystemPressureSettings(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	service := New(repo)
	policy := ac.CircuitBreakerSettings{
		Enabled:                     true,
		DurationSeconds:             30,
		AutoRecover:                 true,
		RecoveryInitialDelaySeconds: 30,
		RecoveryMaxDelaySeconds:     3600,
		SystemPressure: &ac.SystemPressureSettings{
			Enabled:          true,
			MaxCPUPercent:    90,
			MaxMemoryPercent: 90,
		},
	}
	if _, err := service.SaveNodeCircuit(nil, "node-1", policy, 2); err != nil {
		t.Fatal(err)
	}
	settings, source, err := service.GetNodeCircuit(nil, "node-1")
	if err != nil || source != "NODE" || settings.SystemPressure == nil || settings.SystemPressure.MaxCPUPercent != 90 || settings.SystemPressure.MaxMemoryPercent != 90 {
		t.Fatalf("system pressure settings were not persisted: settings=%+v source=%s err=%v", settings, source, err)
	}
}

func TestBatchCircuitPolicyMatchesNodePolicy(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	service := New(repo)

	policy := ac.CircuitBreakerSettings{
		Enabled:                     true,
		MaxCPUCores:                 1,
		MaxMemoryBytes:              512 * 1024 * 1024,
		DurationSeconds:             30,
		AutoRecover:                 true,
		RecoveryInitialDelaySeconds: 30,
		RecoveryMaxDelaySeconds:     3600,
		SystemPressure: &ac.SystemPressureSettings{
			Enabled:          true,
			MaxCPUPercent:    85,
			MaxMemoryPercent: 90,
		},
		ComponentPolicies: map[string]ac.CircuitBreakerSettings{
			"probe-a": {Enabled: true, MaxCPUCores: 0.5, DurationSeconds: 10},
		},
	}
	result, err := service.BatchSaveCircuit(nil, CircuitBatchInput{ClusterIDs: []string{"cluster-a"}, CircuitBreaker: policy}, 2)
	if err != nil {
		t.Fatal(err)
	}
	stored, ok := repo.nodeCircuits["node-1"]
	if !ok || result.SucceededCount != 1 {
		t.Fatalf("cluster system policy was not persisted: result=%+v stored=%+v", result, stored)
	}
	var settings ac.CircuitBreakerSettings
	if err := json.Unmarshal(stored.Settings, &settings); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(settings, policy) {
		t.Fatalf("batch policy differs from node policy: got=%+v want=%+v", settings, policy)
	}
}

func TestCircuitPolicyRejectsSystemPressureOnAComponent(t *testing.T) {
	policy := ac.CircuitBreakerSettings{
		Enabled:         true,
		MaxCPUCores:     1,
		DurationSeconds: 30,
		ComponentPolicies: map[string]ac.CircuitBreakerSettings{
			"probe-a": {
				Enabled:         true,
				MaxCPUCores:     0.5,
				DurationSeconds: 10,
				SystemPressure:  &ac.SystemPressureSettings{Enabled: true, MaxCPUPercent: 90},
			},
		},
	}
	if err := validateCircuitBreaker(&policy); err == nil || !strings.Contains(err.Error(), "must not contain system pressure") {
		t.Fatalf("component system pressure settings were accepted: %v", err)
	}
}

func TestCircuitPolicyAllowsIndependentComponentRule(t *testing.T) {
	policy := ac.CircuitBreakerSettings{
		Enabled: false,
		ComponentPolicies: map[string]ac.CircuitBreakerSettings{
			"probe-a": {
				Enabled:         true,
				MaxCPUCores:     0.5,
				MaxMemoryBytes:  128 * 1024 * 1024,
				DurationSeconds: 10,
			},
		},
	}
	if err := validateCircuitBreaker(&policy); err != nil {
		t.Fatalf("independent component policy was rejected: %v", err)
	}
}

func TestNodeConfigUpdatePersistsDeliveryTargetWithRelease(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", Name: "worker-1", ClusterID: "cluster-a"}
	addBaseline(t, repo, "node-1", "probe-a", "enabled: false\n")
	service := New(repo)

	if _, err := service.UpdateNodeConfigsForUser(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{
		ProbeType: "probe-a", Content: "enabled: true\n", ExpectedGeneration: 0,
	}}}, 42); err != nil {
		t.Fatal(err)
	}
	if len(repo.deliveries) != 1 {
		t.Fatalf("delivery count=%d", len(repo.deliveries))
	}
	delivery := repo.deliveries[0]
	if delivery.Type != deliveryTypeNodeConfig || delivery.Scope != "node-1" || delivery.RequestedBy != 42 || delivery.TargetNodeCount != 1 || len(delivery.Targets) != 1 {
		t.Fatalf("unexpected delivery: %+v", delivery)
	}
	var generations map[string]int64
	if err := json.Unmarshal(delivery.Targets[0].Generations, &generations); err != nil {
		t.Fatal(err)
	}
	if delivery.Targets[0].NodeName != "worker-1" || delivery.Targets[0].ClusterID != "cluster-a" || generations["probe-a"] != 1 {
		t.Fatalf("unexpected delivery target: %+v generations=%v", delivery.Targets[0], generations)
	}
}

func TestListDeliveriesCalculatesEachNodeReceiptFromTargetGeneration(t *testing.T) {
	now := time.Now()
	repo := newConfigSemanticsRepo()
	deliveryID := uuid.New()
	target := func(node string, generation int64) ac.DeliveryTarget {
		data, err := json.Marshal(map[string]int64{"probe-a": generation})
		if err != nil {
			t.Fatal(err)
		}
		return ac.DeliveryTarget{DeliveryID: deliveryID, NodeID: node, NodeName: node, ClusterID: "cluster-a", Generations: data}
	}
	repo.deliveries = []ac.Delivery{{
		ID: deliveryID, Type: deliveryTypeBatchTemplate, Scope: "cluster-a", TargetNodeCount: 4, CreatedAt: now,
		Targets: []ac.DeliveryTarget{target("applied", 2), target("received", 3), target("failed", 4), target("pending", 5)},
	}}
	repo.states = []ac.ApplyState{
		{NodeID: "applied", ProbeType: "probe-a", DesiredGeneration: 3, AppliedGeneration: 3, Status: "RUNNING", ReportedAt: &now},
		{NodeID: "received", ProbeType: "probe-a", DesiredGeneration: 3, AppliedGeneration: 2, Status: "STAGED", ReportedAt: &now},
		{NodeID: "failed", ProbeType: "probe-a", DesiredGeneration: 4, AppliedGeneration: 3, Status: "FAILED", Reason: "checksum mismatch", ReportedAt: &now},
	}

	deliveries, err := New(repo).ListDeliveries(nil, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(deliveries) != 1 {
		t.Fatalf("delivery count=%d", len(deliveries))
	}
	delivery := deliveries[0]
	if delivery.Status != "FAILED" || delivery.AppliedCount != 1 || delivery.ReceivedCount != 1 || delivery.FailedCount != 1 || delivery.PendingCount != 1 {
		t.Fatalf("unexpected delivery summary: %+v", delivery)
	}
	want := []string{"APPLIED", "RECEIVED", "FAILED", "PENDING"}
	for index, target := range delivery.Targets {
		if target.Status != want[index] || len(target.Probes) != 1 || target.Probes[0].Status != want[index] {
			t.Fatalf("target %d=%+v want=%s", index, target, want[index])
		}
	}
	if delivery.Targets[2].Probes[0].Reason != "checksum mismatch" {
		t.Fatalf("failure reason was lost: %+v", delivery.Targets[2])
	}
}

func TestSaveNodeClusterIdentityEnforcesTargetClusterType(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-vm"] = ac.Node{
		ID:                "node-vm",
		ClusterID:         "vm-old",
		InitialClusterID:  "vm-old",
		Environment:       "binary",
		VariableOverrides: ac.JSON(`{"CLUSTER_IP":"10.0.0.8","CUSTOM_VALUE":"keep"}`),
	}
	repo.clusters["vm-old"] = integrationmodel.Cluster{ID: "vm-old", ClusterType: integrationmodel.ClusterTypeVM}
	repo.clusters["vm-next"] = integrationmodel.Cluster{ID: "vm-next", ClusterType: integrationmodel.ClusterTypeVM}
	repo.clusters["k8s-next"] = integrationmodel.Cluster{ID: "k8s-next", ClusterType: integrationmodel.ClusterTypeK8s}
	service := New(repo)

	identity, err := service.SaveNodeClusterIdentity(nil, "node-vm", "vm-next")
	if err != nil {
		t.Fatal(err)
	}
	if identity.ClusterID != "vm-next" || identity.ClusterIP != "10.0.0.8" {
		t.Fatalf("unexpected identity: %+v", identity)
	}
	var overrides map[string]string
	if err := json.Unmarshal(repo.nodes["node-vm"].VariableOverrides, &overrides); err != nil {
		t.Fatal(err)
	}
	if overrides["CLUSTER_IP"] != "10.0.0.8" || overrides["CUSTOM_VALUE"] != "keep" || overrides["CLUSTER_ID"] != "vm-next" {
		t.Fatalf("cluster assignment changed unrelated variables: %+v", overrides)
	}
	if err := service.Register(nil, &ac.Node{ID: "node-vm", ClusterID: "vm-old", Environment: "binary"}); err != nil {
		t.Fatal(err)
	}
	if repo.nodes["node-vm"].ClusterID != "vm-next" {
		t.Fatalf("agent registration discarded the managed cluster assignment: %+v", repo.nodes["node-vm"])
	}
	if _, err := service.RestoreNodeClusterID(nil, "node-vm"); err != nil {
		t.Fatal(err)
	}
	overrides = nil
	if err := json.Unmarshal(repo.nodes["node-vm"].VariableOverrides, &overrides); err != nil {
		t.Fatal(err)
	}
	if overrides["CLUSTER_IP"] != "10.0.0.8" || overrides["CUSTOM_VALUE"] != "keep" {
		t.Fatalf("restoring cluster assignment changed unrelated variables: %+v", overrides)
	}
	if _, ok := overrides["CLUSTER_ID"]; ok {
		t.Fatalf("restoring cluster assignment kept CLUSTER_ID override: %+v", overrides)
	}
	if err := service.Register(nil, &ac.Node{ID: "node-vm", ClusterID: "vm-old", Environment: "binary"}); err != nil {
		t.Fatal(err)
	}
	if repo.nodes["node-vm"].ClusterID != "vm-old" {
		t.Fatalf("restored cluster assignment was not durable: %+v", repo.nodes["node-vm"])
	}

	if _, err := service.SaveNodeClusterIdentity(nil, "node-vm", "k8s-next"); err == nil || !strings.Contains(err.Error(), "cannot assign vm node") {
		t.Fatalf("cross-type assignment was not rejected: %v", err)
	}
}

func TestSaveNodeClusterIdentityRejectsInvalidOverridesWithoutChangingNode(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-vm"] = ac.Node{
		ID:                "node-vm",
		ClusterID:         "vm-old",
		InitialClusterID:  "vm-old",
		Environment:       "binary",
		VariableOverrides: ac.JSON(`{"CLUSTER_IP":`),
	}
	repo.clusters["vm-next"] = integrationmodel.Cluster{ID: "vm-next", ClusterType: integrationmodel.ClusterTypeVM}
	service := New(repo)

	if _, err := service.SaveNodeClusterIdentity(nil, "node-vm", "vm-next"); err == nil || !strings.Contains(err.Error(), "invalid variable overrides") {
		t.Fatalf("invalid overrides were not rejected: %v", err)
	}
	if node := repo.nodes["node-vm"]; node.ClusterID != "vm-old" || string(node.VariableOverrides) != `{"CLUSTER_IP":` {
		t.Fatalf("failed cluster assignment changed the node: %+v", node)
	}
}

func TestMixedClusterIPBlocksExplicitDeliveryButRemainsVisible(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-k8s"] = ac.Node{ID: "node-k8s", ClusterID: "cluster-a", Environment: "kubernetes", Labels: ac.JSON(`{"CLUSTER_IP":"10.0.0.9"}`)}
	repo.nodes["node-vm"] = ac.Node{ID: "node-vm", ClusterID: "cluster-b", Environment: "vm", Labels: ac.JSON(`{"CLUSTER_IP":"10.0.0.9"}`)}
	service := New(repo)

	nodes, err := service.ListNodes(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range nodes {
		if !node.MixedCluster || node.ClusterIP != "10.0.0.9" {
			t.Fatalf("mixed risk was not exposed for node: %+v", node)
		}
		settings, err := service.SettingsForNode(nil, node.ID)
		if err != nil || !settings.UseLocalDefaults || settings.CircuitBreaker == nil {
			t.Fatalf("mixed node %s could not use defaults: settings=%+v err=%v", node.ID, settings, err)
		}
	}
	if _, err := service.BatchSaveCircuit(nil, CircuitBatchInput{
		ClusterIDs: []string{"cluster-a"},
		CircuitBreaker: ac.CircuitBreakerSettings{
			Enabled: true, MaxCPUCores: 1, MaxMemoryBytes: 64 * 1024 * 1024, DurationSeconds: 30,
			RecoveryInitialDelaySeconds: 30, RecoveryMaxDelaySeconds: 60,
		},
	}, 1); err == nil || !strings.Contains(err.Error(), "mixed at cluster_ip 10.0.0.9") {
		t.Fatalf("mixed cluster delivery was not rejected: %v", err)
	}
}
