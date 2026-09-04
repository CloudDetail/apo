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
	"sync"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	integrationmodel "github.com/CloudDetail/apo/backend/pkg/model/integration"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	repo            Repository
	now             func() time.Time
	leaseTTL        time.Duration
	defaultSettings ac.AgentSettings
	reportMu        sync.Mutex
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
	GetCluster(core.Context, string) (integrationmodel.Cluster, error)
	GetNamespaceInjection(core.Context, string) (*ac.NamespaceInjection, error)
	SaveClusterInjectionConfig(core.Context, *ac.ClusterInjectionConfig, *ac.Delivery) error
	SaveClusterWorkloadInventory(core.Context, *ac.ClusterWorkloadInventory) error
	GetClusterWorkloadInventory(core.Context, string) (*ac.ClusterWorkloadInventory, error)
	GetOdigletReadiness(core.Context, string, time.Time) (*bool, error)
	SaveAgentInitialConfigIfUnassigned(core.Context, *ac.InitialConfig) (bool, error)
	GetAgentInitialConfig(core.Context, string, string) (*ac.InitialConfig, error)
	ListAgentInitialConfigs(core.Context, string) ([]ac.InitialConfig, error)
	GetGlobalCircuitConfig(core.Context) (*ac.GlobalCircuitConfig, error)
	SaveGlobalCircuitConfig(core.Context, *ac.GlobalCircuitConfig, *ac.Delivery) error
	GetNodeCircuitConfig(core.Context, string) (*ac.NodeCircuitConfig, error)
	SaveNodeCircuitConfigs(core.Context, []*ac.NodeCircuitConfig, *ac.Delivery) error
	GetNodeLogConfig(core.Context, string, string) (*ac.NodeLogConfig, error)
	SaveNodeLogConfigs(core.Context, []*ac.NodeLogConfig, *ac.Delivery) error
	SaveAgentBinding(core.Context, *ac.Binding, *ac.Release) error
	SaveAgentBindings(core.Context, []*ac.Binding, []*ac.Release, map[string]int64, *ac.Delivery) error
	GetAgentRelease(core.Context, string, string) (*ac.Release, error)
	GetAgentReleaseGeneration(core.Context, string, string, int64) (*ac.Release, error)
	ListAgentReleases(core.Context, string, string, int) ([]ac.Release, error)
	ListLatestAgentReleases(core.Context, string) ([]ac.Release, error)
	ListAgentDeliveries(core.Context, int) ([]ac.Delivery, error)
	ReportAgentState(core.Context, *ac.ApplyState) error
	GetAgentState(core.Context, string, string) (*ac.ApplyState, error)
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
	defaultSystemMaxCPUPercent                = 90
	defaultSystemMaxMemoryPercent             = 90
	defaultSystemRecoveryCPUPercent           = 80
	defaultSystemRecoveryMemoryPercent        = 80
	defaultSystemTriggerDurationSeconds       = 30
	defaultSystemRecoveryDurationSeconds      = 30
	agentStateFreshnessTimeout                = 2 * time.Minute
)

var kubernetesNamespacePattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
var kubernetesWorkloadNamePattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$`)
var serviceNameVariablePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type ClusterInjectionControlInput struct {
	NamespaceInjection ac.NamespaceInjection `json:"namespaceInjection"`
}

type ClusterInjectionControl struct {
	NamespaceInjection  ac.NamespaceInjection `json:"namespaceInjection"`
	Inventory           ac.WorkloadInventory  `json:"inventory"`
	InventoryObservedAt *time.Time            `json:"inventoryObservedAt,omitempty"`
}

type WorkloadInventoryInput struct {
	Namespaces []string         `json:"namespaces"`
	Workloads  []ac.WorkloadRef `json:"workloads"`
}

func normalizeInjectionNamespaces(namespaces []string) ([]string, error) {
	seen := make(map[string]struct{}, len(namespaces))
	result := make([]string, 0, len(namespaces))
	for _, namespace := range namespaces {
		namespace = strings.TrimSpace(namespace)
		if len(namespace) == 0 || len(namespace) > 63 || !kubernetesNamespacePattern.MatchString(namespace) {
			return nil, fmt.Errorf("invalid Kubernetes namespace %q", namespace)
		}
		if _, exists := seen[namespace]; exists {
			continue
		}
		seen[namespace] = struct{}{}
		result = append(result, namespace)
	}
	return result, nil
}

func normalizeClusterInjection(config ac.NamespaceInjection) (ac.NamespaceInjection, error) {
	if len(config.InstrumentNS)+len(config.InstrumentDisabledNS)+len(config.Namespaces)+len(config.Workloads) > 10000 {
		return ac.NamespaceInjection{}, errors.New("namespace injection config exceeds 10000 entries")
	}
	disabled, err := normalizeInjectionNamespaces(config.InstrumentDisabledNS)
	if err != nil {
		return ac.NamespaceInjection{}, err
	}
	enabled, err := normalizeInjectionNamespaces(config.InstrumentNS)
	if err != nil {
		return ac.NamespaceInjection{}, err
	}
	disabledSet := make(map[string]struct{}, len(disabled))
	for _, namespace := range disabled {
		disabledSet[namespace] = struct{}{}
	}
	filteredEnabled := enabled[:0]
	for _, namespace := range enabled {
		if _, excluded := disabledSet[namespace]; !excluded {
			filteredEnabled = append(filteredEnabled, namespace)
		}
	}
	workloads, err := normalizeInjectionWorkloads(config.Workloads)
	if err != nil {
		return ac.NamespaceInjection{}, err
	}
	serviceName, err := normalizeServiceNamePolicy(config.ServiceName)
	if err != nil {
		return ac.NamespaceInjection{}, err
	}
	namespaceRules, err := normalizeNamespaceRules(config.Namespaces)
	if err != nil {
		return ac.NamespaceInjection{}, err
	}
	if len(namespaceRules) == 0 {
		for _, namespace := range filteredEnabled {
			namespaceRules = append(namespaceRules, ac.NamespaceRule{Namespace: namespace, Mode: "enabledFuture"})
		}
		for _, namespace := range disabled {
			namespaceRules = append(namespaceRules, ac.NamespaceRule{Namespace: namespace, Mode: "disabled"})
		}
	}
	return ac.NamespaceInjection{
		InstrumentAll:        config.InstrumentAll,
		InstrumentNS:         filteredEnabled,
		InstrumentDisabledNS: disabled,
		Namespaces:           namespaceRules,
		Workloads:            workloads,
		ServiceName:          serviceName,
	}, nil
}

func normalizeNamespaceRules(rules []ac.NamespaceRule) ([]ac.NamespaceRule, error) {
	result := make([]ac.NamespaceRule, 0, len(rules))
	seen := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		rule.Namespace = strings.TrimSpace(rule.Namespace)
		rule.Mode = strings.TrimSpace(rule.Mode)
		if len(rule.Namespace) == 0 || len(rule.Namespace) > 63 || !kubernetesNamespacePattern.MatchString(rule.Namespace) {
			return nil, fmt.Errorf("invalid Kubernetes namespace %q", rule.Namespace)
		}
		switch rule.Mode {
		case "disabled", "enabled", "enabledFuture":
		default:
			return nil, fmt.Errorf("unsupported namespace injection mode %q", rule.Mode)
		}
		if _, exists := seen[rule.Namespace]; exists {
			return nil, fmt.Errorf("duplicate namespace rule %q", rule.Namespace)
		}
		seen[rule.Namespace] = struct{}{}
		result = append(result, rule)
	}
	return result, nil
}

func normalizeInjectionWorkloads(workloads []ac.WorkloadInjection) ([]ac.WorkloadInjection, error) {
	result := make([]ac.WorkloadInjection, 0, len(workloads))
	seen := make(map[string]struct{}, len(workloads))
	for _, workload := range workloads {
		workload.Namespace = strings.TrimSpace(workload.Namespace)
		workload.Kind = strings.ToLower(strings.TrimSpace(workload.Kind))
		workload.Name = strings.TrimSpace(workload.Name)
		if len(workload.Namespace) == 0 || len(workload.Namespace) > 63 || !kubernetesNamespacePattern.MatchString(workload.Namespace) {
			return nil, fmt.Errorf("invalid Kubernetes namespace %q", workload.Namespace)
		}
		switch workload.Kind {
		case "deployment", "statefulset", "daemonset":
		default:
			return nil, fmt.Errorf("unsupported workload kind %q", workload.Kind)
		}
		if len(workload.Name) == 0 || len(workload.Name) > 253 || !kubernetesWorkloadNamePattern.MatchString(workload.Name) {
			return nil, fmt.Errorf("invalid Kubernetes workload name %q", workload.Name)
		}
		key := workload.Namespace + "\x00" + workload.Kind + "\x00" + workload.Name
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("duplicate workload rule %s/%s/%s", workload.Namespace, workload.Kind, workload.Name)
		}
		seen[key] = struct{}{}
		result = append(result, workload)
	}
	return result, nil
}

func normalizeWorkloadInventory(input WorkloadInventoryInput) (ac.WorkloadInventory, error) {
	if len(input.Namespaces)+len(input.Workloads) > 50000 {
		return ac.WorkloadInventory{}, errors.New("workload inventory exceeds 50000 entries")
	}
	namespaces, err := normalizeInjectionNamespaces(input.Namespaces)
	if err != nil {
		return ac.WorkloadInventory{}, err
	}
	seenWorkloads := make(map[string]struct{}, len(input.Workloads))
	workloads := make([]ac.WorkloadRef, 0, len(input.Workloads))
	for _, item := range input.Workloads {
		normalized, err := normalizeInjectionWorkloads([]ac.WorkloadInjection{{
			Namespace: item.Namespace, Kind: item.Kind, Name: item.Name, Enabled: true,
		}})
		if err != nil {
			return ac.WorkloadInventory{}, err
		}
		item.Namespace = normalized[0].Namespace
		item.Kind = normalized[0].Kind
		item.Name = normalized[0].Name
		key := item.Namespace + "\x00" + item.Kind + "\x00" + item.Name
		if _, exists := seenWorkloads[key]; exists {
			continue
		}
		seenWorkloads[key] = struct{}{}
		workloads = append(workloads, item)
	}
	return ac.WorkloadInventory{Namespaces: namespaces, Workloads: workloads}, nil
}

func normalizeServiceNamePolicy(policy *ac.ServiceNamePolicy) (*ac.ServiceNamePolicy, error) {
	if policy == nil {
		return nil, nil
	}
	if len(policy.Mappings) > 1000 || len(policy.Variables) > 128 || len(policy.DefaultFormat) > 1024 {
		return nil, errors.New("service name policy exceeds supported limits")
	}
	result := &ac.ServiceNamePolicy{DefaultFormat: strings.TrimSpace(policy.DefaultFormat)}
	seenVariables := make(map[string]struct{}, len(policy.Variables))
	for _, variable := range policy.Variables {
		variable.Name = strings.TrimSpace(variable.Name)
		variable.Source = strings.TrimSpace(variable.Source)
		if !serviceNameVariablePattern.MatchString(variable.Name) {
			return nil, fmt.Errorf("invalid service name variable %q", variable.Name)
		}
		switch variable.Source {
		case "workloadName", "workflowName", "deployName", "containerName":
		default:
			return nil, fmt.Errorf("unsupported service name variable source %q", variable.Source)
		}
		if len(variable.Regex) > 1024 || len(variable.Replacement) > 1024 {
			return nil, fmt.Errorf("service name variable %q exceeds supported limits", variable.Name)
		}
		if variable.Regex != "" {
			if _, err := regexp.Compile(variable.Regex); err != nil {
				return nil, fmt.Errorf("invalid service name regex for %q: %w", variable.Name, err)
			}
		}
		if _, exists := seenVariables[variable.Name]; exists {
			return nil, fmt.Errorf("duplicate service name variable %q", variable.Name)
		}
		seenVariables[variable.Name] = struct{}{}
		result.Variables = append(result.Variables, variable)
	}
	for _, mapping := range policy.Mappings {
		mapping.NamespacePattern = strings.TrimSpace(mapping.NamespacePattern)
		mapping.WorkloadPattern = strings.TrimSpace(mapping.WorkloadPattern)
		mapping.ServiceName = strings.TrimSpace(mapping.ServiceName)
		if mapping.NamespacePattern == "" || mapping.WorkloadPattern == "" || mapping.ServiceName == "" || len(mapping.ServiceName) > 255 {
			return nil, errors.New("service name mapping requires namespace pattern, workload pattern, and service name")
		}
		if _, err := regexp.Compile(mapping.NamespacePattern); err != nil {
			return nil, fmt.Errorf("invalid service name namespace regex %q: %w", mapping.NamespacePattern, err)
		}
		if _, err := regexp.Compile(mapping.WorkloadPattern); err != nil {
			return nil, fmt.Errorf("invalid service name workload regex %q: %w", mapping.WorkloadPattern, err)
		}
		result.Mappings = append(result.Mappings, mapping)
	}
	return result, nil
}

func (s *Service) GetClusterInjectionControl(ctx core.Context, clusterID string) (*ClusterInjectionControl, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return nil, errors.New("cluster ID is required")
	}
	config, err := s.repo.GetNamespaceInjection(ctx, clusterID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	control := &ClusterInjectionControl{}
	if config != nil {
		control.NamespaceInjection = *config
	}
	storedInventory, err := s.repo.GetClusterWorkloadInventory(ctx, clusterID)
	if err == nil {
		if err := json.Unmarshal(storedInventory.Inventory, &control.Inventory); err != nil {
			return nil, fmt.Errorf("decode cluster workload inventory: %w", err)
		}
		control.InventoryObservedAt = &storedInventory.ObservedAt
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return control, nil
}

func (s *Service) ReportWorkloadInventory(ctx core.Context, clusterID, nodeID string, input WorkloadInventoryInput) error {
	clusterID = strings.TrimSpace(clusterID)
	nodeID = strings.TrimSpace(nodeID)
	if clusterID == "" || nodeID == "" {
		return errors.New("cluster ID and node ID are required")
	}
	inventory, err := normalizeWorkloadInventory(input)
	if err != nil {
		return err
	}
	data, err := json.Marshal(inventory)
	if err != nil {
		return err
	}
	return s.repo.SaveClusterWorkloadInventory(ctx, &ac.ClusterWorkloadInventory{
		ClusterID: clusterID, Inventory: ac.JSON(data), ReportedBy: nodeID, ObservedAt: s.now(),
	})
}

func (s *Service) SaveClusterInjectionControl(ctx core.Context, clusterID string, input ClusterInjectionControlInput, user int64) (*ClusterInjectionControl, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return nil, errors.New("cluster ID is required")
	}
	config, err := normalizeClusterInjection(input.NamespaceInjection)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	delivery, err := buildRecordedDelivery(deliveryTypeClusterInjection, clusterID, config, nil, user)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveClusterInjectionConfig(ctx, &ac.ClusterInjectionConfig{
		ClusterID: clusterID, Configured: true, Config: ac.JSON(data), UpdatedBy: user,
	}, delivery); err != nil {
		return nil, err
	}
	return s.GetClusterInjectionControl(ctx, clusterID)
}

func loadDefaultAgentSettings() ac.AgentSettings {
	settings := ac.CircuitBreakerSettings{
		Enabled:                     true,
		MaxCPUCores:                 defaultCircuitMaxCPUCores,
		MaxMemoryBytes:              defaultCircuitMaxMemoryBytes,
		DurationSeconds:             defaultCircuitDurationSeconds,
		AutoRecover:                 true,
		RecoveryInitialDelaySeconds: defaultCircuitRecoveryInitialDelaySeconds,
		RecoveryMaxDelaySeconds:     defaultCircuitRecoveryMaxDelaySeconds,
		SystemPressure: &ac.SystemPressureSettings{
			Enabled:                 false,
			MaxCPUPercent:           defaultSystemMaxCPUPercent,
			MaxMemoryPercent:        defaultSystemMaxMemoryPercent,
			RecoveryCPUPercent:      defaultSystemRecoveryCPUPercent,
			RecoveryMemoryPercent:   defaultSystemRecoveryMemoryPercent,
			TriggerDurationSeconds:  defaultSystemTriggerDurationSeconds,
			RecoveryDurationSeconds: defaultSystemRecoveryDurationSeconds,
		},
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
			SystemPressure: &ac.SystemPressureSettings{
				Enabled:                 false,
				MaxCPUPercent:           defaultSystemMaxCPUPercent,
				MaxMemoryPercent:        defaultSystemMaxMemoryPercent,
				RecoveryCPUPercent:      defaultSystemRecoveryCPUPercent,
				RecoveryMemoryPercent:   defaultSystemRecoveryMemoryPercent,
				TriggerDurationSeconds:  defaultSystemTriggerDurationSeconds,
				RecoveryDurationSeconds: defaultSystemRecoveryDurationSeconds,
			},
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
	return s.SettingsForProbeAndCluster(ctx, nodeID, probeType, "")
}

func (s *Service) SettingsForProbeAndCluster(ctx core.Context, nodeID, probeType, clusterID string) (ac.AgentSettings, error) {
	if _, err := s.repo.GetAgentNode(ctx, nodeID); err != nil {
		return ac.AgentSettings{}, err
	}
	settings := ac.AgentSettings{UseLocalDefaults: true}
	if clusterID != "" {
		namespaceInjection, err := s.repo.GetNamespaceInjection(ctx, clusterID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Remote-control mode is fail closed. No persisted selection means an
			// empty remote policy, never a fallback to the instrumentor ConfigMap.
			settings.NamespaceInjection = &ac.NamespaceInjection{}
		} else if err != nil {
			log.Printf("agentconfig: namespace injection settings are unavailable for cluster %s: %v", clusterID, err)
		} else {
			settings.NamespaceInjection = namespaceInjection
		}
		odigletReady, err := s.repo.GetOdigletReadiness(ctx, clusterID, s.now().Add(-agentStateFreshnessTimeout))
		if err != nil {
			log.Printf("agentconfig: odiglet readiness is unavailable for cluster %s: %v", clusterID, err)
		} else {
			settings.OdigletReady = odigletReady
		}
	}
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
	ClusterType string                   `json:"clusterType"`
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
type NodeVariables struct {
	NodeID           string            `json:"nodeId"`
	InitialClusterID string            `json:"initialClusterId"`
	Reported         map[string]string `json:"reported"`
	Overrides        map[string]string `json:"overrides"`
	Effective        map[string]string `json:"effective"`
}

type NodeClusterIdentity struct {
	NodeID           string `json:"nodeId"`
	InstanceID       string `json:"instanceId"`
	ClusterID        string `json:"clusterId"`
	InitialClusterID string `json:"initialClusterId"`
	ClusterIP        string `json:"clusterIp,omitempty"`
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

const (
	deliveryTypeNodeConfig       = "NODE_CONFIG"
	deliveryTypeBatchTemplate    = "BATCH_TEMPLATE"
	deliveryTypeNodeInjection    = "NODE_INJECTION"
	deliveryTypeBatchInjection   = "BATCH_INJECTION"
	deliveryTypeClusterInjection = "CLUSTER_INJECTION"
	deliveryTypeGlobalCircuit    = "GLOBAL_CIRCUIT"
	deliveryTypeNodeCircuit      = "NODE_CIRCUIT"
	deliveryTypeBatchCircuit     = "BATCH_CIRCUIT"
	deliveryTypeNodeLogSources   = "NODE_LOG_SOURCES"
	deliveryTypeBatchLogSources  = "BATCH_LOG_SOURCES"
	deliveryTypeLogDefaults      = "LOG_COLLECTION_DEFAULTS"
	deliveryTypeHostLog          = "HOST_LOG_COLLECTION"
	deliveryTypeContainerLog     = "CONTAINER_LOG_COLLECTION"
)

type deliveryReleaseSnapshot struct {
	NodeID        string                 `json:"nodeId"`
	ProbeType     string                 `json:"probeType"`
	Generation    int64                  `json:"generation"`
	Content       string                 `json:"content,omitempty"`
	Files         []ac.InitialConfigFile `json:"files,omitempty"`
	LogSources    []ac.LogSource         `json:"logSources,omitempty"`
	ContentSHA256 string                 `json:"contentSha256"`
}

func buildRecordedDelivery(deliveryType, scope string, configuration any, nodes []ac.Node, user int64) (*ac.Delivery, error) {
	data, err := json.Marshal(configuration)
	if err != nil {
		return nil, fmt.Errorf("encode delivery configuration: %w", err)
	}
	delivery := &ac.Delivery{
		ID: uuid.New(), Type: deliveryType, Scope: scope, Configuration: ac.JSON(data),
		RequestedBy: user, TargetNodeCount: len(nodes), Targets: make([]ac.DeliveryTarget, 0, len(nodes)),
	}
	for _, node := range nodes {
		delivery.Targets = append(delivery.Targets, ac.DeliveryTarget{
			NodeID: node.ID, NodeName: node.Name, ClusterID: node.ClusterID, Generations: ac.JSON(`{}`),
		})
	}
	return delivery, nil
}

func (s *Service) buildDelivery(ctx core.Context, deliveryType, scope string, template *ac.Template, releases []*ac.Release, user int64) (*ac.Delivery, error) {
	byNode := make(map[string]map[string]int64)
	snapshots := make([]deliveryReleaseSnapshot, 0, len(releases))
	for _, release := range releases {
		if release == nil {
			continue
		}
		generations := byNode[release.NodeID]
		if generations == nil {
			generations = make(map[string]int64)
			byNode[release.NodeID] = generations
		}
		generations[ac.NormalizeProbe(release.ProbeType)] = release.Generation
		snapshot := deliveryReleaseSnapshot{
			NodeID: release.NodeID, ProbeType: ac.NormalizeProbe(release.ProbeType), Generation: release.Generation,
			Content: release.RenderedContent, ContentSHA256: release.ContentSHA256,
		}
		if len(release.Files) > 0 && string(release.Files) != "null" {
			if err := json.Unmarshal(release.Files, &snapshot.Files); err != nil {
				return nil, fmt.Errorf("decode delivery release files for %s/%s: %w", release.NodeID, release.ProbeType, err)
			}
		}
		if len(release.LogSources) > 0 && string(release.LogSources) != "null" {
			if err := json.Unmarshal(release.LogSources, &snapshot.LogSources); err != nil {
				return nil, fmt.Errorf("decode delivery log sources for %s/%s: %w", release.NodeID, release.ProbeType, err)
			}
		}
		snapshots = append(snapshots, snapshot)
	}
	if len(byNode) == 0 {
		return nil, errors.New("delivery has no released configuration")
	}
	nodeIDs := make([]string, 0, len(byNode))
	for nodeID := range byNode {
		nodeIDs = append(nodeIDs, nodeID)
	}
	sort.Strings(nodeIDs)
	configuration, err := json.Marshal(map[string]any{"releases": snapshots})
	if err != nil {
		return nil, fmt.Errorf("encode delivery configuration: %w", err)
	}
	delivery := &ac.Delivery{
		ID: uuid.New(), Type: deliveryType, Scope: scope, RequestedBy: user,
		Configuration: ac.JSON(configuration), ReceiptTracked: true,
		TargetNodeCount: len(nodeIDs), Targets: make([]ac.DeliveryTarget, 0, len(nodeIDs)),
	}
	if template != nil {
		delivery.TemplateID = template.ID
		delivery.TemplateName = template.Name
	}
	for _, nodeID := range nodeIDs {
		node, err := s.repo.GetAgentNode(ctx, nodeID)
		if err != nil {
			return nil, fmt.Errorf("get delivery target %s: %w", nodeID, err)
		}
		generations, err := json.Marshal(byNode[nodeID])
		if err != nil {
			return nil, err
		}
		delivery.Targets = append(delivery.Targets, ac.DeliveryTarget{
			NodeID: nodeID, NodeName: node.Name, ClusterID: node.ClusterID, Generations: ac.JSON(generations),
		})
	}
	return delivery, nil
}

type CurrentNodeConfig struct {
	NodeID             string                 `json:"nodeId"`
	ProbeType          string                 `json:"probeType"`
	Generation         int64                  `json:"generation"`
	ExpectedGeneration int64                  `json:"expectedGeneration"`
	Version            int64                  `json:"version"`
	Content            string                 `json:"content,omitempty"`
	Files              []ac.InitialConfigFile `json:"files"`
	ContentSHA256      string                 `json:"contentSha256"`
	Source             string                 `json:"source"`
	CreatedAt          time.Time              `json:"createdAt,omitempty"`
}
type CurrentNodeConfigs struct {
	NodeID  string              `json:"nodeId"`
	Configs []CurrentNodeConfig `json:"configs"`
}
type DeliveryProbeReceipt struct {
	ProbeType         string     `json:"probeType"`
	TargetGeneration  int64      `json:"targetGeneration"`
	DesiredGeneration int64      `json:"desiredGeneration"`
	AppliedGeneration int64      `json:"appliedGeneration"`
	Status            string     `json:"status"`
	AgentStatus       string     `json:"agentStatus,omitempty"`
	Reason            string     `json:"reason,omitempty"`
	Error             ac.JSON    `json:"error,omitempty"`
	ReportedAt        *time.Time `json:"reportedAt,omitempty"`
}
type DeliveryTargetStatus struct {
	NodeID    string                 `json:"nodeId"`
	NodeName  string                 `json:"nodeName"`
	ClusterID string                 `json:"clusterId"`
	Status    string                 `json:"status"`
	Probes    []DeliveryProbeReceipt `json:"probes"`
}
type DeliveryStatus struct {
	ID              uuid.UUID              `json:"id"`
	Type            string                 `json:"type"`
	Scope           string                 `json:"scope"`
	TemplateID      uuid.UUID              `json:"templateId"`
	TemplateName    string                 `json:"templateName,omitempty"`
	Configuration   ac.JSON                `json:"configuration,omitempty"`
	ReceiptTracked  bool                   `json:"receiptTracked"`
	RequestedBy     int64                  `json:"requestedBy"`
	RequestedByName string                 `json:"requestedByName,omitempty"`
	TargetNodeCount int                    `json:"targetNodeCount"`
	RecordedCount   int                    `json:"recordedCount"`
	AppliedCount    int                    `json:"appliedCount"`
	ReceivedCount   int                    `json:"receivedCount"`
	FailedCount     int                    `json:"failedCount"`
	PendingCount    int                    `json:"pendingCount"`
	Status          string                 `json:"status"`
	CreatedAt       time.Time              `json:"createdAt"`
	Targets         []DeliveryTargetStatus `json:"targets"`
}
type NodeInjectionControlInput struct {
	Content            string `json:"content" binding:"required"`
	ExpectedGeneration int64  `json:"expectedGeneration"`
}
type NodeInjectionControl struct {
	NodeID        string `json:"nodeId"`
	Available     bool   `json:"available"`
	Generation    int64  `json:"generation"`
	FilePath      string `json:"filePath,omitempty"`
	Content       string `json:"content,omitempty"`
	ContentSHA256 string `json:"contentSha256,omitempty"`
}
type BatchNodeInjectionControlInput struct {
	ClusterIDs          []string         `json:"clusterIds" binding:"required"`
	Content             string           `json:"content" binding:"required"`
	ExpectedGenerations map[string]int64 `json:"expectedGenerations" binding:"required"`
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

const (
	templateClusterTypeKubernetes = "kubernetes"
	templateClusterTypeVM         = "vm"
)

func normalizeTemplateClusterType(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "k8s", "kubernetes", "helm":
		return templateClusterTypeKubernetes, nil
	case "vm", "binary":
		return templateClusterTypeVM, nil
	default:
		return "", fmt.Errorf("unsupported template cluster type %q; expected kubernetes or vm", value)
	}
}

func ensureTemplateMatchesNode(template *ac.Template, node *ac.Node) error {
	templateType, err := normalizeTemplateClusterType(template.ClusterType)
	if err != nil {
		return err
	}
	nodeType, err := normalizeTemplateClusterType(node.Environment)
	if err != nil {
		return fmt.Errorf("node %s has unsupported environment %q", node.ID, node.Environment)
	}
	if templateType != nodeType {
		return fmt.Errorf("template cluster type %s does not match node %s environment %s", templateType, node.ID, nodeType)
	}
	return nil
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

func ordinaryConfigFiles(probe string, files []ac.InitialConfigFile) []ac.InitialConfigFile {
	probe = ac.NormalizeProbe(probe)
	if probe != "ilogtail" && probe != "loongcollector" {
		return append([]ac.InitialConfigFile(nil), files...)
	}
	ordinary := make([]ac.InitialConfigFile, 0, len(files))
	for _, file := range files {
		cleaned := path.Clean(strings.TrimSpace(file.Path))
		if strings.HasPrefix(cleaned, "config/local/") {
			continue
		}
		ordinary = append(ordinary, file)
	}
	return ordinary
}

// preserveManagedConfigFiles keeps log collection files in every ordinary
// release. They are hidden from ordinary templates, but daemon-go activates a
// complete file bundle and would otherwise remove the active log pipelines.
func (s *Service) preserveManagedConfigFiles(ctx core.Context, node, probe string, files []ac.InitialConfigFile) ([]ac.InitialConfigFile, error) {
	probe = ac.NormalizeProbe(probe)
	if probe != "ilogtail" && probe != "loongcollector" {
		return files, nil
	}
	managed := []ac.InitialConfigFile{}
	releases, err := s.repo.ListAgentReleases(ctx, node, probe, 50)
	if err != nil {
		return nil, err
	}
	for index := range releases {
		releaseFiles, releaseErr := s.releaseFiles(ctx, node, probe, &releases[index])
		if releaseErr != nil {
			log.Printf("agentconfig: cannot inspect managed log files for %s/%s generation %d; fall back: %v", node, probe, releases[index].Generation, releaseErr)
			continue
		}
		for _, file := range releaseFiles {
			cleaned := path.Clean(strings.TrimSpace(file.Path))
			if strings.HasPrefix(cleaned, "config/local/") {
				file.Path = cleaned
				file.ContentSHA256 = ac.HashContent(file.Content)
				managed = append(managed, file)
			}
		}
		if len(managed) > 0 {
			break
		}
	}
	if len(managed) == 0 {
		baseline, baselineErr := s.repo.GetAgentInitialConfig(ctx, node, probe)
		if baselineErr != nil {
			if errors.Is(baselineErr, gorm.ErrRecordNotFound) {
				return files, nil
			}
			return nil, baselineErr
		}
		if baseline != nil {
			var baselineFiles []ac.InitialConfigFile
			if err := json.Unmarshal(baseline.Files, &baselineFiles); err != nil {
				return nil, fmt.Errorf("decode baseline log files for %s/%s: %w", node, probe, err)
			}
			for _, file := range baselineFiles {
				cleaned := path.Clean(strings.TrimSpace(file.Path))
				if strings.HasPrefix(cleaned, "config/local/") {
					file.Path = cleaned
					file.ContentSHA256 = ac.HashContent(file.Content)
					managed = append(managed, file)
				}
			}
		}
	}
	if len(managed) == 0 {
		return files, nil
	}
	byPath := make(map[string]struct{}, len(files)+len(managed))
	merged := append([]ac.InitialConfigFile(nil), files...)
	for _, file := range merged {
		byPath[file.Path] = struct{}{}
	}
	for _, file := range managed {
		if _, exists := byPath[file.Path]; exists {
			continue
		}
		byPath[file.Path] = struct{}{}
		merged = append(merged, file)
	}
	sort.Slice(merged, func(i, j int) bool { return merged[i].Path < merged[j].Path })
	return merged, nil
}

func normalizeTemplatePayload(in TemplateInput) (ac.JSON, string, string, error) {
	configs := append([]ac.TemplateProbeConfig(nil), in.Configs...)
	if len(configs) == 0 {
		return nil, "", "", fmt.Errorf("template must contain at least one probe configuration")
	}
	seen := make(map[string]struct{}, len(configs))
	normalizedConfigs := make([]ac.TemplateProbeConfig, 0, len(configs))
	var variableSource strings.Builder
	for _, config := range configs {
		probe := ac.NormalizeProbe(config.ProbeType)
		if probe == "" {
			return nil, "", "", fmt.Errorf("template probe type is required")
		}
		ordinaryFiles := ordinaryConfigFiles(probe, config.Files)
		if len(config.Files) > 0 && len(ordinaryFiles) == 0 {
			continue
		}
		if _, duplicate := seen[probe]; duplicate {
			return nil, "", "", fmt.Errorf("template probe %s is duplicated", probe)
		}
		seen[probe] = struct{}{}
		files, err := normalizeConfigFiles(ordinaryFiles, "template "+probe)
		if err != nil {
			return nil, "", "", err
		}
		normalizedConfigs = append(normalizedConfigs, ac.TemplateProbeConfig{ProbeType: probe, Files: files})
		for _, file := range files {
			variableSource.WriteString(file.Content)
			variableSource.WriteByte('\n')
		}
	}
	if len(normalizedConfigs) == 0 {
		return nil, "", "", fmt.Errorf("template must contain at least one ordinary probe configuration")
	}
	sort.Slice(normalizedConfigs, func(i, j int) bool { return normalizedConfigs[i].ProbeType < normalizedConfigs[j].ProbeType })
	data, err := json.Marshal(normalizedConfigs)
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
	clusterType, err := normalizeTemplateClusterType(in.ClusterType)
	if err != nil {
		return nil, err
	}
	t := ac.NewTemplate()
	t.Configs = configs
	t.Name = strings.TrimSpace(in.Name)
	t.Description = in.Description
	t.ClusterType = clusterType
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
	oldConfigs, e := templateProbeConfigs(old)
	if e != nil {
		return nil, e
	}
	configs, checksum, variableSource, e := normalizeTemplatePayload(in)
	if e != nil {
		return nil, e
	}
	oldClusterType, e := normalizeTemplateClusterType(old.ClusterType)
	if e != nil {
		return nil, e
	}
	nextClusterType, e := normalizeTemplateClusterType(in.ClusterType)
	if e != nil {
		return nil, e
	}
	if oldClusterType != nextClusterType {
		return nil, fmt.Errorf("template cluster type cannot be changed; create a new template instead")
	}
	var nextConfigs []ac.TemplateProbeConfig
	if e = json.Unmarshal(configs, &nextConfigs); e != nil {
		return nil, e
	}
	if !sameTemplateConfigFilePaths(oldConfigs, nextConfigs) {
		return nil, fmt.Errorf("template registration cannot be changed; only file template content may be updated")
	}
	old.Configs = configs
	old.Name = strings.TrimSpace(in.Name)
	old.Description = in.Description
	old.ClusterType = oldClusterType
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
		clusterType, err := normalizeTemplateClusterType(templates[index].ClusterType)
		if err != nil {
			return nil, fmt.Errorf("normalize template %s cluster type: %w", templates[index].ID, err)
		}
		templates[index].ClusterType = clusterType
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
	clusterType, err := normalizeTemplateClusterType(template.ClusterType)
	if err != nil {
		return nil, err
	}
	template.ClusterType = clusterType
	return template, nil
}
func (s *Service) DeleteTemplate(ctx core.Context, id uuid.UUID) error {
	return s.repo.DeleteAgentTemplate(ctx, id)
}
func (s *Service) Register(ctx core.Context, n *ac.Node) error {
	if n == nil || strings.TrimSpace(n.ID) == "" {
		return errors.New("node ID is required")
	}
	if existing, err := s.repo.GetAgentNode(ctx, n.ID); err == nil {
		if strings.TrimSpace(existing.InitialClusterID) != "" {
			n.InitialClusterID = existing.InitialClusterID
		}
		if strings.TrimSpace(n.InstanceID) == "" {
			n.InstanceID = existing.InstanceID
		}
		if len(n.VariableOverrides) == 0 || string(n.VariableOverrides) == "{}" {
			n.VariableOverrides = existing.VariableOverrides
		}
		var overrides map[string]string
		if json.Unmarshal(existing.VariableOverrides, &overrides) == nil {
			if assignedClusterID := strings.TrimSpace(overrides["CLUSTER_ID"]); assignedClusterID != "" {
				n.ClusterID = assignedClusterID
			}
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if strings.TrimSpace(n.InitialClusterID) == "" {
		n.InitialClusterID = n.ClusterID
	}
	now := time.Now()
	n.LastSeenAt = &now
	return s.repo.UpsertAgentNode(ctx, n)
}
func (s *Service) ListNodes(ctx core.Context) ([]ac.Node, error) {
	nodes, err := s.repo.ListAgentNodes(ctx)
	if err != nil {
		return nil, err
	}
	markMixedClusterNodes(nodes)
	return nodes, nil
}

func (s *Service) NodeVariables(ctx core.Context, nodeID string) (*NodeVariables, error) {
	node, err := s.repo.GetAgentNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	reported := map[string]string{}
	if len(node.Labels) > 0 {
		if err := json.Unmarshal(node.Labels, &reported); err != nil {
			log.Printf("agentconfig: invalid reported variables for node %s: %v", nodeID, err)
			reported = map[string]string{}
		}
	}
	overrides := map[string]string{}
	if len(node.VariableOverrides) > 0 {
		if err := json.Unmarshal(node.VariableOverrides, &overrides); err != nil {
			log.Printf("agentconfig: invalid variable overrides for node %s: %v", nodeID, err)
			overrides = map[string]string{}
		}
	}
	effective := nodeTemplateVariables(node)
	for key, value := range overrides {
		effective[key] = value
	}
	return &NodeVariables{NodeID: node.ID, InitialClusterID: node.InitialClusterID, Reported: reported, Overrides: overrides, Effective: effective}, nil
}

func (s *Service) SaveNodeVariables(ctx core.Context, nodeID string, overrides map[string]string) (*NodeVariables, error) {
	node, err := s.repo.GetAgentNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	clean := map[string]string{}
	var existingOverrides map[string]string
	if len(node.VariableOverrides) > 0 {
		_ = json.Unmarshal(node.VariableOverrides, &existingOverrides)
	}
	for _, key := range []string{"CLUSTER_IP", "CLUSTER_ID", "INSTANCE_ID"} {
		if value := strings.TrimSpace(existingOverrides[key]); value != "" {
			clean[key] = value
		}
	}
	for key, value := range overrides {
		key = strings.TrimSpace(key)
		if key == "" || !variablePattern.MatchString("{% "+key+" %}") {
			return nil, fmt.Errorf("invalid node variable %q", key)
		}
		if key == "CLUSTER_IP" || key == "CLUSTER_ID" || key == "INSTANCE_ID" {
			return nil, fmt.Errorf("%s must be changed from cluster management", key)
		}
		if strings.TrimSpace(value) != "" {
			clean[key] = value
		}
	}
	data, err := json.Marshal(clean)
	if err != nil {
		return nil, err
	}
	node.VariableOverrides = ac.JSON(data)
	if err := s.repo.UpsertAgentNode(ctx, node); err != nil {
		return nil, err
	}
	return s.NodeVariables(ctx, nodeID)
}

func (s *Service) SaveNodeClusterIdentity(ctx core.Context, nodeID, clusterID string) (*NodeClusterIdentity, error) {
	node, err := s.repo.GetAgentNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return nil, errors.New("cluster ID is required")
	}
	if err := s.validateNodeTargetCluster(ctx, node, clusterID); err != nil {
		return nil, err
	}
	overrides := map[string]string{}
	if len(node.VariableOverrides) > 0 {
		if err := json.Unmarshal(node.VariableOverrides, &overrides); err != nil {
			return nil, fmt.Errorf("invalid variable overrides for node %s: %w", nodeID, err)
		}
	}
	if overrides == nil {
		overrides = map[string]string{}
	}
	overrides["CLUSTER_ID"] = clusterID
	data, err := json.Marshal(overrides)
	if err != nil {
		return nil, err
	}
	node.ClusterID = clusterID
	node.VariableOverrides = ac.JSON(data)
	if err := s.repo.UpsertAgentNode(ctx, node); err != nil {
		return nil, err
	}
	return s.GetNodeClusterIdentity(ctx, nodeID)
}

func (s *Service) GetNodeClusterIdentity(ctx core.Context, nodeID string) (*NodeClusterIdentity, error) {
	node, err := s.repo.GetAgentNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	variables, err := s.NodeVariables(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	return &NodeClusterIdentity{NodeID: node.ID, InstanceID: node.InstanceID, ClusterID: node.ClusterID, InitialClusterID: node.InitialClusterID, ClusterIP: variables.Effective["CLUSTER_IP"]}, nil
}

func (s *Service) RestoreNodeClusterID(ctx core.Context, nodeID string) (*NodeClusterIdentity, error) {
	node, err := s.repo.GetAgentNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(node.InitialClusterID) == "" {
		return nil, errors.New("node has no initial cluster ID")
	}
	if err := s.validateNodeTargetCluster(ctx, node, node.InitialClusterID); err != nil {
		return nil, err
	}
	overrides := map[string]string{}
	if len(node.VariableOverrides) > 0 {
		if err := json.Unmarshal(node.VariableOverrides, &overrides); err != nil {
			return nil, fmt.Errorf("invalid variable overrides for node %s: %w", nodeID, err)
		}
	}
	delete(overrides, "CLUSTER_ID")
	data, err := json.Marshal(overrides)
	if err != nil {
		return nil, err
	}
	node.ClusterID = node.InitialClusterID
	node.VariableOverrides = ac.JSON(data)
	if err := s.repo.UpsertAgentNode(ctx, node); err != nil {
		return nil, err
	}
	return s.GetNodeClusterIdentity(ctx, nodeID)
}

func (s *Service) validateNodeTargetCluster(ctx core.Context, node *ac.Node, clusterID string) error {
	targetCluster, err := s.repo.GetCluster(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("target cluster %q does not exist: %w", clusterID, err)
	}
	nodeType, err := normalizeTemplateClusterType(node.Environment)
	if err != nil {
		return fmt.Errorf("node %s has unsupported environment %q", node.ID, node.Environment)
	}
	targetType, err := normalizeTemplateClusterType(targetCluster.ClusterType)
	if err != nil {
		return fmt.Errorf("target cluster %s has unsupported type %q", targetCluster.ID, targetCluster.ClusterType)
	}
	if nodeType != targetType {
		return fmt.Errorf("cannot assign %s node %s to %s cluster %s", nodeType, node.ID, targetType, targetCluster.ID)
	}
	return nil
}

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

func defaultSystemRecoveryPercent(maximum float64) float64 {
	if maximum > 10 {
		return maximum - 10
	}
	return maximum * 0.8
}

func normalizeSystemPressureSettings(settings *ac.SystemPressureSettings) {
	if settings == nil {
		return
	}
	if settings.MaxCPUPercent > 0 && settings.RecoveryCPUPercent == 0 {
		settings.RecoveryCPUPercent = defaultSystemRecoveryPercent(settings.MaxCPUPercent)
		if settings.Enabled {
			log.Printf("agentconfig: system CPU recovery threshold is missing; use %.1f%%", settings.RecoveryCPUPercent)
		}
	}
	if settings.MaxMemoryPercent > 0 && settings.RecoveryMemoryPercent == 0 {
		settings.RecoveryMemoryPercent = defaultSystemRecoveryPercent(settings.MaxMemoryPercent)
		if settings.Enabled {
			log.Printf("agentconfig: system memory recovery threshold is missing; use %.1f%%", settings.RecoveryMemoryPercent)
		}
	}
	if settings.TriggerDurationSeconds == 0 {
		settings.TriggerDurationSeconds = defaultSystemTriggerDurationSeconds
		if settings.Enabled {
			log.Printf("agentconfig: system circuit trigger duration is missing; use %d seconds", settings.TriggerDurationSeconds)
		}
	}
	if settings.RecoveryDurationSeconds == 0 {
		settings.RecoveryDurationSeconds = defaultSystemRecoveryDurationSeconds
		if settings.Enabled {
			log.Printf("agentconfig: system circuit recovery duration is missing; use %d seconds", settings.RecoveryDurationSeconds)
		}
	}
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
		if policy.SystemPressure != nil {
			return fmt.Errorf("component policy %s must not contain system pressure settings", componentType)
		}
		if err := validateCircuitBreaker(&policy); err != nil {
			return fmt.Errorf("component policy %s: %w", componentType, err)
		}
	}
	if settings.SystemPressure != nil {
		normalizeSystemPressureSettings(settings.SystemPressure)
	}
	if settings.SystemPressure != nil && settings.SystemPressure.Enabled {
		pressure := settings.SystemPressure
		if pressure.MaxCPUPercent != 0 && (pressure.MaxCPUPercent < 1 || pressure.MaxCPUPercent > 100) {
			return fmt.Errorf("system max CPU percent must be 0 or between 1 and 100")
		}
		if pressure.MaxMemoryPercent != 0 && (pressure.MaxMemoryPercent < 1 || pressure.MaxMemoryPercent > 100) {
			return fmt.Errorf("system max memory percent must be 0 or between 1 and 100")
		}
		if pressure.MaxCPUPercent == 0 && pressure.MaxMemoryPercent == 0 {
			return fmt.Errorf("enabled system pressure breaker requires a CPU or memory threshold")
		}
		if (pressure.MaxCPUPercent == 0 && pressure.RecoveryCPUPercent != 0) ||
			(pressure.MaxCPUPercent != 0 && (pressure.RecoveryCPUPercent <= 0 || pressure.RecoveryCPUPercent >= pressure.MaxCPUPercent)) {
			return fmt.Errorf("system recovery CPU percent must be 0 when CPU protection is disabled, otherwise positive and below the CPU circuit threshold")
		}
		if (pressure.MaxMemoryPercent == 0 && pressure.RecoveryMemoryPercent != 0) ||
			(pressure.MaxMemoryPercent != 0 && (pressure.RecoveryMemoryPercent <= 0 || pressure.RecoveryMemoryPercent >= pressure.MaxMemoryPercent)) {
			return fmt.Errorf("system recovery memory percent must be 0 when memory protection is disabled, otherwise positive and below the memory circuit threshold")
		}
		if pressure.TriggerDurationSeconds < 5 || pressure.TriggerDurationSeconds > 600 {
			return fmt.Errorf("system trigger duration seconds must be between 5 and 600")
		}
		if pressure.RecoveryDurationSeconds < 5 || pressure.RecoveryDurationSeconds > 600 {
			return fmt.Errorf("system recovery duration seconds must be between 5 and 600")
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
	delivery, err := buildRecordedDelivery(deliveryTypeGlobalCircuit, "全局默认", settings, nil, user)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveGlobalCircuitConfig(ctx, config, delivery); err != nil {
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

func (s *Service) saveNodeCircuitConfigs(ctx core.Context, nodes []ac.Node, settings ac.CircuitBreakerSettings, user int64, deliveryType, scope string) (*BatchResult, error) {
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
	delivery, err := buildRecordedDelivery(deliveryType, scope, settings, nodes, user)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveNodeCircuitConfigs(ctx, configs, delivery); err != nil {
		return nil, err
	}
	return &BatchResult{TargetNodeCount: len(nodes), SucceededCount: len(nodes)}, nil
}

func (s *Service) SaveNodeCircuit(ctx core.Context, node string, settings ac.CircuitBreakerSettings, user int64) (*ac.CircuitBreakerSettings, error) {
	if err := s.ensureNodeAllowsExplicitDelivery(ctx, node); err != nil {
		return nil, err
	}
	target, err := s.repo.GetAgentNode(ctx, node)
	if err != nil {
		return nil, err
	}
	if _, err := s.saveNodeCircuitConfigs(ctx, []ac.Node{*target}, settings, user, deliveryTypeNodeCircuit, node); err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *Service) BatchSaveCircuit(ctx core.Context, input CircuitBatchInput, user int64) (*BatchResult, error) {
	if err := validateCircuitBreaker(&input.CircuitBreaker); err != nil {
		return nil, err
	}
	nodes, err := s.nodesInClusters(ctx, input.ClusterIDs)
	if err != nil {
		return nil, err
	}
	return s.saveNodeCircuitConfigs(ctx, nodes, input.CircuitBreaker, user, deliveryTypeBatchCircuit, strings.Join(input.ClusterIDs, ","))
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

func (s *Service) saveNodeLogSources(ctx core.Context, nodes []ac.Node, probe string, sources []ac.LogSource, user int64, deliveryType, scope string) (*BatchResult, error) {
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
	delivery, err := buildRecordedDelivery(deliveryType, scope, map[string]any{
		"probeType": probe, "logSources": sources,
	}, nodes, user)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveNodeLogConfigs(ctx, configs, delivery); err != nil {
		return nil, err
	}
	return &BatchResult{TargetNodeCount: len(nodes), SucceededCount: len(nodes)}, nil
}

func (s *Service) SaveNodeLogSources(ctx core.Context, node, probe string, sources []ac.LogSource, user int64) ([]ac.LogSource, error) {
	if err := s.ensureNodeAllowsExplicitDelivery(ctx, node); err != nil {
		return nil, err
	}
	target, err := s.repo.GetAgentNode(ctx, node)
	if err != nil {
		return nil, err
	}
	if _, err := s.saveNodeLogSources(ctx, []ac.Node{*target}, probe, sources, user, deliveryTypeNodeLogSources, node); err != nil {
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
	return s.saveNodeLogSources(ctx, nodes, input.ProbeType, input.LogSources, user, deliveryTypeBatchLogSources, strings.Join(input.ClusterIDs, ","))
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
	markMixedClusterNodes(nodes)
	found := map[string]bool{}
	targets := make([]ac.Node, 0)
	for _, node := range nodes {
		if wanted[node.ClusterID] {
			if node.MixedCluster {
				return nil, fmt.Errorf("cluster %s is mixed at cluster_ip %s; resolve node ownership in cluster management before dispatching configuration", node.ClusterID, node.ClusterIP)
			}
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

func markMixedClusterNodes(nodes []ac.Node) {
	typesByClusterIP := map[string]map[string]struct{}{}
	for index := range nodes {
		clusterIP := strings.TrimSpace(nodeTemplateVariables(&nodes[index])["CLUSTER_IP"])
		nodes[index].ClusterIP = clusterIP
		if clusterIP == "" {
			continue
		}
		nodeType, err := normalizeTemplateClusterType(nodes[index].Environment)
		if err != nil {
			log.Printf("agentconfig: cannot classify node %s environment %q for mixed-cluster detection: %v", nodes[index].ID, nodes[index].Environment, err)
			continue
		}
		if typesByClusterIP[clusterIP] == nil {
			typesByClusterIP[clusterIP] = map[string]struct{}{}
		}
		typesByClusterIP[clusterIP][nodeType] = struct{}{}
	}
	for index := range nodes {
		nodes[index].MixedCluster = nodes[index].ClusterIP != "" && len(typesByClusterIP[nodes[index].ClusterIP]) > 1
	}
}

func (s *Service) ensureNodeAllowsExplicitDelivery(ctx core.Context, nodeID string) error {
	nodes, err := s.repo.ListAgentNodes(ctx)
	if err != nil {
		return err
	}
	markMixedClusterNodes(nodes)
	for _, node := range nodes {
		if node.ID != nodeID {
			continue
		}
		if node.MixedCluster {
			return fmt.Errorf("node %s belongs to a mixed cluster at cluster_ip %s; resolve node ownership in cluster management before dispatching configuration", node.ID, node.ClusterIP)
		}
		return nil
	}
	return fmt.Errorf("node %s does not exist", nodeID)
}

func (s *Service) CurrentNodeConfig(ctx core.Context, node, probe string) (*CurrentNodeConfig, error) {
	if _, err := s.repo.GetAgentNode(ctx, node); err != nil {
		return nil, err
	}
	probe = ac.NormalizeProbe(probe)
	state, stateErr := s.repo.GetAgentState(ctx, node, probe)
	if stateErr == nil {
		if current, ok := currentNodeConfigFromSnapshot(node, probe, state.EffectiveConfig); ok {
			return current, nil
		}
		if state.AppliedGeneration > 0 {
			release, err := s.repo.GetAgentReleaseGeneration(ctx, node, probe, state.AppliedGeneration)
			if err != nil {
				return nil, err
			}
			files, err := s.releaseFiles(ctx, node, probe, release)
			if err != nil {
				return nil, err
			}
			return currentNodeConfig(node, probe, release.Generation, files, release.ContentSHA256, "RELEASE"), nil
		}
	} else if !errors.Is(stateErr, gorm.ErrRecordNotFound) {
		return nil, stateErr
	}
	current, err := s.initialNodeConfig(ctx, node, probe)
	if err == nil {
		return current, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Some agents can receive a release before they report an initial config.
	// Keep the node editor usable in that window, while preserving real storage
	// errors and making the fallback visible in the service log.
	release, releaseErr := s.repo.GetAgentRelease(ctx, node, probe)
	if releaseErr == nil {
		files, filesErr := s.releaseFiles(ctx, node, probe, release)
		if filesErr != nil {
			return nil, filesErr
		}
		log.Printf("agentconfig: initial config missing for %s/%s; using latest release generation %d", node, probe, release.Generation)
		return currentNodeConfig(node, probe, release.Generation, files, release.ContentSHA256, "RELEASE"), nil
	}
	if !errors.Is(releaseErr, gorm.ErrRecordNotFound) {
		return nil, releaseErr
	}
	return nil, err
}

func (s *Service) initialNodeConfig(ctx core.Context, node, probe string) (*CurrentNodeConfig, error) {
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

func currentNodeConfigFromSnapshot(node, probe string, effectiveConfig ac.JSON) (*CurrentNodeConfig, bool) {
	var snapshot struct {
		ConfigSource         string                 `json:"configSource"`
		ConfigGeneration     int64                  `json:"configGeneration"`
		EffectiveConfigFiles []ac.InitialConfigFile `json:"effectiveConfigFiles"`
	}
	if err := json.Unmarshal(effectiveConfig, &snapshot); err != nil || len(snapshot.EffectiveConfigFiles) == 0 {
		return nil, false
	}
	for index := range snapshot.EffectiveConfigFiles {
		if snapshot.EffectiveConfigFiles[index].ContentSHA256 == "" {
			snapshot.EffectiveConfigFiles[index].ContentSHA256 = ac.HashContent(snapshot.EffectiveConfigFiles[index].Content)
		}
	}
	source := "INITIAL"
	if strings.EqualFold(snapshot.ConfigSource, "REMOTE") {
		source = "RELEASE"
	}
	return currentNodeConfig(node, probe, snapshot.ConfigGeneration, snapshot.EffectiveConfigFiles, ac.HashInitialConfigFiles(snapshot.EffectiveConfigFiles), source), true
}

func currentNodeConfig(node, probe string, generation int64, files []ac.InitialConfigFile, sha, source string) *CurrentNodeConfig {
	files = ordinaryConfigFiles(probe, files)
	sha = ac.HashInitialConfigFiles(files)
	config := &CurrentNodeConfig{NodeID: node, ProbeType: probe, Generation: generation, ExpectedGeneration: generation, Version: generation + 1, Files: files, ContentSHA256: sha, Source: source}
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
	if probe == "preload" {
		return nil, errors.New("injection control does not support configuration history or rollback")
	}
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
		if len(config.Files) == 0 {
			continue
		}
		config.CreatedAt = releases[index].CreatedAt
		result = append(result, *config)
	}
	if len(result) < limit {
		baseline, baselineErr := s.repo.GetAgentInitialConfig(ctx, node, probe)
		if baselineErr == nil {
			var files []ac.InitialConfigFile
			if err := json.Unmarshal(baseline.Files, &files); err != nil {
				return nil, err
			}
			if len(files) == 0 {
				return nil, fmt.Errorf("node has no initial configuration for %s", probe)
			}
			config := currentNodeConfig(node, probe, 0, files, baseline.ContentSHA256, "INITIAL")
			if len(config.Files) == 0 {
				return result, nil
			}
			config.CreatedAt = baseline.CapturedAt
			result = append(result, *config)
		} else if !errors.Is(baselineErr, gorm.ErrRecordNotFound) {
			return nil, baselineErr
		}
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
	latestGeneration := make(map[string]int64, len(releases))
	for _, release := range releases {
		probe := ac.NormalizeProbe(release.ProbeType)
		probeTypes[probe] = struct{}{}
		latestGeneration[probe] = release.Generation
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
		if len(current.Files) == 0 {
			continue
		}
		if latestGeneration[probe] > current.ExpectedGeneration {
			current.ExpectedGeneration = latestGeneration[probe]
		}
		result.Configs = append(result.Configs, *current)
	}
	return result, nil
}

func releaseKey(node, probe string) string { return node + "\x00" + ac.NormalizeProbe(probe) }

func (s *Service) UpdateNodeConfigs(ctx core.Context, node string, input NodeConfigsInput) (*CurrentNodeConfigs, error) {
	return s.UpdateNodeConfigsForUser(ctx, node, input, 0)
}

func (s *Service) UpdateNodeConfigsForUser(ctx core.Context, node string, input NodeConfigsInput, user int64) (*CurrentNodeConfigs, error) {
	if err := s.ensureNodeAllowsExplicitDelivery(ctx, node); err != nil {
		return nil, err
	}
	if err := rejectSpecialControlConfigs(input); err != nil {
		return nil, err
	}
	return s.updateNodeConfigGroup(ctx, node, input, true, deliveryTypeNodeConfig, user)
}

func (s *Service) UpdateNodeConfigGroup(ctx core.Context, node string, input NodeConfigsInput) (*CurrentNodeConfigs, error) {
	return s.UpdateNodeConfigGroupForUser(ctx, node, input, 0)
}

func (s *Service) UpdateNodeConfigGroupForUser(ctx core.Context, node string, input NodeConfigsInput, user int64) (*CurrentNodeConfigs, error) {
	if err := s.ensureNodeAllowsExplicitDelivery(ctx, node); err != nil {
		return nil, err
	}
	if err := rejectSpecialControlConfigs(input); err != nil {
		return nil, err
	}
	return s.updateNodeConfigGroup(ctx, node, input, false, deliveryTypeNodeConfig, user)
}

func rejectSpecialControlConfigs(input NodeConfigsInput) error {
	for _, config := range input.Configs {
		if ac.NormalizeProbe(config.ProbeType) == "preload" {
			return errors.New("preload must be updated through the injection control API")
		}
	}
	return nil
}

func (s *Service) GetNodeInjectionControl(ctx core.Context, node string) (*NodeInjectionControl, error) {
	if _, err := s.repo.GetAgentNode(ctx, node); err != nil {
		return nil, err
	}
	current, err := s.latestNodeConfig(ctx, node, "preload")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &NodeInjectionControl{NodeID: node}, nil
	}
	if err != nil {
		return nil, err
	}
	return nodeInjectionControl(current)
}

func (s *Service) SaveNodeInjectionControl(ctx core.Context, node string, input NodeInjectionControlInput) (*NodeInjectionControl, error) {
	return s.SaveNodeInjectionControlForUser(ctx, node, input, 0)
}

func (s *Service) SaveNodeInjectionControlForUser(ctx core.Context, node string, input NodeInjectionControlInput, user int64) (*NodeInjectionControl, error) {
	if err := s.ensureNodeAllowsExplicitDelivery(ctx, node); err != nil {
		return nil, err
	}
	current, err := s.GetNodeInjectionControl(ctx, node)
	if err != nil {
		return nil, err
	}
	if !current.Available {
		return nil, errors.New("node has no registered preload injection configuration")
	}
	if input.ExpectedGeneration != current.Generation {
		return nil, fmt.Errorf("generation conflict for preload: expected %d, actual %d", input.ExpectedGeneration, current.Generation)
	}
	if input.Content == current.Content {
		return current, nil
	}
	_, err = s.updateNodeConfigGroup(ctx, node, NodeConfigsInput{Configs: []NodeProbeConfigInput{{
		ProbeType:          "preload",
		Files:              []ac.InitialConfigFile{{Path: current.FilePath, Content: input.Content}},
		ExpectedGeneration: input.ExpectedGeneration,
	}}}, false, deliveryTypeNodeInjection, user)
	if err != nil {
		return nil, err
	}
	return s.GetNodeInjectionControl(ctx, node)
}

func (s *Service) latestNodeConfig(ctx core.Context, node, probe string) (*CurrentNodeConfig, error) {
	probe = ac.NormalizeProbe(probe)
	release, err := s.repo.GetAgentRelease(ctx, node, probe)
	if err == nil {
		files, fileErr := s.releaseFiles(ctx, node, probe, release)
		if fileErr != nil {
			return nil, fileErr
		}
		return currentNodeConfig(node, probe, release.Generation, files, release.ContentSHA256, "RELEASE"), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return s.initialNodeConfig(ctx, node, probe)
}

func (s *Service) BatchSaveNodeInjectionControl(ctx core.Context, input BatchNodeInjectionControlInput) (*BatchResult, error) {
	return s.BatchSaveNodeInjectionControlForUser(ctx, input, 0)
}

func (s *Service) BatchSaveNodeInjectionControlForUser(ctx core.Context, input BatchNodeInjectionControlInput, user int64) (*BatchResult, error) {
	targets, err := s.nodesInClusters(ctx, input.ClusterIDs)
	if err != nil {
		return nil, err
	}
	bindings := make([]*ac.Binding, 0, len(targets))
	releases := make([]*ac.Release, 0, len(targets))
	expected := make(map[string]int64, len(targets))
	for _, target := range targets {
		control, err := s.GetNodeInjectionControl(ctx, target.ID)
		if err != nil {
			return nil, fmt.Errorf("get injection control for %s: %w", target.ID, err)
		}
		if !control.Available {
			return nil, fmt.Errorf("node %s has no registered preload injection configuration", target.ID)
		}
		expectedGeneration, exists := input.ExpectedGenerations[target.ID]
		if !exists {
			return nil, fmt.Errorf("expected generation for node %s is required", target.ID)
		}
		if expectedGeneration != control.Generation {
			return nil, fmt.Errorf("generation conflict for %s/preload: expected %d, actual %d", target.ID, expectedGeneration, control.Generation)
		}
		if input.Content == control.Content {
			continue
		}
		files, err := normalizeConfigFiles([]ac.InitialConfigFile{{Path: control.FilePath, Content: input.Content}}, "node injection control")
		if err != nil {
			return nil, fmt.Errorf("invalid injection control for %s: %w", target.ID, err)
		}
		filesJSON, err := json.Marshal(files)
		if err != nil {
			return nil, err
		}
		logSources := ac.JSON("null")
		if control.Generation > 0 {
			last, err := s.repo.GetAgentRelease(ctx, target.ID, "preload")
			if err != nil {
				return nil, err
			}
			logSources = append(ac.JSON(nil), last.LogSources...)
		}
		generation := control.Generation + 1
		releases = append(releases, &ac.Release{
			ID: uuid.New(), NodeID: target.ID, ProbeType: "preload", Generation: generation,
			TemplateID: uuid.Nil, TemplateVersion: 1, Variables: ac.JSON("{}"),
			RenderedContent: input.Content, ContentSHA256: ac.HashInitialConfigFiles(files),
			Files: ac.JSON(filesJSON), LogSources: logSources,
		})
		bindings = append(bindings, &ac.Binding{
			NodeID: target.ID, ProbeType: "preload", TemplateID: uuid.Nil,
			DesiredGeneration: generation, Enabled: true, Variables: ac.JSON("{}"),
		})
		expected[releaseKey(target.ID, "preload")] = control.Generation
	}
	if len(releases) > 0 {
		delivery, err := s.buildDelivery(ctx, deliveryTypeBatchInjection, strings.Join(input.ClusterIDs, ","), nil, releases, user)
		if err != nil {
			return nil, err
		}
		if err := s.repo.SaveAgentBindings(ctx, bindings, releases, expected, delivery); err != nil {
			return nil, err
		}
	}
	return &BatchResult{
		TargetNodeCount: len(targets),
		SucceededCount:  len(targets),
	}, nil
}

func nodeInjectionControl(current *CurrentNodeConfig) (*NodeInjectionControl, error) {
	for _, file := range current.Files {
		if path.Base(file.Path) == "libapoinstrument.conf" {
			return &NodeInjectionControl{
				NodeID:        current.NodeID,
				Available:     true,
				Generation:    current.Generation,
				FilePath:      file.Path,
				Content:       file.Content,
				ContentSHA256: file.ContentSHA256,
			}, nil
		}
	}
	return nil, errors.New("preload has no registered libapoinstrument.conf")
}

func (s *Service) updateNodeConfigGroup(ctx core.Context, node string, input NodeConfigsInput, requireComplete bool, deliveryType string, user int64) (*CurrentNodeConfigs, error) {
	if len(input.Configs) == 0 {
		return nil, fmt.Errorf("at least one probe configuration is required")
	}
	currentSet, err := s.CurrentNodeConfigs(ctx, node)
	if err != nil {
		return nil, err
	}
	ordinaryConfigCount := 0
	for _, current := range currentSet.Configs {
		if ac.NormalizeProbe(current.ProbeType) != "preload" {
			ordinaryConfigCount++
		}
	}
	if requireComplete && len(input.Configs) != ordinaryConfigCount {
		return nil, fmt.Errorf("node update must include all %d ordinary probe configurations", ordinaryConfigCount)
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
		if update.ExpectedGeneration != current.ExpectedGeneration {
			return nil, fmt.Errorf("generation conflict for %s: expected %d, actual %d", probe, update.ExpectedGeneration, current.ExpectedGeneration)
		}
		expected[releaseKey(node, probe)] = current.ExpectedGeneration
		files, err := normalizeNodeUpdateFiles(update, current)
		if err != nil {
			return nil, fmt.Errorf("invalid files for %s: %w", probe, err)
		}
		if !sameConfigFilePaths(files, current.Files) {
			return nil, fmt.Errorf("registered configuration files for %s cannot be added, removed, or renamed", probe)
		}
		if sameConfigFiles(files, current.Files) {
			continue
		}
		files, err = s.preserveManagedConfigFiles(ctx, node, probe, files)
		if err != nil {
			return nil, fmt.Errorf("preserve log collection files for %s: %w", probe, err)
		}
		logSources := ac.JSON("null")
		if current.ExpectedGeneration > 0 {
			last, err := s.repo.GetAgentRelease(ctx, node, probe)
			if err != nil {
				return nil, err
			}
			logSources = append(ac.JSON(nil), last.LogSources...)
		}
		generation := current.ExpectedGeneration + 1
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
		delivery, err := s.buildDelivery(ctx, deliveryType, node, nil, releases, user)
		if err != nil {
			return nil, err
		}
		if err := s.repo.SaveAgentBindings(ctx, bindings, releases, expected, delivery); err != nil {
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
	return normalizeConfigFiles(files, "node configuration")
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
	ordinaryConfigs := configs[:0]
	for _, config := range configs {
		if ac.NormalizeProbe(config.ProbeType) != "preload" {
			ordinaryConfigs = append(ordinaryConfigs, config)
		}
	}
	configs = ordinaryConfigs
	if len(configs) == 0 {
		return nil, errors.New("template contains no ordinary probe configuration")
	}
	bindings := make([]*ac.Binding, 0, len(targets)*len(configs))
	releases := make([]*ac.Release, 0, len(targets)*len(configs))
	expected := make(map[string]int64, len(targets)*len(configs))
	for _, node := range targets {
		if err := ensureTemplateMatchesNode(template, &node); err != nil {
			return nil, err
		}
		for _, config := range configs {
			release, binding, previous, err := s.prepareAssignment(ctx, node.ID, config.ProbeType, template, config.Files, AssignmentInput{TemplateID: template.ID, Variables: input.Variables}, nodeTemplateVariables(&node))
			if err != nil {
				return nil, fmt.Errorf("prepare template assignment for %s/%s: %w", node.ID, config.ProbeType, err)
			}
			releases = append(releases, release)
			bindings = append(bindings, binding)
			expected[releaseKey(node.ID, config.ProbeType)] = previous
		}
	}
	delivery, err := s.buildDelivery(ctx, deliveryTypeBatchTemplate, strings.Join(input.ClusterIDs, ","), template, releases, user)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveAgentBindings(ctx, bindings, releases, expected, delivery); err != nil {
		return nil, err
	}
	return &BatchResult{TargetNodeCount: len(targets), SucceededCount: len(targets)}, nil
}

func (s *Service) Assign(ctx core.Context, node, probe string, in AssignmentInput) (*ac.Release, error) {
	targetNode, e := s.repo.GetAgentNode(ctx, node)
	if e != nil {
		return nil, e
	}
	t, e := s.repo.GetAgentTemplate(ctx, in.TemplateID)
	if e != nil {
		return nil, e
	}
	if e = ensureTemplateMatchesNode(t, targetNode); e != nil {
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
	current, e := s.CurrentNodeConfig(ctx, node, probe)
	if e != nil {
		return nil, e
	}
	if !sameConfigFilePaths(selected.Files, current.Files) {
		return nil, fmt.Errorf("template files for %s do not match registered configuration files on node %s", probe, node)
	}
	release, binding, previous, err := s.prepareAssignment(ctx, node, probe, t, selected.Files, in, nodeTemplateVariables(targetNode))
	if err != nil {
		return nil, err
	}
	if in.ExpectedGeneration > 0 && previous != in.ExpectedGeneration {
		return nil, fmt.Errorf("generation conflict: expected %d, actual %d", in.ExpectedGeneration, previous)
	}
	delivery, err := s.buildDelivery(ctx, deliveryTypeNodeConfig, node, t, []*ac.Release{release}, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.SaveAgentBindings(ctx, []*ac.Binding{binding}, []*ac.Release{release}, map[string]int64{releaseKey(node, probe): previous}, delivery); err != nil {
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
	normalized, _, variableSource, err := normalizeTemplatePayload(TemplateInput{Configs: configs})
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(normalized, &configs); err != nil {
		return nil, err
	}
	t.Variables = mergeVariables(variableSource, t.Variables)
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

func sameTemplateConfigFilePaths(left, right []ac.TemplateProbeConfig) bool {
	if len(left) != len(right) {
		return false
	}
	rightByProbe := make(map[string][]ac.InitialConfigFile, len(right))
	for _, config := range right {
		rightByProbe[config.ProbeType] = config.Files
	}
	for _, config := range left {
		files, ok := rightByProbe[config.ProbeType]
		if !ok || !sameConfigFilePaths(config.Files, files) {
			return false
		}
	}
	return true
}

func nodeTemplateVariables(node *ac.Node) map[string]string {
	if node == nil {
		return nil
	}
	values := map[string]string{}
	if len(node.Labels) > 0 {
		_ = json.Unmarshal(node.Labels, &values)
	}
	if node.ClusterID != "" {
		values["CLUSTER_ID"] = node.ClusterID
	}
	if node.InitialClusterID != "" {
		values["INITIAL_CLUSTER_ID"] = node.InitialClusterID
	}
	if node.InstanceID != "" {
		values["INSTANCE_ID"] = node.InstanceID
	}
	if len(node.VariableOverrides) > 0 {
		var overrides map[string]string
		if json.Unmarshal(node.VariableOverrides, &overrides) == nil {
			for key, value := range overrides {
				values[key] = value
			}
		}
	}
	return values
}

func (s *Service) prepareAssignment(ctx core.Context, node, probe string, t *ac.Template, definedFiles []ac.InitialConfigFile, in AssignmentInput, builtins map[string]string) (*ac.Release, *ac.Binding, int64, error) {
	missing := []string{}
	invalid := []string{}
	used := map[string]bool{}
	values := map[string]string{}
	for _, v := range t.Variables {
		value, ok := builtins[v.Name]
		if !ok || value == "" {
			value, ok = in.Variables[v.Name]
		}
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
	var e error
	renderedContent := ""
	releaseFiles := ac.JSON("null")
	contentSHA256 := ""
	definedFiles = append([]ac.InitialConfigFile(nil), definedFiles...)
	for index := range definedFiles {
		definedFiles[index].Content, e = Render(definedFiles[index].Content, values)
		if e != nil {
			return nil, nil, 0, fmt.Errorf("render template file %s: %w", definedFiles[index].Path, e)
		}
		definedFiles[index].ContentSHA256 = ac.HashContent(definedFiles[index].Content)
	}
	definedFiles, e = s.preserveManagedConfigFiles(ctx, node, probe, definedFiles)
	if e != nil {
		return nil, nil, 0, e
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
	s.reportMu.Lock()
	defer s.reportMu.Unlock()

	state.ProbeType = ac.NormalizeProbe(state.ProbeType)
	previous, err := s.repo.GetAgentState(ctx, state.NodeID, state.ProbeType)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		previous = nil
	}
	state.Runtime, state.EffectiveConfig = s.prepareReportedState(ctx, state, previous)
	return s.repo.ReportAgentState(ctx, state)
}

var effectiveConfigRuntimeFields = []string{
	"configSource",
	"configGeneration",
	"contentId",
	"effectiveConfigFiles",
}

func (s *Service) prepareReportedState(ctx core.Context, state, previous *ac.ApplyState) (ac.JSON, ac.JSON) {
	fields, err := runtimeFields(state.Runtime)
	if err != nil {
		log.Printf("agentconfig: invalid runtime report for %s/%s; store runtime unchanged: %v", state.NodeID, state.ProbeType, err)
		return state.Runtime, nil
	}
	if _, reported := fields["effectiveConfigFiles"]; reported {
		effectiveConfig := make(map[string]json.RawMessage, len(effectiveConfigRuntimeFields))
		for _, name := range effectiveConfigRuntimeFields {
			if value, ok := fields[name]; ok {
				effectiveConfig[name] = value
				delete(fields, name)
			}
		}
		return marshalRuntimeFields(fields, state.Runtime), marshalRuntimeFields(effectiveConfig, nil)
	}

	if previous != nil {
		previousFields, previousErr := runtimeFields(previous.EffectiveConfig)
		if previousErr == nil {
			if generation, ok := runtimeConfigGeneration(previousFields); ok && generation == state.AppliedGeneration {
				return state.Runtime, nil
			}
		} else {
			log.Printf("agentconfig: invalid stored effective config for %s/%s; rebuild it: %v", state.NodeID, state.ProbeType, previousErr)
		}
	}

	var source, contentID string
	var files []ac.InitialConfigFile
	if state.AppliedGeneration == 0 {
		baseline, baselineErr := s.repo.GetAgentInitialConfig(ctx, state.NodeID, state.ProbeType)
		if baselineErr == nil {
			source = "DEFAULT"
			contentID = ""
			if decodeErr := json.Unmarshal(baseline.Files, &files); decodeErr != nil {
				baselineErr = decodeErr
			}
		}
		if baselineErr != nil {
			log.Printf("agentconfig: cannot build Generation 0 effective config for %s/%s; defer until the next report: %v", state.NodeID, state.ProbeType, baselineErr)
			return state.Runtime, nil
		}
	} else {
		release, releaseErr := s.repo.GetAgentReleaseGeneration(ctx, state.NodeID, state.ProbeType, state.AppliedGeneration)
		if releaseErr == nil {
			files, releaseErr = s.releaseFiles(ctx, state.NodeID, state.ProbeType, release)
			if releaseErr == nil {
				source = "REMOTE"
				contentID = release.ContentSHA256
			}
		}
		if releaseErr != nil {
			log.Printf("agentconfig: cannot build applied Generation %d config for %s/%s; defer until the next report: %v", state.AppliedGeneration, state.NodeID, state.ProbeType, releaseErr)
			return state.Runtime, nil
		}
	}
	effectiveConfig := map[string]json.RawMessage{}
	setRuntimeField(effectiveConfig, "configSource", source)
	setRuntimeField(effectiveConfig, "configGeneration", state.AppliedGeneration)
	setRuntimeField(effectiveConfig, "contentId", contentID)
	setRuntimeField(effectiveConfig, "effectiveConfigFiles", files)
	return state.Runtime, marshalRuntimeFields(effectiveConfig, nil)
}

func runtimeFields(runtime ac.JSON) (map[string]json.RawMessage, error) {
	fields := map[string]json.RawMessage{}
	if len(runtime) == 0 || string(runtime) == "null" {
		return fields, nil
	}
	if err := json.Unmarshal(runtime, &fields); err != nil {
		return nil, err
	}
	return fields, nil
}

func runtimeConfigGeneration(fields map[string]json.RawMessage) (int64, bool) {
	value, ok := fields["configGeneration"]
	if !ok {
		return 0, false
	}
	var generation int64
	if err := json.Unmarshal(value, &generation); err != nil {
		return 0, false
	}
	return generation, true
}

func setRuntimeField(fields map[string]json.RawMessage, name string, value interface{}) {
	encoded, _ := json.Marshal(value)
	fields[name] = encoded
}

func marshalRuntimeFields(fields map[string]json.RawMessage, fallback ac.JSON) ac.JSON {
	encoded, err := json.Marshal(fields)
	if err != nil {
		return fallback
	}
	return ac.JSON(encoded)
}
func (s *Service) ListStates(ctx core.Context) ([]ac.ApplyState, error) {
	states, err := s.repo.ListAgentStates(ctx)
	if err != nil {
		return nil, err
	}
	cutoff := s.now().Add(-agentStateFreshnessTimeout)
	for index := range states {
		if states[index].ReportedAt == nil || states[index].ReportedAt.Before(cutoff) {
			states[index].Status = "OFFLINE"
			states[index].Reason = "status report expired"
		}
	}
	return states, nil
}

func (s *Service) ListDeliveries(ctx core.Context, limit int) ([]DeliveryStatus, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	deliveries, err := s.repo.ListAgentDeliveries(ctx, limit)
	if err != nil {
		return nil, err
	}
	states, err := s.ListStates(ctx)
	if err != nil {
		return nil, err
	}
	stateByTarget := make(map[string]ac.ApplyState, len(states))
	for _, state := range states {
		stateByTarget[releaseKey(state.NodeID, state.ProbeType)] = state
	}
	result := make([]DeliveryStatus, 0, len(deliveries))
	for _, delivery := range deliveries {
		receiptTracked := delivery.ReceiptTracked
		if !receiptTracked {
			for _, target := range delivery.Targets {
				generations := map[string]int64{}
				if err := json.Unmarshal(target.Generations, &generations); err == nil && len(generations) > 0 {
					receiptTracked = true
					break
				}
			}
		}
		item := DeliveryStatus{
			ID: delivery.ID, Type: delivery.Type, Scope: delivery.Scope,
			TemplateID: delivery.TemplateID, TemplateName: delivery.TemplateName,
			Configuration: delivery.Configuration, ReceiptTracked: receiptTracked,
			RequestedBy: delivery.RequestedBy, RequestedByName: delivery.RequestedByName,
			TargetNodeCount: delivery.TargetNodeCount,
			CreatedAt:       delivery.CreatedAt, Targets: make([]DeliveryTargetStatus, 0, len(delivery.Targets)),
		}
		for _, target := range delivery.Targets {
			generations := map[string]int64{}
			if err := json.Unmarshal(target.Generations, &generations); err != nil {
				return nil, fmt.Errorf("decode delivery %s target %s generations: %w", delivery.ID, target.NodeID, err)
			}
			probeTypes := make([]string, 0, len(generations))
			for probeType := range generations {
				probeTypes = append(probeTypes, probeType)
			}
			sort.Strings(probeTypes)
			targetStatus := DeliveryTargetStatus{
				NodeID: target.NodeID, NodeName: target.NodeName, ClusterID: target.ClusterID,
				Probes: make([]DeliveryProbeReceipt, 0, len(probeTypes)),
			}
			if !receiptTracked {
				targetStatus.Status = "RECORDED"
				item.RecordedCount++
				item.Targets = append(item.Targets, targetStatus)
				continue
			}
			applied, received, failed := 0, 0, 0
			for _, probeType := range probeTypes {
				targetGeneration := generations[probeType]
				state, exists := stateByTarget[releaseKey(target.NodeID, probeType)]
				receipt := DeliveryProbeReceipt{ProbeType: probeType, TargetGeneration: targetGeneration, Status: "PENDING"}
				if exists {
					receipt.DesiredGeneration = state.DesiredGeneration
					receipt.AppliedGeneration = state.AppliedGeneration
					receipt.AgentStatus = state.Status
					receipt.Reason = state.Reason
					receipt.Error = state.Error
					receipt.ReportedAt = state.ReportedAt
				}
				switch {
				case receipt.AppliedGeneration >= targetGeneration:
					receipt.Status = "APPLIED"
					applied++
					received++
				case receipt.DesiredGeneration >= targetGeneration && (state.Status == "FAILED" || state.Status == "REJECTED" || state.Status == "UNAVAILABLE"):
					receipt.Status = "FAILED"
					failed++
					received++
				case receipt.DesiredGeneration >= targetGeneration:
					receipt.Status = "RECEIVED"
					received++
				}
				targetStatus.Probes = append(targetStatus.Probes, receipt)
			}
			switch {
			case failed > 0:
				targetStatus.Status = "FAILED"
				item.FailedCount++
			case applied == len(probeTypes):
				targetStatus.Status = "APPLIED"
				item.AppliedCount++
			case received == len(probeTypes):
				targetStatus.Status = "RECEIVED"
				item.ReceivedCount++
			case received > 0:
				targetStatus.Status = "PARTIAL"
				item.ReceivedCount++
			default:
				targetStatus.Status = "PENDING"
				item.PendingCount++
			}
			item.Targets = append(item.Targets, targetStatus)
		}
		switch {
		case !receiptTracked:
			item.Status = "RECORDED"
		case item.FailedCount > 0:
			item.Status = "FAILED"
		case item.AppliedCount == item.TargetNodeCount:
			item.Status = "APPLIED"
		case item.AppliedCount+item.ReceivedCount+item.FailedCount == item.TargetNodeCount:
			item.Status = "RECEIVED"
		case item.AppliedCount+item.ReceivedCount+item.FailedCount > 0:
			item.Status = "PARTIAL"
		default:
			item.Status = "PENDING"
		}
		result = append(result, item)
	}
	return result, nil
}
func (s *Service) CreateCommand(ctx core.Context, node, probe string, in CommandInput, user int64) (*ac.Command, error) {
	typ := strings.ToUpper(strings.TrimSpace(in.Type))
	if typ != "RESTART" && typ != "OPEN_CIRCUIT" && typ != "RESET_CIRCUIT" && typ != "LIST_LOG_FILES" && typ != "PREVIEW_LOG_FILE" && typ != "DOWNLOAD_LOG_FILE" && typ != "SCAN_JAVA_LOG_FILES" && typ != "ENABLE_PRELOAD" && typ != "DISABLE_PRELOAD" {
		return nil, fmt.Errorf("unsupported command type")
	}
	if (typ == "ENABLE_PRELOAD" || typ == "DISABLE_PRELOAD") && ac.NormalizeProbe(probe) != "preload" {
		return nil, fmt.Errorf("preload control command requires the preload component")
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
