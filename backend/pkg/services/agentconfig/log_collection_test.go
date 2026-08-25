package agentconfig

import (
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

func (r *logCollectionTestRepo) SaveLogCollectionConfig(_ core.Context, config *ac.LogCollectionConfig) error {
	key := config.ScopeType + "\x00" + config.ScopeID
	if old, ok := r.configs[key]; ok {
		config.Version = old.Version + 1
	}
	copy := *config
	r.configs[key] = copy
	return nil
}

func TestHostFileCollectionRendersAndPersistsIlogtailInputs(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	service := New(repo)
	input := HostFileCollectionInput{HostFiles: []HostFileSource{
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
	for _, expected := range []string{"Type: input_file", "/var/log/application/**/*.log", "MaxDirSearchDepth: 4", "AllowingIncludedByMultiConfigs: true"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("preview does not contain %q:\n%s", expected, content)
		}
	}

	saved, err := service.SaveHostFileCollection(nil, "node-1", input, 7)
	if err != nil || saved.Version != 1 || saved.HostFiles[0].Path != "/var/log/application" {
		t.Fatalf("saved=%+v err=%v", saved, err)
	}
	loaded, err := service.GetHostFileCollection(nil, "node-1")
	if err != nil || len(loaded.HostFiles) != 2 || loaded.HostFiles[0].ID != "app" {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
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
	service := New(repo)
	input := ContainerStdoutCollectionInput{ContainerStdout: []ContainerStdoutSource{{
		ID: "api", Name: "api containers", Enabled: true, NamespaceRegex: "^(prod)$", PodRegex: "^api-.*", ContainerRegex: "^(app)$", IncludeEnv: "SERVICE=^api$", IncludeContainerLabel: "tier=backend", CollectStdout: true,
	}}}

	preview, err := service.PreviewContainerStdoutCollection(nil, "cluster-a", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Files) != 1 || preview.Files[0].Path != "config/local/apo-container-stdout-api.yaml" {
		t.Fatalf("preview files=%+v", preview.Files)
	}
	content := preview.Files[0].Content
	for _, expected := range []string{"Type: input_container_stdio", "IgnoringStdout: false", "IgnoringStderr: true", "K8sNamespaceRegex: ^(prod)$", "K8sPodRegex: ^api-.*", "K8sContainerRegex: ^(app)$", "IncludeEnv:", "SERVICE: ^api$", "IncludeContainerLabel:", "tier: backend"} {
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

func TestContainerStdoutCollectionRejectsInvalidFilters(t *testing.T) {
	repo := newLogCollectionTestRepo()
	repo.nodes["node-1"] = ac.Node{ID: "node-1", ClusterID: "cluster-a"}
	service := New(repo)

	invalid := []ContainerStdoutSource{
		{Name: "no stream", Enabled: true},
		{Name: "bad regex", Enabled: true, CollectStdout: true, PodRegex: "["},
		{Name: "duplicate map key", Enabled: true, CollectStdout: true, IncludeEnv: "SERVICE=api\nSERVICE=worker"},
	}
	for _, source := range invalid {
		if _, err := service.PreviewContainerStdoutCollection(nil, "cluster-a", ContainerStdoutCollectionInput{ContainerStdout: []ContainerStdoutSource{source}}); err == nil {
			t.Fatalf("invalid source was accepted: %+v", source)
		}
	}
	repo.nodes["node-2"] = ac.Node{ID: "node-2", ClusterID: "binary-cluster", Environment: "binary"}
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
