package agentconfig

import (
	"encoding/json"
	"strings"
	"testing"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	"gorm.io/gorm"
)

type logCollectionTestRepo struct {
	*configSemanticsRepo
	configs map[string]ac.LogCollectionConfig
}

func newLogCollectionTestRepo() *logCollectionTestRepo {
	return &logCollectionTestRepo{
		configSemanticsRepo: newConfigSemanticsRepo(),
		configs:             map[string]ac.LogCollectionConfig{},
	}
}

func (r *logCollectionTestRepo) GetLogCollectionConfig(_ core.Context, scopeType, scopeID string) (*ac.LogCollectionConfig, error) {
	config, ok := r.configs[scopeType+"\x00"+scopeID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copy := config
	return &copy, nil
}

func (r *logCollectionTestRepo) SaveLogCollectionConfig(_ core.Context, config *ac.LogCollectionConfig, delivery *ac.Delivery) error {
	key := config.ScopeType + "\x00" + config.ScopeID
	if old, ok := r.configs[key]; ok {
		config.Version = old.Version + 1
	}
	copy := *config
	r.configs[key] = copy
	if delivery != nil {
		r.deliveries = append(r.deliveries, *delivery)
	}
	return nil
}

func addIlogtailBaseline(t *testing.T, repo *logCollectionTestRepo, nodeID string) {
	addLogCollectorBaseline(t, repo, nodeID, "ilogtail")
}

func addLogCollectorBaseline(t *testing.T, repo *logCollectionTestRepo, nodeID, probeType string) {
	t.Helper()
	const common = `global:
  UsingOldContentTag: true
processors:
  - Type: processor_log_to_prom_metric
    PromPort: ":9499"
flushers:
  - Type: flusher_http
    RemoteURL: http://collector.example:30310
    Convert:
      Protocol: custom_single
      Encoding: json
`
	addBaselineFiles(t, repo.configSemanticsRepo, nodeID, probeType, []ac.InitialConfigFile{
		{Path: "config/local/container.yaml", Content: "enable: true\ninputs:\n  - Type: service_docker_stdout\n" + common},
		{Path: "config/local/host.yaml", Content: "enable: true\ninputs:\n  - Type: input_file\n" + common},
	})
}

func TestVMClusterCollectionUsesLoongcollectorBaseline(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-43"] = ac.Node{ID: "node-43", ClusterID: "cluster-binary-43", Environment: "binary"}
	addLogCollectorBaseline(t, repo, "node-43", "loongcollector")
	service := New(repo)
	input := ContainerStdoutCollectionInput{HostFiles: []HostFileSource{{
		ID: "source-1", Name: "host-file-source-1", Enabled: true,
		Path: "/var/log", FilePattern: "*.nolog", MaxDirSearchDepth: 0,
	}}}

	preview, err := service.PreviewContainerStdoutCollection(nil, "cluster-binary-43", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Files) != 1 || preview.Files[0].Path != "config/local/apo-host-file-source-1.yaml" ||
		!strings.Contains(preview.Files[0].Content, "/var/log/*.nolog") {
		t.Fatalf("unexpected VM log collection preview: %+v", preview.Files)
	}
	saved, err := service.SaveContainerStdoutCollection(nil, "cluster-binary-43", input, 7)
	if err != nil || saved.Version != 1 || len(saved.HostFiles) != 1 {
		t.Fatalf("saved=%+v err=%v", saved, err)
	}
}

func TestHostFileCollectionRendersAndPersistsIlogtailInputs(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addIlogtailBaseline(t, repo, "node-1")
	service := New(repo)
	input := HostFileCollectionInput{DisableClusterRules: true, HostFiles: []HostFileSource{
		{ID: "app", Name: "application", Enabled: true, Path: " /var/log/application ", MaxDirSearchDepth: 4, FilePattern: "*.log"},
		{ID: "disabled", Name: "disabled", Enabled: false, Path: "/var/log/disabled", FilePattern: "*.txt"},
	}}

	preview, err := service.PreviewHostFileCollection(nil, "node-1", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Files) != 1 || preview.Files[0].Path != "config/local/apo-host-file-app.yaml" {
		t.Fatalf("preview files=%+v", preview.Files)
	}
	content := preview.Files[0].Content
	for _, expected := range []string{"UsingOldContentTag: true", "Type: input_file", "/var/log/application/**/*.log", "MaxDirSearchDepth: 4", "AllowingIncludedByMultiConfigs: true", "Type: processor_log_to_prom_metric", "Type: flusher_http", "RemoteURL: http://collector.example:30310"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("preview does not contain %q:\n%s", expected, content)
		}
	}
	if preview.Files[0].InputCount != 1 || preview.Files[0].ProcessorCount != 1 || preview.Files[0].FlusherCount != 1 {
		t.Fatalf("unexpected preview counts: %+v", preview.Files[0])
	}

	saved, err := service.SaveHostFileCollection(nil, "node-1", input, 7)
	if err != nil || saved.Version != 1 || saved.HostFiles[0].Path != "/var/log/application" {
		t.Fatalf("saved=%+v err=%v", saved, err)
	}
	loaded, err := service.GetHostFileCollection(nil, "node-1")
	if err != nil || len(loaded.HostFiles) != 2 || loaded.HostFiles[0].ID != "app" {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
	}
	if loaded.ProcessorCount != 1 || loaded.OutputCount != 1 ||
		!strings.Contains(loaded.ProcessingRule, "processors:") ||
		!strings.Contains(loaded.OutputRule, "RemoteURL: http://collector.example:30310") {
		t.Fatalf("host pipeline defaults were not returned: %+v", loaded)
	}
	if !loaded.DisableClusterRules {
		t.Fatal("node cluster-rule override was not persisted")
	}
}

