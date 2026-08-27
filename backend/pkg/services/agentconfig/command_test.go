package agentconfig

import (
	"strings"
	"testing"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	"github.com/google/uuid"
)

type commandRepo struct{ created *ac.Command }

func (r *commandRepo) CreateAgentCommand(_ core.Context, command *ac.Command) error {
	r.created = command
	return nil
}
func (*commandRepo) CreateAgentTemplate(core.Context, *ac.Template) error           { return nil }
func (*commandRepo) UpdateAgentTemplate(core.Context, *ac.Template) error           { return nil }
func (*commandRepo) DeleteAgentTemplate(core.Context, uuid.UUID) error              { return nil }
func (*commandRepo) GetAgentTemplate(core.Context, uuid.UUID) (*ac.Template, error) { return nil, nil }
func (*commandRepo) ListAgentTemplates(core.Context) ([]ac.Template, error)         { return nil, nil }
func (*commandRepo) UpsertAgentNode(core.Context, *ac.Node) error                   { return nil }
func (*commandRepo) ListAgentNodes(core.Context) ([]ac.Node, error)                 { return nil, nil }
func (*commandRepo) GetAgentNode(core.Context, string) (*ac.Node, error)            { return nil, nil }
func (*commandRepo) GetNamespaceInjection(core.Context, string) (*ac.NamespaceInjection, error) {
	return nil, nil
}
func (*commandRepo) SaveClusterInjectionConfig(core.Context, *ac.ClusterInjectionConfig) error {
	return nil
}
func (*commandRepo) SaveClusterWorkloadInventory(core.Context, *ac.ClusterWorkloadInventory) error {
	return nil
}
func (*commandRepo) GetClusterWorkloadInventory(core.Context, string) (*ac.ClusterWorkloadInventory, error) {
	return nil, nil
}
func (*commandRepo) GetOdigletReadiness(core.Context, string, time.Time) (*bool, error) {
	return nil, nil
}
func (*commandRepo) SaveAgentInitialConfigIfUnassigned(core.Context, *ac.InitialConfig) (bool, error) {
	return false, nil
}
func (*commandRepo) GetAgentInitialConfig(core.Context, string, string) (*ac.InitialConfig, error) {
	return nil, nil
}
func (*commandRepo) ListAgentInitialConfigs(core.Context, string) ([]ac.InitialConfig, error) {
	return nil, nil
}
func (*commandRepo) GetGlobalCircuitConfig(core.Context) (*ac.GlobalCircuitConfig, error) {
	return nil, nil
}
func (*commandRepo) SaveGlobalCircuitConfig(core.Context, *ac.GlobalCircuitConfig) error { return nil }
func (*commandRepo) GetNodeCircuitConfig(core.Context, string) (*ac.NodeCircuitConfig, error) {
	return nil, nil
}
func (*commandRepo) SaveNodeCircuitConfigs(core.Context, []*ac.NodeCircuitConfig) error { return nil }
func (*commandRepo) GetNodeLogConfig(core.Context, string, string) (*ac.NodeLogConfig, error) {
	return nil, nil
}
func (*commandRepo) SaveNodeLogConfigs(core.Context, []*ac.NodeLogConfig) error    { return nil }
func (*commandRepo) SaveAgentBinding(core.Context, *ac.Binding, *ac.Release) error { return nil }
func (*commandRepo) SaveAgentBindings(core.Context, []*ac.Binding, []*ac.Release, map[string]int64, *ac.Delivery) error {
	return nil
}
func (*commandRepo) GetAgentRelease(core.Context, string, string) (*ac.Release, error) {
	return nil, nil
}
func (*commandRepo) ListAgentReleases(core.Context, string, string, int) ([]ac.Release, error) {
	return nil, nil
}
func (*commandRepo) ListLatestAgentReleases(core.Context, string) ([]ac.Release, error) {
	return nil, nil
}
func (*commandRepo) ListAgentDeliveries(core.Context, int) ([]ac.Delivery, error) {
	return nil, nil
}
func (*commandRepo) ReportAgentState(core.Context, *ac.ApplyState) error   { return nil }
func (*commandRepo) ListAgentStates(core.Context) ([]ac.ApplyState, error) { return nil, nil }
func (*commandRepo) ListAgentCommands(core.Context, string, string, int) ([]ac.Command, error) {
	return nil, nil
}
func (*commandRepo) ClaimAgentCommands(core.Context, string, string, string, time.Time, time.Duration) ([]ac.Command, error) {
	return nil, nil
}
func (*commandRepo) TransitionAgentCommand(core.Context, uuid.UUID, string, string, string, string, string, time.Time) (*ac.Command, error) {
	return nil, nil
}

