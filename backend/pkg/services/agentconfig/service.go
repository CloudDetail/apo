package agentconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	repo            Repository
	now             func() time.Time
	leaseTTL        time.Duration
	defaultSettings ac.AgentSettings
}

// Repository is the narrow persistence contract owned by APO's remote-agent
// configuration and command-control domain.
type Repository interface {
	CreateAgentTemplate(core.Context, *ac.Template) error
	UpdateAgentTemplate(core.Context, *ac.Template) error
	DeleteAgentTemplate(core.Context, uuid.UUID) error
	GetAgentTemplate(core.Context, uuid.UUID) (*ac.Template, error)
	ListAgentTemplates(core.Context) ([]ac.Template, error)
	UpsertAgentNode(core.Context, *ac.Node) error
	ListAgentNodes(core.Context) ([]ac.Node, error)
	GetAgentNode(core.Context, string) (*ac.Node, error)
	SaveAgentInitialConfigIfUnassigned(core.Context, *ac.InitialConfig) (bool, error)
	GetAgentInitialConfig(core.Context, string, string) (*ac.InitialConfig, error)
	ListAgentInitialConfigs(core.Context, string) ([]ac.InitialConfig, error)
	GetGlobalCircuitConfig(core.Context) (*ac.GlobalCircuitConfig, error)
	SaveGlobalCircuitConfig(core.Context, *ac.GlobalCircuitConfig) error
	GetNodeCircuitConfig(core.Context, string) (*ac.NodeCircuitConfig, error)
	SaveNodeCircuitConfigs(core.Context, []*ac.NodeCircuitConfig) error
	GetNodeLogConfig(core.Context, string, string) (*ac.NodeLogConfig, error)
	SaveNodeLogConfigs(core.Context, []*ac.NodeLogConfig) error
	SaveAgentBinding(core.Context, *ac.Binding, *ac.Release) error
	SaveAgentBindings(core.Context, []*ac.Binding, []*ac.Release, map[string]int64) error
	GetAgentRelease(core.Context, string, string) (*ac.Release, error)
	ListAgentReleases(core.Context, string, string, int) ([]ac.Release, error)
	ListLatestAgentReleases(core.Context, string) ([]ac.Release, error)
	ReportAgentState(core.Context, *ac.ApplyState) error
	ListAgentStates(core.Context) ([]ac.ApplyState, error)
	CreateAgentCommand(core.Context, *ac.Command) error
	ListAgentCommands(core.Context, string, string, int) ([]ac.Command, error)
	ClaimAgentCommands(core.Context, string, string, string, time.Time, time.Duration) ([]ac.Command, error)
	TransitionAgentCommand(core.Context, uuid.UUID, string, string, string, string, string, time.Time) (*ac.Command, error)
}

func New(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now, leaseTTL: 30 * time.Second, defaultSettings: loadDefaultAgentSettings()}
}

const (
	defaultCircuitMaxCPUCores                 = 1
	defaultCircuitMaxMemoryBytes              = 2 * 1024 * 1024 * 1024
	defaultCircuitDurationSeconds             = 30
	defaultCircuitRecoveryInitialDelaySeconds = 30
	defaultCircuitRecoveryMaxDelaySeconds     = 60 * 60
)

func loadDefaultAgentSettings() ac.AgentSettings {
	settings := ac.CircuitBreakerSettings{
		Enabled:                     true,
		MaxCPUCores:                 defaultCircuitMaxCPUCores,
		MaxMemoryBytes:              defaultCircuitMaxMemoryBytes,
		DurationSeconds:             defaultCircuitDurationSeconds,
		AutoRecover:                 true,
		RecoveryInitialDelaySeconds: defaultCircuitRecoveryInitialDelaySeconds,
		RecoveryMaxDelaySeconds:     defaultCircuitRecoveryMaxDelaySeconds,
	}
	if value := strings.TrimSpace(os.Getenv("AGENT_DEFAULT_CIRCUIT_BREAKER_ENABLED")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			log.Printf("agentconfig: invalid AGENT_DEFAULT_CIRCUIT_BREAKER_ENABLED=%q; use %t: %v", value, settings.Enabled, err)
		} else {
			settings.Enabled = parsed
		}
	}
	if value := strings.TrimSpace(os.Getenv("AGENT_DEFAULT_CIRCUIT_MAX_CPU_CORES")); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil || parsed < 0 {
			log.Printf("agentconfig: invalid AGENT_DEFAULT_CIRCUIT_MAX_CPU_CORES=%q; use %.2f", value, settings.MaxCPUCores)
		} else {
			settings.MaxCPUCores = parsed
		}
	}
	if value := strings.TrimSpace(os.Getenv("AGENT_DEFAULT_CIRCUIT_MAX_MEMORY_BYTES")); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			log.Printf("agentconfig: invalid AGENT_DEFAULT_CIRCUIT_MAX_MEMORY_BYTES=%q; use %d", value, settings.MaxMemoryBytes)
		} else {
			settings.MaxMemoryBytes = parsed
		}
	}
	if value := strings.TrimSpace(os.Getenv("AGENT_DEFAULT_CIRCUIT_DURATION_SECONDS")); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil || parsed == 0 {
			log.Printf("agentconfig: invalid AGENT_DEFAULT_CIRCUIT_DURATION_SECONDS=%q; use %d", value, settings.DurationSeconds)
		} else {
			settings.DurationSeconds = parsed
		}
	}
	if value := strings.TrimSpace(os.Getenv("AGENT_DEFAULT_CIRCUIT_AUTO_RECOVER")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			log.Printf("agentconfig: invalid AGENT_DEFAULT_CIRCUIT_AUTO_RECOVER=%q; use %t: %v", value, settings.AutoRecover, err)
		} else {
			settings.AutoRecover = parsed
		}
	}
	if value := strings.TrimSpace(os.Getenv("AGENT_DEFAULT_CIRCUIT_RECOVERY_INITIAL_DELAY_SECONDS")); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil || parsed == 0 {
			log.Printf("agentconfig: invalid AGENT_DEFAULT_CIRCUIT_RECOVERY_INITIAL_DELAY_SECONDS=%q; use %d", value, settings.RecoveryInitialDelaySeconds)
		} else {
			settings.RecoveryInitialDelaySeconds = parsed
		}
	}
	if value := strings.TrimSpace(os.Getenv("AGENT_DEFAULT_CIRCUIT_RECOVERY_MAX_DELAY_SECONDS")); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil || parsed == 0 {
			log.Printf("agentconfig: invalid AGENT_DEFAULT_CIRCUIT_RECOVERY_MAX_DELAY_SECONDS=%q; use %d", value, settings.RecoveryMaxDelaySeconds)
		} else {
			settings.RecoveryMaxDelaySeconds = parsed
		}
	}
	if settings.Enabled && settings.MaxCPUCores == 0 && settings.MaxMemoryBytes == 0 {
		log.Printf("agentconfig: default circuit breaker has no active threshold; use built-in CPU and memory limits")
		settings.MaxCPUCores = defaultCircuitMaxCPUCores
		settings.MaxMemoryBytes = defaultCircuitMaxMemoryBytes
	}
	if err := validateCircuitBreaker(&settings); err != nil {
		log.Printf("agentconfig: invalid configured default circuit breaker; use built-in settings: %v", err)
		settings = ac.CircuitBreakerSettings{
			Enabled:                     true,
			MaxCPUCores:                 defaultCircuitMaxCPUCores,
			MaxMemoryBytes:              defaultCircuitMaxMemoryBytes,
			DurationSeconds:             defaultCircuitDurationSeconds,
			AutoRecover:                 true,
			RecoveryInitialDelaySeconds: defaultCircuitRecoveryInitialDelaySeconds,
			RecoveryMaxDelaySeconds:     defaultCircuitRecoveryMaxDelaySeconds,
		}
	}
	return ac.AgentSettings{UseLocalDefaults: true, CircuitBreaker: &settings}
}