func TestLogCollectionPersistsEditablePipelineRules(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addIlogtailBaseline(t, repo, "node-1")
	service := New(repo)
	processingRule := "processors:\n  - Type: processor_filter_regex\n    Include:\n      Level: ERROR\n"
	outputRule := "flushers:\n  - Type: flusher_stdout\n    OnlyStdout: true\n"
	input := HostFileCollectionInput{
		HostFiles: []HostFileSource{{
			ID: "app", Name: "application", Enabled: true, Path: "/var/log/app", FilePattern: "*.log",
		}},
		ProcessingRule: &processingRule,
		OutputRule:     &outputRule,
	}

	preview, err := service.PreviewHostFileCollection(nil, "node-1", input)
	if err != nil {
		t.Fatal(err)
	}
	content := preview.Files[0].Content
	for _, expected := range []string{"Type: processor_filter_regex", "Level: ERROR", "Type: flusher_stdout", "OnlyStdout: true"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("editable pipeline preview does not contain %q:\n%s", expected, content)
		}
	}
	if strings.Contains(content, "processor_log_to_prom_metric") || strings.Contains(content, "collector.example") {
		t.Fatalf("baseline pipeline was not replaced by editable rules:\n%s", content)
	}

	if _, err := service.SaveHostFileCollection(nil, "node-1", input, 7); err != nil {
		t.Fatal(err)
	}
	loaded, err := service.GetHostFileCollection(nil, "node-1")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ProcessorCount != 1 || loaded.OutputCount != 1 ||
		!strings.Contains(loaded.ProcessingRule, "processor_filter_regex") ||
		!strings.Contains(loaded.OutputRule, "flusher_stdout") {
		t.Fatalf("editable pipeline rules were not persisted: %+v", loaded)
	}
	stored := string(repo.configs[logCollectionScopeNode+"\x00node-1"].Config)
	if !strings.Contains(stored, `"processingRule"`) || !strings.Contains(stored, `"outputRule"`) {
		t.Fatalf("editable pipeline fields missing from stored config: %s", stored)
	}
}