func TestCreateOpenCircuitCommand(t *testing.T) {
	repo := &commandRepo{}
	service := New(repo)
	epoch := uint64(7)
	command, err := service.CreateCommand(nil, "node-1", "EBPF-Agent", CommandInput{Type: "open_circuit", ExpectedCircuitEpoch: &epoch, Reason: " memory_limit "}, 9)
	if err != nil {
		t.Fatal(err)
	}
	if command.Type != "OPEN_CIRCUIT" || command.ProbeType != "ebpf-agent" || command.ExpectedCircuitEpoch == nil || *command.ExpectedCircuitEpoch != 7 || command.Reason != "memory_limit" {
		t.Fatalf("unexpected command: %+v", command)
	}
	if repo.created != command {
		t.Fatal("command was not persisted")
	}
}

func TestCircuitCommandRequiresEpochAndOpenReason(t *testing.T) {
	service := New(&commandRepo{})
	if _, err := service.CreateCommand(nil, "node", "probe", CommandInput{Type: "OPEN_CIRCUIT"}, 1); err == nil {
		t.Fatal("OPEN_CIRCUIT without epoch was accepted")
	}
	epoch := uint64(0)
	if _, err := service.CreateCommand(nil, "node", "probe", CommandInput{Type: "OPEN_CIRCUIT", ExpectedCircuitEpoch: &epoch}, 1); err == nil {
		t.Fatal("OPEN_CIRCUIT without reason was accepted")
	}
	if _, err := service.CreateCommand(nil, "node", "probe", CommandInput{Type: "RESET_CIRCUIT"}, 1); err == nil {
		t.Fatal("RESET_CIRCUIT without epoch was accepted")
	}
}

func TestCreateLogDiagnosticCommands(t *testing.T) {
	repo := &commandRepo{}
	service := New(repo)
	command, err := service.CreateCommand(nil, "node", "probe", CommandInput{Type: "preview_log_file", FilePath: "/var/log/probe.log", TailLines: 100}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if command.Type != "PREVIEW_LOG_FILE" || command.FilePath != "/var/log/probe.log" || command.TailLines != 100 {
		t.Fatalf("unexpected log command: %+v", command)
	}
	if _, err := service.CreateCommand(nil, "node", "probe", CommandInput{Type: "DOWNLOAD_LOG_FILE"}, 1); err == nil {
		t.Fatal("DOWNLOAD_LOG_FILE without a path was accepted")
	}
	if _, err := service.CreateCommand(nil, "node", "probe", CommandInput{Type: "PREVIEW_LOG_FILE", FilePath: "/tmp/a", TailLines: 101}, 1); err == nil {
		t.Fatal("preview over 100 lines was accepted")
	}
}

func TestCreateJavaLogScanCommand(t *testing.T) {
	repo := &commandRepo{}
	service := New(repo)
	command, err := service.CreateCommand(nil, "node", "ilogtail", CommandInput{Type: "scan_java_log_files"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if command.Type != "SCAN_JAVA_LOG_FILES" || command.ProbeType != "ilogtail" {
		t.Fatalf("unexpected Java log scan command: %+v", command)
	}
}

func TestCreatePreloadControlCommands(t *testing.T) {
	for _, commandType := range []string{"enable_preload", "DISABLE_PRELOAD"} {
		repo := &commandRepo{}
		service := New(repo)
		command, err := service.CreateCommand(nil, "node", "preload", CommandInput{Type: commandType}, 1)
		if err != nil {
			t.Fatal(err)
		}
		if command.Type != strings.ToUpper(commandType) || command.ProbeType != "preload" {
			t.Fatalf("unexpected preload command: %+v", command)
		}
	}
	if _, err := New(&commandRepo{}).CreateCommand(nil, "node", "ebpf-agent", CommandInput{Type: "ENABLE_PRELOAD"}, 1); err == nil {
		t.Fatal("preload control was accepted for another component")
	}
}
