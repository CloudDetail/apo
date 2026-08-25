package agentconfig

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type configSemanticsRepo struct {
	*commandRepo
	nodes          map[string]ac.Node
	baselines      map[string]ac.InitialConfig
	releases       map[string]ac.Release
	templates      map[uuid.UUID]ac.Template
	globalCircuit  *ac.GlobalCircuitConfig
	nodeCircuits   map[string]ac.NodeCircuitConfig
	nodeLogs       map[string]ac.NodeLogConfig
	batchSaveCalls int
	lastExpected   map[string]int64
}

func newConfigSemanticsRepo() *configSemanticsRepo {
	return &configSemanticsRepo{
		commandRepo:  &commandRepo{},
		nodes:        map[string]ac.Node{},
		baselines:    map[string]ac.InitialConfig{},
		releases:     map[string]ac.Release{},
		templates:    map[uuid.UUID]ac.Template{},
		nodeCircuits: map[string]ac.NodeCircuitConfig{},
		nodeLogs:     map[string]ac.NodeLogConfig{},
	}
}

func (r *configSemanticsRepo) GetAgentNode(_ core.Context, node string) (*ac.Node, error) {
	value, ok := r.nodes[node]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &value, nil
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
func (r *configSemanticsRepo) SaveAgentBindings(_ core.Context, bindings []*ac.Binding, releases []*ac.Release, expected map[string]int64) error {
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

func TestNodeConfigUpdatePreservesCompleteMultiFileProbe(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addBaselineFiles(t, repo, "node-1", "ilogtail", []ac.InitialConfigFile{
		{Path: "conf/plugin_logger.xml", Content: "<seelog/>\n"},
		{Path: "config/local/pod_stdout_all.yaml", Content: "enable: true\n"},
		{Path: "config/local/pod_stdout_file.yaml", Content: "enable: false\n"},
	})
	service := New(repo)

	current, err := service.CurrentNodeConfigs(nil, "node-1")
	if err != nil || len(current.Configs) != 1 || len(current.Configs[0].Files) != 3 || current.Configs[0].Content != "" {
		t.Fatalf("unexpected multi-file baseline: configs=%+v err=%v", current.Configs, err)
	}
	files := append([]ac.InitialConfigFile(nil), current.Configs[0].Files...)
	files[1].Content = "enable: changed\n"
	updated, err := service.UpdateNodeConfigs(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{ProbeType: "ilogtail", Files: files, ExpectedGeneration: 0}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Configs[0].Files) != 3 || updated.Configs[0].Generation != 1 {
		t.Fatalf("multi-file release was incomplete: %+v", updated.Configs[0])
	}
	release := repo.releases[releaseKey("node-1", "ilogtail")]
	var persisted []ac.InitialConfigFile
	if err := json.Unmarshal(release.Files, &persisted); err != nil || len(persisted) != 3 {
		t.Fatalf("persisted files=%+v err=%v", persisted, err)
	}
	if release.ContentSHA256 != ac.HashInitialConfigFiles(persisted) {
		t.Fatalf("bundle sha=%s want=%s", release.ContentSHA256, ac.HashInitialConfigFiles(persisted))
	}

	_, err = service.UpdateNodeConfigs(nil, "node-1", NodeConfigsInput{Configs: []NodeProbeConfigInput{{ProbeType: "ilogtail", Files: persisted[:2], ExpectedGeneration: 1}}})
	if err == nil || repo.batchSaveCalls != 1 {
		t.Fatalf("incomplete multi-file update was not rejected atomically: err=%v calls=%d", err, repo.batchSaveCalls)
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
		if config.Generation != 1 || config.Source != "RELEASE" {
			t.Fatalf("unexpected updated config: %+v", config)
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
		{ProbeType: "probe-a", Files: []ac.InitialConfigFile{{Path: "config.yaml", Content: "value: ${VALUE}\n"}}},
		{ProbeType: "probe-b", Files: []ac.InitialConfigFile{{Path: "config.yaml", Content: "enabled: ${ENABLED}\n"}}},
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

func TestBatchAssignmentOverwritesCompleteMultiFileProbeFromTemplate(t *testing.T) {
	repo := newConfigSemanticsRepo()
	paths := []string{
		"conf/plugin_logger.xml",
		"config/local/pod_stdout_all.yaml",
		"config/local/pod_stdout_file.yaml",
	}
	for _, node := range []string{"node-a", "node-b"} {
		repo.nodes[node] = ac.Node{ID: node, ClusterID: "cluster-a"}
		addBaselineFiles(t, repo, node, "ilogtail", []ac.InitialConfigFile{
			{Path: paths[0], Content: "old logger\n"},
			{Path: paths[1], Content: "old all\n"},
			{Path: paths[2], Content: "old file\n"},
		})
	}
	templateFiles := []ac.InitialConfigFile{
		{Path: paths[0], Content: "level: ${LEVEL}\n"},
		{Path: paths[1], Content: "all: ${LEVEL}\n"},
		{Path: paths[2], Content: "file: ${LEVEL}\n"},
	}
	for index := range templateFiles {
		templateFiles[index].ContentSHA256 = ac.HashContent(templateFiles[index].Content)
	}
	templateID := uuid.New()
	addBundleTemplate(t, repo, templateID, []ac.TemplateProbeConfig{{ProbeType: "ilogtail", Files: templateFiles}}, []ac.TemplateVariable{{Name: "LEVEL", Required: true}})
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
		release := repo.releases[releaseKey(node, "ilogtail")]
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

func TestBatchAssignmentRejectsIncompleteFileSet(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-a"] = ac.Node{ID: "node-a", ClusterID: "cluster-a"}
	addBaselineFiles(t, repo, "node-a", "ilogtail", []ac.InitialConfigFile{
		{Path: "a.yaml", Content: "a: old\n"},
		{Path: "b.yaml", Content: "b: old\n"},
	})
	templateID := uuid.New()
	addBundleTemplate(t, repo, templateID, []ac.TemplateProbeConfig{{ProbeType: "ilogtail", Files: []ac.InitialConfigFile{{Path: "a.yaml", Content: "a: new\n"}}}}, nil)
	service := New(repo)

	_, err := service.BatchAssign(nil, BatchAssignmentInput{ClusterIDs: []string{"cluster-a"}, TemplateID: templateID}, 7)
	if err == nil || !strings.Contains(err.Error(), "file set does not match") {
		t.Fatalf("incomplete template was accepted for a multi-file probe: %v", err)
	}
	if repo.batchSaveCalls != 0 {
		t.Fatalf("rejected batch wrote releases: calls=%d", repo.batchSaveCalls)
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