func TestNodeCollectionSupportsContainerAndFileRulesWithIndependentPipelines(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a", Environment: "kubernetes"}
	addIlogtailBaseline(t, repo, "node-1")
	service := New(repo)
	fileProcessing := "processors:\n  - Type: processor_file_only\n"
	fileOutput := "flushers:\n  - Type: flusher_file_only\n"
	containerProcessing := "processors:\n  - Type: processor_container_only\n"
	containerOutput := "flushers:\n  - Type: flusher_container_only\n"
	input := HostFileCollectionInput{
		HostFiles: []HostFileSource{{
			ID: "file", Name: "file", Enabled: true, Path: "/var/log/app", FilePattern: "*.log",
			ProcessingRule: &fileProcessing, OutputRule: &fileOutput,
		}},
		ContainerStdout: []ContainerStdoutSource{{
			ID: "container", Name: "container", Enabled: true, CollectStdout: true,
			ProcessingRule: &containerProcessing, OutputRule: &containerOutput,
		}},
	}

	preview, err := service.PreviewHostFileCollection(nil, "node-1", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Files) != 2 {
		t.Fatalf("preview files=%+v", preview.Files)
	}
	for _, file := range preview.Files {
		if strings.Contains(file.Path, "host-file") {
			if !strings.Contains(file.Content, "processor_file_only") || !strings.Contains(file.Content, "flusher_file_only") || strings.Contains(file.Content, "container_only") {
				t.Fatalf("file rule pipeline was mixed:\n%s", file.Content)
			}
		} else if !strings.Contains(file.Content, "processor_container_only") || !strings.Contains(file.Content, "flusher_container_only") || strings.Contains(file.Content, "file_only") {
			t.Fatalf("container rule pipeline was mixed:\n%s", file.Content)
		}
	}

	saved, err := service.SaveHostFileCollection(nil, "node-1", input, 7)
	if err != nil || len(saved.ContainerStdout) != 1 || len(saved.HostFiles) != 1 {
		t.Fatalf("saved=%+v err=%v", saved, err)
	}
}

