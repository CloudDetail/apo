package agentconfig

import (
	"encoding/json"
	"testing"

	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
)

func TestNormalizeLogSources(t *testing.T) {
	sources := []ac.LogSource{
		{Path: " /var/log/probe ", Pattern: ` ^probe-.*\.log$ `},
		{Type: "stdout", Path: "/ignored", Pattern: ".*", MaxBytes: 32 * 1024 * 1024},
	}
	data, err := normalizeLogSources(&sources)
	if err != nil {
		t.Fatal(err)
	}
	var normalized []ac.LogSource
	if err := json.Unmarshal(data, &normalized); err != nil {
		t.Fatal(err)
	}
	if len(normalized) != 2 || normalized[0].Type != "file" || normalized[0].Path != "/var/log/probe" || normalized[0].Pattern != `^probe-.*\.log$` || normalized[1].Type != "stdout" || normalized[1].Path != "" || normalized[1].Pattern != "" || normalized[1].MaxBytes != 32*1024*1024 {
		t.Fatalf("normalized=%+v", normalized)
	}

	invalid := []ac.LogSource{{Path: "relative.log"}}
	if _, err := normalizeLogSources(&invalid); err == nil {
		t.Fatal("relative log source was accepted")
	}

	invalid = []ac.LogSource{{Path: "/var/log", Pattern: "["}}
	if _, err := normalizeLogSources(&invalid); err == nil {
		t.Fatal("invalid regular expression was accepted")
	}

	invalid = []ac.LogSource{{Type: "container", Path: "/var/log/probe.log"}}
	if _, err := normalizeLogSources(&invalid); err == nil {
		t.Fatal("unsupported log source type was accepted")
	}

	invalid = []ac.LogSource{{Type: "stdout", MaxBytes: 1024}}
	if _, err := normalizeLogSources(&invalid); err == nil {
		t.Fatal("stdout maximum below the minimum was accepted")
	}

	invalid = []ac.LogSource{{Type: "file", Path: "/var/log/probe.log", MaxBytes: 1024 * 1024}}
	if _, err := normalizeLogSources(&invalid); err == nil {
		t.Fatal("file log source accepted maxBytes")
	}
}

func TestLogSourcesAreSavedAndResolvedIndependently(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	service := New(repo)

	settings, err := service.SettingsForProbe(nil, "node-1", "ilogtail")
	if err != nil {
		t.Fatal(err)
	}
	if !settings.UseLocalDefaults || settings.CircuitBreaker == nil {
		t.Fatalf("missing deployment fallback or circuit policy: %+v", settings)
	}

	sources := []ac.LogSource{{Path: "/var/log/pods", Pattern: `.*\.log$`}, {Type: "stdout", MaxBytes: 8 * 1024 * 1024}}
	saved, err := service.SaveNodeLogSources(nil, "node-1", "ilogtail", sources, 7)
	if err != nil || len(saved) != 2 {
		t.Fatalf("save log sources: sources=%+v err=%v", saved, err)
	}
	settings, err = service.SettingsForProbe(nil, "node-1", "ilogtail")
	if err != nil {
		t.Fatal(err)
	}
	if settings.UseLocalDefaults || len(settings.LogSources) != 2 || settings.CircuitBreaker == nil {
		t.Fatalf("independent settings were not resolved: %+v", settings)
	}

	global := *settings.CircuitBreaker
	global.MaxCPUCores = 2
	if _, err := service.SaveGlobalCircuit(nil, global, 8); err != nil {
		t.Fatal(err)
	}
	configured, _, err := service.GetNodeLogSources(nil, "node-1", "ilogtail")
	if err != nil || len(configured) != 2 {
		t.Fatalf("circuit save changed log sources: sources=%+v err=%v", configured, err)
	}
}

func TestSettingsIncludeClusterNamespaceInjection(t *testing.T) {
	repo := newConfigSemanticsRepo()
	ready := true
	repo.odigletReady = &ready
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	repo.namespaceConfigs["cluster-a"] = ac.NamespaceInjection{
		InstrumentAll:        true,
		InstrumentNS:         []string{"orders"},
		InstrumentDisabledNS: []string{"monitoring"},
	}

	settings, err := New(repo).SettingsForProbeAndCluster(nil, "node-1", "odigos-instrumentor", "cluster-a")
	if err != nil {
		t.Fatal(err)
	}
	if settings.NamespaceInjection == nil || !settings.NamespaceInjection.InstrumentAll ||
		len(settings.NamespaceInjection.InstrumentNS) != 1 || settings.NamespaceInjection.InstrumentNS[0] != "orders" ||
		len(settings.NamespaceInjection.InstrumentDisabledNS) != 1 || settings.NamespaceInjection.InstrumentDisabledNS[0] != "monitoring" {
		t.Fatalf("namespace injection settings=%+v", settings.NamespaceInjection)
	}
	if settings.OdigletReady == nil || !*settings.OdigletReady {
		t.Fatalf("odiglet readiness=%v", settings.OdigletReady)
	}
}

func TestDirectInstallWithoutClusterIntegrationUsesLocalNamespaceConfig(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	settings, err := New(repo).SettingsForProbeAndCluster(nil, "node-1", "ilogtail", "cluster-a")
	if err != nil {
		t.Fatal(err)
	}
	if settings.NamespaceInjection != nil || settings.OdigletReady != nil || settings.CircuitBreaker == nil || !settings.UseLocalDefaults {
		t.Fatalf("unexpected settings fallback: %+v", settings)
	}
}

func TestSettingsReportAllOdigletsUnavailable(t *testing.T) {
	repo := newConfigSemanticsRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	ready := false
	repo.odigletReady = &ready
	settings, err := New(repo).SettingsForProbeAndCluster(nil, "node-1", "odigos-instrumentor", "cluster-a")
	if err != nil {
		t.Fatal(err)
	}
	if settings.OdigletReady == nil || *settings.OdigletReady {
		t.Fatalf("odiglet readiness=%v", settings.OdigletReady)
	}
}

func TestBatchLogSourcesOverwriteEveryTargetNode(t *testing.T) {
	repo := newConfigSemanticsRepo()
	for _, node := range []string{"node-a", "node-b"} {
		repo.nodes[node] = ac.Node{ID: node, ClusterID: "cluster-a"}
	}
	service := New(repo)
	result, err := service.BatchSaveLogSources(nil, LogSourcesBatchInput{
		ClusterIDs: []string{"cluster-a"},
		ProbeType:  "ilogtail",
		LogSources: []ac.LogSource{{Path: "/var/log/containers", Pattern: `.*\.log$`}},
	}, 9)
	if err != nil {
		t.Fatal(err)
	}
	if result.TargetNodeCount != 2 || result.SucceededCount != 2 {
		t.Fatalf("unexpected batch result: %+v", result)
	}
	for _, node := range []string{"node-a", "node-b"} {
		sources, configured, err := service.GetNodeLogSources(nil, node, "ilogtail")
		if err != nil || !configured || len(sources) != 1 || sources[0].Path != "/var/log/containers" {
			t.Fatalf("%s log sources=%+v configured=%t err=%v", node, sources, configured, err)
		}
	}
}