func (s *Service) Settings() ac.AgentSettings {
	settings := s.defaultSettings
	if settings.CircuitBreaker != nil {
		copy := *settings.CircuitBreaker
		settings.CircuitBreaker = &copy
	}
	return settings
}

func (s *Service) SettingsForNode(ctx core.Context, nodeID string) (ac.AgentSettings, error) {
	return s.SettingsForProbe(ctx, nodeID, "")
}

func (s *Service) SettingsForProbe(ctx core.Context, nodeID, probeType string) (ac.AgentSettings, error) {
	if _, err := s.repo.GetAgentNode(ctx, nodeID); err != nil {
		return ac.AgentSettings{}, err
	}
	settings := ac.AgentSettings{UseLocalDefaults: true}
	probeType = ac.NormalizeProbe(probeType)
	if assigned, err := s.repo.GetNodeLogConfig(ctx, nodeID, probeType); err == nil {
		var sources []ac.LogSource
		if decodeErr := json.Unmarshal(assigned.Sources, &sources); decodeErr != nil {
			log.Printf("agentconfig: invalid stored log sources for %s/%s; use deployment defaults: %v", nodeID, probeType, decodeErr)
		} else if normalized, normalizeErr := normalizeLogSources(&sources); normalizeErr != nil {
			log.Printf("agentconfig: invalid stored log sources for %s/%s; use deployment defaults: %v", nodeID, probeType, normalizeErr)
		} else if decodeErr := json.Unmarshal(normalized, &settings.LogSources); decodeErr != nil {
			log.Printf("agentconfig: cannot decode normalized log sources for %s/%s; use deployment defaults: %v", nodeID, probeType, decodeErr)
		} else {
			settings.UseLocalDefaults = false
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return ac.AgentSettings{}, err
	}
	if assigned, err := s.repo.GetNodeCircuitConfig(ctx, nodeID); err == nil {
		circuit, decodeErr := decodeCircuitSettings(assigned.Settings)
		validationErr := validateCircuitBreaker(&circuit)
		if decodeErr == nil && validationErr == nil {
			settings.CircuitBreaker = &circuit
			return settings, nil
		}
		if decodeErr != nil {
			log.Printf("agentconfig: invalid stored node circuit settings for %s; use global default: %v", nodeID, decodeErr)
		} else {
			log.Printf("agentconfig: invalid stored node circuit settings for %s; use global default: %v", nodeID, validationErr)
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return ac.AgentSettings{}, err
	}
	circuit, err := s.GetGlobalCircuit(ctx)
	if err != nil {
		return ac.AgentSettings{}, err
	}
	settings.CircuitBreaker = circuit
	return settings, nil
}

type TemplateInput struct {
	Configs     []ac.TemplateProbeConfig `json:"configs"`
	Name        string                   `json:"name" binding:"required"`
	Description string                   `json:"description"`
	Variables   []ac.TemplateVariable    `json:"variables"`
}
type AssignmentInput struct {
	TemplateID         uuid.UUID         `json:"templateId" binding:"required"`
	Variables          map[string]string `json:"variables"`
	ExpectedGeneration int64             `json:"expectedGeneration"`
	LogSources         *[]ac.LogSource   `json:"logSources"`
}
type NodeProbeConfigInput struct {
	ProbeType          string                 `json:"probeType" binding:"required"`
	Content            string                 `json:"content"`
	Files              []ac.InitialConfigFile `json:"files"`
	ExpectedGeneration int64                  `json:"expectedGeneration"`
}
type NodeConfigsInput struct {
	Configs []NodeProbeConfigInput `json:"configs" binding:"required"`
}
type BatchAssignmentInput struct {
	ClusterIDs []string          `json:"clusterIds" binding:"required"`
	TemplateID uuid.UUID         `json:"templateId" binding:"required"`
	Variables  map[string]string `json:"variables"`
}
type CircuitBatchInput struct {
	ClusterIDs     []string                  `json:"clusterIds" binding:"required"`
	CircuitBreaker ac.CircuitBreakerSettings `json:"circuitBreaker" binding:"required"`
}
type LogSourcesBatchInput struct {
	ClusterIDs []string       `json:"clusterIds" binding:"required"`
	ProbeType  string         `json:"probeType" binding:"required"`
	LogSources []ac.LogSource `json:"logSources"`
}
type BatchResult struct {
	TargetNodeCount int      `json:"targetNodeCount"`
	SucceededCount  int      `json:"succeededCount"`
	FailedNodeIDs   []string `json:"failedNodeIds,omitempty"`
}
type CurrentNodeConfig struct {
	NodeID        string                 `json:"nodeId"`
	ProbeType     string                 `json:"probeType"`
	Generation    int64                  `json:"generation"`
	Content       string                 `json:"content,omitempty"`
	Files         []ac.InitialConfigFile `json:"files"`
	ContentSHA256 string                 `json:"contentSha256"`
	Source        string                 `json:"source"`
	CreatedAt     time.Time              `json:"createdAt,omitempty"`
}
type CurrentNodeConfigs struct {
	NodeID  string              `json:"nodeId"`
	Configs []CurrentNodeConfig `json:"configs"`
}
type InitialConfigInput struct {
	ProbeType string                 `json:"probeType" binding:"required"`
	Files     []ac.InitialConfigFile `json:"files" binding:"required"`
}
type CommandInput struct {
	Type                 string  `json:"type" binding:"required"`
	ExpectedRuntimeEpoch *uint64 `json:"expectedRuntimeEpoch"`
	ExpectedCircuitEpoch *uint64 `json:"expectedCircuitEpoch"`
	Reason               string  `json:"reason"`
	FilePath             string  `json:"filePath"`
	TailLines            int     `json:"tailLines"`
}
type ValidationError struct {
	MissingVariables []string `json:"missingVariables"`
	InvalidVariables []string `json:"invalidVariables"`
	UnusedVariables  []string `json:"unusedVariables"`
}

func (e *ValidationError) Error() string {
	b, _ := json.Marshal(e)
	return "template variables are incomplete: " + string(b)
}

func normalizeConfigFiles(input []ac.InitialConfigFile, description string) ([]ac.InitialConfigFile, error) {
	if len(input) == 0 || len(input) > 32 {
		return nil, fmt.Errorf("%s must contain between 1 and 32 files", description)
	}
	seen := make(map[string]struct{}, len(input))
	total := 0
	files := make([]ac.InitialConfigFile, 0, len(input))
	for _, file := range input {
		cleaned := path.Clean(strings.TrimSpace(file.Path))
		if cleaned == "." || cleaned == ".." || path.IsAbs(cleaned) || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "\\") {
			return nil, fmt.Errorf("invalid %s path %q", description, file.Path)
		}
		if _, exists := seen[cleaned]; exists {
			return nil, fmt.Errorf("duplicate %s path %s", description, cleaned)
		}
		seen[cleaned] = struct{}{}
		total += len(cleaned) + len(file.Content)
		if total > 1024*1024 {
			return nil, fmt.Errorf("%s exceeds 1 MiB", description)
		}
		files = append(files, ac.InitialConfigFile{Path: cleaned, Content: file.Content, ContentSHA256: ac.HashContent(file.Content)})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func normalizeTemplatePayload(in TemplateInput) (ac.JSON, string, string, error) {
	configs := append([]ac.TemplateProbeConfig(nil), in.Configs...)
	if len(configs) == 0 {
		return nil, "", "", fmt.Errorf("template must contain at least one probe configuration")
	}
	seen := make(map[string]struct{}, len(configs))
	var variableSource strings.Builder
	for index := range configs {
		probe := ac.NormalizeProbe(configs[index].ProbeType)
		if probe == "" {
			return nil, "", "", fmt.Errorf("template probe type is required")
		}
		if _, duplicate := seen[probe]; duplicate {
			return nil, "", "", fmt.Errorf("template probe %s is duplicated", probe)
		}
		seen[probe] = struct{}{}
		files, err := normalizeConfigFiles(configs[index].Files, "template "+probe)
		if err != nil {
			return nil, "", "", err
		}
		configs[index] = ac.TemplateProbeConfig{ProbeType: probe, Files: files}
		for _, file := range files {
			variableSource.WriteString(file.Content)
			variableSource.WriteByte('\n')
		}
	}
	sort.Slice(configs, func(i, j int) bool { return configs[i].ProbeType < configs[j].ProbeType })
	data, err := json.Marshal(configs)
	if err != nil {
		return nil, "", "", err
	}
	if len(data) > 1024*1024 {
		return nil, "", "", fmt.Errorf("template exceeds 1 MiB")
	}
	return ac.JSON(data), ac.HashContent(string(data)), variableSource.String(), nil
}

func mergeVariables(content string, supplied []ac.TemplateVariable) []ac.TemplateVariable {
	meta := map[string]ac.TemplateVariable{}
	for _, v := range supplied {
		meta[v.Name] = v
	}
	out := []ac.TemplateVariable{}
	for _, v := range ExtractVariables(content) {
		x := meta[v.Name]
		x.Name = v.Name
		if x.DefaultValue == "" {
			x.DefaultValue = v.DefaultValue
		}
		x.Required = v.Required && x.DefaultValue == ""
		out = append(out, x)
	}
	return out
}
func (s *Service) Inspect(content string) []VariableSpec { return ExtractVariables(content) }
func (s *Service) CreateTemplate(ctx core.Context, in TemplateInput, user int64) (*ac.Template, error) {
	configs, checksum, variableSource, err := normalizeTemplatePayload(in)
	if err != nil {
		return nil, err
	}
	t := ac.NewTemplate()
	t.Configs = configs
	t.Name = strings.TrimSpace(in.Name)
	t.Description = in.Description
	t.ContentSHA256 = checksum
	t.CreatedBy = user
	t.UpdatedBy = user
	t.Variables = mergeVariables(variableSource, in.Variables)
	for i := range t.Variables {
		t.Variables[i].TemplateID = t.ID
	}
	return t, s.repo.CreateAgentTemplate(ctx, t)
}
func (s *Service) UpdateTemplate(ctx core.Context, id uuid.UUID, in TemplateInput, user int64) (*ac.Template, error) {
	old, e := s.repo.GetAgentTemplate(ctx, id)
	if e != nil {
		return nil, e
	}
	configs, checksum, variableSource, e := normalizeTemplatePayload(in)
	if e != nil {
		return nil, e
	}
	old.Configs = configs
	old.Name = strings.TrimSpace(in.Name)
	old.Description = in.Description
	old.Version++
	old.ContentSHA256 = checksum
	old.UpdatedBy = user
	old.Variables = mergeVariables(variableSource, in.Variables)
	for i := range old.Variables {
		old.Variables[i].ID = 0
		old.Variables[i].TemplateID = id
	}
	e = s.repo.UpdateAgentTemplate(ctx, old)
	return old, e
}
func (s *Service) ListTemplates(ctx core.Context) ([]ac.Template, error) {
	templates, err := s.repo.ListAgentTemplates(ctx)
	if err != nil {
		return nil, err
	}
	for index := range templates {
		if err := hydrateTemplateConfigs(&templates[index]); err != nil {
			return nil, fmt.Errorf("decode template %s: %w", templates[index].ID, err)
		}
	}
	return templates, nil
}
func (s *Service) GetTemplate(ctx core.Context, id uuid.UUID) (*ac.Template, error) {
	template, err := s.repo.GetAgentTemplate(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := hydrateTemplateConfigs(template); err != nil {
		return nil, err
	}
	return template, nil
}
func (s *Service) DeleteTemplate(ctx core.Context, id uuid.UUID) error {
	return s.repo.DeleteAgentTemplate(ctx, id)
}
func (s *Service) Register(ctx core.Context, n *ac.Node) error {
	now := time.Now()
	n.LastSeenAt = &now
	return s.repo.UpsertAgentNode(ctx, n)
}
func (s *Service) ListNodes(ctx core.Context) ([]ac.Node, error) { return s.repo.ListAgentNodes(ctx) }

func (s *Service) SaveInitialConfig(ctx core.Context, node string, in InitialConfigInput) (*ac.InitialConfig, bool, error) {
	if _, err := s.repo.GetAgentNode(ctx, node); err != nil {
		return nil, false, err
	}
	probe := ac.NormalizeProbe(in.ProbeType)
	if probe == "" {
		return nil, false, fmt.Errorf("probe type is required")
	}
	files, err := normalizeConfigFiles(in.Files, "initial config")
	if err != nil {
		return nil, false, err
	}
	data, err := json.Marshal(files)
	if err != nil {
		return nil, false, err
	}
	baseline := &ac.InitialConfig{NodeID: node, ProbeType: probe, Files: ac.JSON(data), ContentSHA256: ac.HashInitialConfigFiles(files), CapturedAt: s.now()}
	saved, err := s.repo.SaveAgentInitialConfigIfUnassigned(ctx, baseline)
	return baseline, saved, err
}

func (s *Service) GetInitialConfig(ctx core.Context, node, probe string) (*ac.InitialConfig, error) {
	return s.repo.GetAgentInitialConfig(ctx, node, ac.NormalizeProbe(probe))
}

func validateCircuitBreaker(settings *ac.CircuitBreakerSettings) error {
	if settings == nil {
		return nil
	}
	for componentType, policy := range settings.ComponentPolicies {
		normalized := ac.NormalizeProbe(componentType)
		if normalized == "" || normalized != componentType {
			return fmt.Errorf("component policy key %q must be normalized", componentType)
		}
		if len(policy.ComponentPolicies) > 0 {
			return fmt.Errorf("component policy %s must not contain nested component policies", componentType)
		}
		if err := validateCircuitBreaker(&policy); err != nil {
			return fmt.Errorf("component policy %s: %w", componentType, err)
		}
	}
	if !settings.Enabled {
		return nil
	}
	if settings.MaxCPUCores != 0 && (settings.MaxCPUCores < 0.1 || settings.MaxCPUCores > 256) {
		return fmt.Errorf("max CPU cores must be 0 or between 0.1 and 256")
	}
	const (
		minimumMemory = 64 * 1024 * 1024
		maximumMemory = 64 * 1024 * 1024 * 1024
	)
	if settings.MaxMemoryBytes != 0 && (settings.MaxMemoryBytes < minimumMemory || settings.MaxMemoryBytes > maximumMemory) {
		return fmt.Errorf("max memory bytes must be 0 or between 64 MiB and 64 GiB")
	}
	if settings.MaxCPUCores == 0 && settings.MaxMemoryBytes == 0 {
		return fmt.Errorf("enabled circuit breaker requires a CPU or memory threshold")
	}
	if settings.DurationSeconds < 5 || settings.DurationSeconds > 600 {
		return fmt.Errorf("duration seconds must be between 5 and 600")
	}
	if settings.AutoRecover {
		if settings.RecoveryInitialDelaySeconds < 5 || settings.RecoveryInitialDelaySeconds > 600 {
			return fmt.Errorf("recovery initial delay seconds must be between 5 and 600")
		}
		if settings.RecoveryMaxDelaySeconds < settings.RecoveryInitialDelaySeconds || settings.RecoveryMaxDelaySeconds > 3600 {
			return fmt.Errorf("recovery max delay seconds must be between the initial delay and 3600")
		}
	}
	return nil
}

func decodeCircuitSettings(data ac.JSON) (ac.CircuitBreakerSettings, error) {
	var settings ac.CircuitBreakerSettings
	err := json.Unmarshal(data, &settings)
	return settings, err
}

func (s *Service) GetGlobalCircuit(ctx core.Context) (*ac.CircuitBreakerSettings, error) {
	config, err := s.repo.GetGlobalCircuitConfig(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.Settings().CircuitBreaker, nil
	}
	if err != nil {
		return nil, err
	}
	settings, err := decodeCircuitSettings(config.Settings)
	validationErr := validateCircuitBreaker(&settings)
	if err != nil || validationErr != nil {
		if err != nil {
			log.Printf("agentconfig: invalid stored global circuit settings; use built-in default: %v", err)
		} else {
			log.Printf("agentconfig: invalid stored global circuit settings; use built-in default: %v", validationErr)
		}
		return s.Settings().CircuitBreaker, nil
	}
	return &settings, nil
}

func (s *Service) SaveGlobalCircuit(ctx core.Context, settings ac.CircuitBreakerSettings, user int64) (*ac.CircuitBreakerSettings, error) {
	if err := validateCircuitBreaker(&settings); err != nil {
		return nil, err
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	config := &ac.GlobalCircuitConfig{ID: 1, Settings: ac.JSON(data), UpdatedBy: user}
	if err := s.repo.SaveGlobalCircuitConfig(ctx, config); err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *Service) GetNodeCircuit(ctx core.Context, node string) (*ac.CircuitBreakerSettings, string, error) {
	if _, err := s.repo.GetAgentNode(ctx, node); err != nil {
		return nil, "", err
	}
	config, err := s.repo.GetNodeCircuitConfig(ctx, node)
	if err == nil {
		settings, err := decodeCircuitSettings(config.Settings)
		if err == nil {
			err = validateCircuitBreaker(&settings)
		}
		return &settings, "NODE", err
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}
	settings, err := s.GetGlobalCircuit(ctx)
	return settings, "GLOBAL_DEFAULT", err
}

func (s *Service) saveNodeCircuitConfigs(ctx core.Context, nodes []ac.Node, settings ac.CircuitBreakerSettings, user int64) (*BatchResult, error) {
	if err := validateCircuitBreaker(&settings); err != nil {
		return nil, err
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	configs := make([]*ac.NodeCircuitConfig, 0, len(nodes))
	for _, node := range nodes {
		configs = append(configs, &ac.NodeCircuitConfig{NodeID: node.ID, Settings: append(ac.JSON(nil), data...), Version: 1, UpdatedBy: user})
	}
	if err := s.repo.SaveNodeCircuitConfigs(ctx, configs); err != nil {
		return nil, err
	}
	return &BatchResult{TargetNodeCount: len(nodes), SucceededCount: len(nodes)}, nil
}

func (s *Service) SaveNodeCircuit(ctx core.Context, node string, settings ac.CircuitBreakerSettings, user int64) (*ac.CircuitBreakerSettings, error) {
	target, err := s.repo.GetAgentNode(ctx, node)
	if err != nil {
		return nil, err
	}
	if _, err := s.saveNodeCircuitConfigs(ctx, []ac.Node{*target}, settings, user); err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *Service) BatchSaveCircuit(ctx core.Context, input CircuitBatchInput, user int64) (*BatchResult, error) {
	nodes, err := s.nodesInClusters(ctx, input.ClusterIDs)
	if err != nil {
		return nil, err
	}
	return s.saveNodeCircuitConfigs(ctx, nodes, input.CircuitBreaker, user)
}

func (s *Service) GetNodeLogSources(ctx core.Context, node, probe string) ([]ac.LogSource, bool, error) {
	if _, err := s.repo.GetAgentNode(ctx, node); err != nil {
		return nil, false, err
	}
	probe = ac.NormalizeProbe(probe)
	if probe == "" {
		return nil, false, fmt.Errorf("probe type is required")
	}
	config, err := s.repo.GetNodeLogConfig(ctx, node, probe)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []ac.LogSource{}, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var sources []ac.LogSource
	if err := json.Unmarshal(config.Sources, &sources); err != nil {
		return nil, false, err
	}
	return sources, true, nil
}

func (s *Service) saveNodeLogSources(ctx core.Context, nodes []ac.Node, probe string, sources []ac.LogSource, user int64) (*BatchResult, error) {
	probe = ac.NormalizeProbe(probe)
	if probe == "" {
		return nil, fmt.Errorf("probe type is required")
	}
	normalized, err := normalizeLogSources(&sources)
	if err != nil {
		return nil, err
	}
	configs := make([]*ac.NodeLogConfig, 0, len(nodes))
	for _, node := range nodes {
		configs = append(configs, &ac.NodeLogConfig{
			NodeID: node.ID, ProbeType: probe, Sources: append(ac.JSON(nil), normalized...), Version: 1, UpdatedBy: user,
		})
	}
	if err := s.repo.SaveNodeLogConfigs(ctx, configs); err != nil {
		return nil, err
	}
	return &BatchResult{TargetNodeCount: len(nodes), SucceededCount: len(nodes)}, nil
}

func (s *Service) SaveNodeLogSources(ctx core.Context, node, probe string, sources []ac.LogSource, user int64) ([]ac.LogSource, error) {
	target, err := s.repo.GetAgentNode(ctx, node)
	if err != nil {
		return nil, err
	}
	if _, err := s.saveNodeLogSources(ctx, []ac.Node{*target}, probe, sources, user); err != nil {
		return nil, err
	}
	stored, _, err := s.GetNodeLogSources(ctx, node, probe)
	return stored, err
}

func (s *Service) BatchSaveLogSources(ctx core.Context, input LogSourcesBatchInput, user int64) (*BatchResult, error) {
	nodes, err := s.nodesInClusters(ctx, input.ClusterIDs)
	if err != nil {
		return nil, err
	}
	return s.saveNodeLogSources(ctx, nodes, input.ProbeType, input.LogSources, user)
}

func (s *Service) nodesInClusters(ctx core.Context, clusterIDs []string) ([]ac.Node, error) {
	wanted := map[string]bool{}
	for _, clusterID := range clusterIDs {
		clusterID = strings.TrimSpace(clusterID)
		if clusterID != "" {
			wanted[clusterID] = true
		}
	}
	if len(wanted) == 0 {
		return nil, fmt.Errorf("at least one cluster is required")
	}
	nodes, err := s.repo.ListAgentNodes(ctx)
	if err != nil {
		return nil, err
	}
	found := map[string]bool{}
	targets := make([]ac.Node, 0)
	for _, node := range nodes {
		if wanted[node.ClusterID] {
			found[node.ClusterID] = true
			targets = append(targets, node)
		}
	}
	missing := make([]string, 0)
	for clusterID := range wanted {
		if !found[clusterID] {
			missing = append(missing, clusterID)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("clusters have no registered nodes: %s", strings.Join(missing, ", "))
	}
	return targets, nil
}

func (s *Service) CurrentNodeConfig(ctx core.Context, node, probe string) (*CurrentNodeConfig, error) {
	if _, err := s.repo.GetAgentNode(ctx, node); err != nil {
		return nil, err
	}
	probe = ac.NormalizeProbe(probe)
	if release, err := s.repo.GetAgentRelease(ctx, node, probe); err == nil {
		files, fileErr := s.releaseFiles(ctx, node, probe, release)
		if fileErr != nil {
			return nil, fileErr
		}
		return currentNodeConfig(node, probe, release.Generation, files, release.ContentSHA256, "RELEASE"), nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	baseline, err := s.repo.GetAgentInitialConfig(ctx, node, probe)
	if err != nil {
		return nil, err
	}
	var files []ac.InitialConfigFile
	if err := json.Unmarshal(baseline.Files, &files); err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("node has no initial configuration for %s", probe)
	}
	return currentNodeConfig(node, probe, 0, files, baseline.ContentSHA256, "INITIAL"), nil
}

func currentNodeConfig(node, probe string, generation int64, files []ac.InitialConfigFile, sha, source string) *CurrentNodeConfig {
	config := &CurrentNodeConfig{NodeID: node, ProbeType: probe, Generation: generation, Files: files, ContentSHA256: sha, Source: source}
	if len(files) == 1 {
		config.Content = files[0].Content
	}
	return config
}

func (s *Service) ListNodeConfigHistory(ctx core.Context, node, probe string, limit int) ([]CurrentNodeConfig, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	probe = ac.NormalizeProbe(probe)
	releases, err := s.repo.ListAgentReleases(ctx, node, probe, limit)
	if err != nil {
		return nil, err
	}
	result := make([]CurrentNodeConfig, 0, len(releases))
	for index := range releases {
		files, fileErr := s.releaseFiles(ctx, node, probe, &releases[index])
		if fileErr != nil {
			return nil, fileErr
		}
		config := currentNodeConfig(node, probe, releases[index].Generation, files, releases[index].ContentSHA256, "RELEASE")
		config.CreatedAt = releases[index].CreatedAt
		result = append(result, *config)
	}
	return result, nil
}

func (s *Service) releaseFiles(ctx core.Context, node, probe string, release *ac.Release) ([]ac.InitialConfigFile, error) {
	if len(release.Files) > 0 && string(release.Files) != "null" && string(release.Files) != "{}" {
		var files []ac.InitialConfigFile
		if err := json.Unmarshal(release.Files, &files); err != nil {
			return nil, fmt.Errorf("decode release files for %s: %w", probe, err)
		}
		if len(files) == 0 {
			return nil, fmt.Errorf("release %s/%d has no configuration files", probe, release.Generation)
		}
		return files, nil
	}

	path := "config.yaml"
	if baseline, err := s.repo.GetAgentInitialConfig(ctx, node, probe); err == nil {
		var files []ac.InitialConfigFile
		if err := json.Unmarshal(baseline.Files, &files); err != nil {
			return nil, err
		}
		if len(files) != 1 {
			return nil, fmt.Errorf("legacy release %s/%d cannot represent %d configuration files", probe, release.Generation, len(files))
		}
		path = files[0].Path
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return []ac.InitialConfigFile{{Path: path, Content: release.RenderedContent, ContentSHA256: ac.HashContent(release.RenderedContent)}}, nil
}

func (s *Service) CurrentNodeConfigs(ctx core.Context, node string) (*CurrentNodeConfigs, error) {
	if _, err := s.repo.GetAgentNode(ctx, node); err != nil {
		return nil, err
	}
	probeTypes := map[string]struct{}{}
	baselines, err := s.repo.ListAgentInitialConfigs(ctx, node)
	if err != nil {
		return nil, err
	}
	for _, baseline := range baselines {
		probeTypes[ac.NormalizeProbe(baseline.ProbeType)] = struct{}{}
	}
	releases, err := s.repo.ListLatestAgentReleases(ctx, node)
	if err != nil {
		return nil, err
	}
	for _, release := range releases {
		probeTypes[ac.NormalizeProbe(release.ProbeType)] = struct{}{}
	}
	probes := make([]string, 0, len(probeTypes))
	for probe := range probeTypes {
		if probe != "" {
			probes = append(probes, probe)
		}
	}
	sort.Strings(probes)
	result := &CurrentNodeConfigs{NodeID: node, Configs: make([]CurrentNodeConfig, 0, len(probes))}
	for _, probe := range probes {
		current, err := s.CurrentNodeConfig(ctx, node, probe)
		if err != nil {
			return nil, err
		}
		result.Configs = append(result.Configs, *current)
	}
	return result, nil
}

func releaseKey(node, probe string) string { return node + "\x00" + ac.NormalizeProbe(probe) }

func (s *Service) UpdateNodeConfigs(ctx core.Context, node string, input NodeConfigsInput) (*CurrentNodeConfigs, error) {
	if len(input.Configs) == 0 {
		return nil, fmt.Errorf("at least one probe configuration is required")
	}
	currentSet, err := s.CurrentNodeConfigs(ctx, node)
	if err != nil {
		return nil, err
	}
	if len(input.Configs) != len(currentSet.Configs) {
		return nil, fmt.Errorf("node update must include all %d current probe configurations", len(currentSet.Configs))
	}
	currentByProbe := make(map[string]CurrentNodeConfig, len(currentSet.Configs))
	for _, current := range currentSet.Configs {
		currentByProbe[current.ProbeType] = current
	}
	seen := map[string]bool{}
	bindings := make([]*ac.Binding, 0, len(input.Configs))
	releases := make([]*ac.Release, 0, len(input.Configs))
	expected := make(map[string]int64, len(input.Configs))
	for _, update := range input.Configs {
		probe := ac.NormalizeProbe(update.ProbeType)
		if probe == "" || seen[probe] {
			return nil, fmt.Errorf("probe types must be non-empty and unique")
		}
		seen[probe] = true
		current, exists := currentByProbe[probe]
		if !exists {
			return nil, fmt.Errorf("probe %s is not part of the node's current configuration", probe)
		}
		if update.ExpectedGeneration != current.Generation {
			return nil, fmt.Errorf("generation conflict for %s: expected %d, actual %d", probe, update.ExpectedGeneration, current.Generation)
		}
		expected[releaseKey(node, probe)] = current.Generation
		files, err := normalizeNodeUpdateFiles(update, current)
		if err != nil {
			return nil, fmt.Errorf("invalid files for %s: %w", probe, err)
		}
		if sameConfigFiles(files, current.Files) {
			continue
		}
		logSources := ac.JSON("null")
		if current.Generation > 0 {
			last, err := s.repo.GetAgentRelease(ctx, node, probe)
			if err != nil {
				return nil, err
			}
			logSources = append(ac.JSON(nil), last.LogSources...)
		}
		generation := current.Generation + 1
		filesJSON, err := json.Marshal(files)
		if err != nil {
			return nil, err
		}
		renderedContent := ""
		if len(files) == 1 {
			renderedContent = files[0].Content
		}
		release := &ac.Release{ID: uuid.New(), NodeID: node, ProbeType: probe, Generation: generation, TemplateID: uuid.Nil, TemplateVersion: 1, Variables: ac.JSON("{}"), RenderedContent: renderedContent, ContentSHA256: ac.HashInitialConfigFiles(files), Files: ac.JSON(filesJSON), LogSources: logSources}
		releases = append(releases, release)
		bindings = append(bindings, &ac.Binding{NodeID: node, ProbeType: probe, TemplateID: uuid.Nil, DesiredGeneration: generation, Enabled: true, Variables: ac.JSON("{}")})
	}
	if len(releases) > 0 {
		if err := s.repo.SaveAgentBindings(ctx, bindings, releases, expected); err != nil {
			return nil, err
		}
	}
	return s.CurrentNodeConfigs(ctx, node)
}

func normalizeNodeUpdateFiles(update NodeProbeConfigInput, current CurrentNodeConfig) ([]ac.InitialConfigFile, error) {
	files := update.Files
	if len(files) == 0 && len(current.Files) == 1 {
		files = []ac.InitialConfigFile{{Path: current.Files[0].Path, Content: update.Content}}
	}
	if len(files) != len(current.Files) {
		return nil, fmt.Errorf("all %d current files are required", len(current.Files))
	}
	currentPaths := make(map[string]struct{}, len(current.Files))
	for _, file := range current.Files {
		currentPaths[file.Path] = struct{}{}
	}
	seen := make(map[string]struct{}, len(files))
	normalized := make([]ac.InitialConfigFile, 0, len(files))
	for _, file := range files {
		path := filepath.Clean(strings.TrimSpace(file.Path))
		if _, exists := currentPaths[path]; !exists {
			return nil, fmt.Errorf("path %s is not part of the current configuration", path)
		}
		if _, duplicate := seen[path]; duplicate {
			return nil, fmt.Errorf("path %s is duplicated", path)
		}
		seen[path] = struct{}{}
		normalized = append(normalized, ac.InitialConfigFile{Path: path, Content: file.Content, ContentSHA256: ac.HashContent(file.Content)})
	}
	sort.Slice(normalized, func(i, j int) bool { return normalized[i].Path < normalized[j].Path })
	return normalized, nil
}

func sameConfigFiles(left, right []ac.InitialConfigFile) bool {
	if len(left) != len(right) {
		return false
	}
	rightByPath := make(map[string]string, len(right))
	for _, file := range right {
		rightByPath[file.Path] = file.Content
	}
	for _, file := range left {
		if content, ok := rightByPath[file.Path]; !ok || content != file.Content {
			return false
		}
	}
	return true
}

func (s *Service) BatchAssign(ctx core.Context, input BatchAssignmentInput, user int64) (*BatchResult, error) {
	targets, err := s.nodesInClusters(ctx, input.ClusterIDs)
	if err != nil {
		return nil, err
	}
	template, err := s.repo.GetAgentTemplate(ctx, input.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("get batch template: %w", err)
	}
	configs, err := templateProbeConfigs(template)
	if err != nil {
		return nil, err
	}
	byProbe := make(map[string]ac.TemplateProbeConfig, len(configs))
	for _, config := range configs {
		byProbe[config.ProbeType] = config
	}
	bindings := make([]*ac.Binding, 0, len(targets)*len(configs))
	releases := make([]*ac.Release, 0, len(targets)*len(configs))
	expected := make(map[string]int64, len(targets)*len(configs))
	for _, node := range targets {
		current, err := s.CurrentNodeConfigs(ctx, node.ID)
		if err != nil {
			return nil, fmt.Errorf("get current configuration for %s: %w", node.ID, err)
		}
		if len(current.Configs) == 0 {
			return nil, fmt.Errorf("node %s has no probe configuration", node.ID)
		}
		if len(current.Configs) != len(configs) {
			return nil, fmt.Errorf("template contains %d probes but node %s has %d", len(configs), node.ID, len(current.Configs))
		}
		for _, currentConfig := range current.Configs {
			probe := currentConfig.ProbeType
			selected, exists := byProbe[probe]
			if !exists {
				return nil, fmt.Errorf("template does not contain probe %s required by node %s", probe, node.ID)
			}
			release, binding, previous, err := s.prepareAssignment(ctx, node.ID, probe, template, selected.Files, AssignmentInput{TemplateID: template.ID, Variables: input.Variables})
			if err != nil {
				return nil, fmt.Errorf("prepare template assignment for %s/%s: %w", node.ID, probe, err)
			}
			releases = append(releases, release)
			bindings = append(bindings, binding)
			expected[releaseKey(node.ID, probe)] = previous
		}
	}
	if err := s.repo.SaveAgentBindings(ctx, bindings, releases, expected); err != nil {
		return nil, err
	}
	return &BatchResult{TargetNodeCount: len(targets), SucceededCount: len(targets)}, nil
}

func (s *Service) Assign(ctx core.Context, node, probe string, in AssignmentInput) (*ac.Release, error) {
	if _, e := s.repo.GetAgentNode(ctx, node); e != nil {
		return nil, e
	}
	t, e := s.repo.GetAgentTemplate(ctx, in.TemplateID)
	if e != nil {
		return nil, e
	}
	probe = ac.NormalizeProbe(probe)
	configs, e := templateProbeConfigs(t)
	if e != nil {
		return nil, e
	}
	var selected *ac.TemplateProbeConfig
	for index := range configs {
		if configs[index].ProbeType == probe {
			selected = &configs[index]
			break
		}
	}
	if selected == nil {
		return nil, fmt.Errorf("template does not contain probe %s", probe)
	}
	release, binding, previous, err := s.prepareAssignment(ctx, node, probe, t, selected.Files, in)
	if err != nil {
		return nil, err
	}
	if in.ExpectedGeneration > 0 && previous != in.ExpectedGeneration {
		return nil, fmt.Errorf("generation conflict: expected %d, actual %d", in.ExpectedGeneration, previous)
	}
	if err := s.repo.SaveAgentBindings(ctx, []*ac.Binding{binding}, []*ac.Release{release}, map[string]int64{releaseKey(node, probe): previous}); err != nil {
		return nil, err
	}
	return release, nil
}

func templateProbeConfigs(t *ac.Template) ([]ac.TemplateProbeConfig, error) {
	if t == nil {
		return nil, fmt.Errorf("template is required")
	}
	if len(t.Configs) == 0 || string(t.Configs) == "null" || string(t.Configs) == "[]" || string(t.Configs) == "{}" {
		return nil, fmt.Errorf("template contains no probe configuration")
	}
	var configs []ac.TemplateProbeConfig
	if err := json.Unmarshal(t.Configs, &configs); err != nil {
		return nil, fmt.Errorf("decode template configurations: %w", err)
	}
	normalized, _, _, err := normalizeTemplatePayload(TemplateInput{Configs: configs})
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(normalized, &configs); err != nil {
		return nil, err
	}
	return configs, nil
}

func hydrateTemplateConfigs(t *ac.Template) error {
	configs, err := templateProbeConfigs(t)
	if err != nil {
		return err
	}
	data, err := json.Marshal(configs)
	if err != nil {
		return err
	}
	t.Configs = ac.JSON(data)
	return nil
}

func sameConfigFilePaths(left, right []ac.InitialConfigFile) bool {
	if len(left) != len(right) {
		return false
	}
	rightPaths := make(map[string]struct{}, len(right))
	for _, file := range right {
		rightPaths[file.Path] = struct{}{}
	}
	for _, file := range left {
		if _, ok := rightPaths[file.Path]; !ok {
			return false
		}
	}
	return true
}

func (s *Service) prepareAssignment(ctx core.Context, node, probe string, t *ac.Template, definedFiles []ac.InitialConfigFile, in AssignmentInput) (*ac.Release, *ac.Binding, int64, error) {
	missing := []string{}
	invalid := []string{}
	used := map[string]bool{}
	values := map[string]string{}
	for _, v := range t.Variables {
		value, ok := in.Variables[v.Name]
		if !ok || value == "" {
			value = v.DefaultValue
		}
		if value == "" && v.Required {
			missing = append(missing, v.Name)
			continue
		}
		if v.ValidationRegex != "" && value != "" {
			re, x := regexp.Compile(v.ValidationRegex)
			if x != nil || !re.MatchString(value) {
				invalid = append(invalid, v.Name)
				continue
			}
		}
		values[v.Name] = value
		used[v.Name] = true
	}
	unused := []string{}
	for k := range in.Variables {
		if !used[k] {
			unused = append(unused, k)
		}
	}
	if len(missing)+len(invalid) > 0 {
		return nil, nil, 0, &ValidationError{MissingVariables: missing, InvalidVariables: invalid, UnusedVariables: unused}
	}
	current, e := s.CurrentNodeConfig(ctx, node, probe)
	if e != nil {
		return nil, nil, 0, e
	}
	renderedContent := ""
	releaseFiles := ac.JSON("null")
	contentSHA256 := ""
	if !sameConfigFilePaths(definedFiles, current.Files) {
		return nil, nil, 0, fmt.Errorf("template file set does not match the current %d-file configuration", len(current.Files))
	}
	definedFiles = append([]ac.InitialConfigFile(nil), definedFiles...)
	for index := range definedFiles {
		definedFiles[index].Content, e = Render(definedFiles[index].Content, values)
		if e != nil {
			return nil, nil, 0, fmt.Errorf("render template file %s: %w", definedFiles[index].Path, e)
		}
		definedFiles[index].ContentSHA256 = ac.HashContent(definedFiles[index].Content)
	}
	data, marshalErr := json.Marshal(definedFiles)
	if marshalErr != nil {
		return nil, nil, 0, marshalErr
	}
	releaseFiles = ac.JSON(data)
	contentSHA256 = ac.HashInitialConfigFiles(definedFiles)
	if len(definedFiles) == 1 {
		renderedContent = definedFiles[0].Content
	}
	last, e := s.repo.GetAgentRelease(ctx, node, probe)
	generation := int64(1)
	previous := int64(0)
	if e == nil {
		previous = last.Generation
		generation = last.Generation + 1
	} else if !errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, nil, 0, e
	}
	logSources, e := normalizeLogSources(in.LogSources)
	if e != nil {
		return nil, nil, 0, e
	}
	if in.LogSources == nil && last != nil && len(last.LogSources) > 0 {
		logSources = append(ac.JSON(nil), last.LogSources...)
	}
	vars, _ := json.Marshal(values)
	r := &ac.Release{ID: uuid.New(), NodeID: node, ProbeType: probe, Generation: generation, TemplateID: t.ID, TemplateVersion: t.Version, Variables: ac.JSON(vars), RenderedContent: renderedContent, ContentSHA256: contentSHA256, Files: releaseFiles, LogSources: logSources}
	b := &ac.Binding{NodeID: node, ProbeType: probe, TemplateID: t.ID, DesiredGeneration: generation, Enabled: true, Variables: ac.JSON(vars)}
	return r, b, previous, nil
}

func normalizeLogSources(input *[]ac.LogSource) (ac.JSON, error) {
	const (
		defaultStdoutMaxBytes int64 = 10 * 1024 * 1024
		minimumStdoutMaxBytes int64 = 64 * 1024
		maximumStdoutMaxBytes int64 = 1024 * 1024 * 1024
	)
	if input == nil {
		return ac.JSON("null"), nil
	}
	if len(*input) > 20 {
		return nil, fmt.Errorf("at most 20 log sources are allowed")
	}
	normalized := make([]ac.LogSource, 0, len(*input))
	for _, source := range *input {
		source.Type = strings.ToLower(strings.TrimSpace(source.Type))
		if source.Type == "" {
			source.Type = "file"
		}
		if source.Type == "stdout" {
			source.Path = ""
			source.Pattern = ""
			if source.MaxBytes == 0 {
				source.MaxBytes = defaultStdoutMaxBytes
			}
			if source.MaxBytes < minimumStdoutMaxBytes || source.MaxBytes > maximumStdoutMaxBytes {
				return nil, fmt.Errorf("stdout maxBytes must be between %d and %d", minimumStdoutMaxBytes, maximumStdoutMaxBytes)
			}
			normalized = append(normalized, source)
			continue
		}
		if source.Type != "file" {
			return nil, fmt.Errorf("unsupported log source type: %s", source.Type)
		}
		if source.MaxBytes != 0 {
			return nil, fmt.Errorf("maxBytes is only supported for stdout log sources")
		}
		source.Path = filepath.Clean(strings.TrimSpace(source.Path))
		source.Pattern = strings.TrimSpace(source.Pattern)
		if !filepath.IsAbs(source.Path) {
			return nil, fmt.Errorf("log source must be an absolute path: %s", source.Path)
		}
		if source.Pattern != "" {
			if _, err := regexp.Compile(source.Pattern); err != nil {
				return nil, fmt.Errorf("invalid log source pattern for %s: %v", source.Path, err)
			}
		}
		normalized = append(normalized, source)
	}
	data, _ := json.Marshal(normalized)
	return ac.JSON(data), nil
}
func (s *Service) GetConfig(ctx core.Context, node, probe string, known int64) (*ac.Release, error) {
	r, e := s.repo.GetAgentRelease(ctx, node, ac.NormalizeProbe(probe))
	if e != nil {
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, e
	}
	if r.Generation <= known {
		return nil, nil
	}
	return r, nil
}
func (s *Service) Report(ctx core.Context, state *ac.ApplyState) error {
	return s.repo.ReportAgentState(ctx, state)
}
func (s *Service) ListStates(ctx core.Context) ([]ac.ApplyState, error) {
	return s.repo.ListAgentStates(ctx)
}
func (s *Service) CreateCommand(ctx core.Context, node, probe string, in CommandInput, user int64) (*ac.Command, error) {
	typ := strings.ToUpper(strings.TrimSpace(in.Type))
	if typ != "RESTART" && typ != "OPEN_CIRCUIT" && typ != "RESET_CIRCUIT" && typ != "LIST_LOG_FILES" && typ != "PREVIEW_LOG_FILE" && typ != "DOWNLOAD_LOG_FILE" {
		return nil, fmt.Errorf("unsupported command type")
	}
	if typ == "OPEN_CIRCUIT" || typ == "RESET_CIRCUIT" {
		if in.ExpectedCircuitEpoch == nil {
			return nil, fmt.Errorf("expected circuit epoch is required")
		}
	}
	if typ == "OPEN_CIRCUIT" && strings.TrimSpace(in.Reason) == "" {
		return nil, fmt.Errorf("circuit reason is required")
	}
	if (typ == "PREVIEW_LOG_FILE" || typ == "DOWNLOAD_LOG_FILE") && strings.TrimSpace(in.FilePath) == "" {
		return nil, fmt.Errorf("log file path is required")
	}
	if in.TailLines < 0 || in.TailLines > 100 {
		return nil, fmt.Errorf("tail lines must be between 0 and 100")
	}
	c := &ac.Command{ID: uuid.New(), NodeID: node, ProbeType: ac.NormalizeProbe(probe), Type: typ, Status: "PENDING", RequestedBy: user, ExpectedRuntimeEpoch: in.ExpectedRuntimeEpoch, ExpectedCircuitEpoch: in.ExpectedCircuitEpoch, Reason: strings.TrimSpace(in.Reason), FilePath: strings.TrimSpace(in.FilePath), TailLines: in.TailLines}
	return c, s.repo.CreateAgentCommand(ctx, c)
}
func (s *Service) ListCommands(ctx core.Context, node, probe string, limit int) ([]ac.Command, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	commands, err := s.repo.ListAgentCommands(ctx, node, ac.NormalizeProbe(probe), limit)
	if err != nil {
		return nil, err
	}
	for index := range commands {
		commands[index].LeaseToken = ""
	}
	return commands, nil
}
func (s *Service) Commands(ctx core.Context, node, probe, claimant string) ([]ac.Command, error) {
	if strings.TrimSpace(claimant) == "" {
		return nil, fmt.Errorf("claimant id is required")
	}
	return s.repo.ClaimAgentCommands(ctx, node, ac.NormalizeProbe(probe), claimant, s.now(), s.leaseTTL)
}
func (s *Service) TransitionCommand(ctx core.Context, id uuid.UUID, nodeID, claimant, token, status, result string) (*ac.Command, error) {
	return s.repo.TransitionAgentCommand(ctx, id, nodeID, claimant, token, strings.ToUpper(status), result, s.now())
}