func TestLogCollectionTypeDefaultsAndRuleOverridePrecedence(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a", Environment: "kubernetes"}
	addIlogtailBaseline(t, repo, "node-1")
	service := New(repo)
	fileDefaults := PipelineRuleDefaults{
		ProcessingRule: "processors:\n  - Type: processor_file_default\n",
		OutputRule:     "flushers:\n  - Type: flusher_file_default\n",
	}
	containerDefaults := PipelineRuleDefaults{
		ProcessingRule: "processors:\n  - Type: processor_container_default\n",
		OutputRule:     "flushers:\n  - Type: flusher_container_default\n",
	}
	if _, err := service.SaveLogCollectionDefaults(nil, LogCollectionDefaultsInput{File: fileDefaults, Container: containerDefaults}, 7); err != nil {
		t.Fatal(err)
	}
	override := "processors:\n  - Type: processor_rule_override\n"
	preview, err := service.PreviewContainerStdoutCollection(nil, "cluster-a", ContainerStdoutCollectionInput{
		ContainerStdout: []ContainerStdoutSource{{
			ID: "container", Name: "container", Enabled: true, CollectStdout: true,
			ProcessingRule: &override,
		}},
		HostFiles: []HostFileSource{{
			ID: "file", Name: "file", Enabled: true, Path: "/var/log/app", FilePattern: "*.log",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range preview.Files {
		if strings.Contains(file.Path, "host-file") {
			if !strings.Contains(file.Content, "processor_file_default") || !strings.Contains(file.Content, "flusher_file_default") {
				t.Fatalf("file defaults were not applied:\n%s", file.Content)
			}
		} else {
			if !strings.Contains(file.Content, "processor_rule_override") || strings.Contains(file.Content, "processor_container_default") || !strings.Contains(file.Content, "flusher_container_default") {
				t.Fatalf("rule override precedence is incorrect:\n%s", file.Content)
			}
		}
	}
}

func TestLogCollectionDefaultsRejectInvalidOutput(t *testing.T) {
	repo := newLogCollectionTestRepo()
	service := New(repo)
	emptyOutput := "flushers: []\n"
	validOutput := "flushers:\n  - Type: flusher_stdout\n"
	_, err := service.SaveLogCollectionDefaults(nil, LogCollectionDefaultsInput{
		File:      PipelineRuleDefaults{ProcessingRule: "processors: []\n", OutputRule: emptyOutput},
		Container: PipelineRuleDefaults{ProcessingRule: "processors: []\n", OutputRule: validOutput},
	}, 7)
	if err == nil {
		t.Fatal("empty default output was accepted")
	}
}

func TestLegacyScopePipelineMigratesIntoRulesOnRead(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addIlogtailBaseline(t, repo, "node-1")
	repo.configs[logCollectionScopeNode+"\x00node-1"] = ac.LogCollectionConfig{
		ScopeType: logCollectionScopeNode,
		ScopeID:   "node-1",
		Config: ac.JSON(`{
          "hostFiles":[{"id":"app","name":"app","enabled":true,"path":"/var/log/app","filePattern":"*.log"}],
          "processingRule":"processors:\n  - Type: processor_legacy\n",
          "outputRule":"flushers:\n  - Type: flusher_legacy\n"
        }`),
	}
	loaded, err := New(repo).GetHostFileCollection(nil, "node-1")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.HostFiles[0].ProcessingRule == nil || !strings.Contains(*loaded.HostFiles[0].ProcessingRule, "processor_legacy") ||
		loaded.HostFiles[0].OutputRule == nil || !strings.Contains(*loaded.HostFiles[0].OutputRule, "flusher_legacy") {
		t.Fatalf("legacy pipeline was not migrated into the source: %+v", loaded.HostFiles[0])
	}
}

func TestLogCollectionRejectsInvalidEditablePipelineRules(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	addIlogtailBaseline(t, repo, "node-1")
	service := New(repo)
	emptyOutput := "flushers: []\n"
	wrongProcessingRoot := "flushers:\n  - Type: flusher_stdout\n"
	invalidYAML := "flushers: [\n"
	missingOutputType := "flushers:\n  - RemoteURL: http://collector.example\n"

	for _, input := range []HostFileCollectionInput{
		{OutputRule: &emptyOutput},
		{ProcessingRule: &wrongProcessingRoot},
		{OutputRule: &invalidYAML},
		{OutputRule: &missingOutputType},
	} {
		if _, err := service.PreviewHostFileCollection(nil, "node-1", input); err == nil {
			t.Fatalf("invalid editable pipeline was accepted: %+v", input)
		}
	}
}

func TestHostFileCollectionRejectsInvalidPathAndPattern(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1"}
	service := New(repo)

	for _, source := range []HostFileSource{
		{Name: "relative", Enabled: true, Path: "var/log", FilePattern: "*.log"},
		{Name: "nested pattern", Enabled: true, Path: "/var/log", FilePattern: "app/*.log"},
		{Name: "invalid pattern", Enabled: true, Path: "/var/log", FilePattern: "["},
		{Name: "depth", Enabled: true, Path: "/var/log", FilePattern: "*.log", MaxDirSearchDepth: 1001},
	} {
		if _, err := service.PreviewHostFileCollection(nil, "node-1", HostFileCollectionInput{HostFiles: []HostFileSource{source}}); err == nil {
			t.Fatalf("invalid source was accepted: %+v", source)
		}
	}
	preview, err := service.PreviewHostFileCollection(nil, "node-1", HostFileCollectionInput{HostFiles: []HostFileSource{{ID: "disabled", Enabled: false}}})
	if err != nil || len(preview.Files) != 0 {
		t.Fatalf("disabled incomplete source blocked preview: preview=%+v err=%v", preview, err)
	}
}

func TestContainerStdoutCollectionUsesClusterScopeAndRegexFilters(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a", Environment: "kubernetes"}
	addIlogtailBaseline(t, repo, "node-1")
	service := New(repo)
	input := ContainerStdoutCollectionInput{ContainerStdout: []ContainerStdoutSource{{
		ID: "api", Name: "api containers", Enabled: true, NamespaceRegex: "^(prod)$", PodRegex: "^api-.*", ContainerRegex: "^(app)$", ExcludeK8sLabel: FilterMap{"app": "api", "team": "platform"}, IncludeEnv: "SERVICE=^api$", IncludeContainerLabel: "tier=backend", CollectStdout: true,
	}}}

	preview, err := service.PreviewContainerStdoutCollection(nil, "cluster-a", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Files) != 1 || preview.Files[0].Path != "config/local/apo-container-stdout-api.yaml" {
		t.Fatalf("preview files=%+v", preview.Files)
	}
	content := preview.Files[0].Content
	for _, expected := range []string{"UsingOldContentTag: true", "Type: input_container_stdio", "IgnoringStdout: false", "IgnoringStderr: true", "K8sNamespaceRegex: ^(prod)$", "K8sPodRegex: ^api-.*", "K8sContainerRegex: ^(app)$", "ExcludeK8sLabel:", "app: api", "team: platform", "IncludeEnv:", "SERVICE: ^api$", "IncludeContainerLabel:", "tier: backend", "Type: processor_log_to_prom_metric", "Type: flusher_http"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("preview does not contain %q:\n%s", expected, content)
		}
	}

	if _, err := service.SaveContainerStdoutCollection(nil, "cluster-a", input, 8); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PreviewContainerStdoutCollection(nil, "missing", input); err == nil {
		t.Fatal("missing cluster was accepted")
	}
}

func TestClusterCollectionPersistsContainerAndFileRules(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a", Environment: "kubernetes"}
	addIlogtailBaseline(t, repo, "node-1")
	service := New(repo)
	processingRule := "processors:\n  - Type: processor_filter_regex\n    Include:\n      Source: container\n"
	outputRule := "flushers:\n  - Type: flusher_stdout\n"
	input := ContainerStdoutCollectionInput{
		ContainerStdout: []ContainerStdoutSource{{
			ID: "pods", Name: "pods", Enabled: true, CollectStdout: true, CollectStderr: true,
		}},
		HostFiles: []HostFileSource{{
			ID: "nginx", Name: "nginx", Enabled: true, Path: "/var/log/nginx", FilePattern: "*.log", MaxDirSearchDepth: 1,
		}},
		ProcessingRule: &processingRule,
		OutputRule:     &outputRule,
	}

	preview, err := service.PreviewContainerStdoutCollection(nil, "cluster-a", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Files) != 2 {
		t.Fatalf("preview files=%+v", preview.Files)
	}
	joined := preview.Files[0].Content + preview.Files[1].Content
	for _, expected := range []string{"Type: input_container_stdio", "Type: input_file", "/var/log/nginx/**/*.log"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("cluster preview does not contain %q:\n%s", expected, joined)
		}
	}

	if _, err := service.SaveContainerStdoutCollection(nil, "cluster-a", input, 8); err != nil {
		t.Fatal(err)
	}
	loaded, err := service.GetContainerStdoutCollection(nil, "cluster-a")
	if err != nil || len(loaded.ContainerStdout) != 1 || len(loaded.HostFiles) != 1 {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
	}
	if !strings.Contains(loaded.ProcessingRule, "processor_filter_regex") ||
		!strings.Contains(loaded.OutputRule, "flusher_stdout") {
		t.Fatalf("cluster pipeline rules were not persisted: %+v", loaded)
	}
}

func TestContainerStdoutCollectionRejectsInvalidFilters(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	service := New(repo)

	invalid := []ContainerStdoutSource{
		{Name: "no stream", Enabled: true},
		{Name: "bad regex", Enabled: true, CollectStdout: true, PodRegex: "["},
		{Name: "duplicate map key", Enabled: true, CollectStdout: true, IncludeEnv: "SERVICE=api\nSERVICE=worker"},
		{Name: "invalid event limit", Enabled: true, CollectStdout: true, MaxEventsPerSecond: maxLogEventsPerSecond + 1},
	}
	for _, source := range invalid {
		if _, err := service.PreviewContainerStdoutCollection(nil, "cluster-a", ContainerStdoutCollectionInput{ContainerStdout: []ContainerStdoutSource{source}}); err == nil {
			t.Fatalf("invalid source was accepted: %+v", source)
		}
	}
	repo.nodes["node-2"] = ac.Node{ID: "node-2", ClusterID: "binary-cluster", Environment: "binary"}
	addIlogtailBaseline(t, repo, "node-2")
	preview, err := service.PreviewContainerStdoutCollection(nil, "binary-cluster", ContainerStdoutCollectionInput{ContainerStdout: []ContainerStdoutSource{{
		ID: "all", Name: "all containers", Enabled: true, CollectStdout: true,
	}}})
	if err != nil {
		t.Fatalf("non-Kubernetes cluster with generic container collection was rejected: %v", err)
	}
	if len(preview.Files) != 1 || strings.Contains(preview.Files[0].Content, "ContainerFilters") {
		t.Fatalf("unused filters should be absent from generated config: %+v", preview.Files)
	}
}

func TestDefaultK8sContainerStdoutCollection(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a", Environment: "kubernetes"}
	addIlogtailBaseline(t, repo, "node-1")
	service := New(repo)

	collection, err := service.GetContainerStdoutCollection(nil, "cluster-a")
	if err != nil {
		t.Fatal(err)
	}
	if collection.Version != 0 || len(collection.ContainerStdout) != 1 {
		t.Fatalf("unexpected default collection: %+v", collection)
	}
	if collection.ProcessorCount != 1 || collection.OutputCount != 1 ||
		!strings.Contains(collection.ProcessingRule, "processor_log_to_prom_metric") ||
		!strings.Contains(collection.OutputRule, "flusher_http") {
		t.Fatalf("container pipeline defaults were not returned: %+v", collection)
	}
	source := collection.ContainerStdout[0]
	if source.ID != defaultK8sContainerSourceID || !source.CollectStdout || !source.CollectStderr {
		t.Fatalf("unexpected default source: %+v", source)
	}
	if source.ExcludeK8sLabel["app.kubernetes.io/instance"] != "^(syncause|syncause-agent)$" {
		t.Fatalf("unexpected self exclusion: %+v", source.ExcludeK8sLabel)
	}
	encoded, err := json.Marshal(collection)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"excludeK8sLabel":{"app.kubernetes.io/instance":"^(syncause|syncause-agent)$"}`) {
		t.Fatalf("excludeK8sLabel was not returned as a JSON object: %s", encoded)
	}

	preview, err := service.PreviewContainerStdoutCollection(nil, "cluster-a", ContainerStdoutCollectionInput{ContainerStdout: collection.ContainerStdout})
	if err != nil {
		t.Fatal(err)
	}
	content := preview.Files[0].Content
	for _, expected := range []string{"ExcludeK8sLabel:", "app.kubernetes.io/instance: ^(syncause|syncause-agent)$"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("default preview does not contain %q:\n%s", expected, content)
		}
	}
	if strings.Contains(content, "K8sNamespaceRegex:") {
		t.Fatalf("default preview must not contain K8sNamespaceRegex:\n%s", content)
	}
}

func TestContainerStdoutCollectionReadsLegacyExcludeK8sLabelString(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a", Environment: "kubernetes"}
	repo.configs[logCollectionScopeCluster+"\x00cluster-a"] = ac.LogCollectionConfig{
		ScopeType: logCollectionScopeCluster,
		ScopeID:   "cluster-a",
		Config:    ac.JSON(`{"containerStdout":[{"id":"legacy","name":"legacy","enabled":true,"collectStdout":true,"excludeK8sLabel":"app=api\nteam=platform"}]}`),
	}

	collection, err := New(repo).GetContainerStdoutCollection(nil, "cluster-a")
	if err != nil {
		t.Fatal(err)
	}
	labels := collection.ContainerStdout[0].ExcludeK8sLabel
	if labels["app"] != "api" || labels["team"] != "platform" {
		t.Fatalf("legacy labels were not converted to a map: %+v", labels)
	}
}

func TestContainerStdoutCollectionRendersEventLimit(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a", Environment: "k8s"}
	addIlogtailBaseline(t, repo, "node-1")
	service := New(repo)
	preview, err := service.PreviewContainerStdoutCollection(nil, "cluster-a", ContainerStdoutCollectionInput{ContainerStdout: []ContainerStdoutSource{{
		ID: "limited", Name: "limited", Enabled: true, CollectStdout: true, MaxEventsPerSecond: 100,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"processors:", "Type: processor_log_to_prom_metric", "Type: processor_rate_limit", "Limit: 100/s", "flushers:", "Type: flusher_http"} {
		if !strings.Contains(preview.Files[0].Content, expected) {
			t.Fatalf("limited preview does not contain %q:\n%s", expected, preview.Files[0].Content)
		}
	}
	if preview.Files[0].ProcessorCount != 2 || preview.Files[0].FlusherCount != 1 {
		t.Fatalf("unexpected preview counts: %+v", preview.Files[0])
	}
}

func TestLogCollectionPreviewRejectsMissingOrInconsistentOutputs(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a", Environment: "kubernetes"}
	service := New(repo)
	input := ContainerStdoutCollectionInput{ContainerStdout: []ContainerStdoutSource{{
		ID: "all", Name: "all", Enabled: true, CollectStdout: true,
	}}}
	if _, err := service.PreviewContainerStdoutCollection(nil, "cluster-a", input); err == nil {
		t.Fatal("preview without an iLogtail output pipeline was accepted")
	}

	addIlogtailBaseline(t, repo, "node-1")
	repo.nodes["node-2"] = ac.Node{ID: "node-2", ClusterID: "cluster-a", Environment: "kubernetes"}
	addBaselineFiles(t, repo.configSemanticsRepo, "node-2", "ilogtail", []ac.InitialConfigFile{{
		Path: "config/local/container.yaml",
		Content: `enable: true
inputs:
  - Type: input_container_stdio
flushers:
  - Type: flusher_http
    RemoteURL: http://different-collector.example:30310
`,
	}})
	if _, err := service.PreviewContainerStdoutCollection(nil, "cluster-a", input); err == nil || !strings.Contains(err.Error(), "inconsistent") {
		t.Fatalf("inconsistent cluster output pipeline was accepted: %v", err)
	}
}
