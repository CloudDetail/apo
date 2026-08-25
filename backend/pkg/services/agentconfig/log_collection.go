package agentconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

const (
	logCollectionScopeNode    = "node"
	logCollectionScopeCluster = "cluster"
)

type HostFileSource struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Enabled           bool   `json:"enabled"`
	Path              string `json:"path"`
	MaxDirSearchDepth int    `json:"maxDirSearchDepth"`
	FilePattern       string `json:"filePattern"`
}

type ContainerStdoutSource struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Enabled               bool   `json:"enabled"`
	NamespaceRegex        string `json:"namespaceRegex"`
	PodRegex              string `json:"podRegex"`
	ContainerRegex        string `json:"containerRegex"`
	IncludeK8sLabel       string `json:"includeK8sLabel"`
	ExcludeK8sLabel       string `json:"excludeK8sLabel"`
	IncludeEnv            string `json:"includeEnv"`
	ExcludeEnv            string `json:"excludeEnv"`
	IncludeContainerLabel string `json:"includeContainerLabel"`
	ExcludeContainerLabel string `json:"excludeContainerLabel"`
	CollectStdout         bool   `json:"collectStdout"`
	CollectStderr         bool   `json:"collectStderr"`
}

type HostFileCollectionInput struct {
	HostFiles []HostFileSource `json:"hostFiles"`
}

type ContainerStdoutCollectionInput struct {
	ContainerStdout []ContainerStdoutSource `json:"containerStdout"`
}

type RenderedLogCollectionFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type LogCollectionPreview struct {
	ScopeType string                      `json:"scopeType"`
	ScopeID   string                      `json:"scopeId"`
	Files     []RenderedLogCollectionFile `json:"files"`
}

type HostFileCollection struct {
	ScopeType string           `json:"scopeType"`
	ScopeID   string           `json:"scopeId"`
	Version   int64            `json:"version"`
	HostFiles []HostFileSource `json:"hostFiles"`
}

type ContainerStdoutCollection struct {
	ScopeType       string                  `json:"scopeType"`
	ScopeID         string                  `json:"scopeId"`
	Version         int64                   `json:"version"`
	ContainerStdout []ContainerStdoutSource `json:"containerStdout"`
}

type logCollectionRepository interface {
	GetLogCollectionConfig(core.Context, string, string) (*ac.LogCollectionConfig, error)
	SaveLogCollectionConfig(core.Context, *ac.LogCollectionConfig) error
}

type hostFilePipeline struct {
	Enable bool            `yaml:"enable"`
	Inputs []hostFileInput `yaml:"inputs"`
}

type hostFileInput struct {
	Type                           string   `yaml:"Type"`
	FilePaths                      []string `yaml:"FilePaths"`
	MaxDirSearchDepth              int      `yaml:"MaxDirSearchDepth,omitempty"`
	AllowingIncludedByMultiConfigs bool     `yaml:"AllowingIncludedByMultiConfigs"`
}

type containerStdoutPipeline struct {
	Enable bool                   `yaml:"enable"`
	Inputs []containerStdoutInput `yaml:"inputs"`
}

type containerStdoutInput struct {
	Type            string            `yaml:"Type"`
	IgnoringStdout  bool              `yaml:"IgnoringStdout"`
	IgnoringStderr  bool              `yaml:"IgnoringStderr"`
	ContainerFilter *containerFilters `yaml:"ContainerFilters,omitempty"`
}

type containerFilters struct {
	K8sNamespaceRegex     string            `yaml:"K8sNamespaceRegex,omitempty"`
	K8sPodRegex           string            `yaml:"K8sPodRegex,omitempty"`
	K8sContainerRegex     string            `yaml:"K8sContainerRegex,omitempty"`
	IncludeK8sLabel       map[string]string `yaml:"IncludeK8sLabel,omitempty"`
	ExcludeK8sLabel       map[string]string `yaml:"ExcludeK8sLabel,omitempty"`
	IncludeEnv            map[string]string `yaml:"IncludeEnv,omitempty"`
	ExcludeEnv            map[string]string `yaml:"ExcludeEnv,omitempty"`
	IncludeContainerLabel map[string]string `yaml:"IncludeContainerLabel,omitempty"`
	ExcludeContainerLabel map[string]string `yaml:"ExcludeContainerLabel,omitempty"`
}

func (s *Service) logCollectionRepo() (logCollectionRepository, error) {
	repo, ok := s.repo.(logCollectionRepository)
	if !ok {
		return nil, errors.New("log collection repository is not configured")
	}
	return repo, nil
}

func normalizeSourceID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return uuid.NewString()
	}
	return value
}

func safeSourceID(value string) string {
	value = strings.ToLower(value)
	value = regexp.MustCompile(`[^a-z0-9_-]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "source"
	}
	return value
}

func parseFilterMap(label, value string) (map[string]string, error) {
	result := map[string]string{}
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, fmt.Errorf("%s filter key is required", label)
		}
		if _, exists := result[key]; exists {
			return nil, fmt.Errorf("%s filter key %s is duplicated", label, key)
		}
		filterValue := ""
		if len(parts) == 2 {
			filterValue = strings.TrimSpace(parts[1])
		}
		result[key] = filterValue
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

func (filters *containerFilters) empty() bool {
	return filters.K8sNamespaceRegex == "" &&
		filters.K8sPodRegex == "" &&
		filters.K8sContainerRegex == "" &&
		len(filters.IncludeK8sLabel) == 0 &&
		len(filters.ExcludeK8sLabel) == 0 &&
		len(filters.IncludeEnv) == 0 &&
		len(filters.ExcludeEnv) == 0 &&
		len(filters.IncludeContainerLabel) == 0 &&
		len(filters.ExcludeContainerLabel) == 0
}

func normalizeHostFiles(input []HostFileSource) ([]HostFileSource, error) {
	if len(input) > 20 {
		return nil, errors.New("at most 20 host file sources are allowed")
	}
	result := make([]HostFileSource, 0, len(input))
	seen := map[string]bool{}
	seenFileIDs := map[string]bool{}
	for _, source := range input {
		source.ID = normalizeSourceID(source.ID)
		source.Name = strings.TrimSpace(source.Name)
		source.Path = filepath.Clean(strings.TrimSpace(source.Path))
		source.FilePattern = strings.TrimSpace(source.FilePattern)
		if seen[source.ID] {
			return nil, fmt.Errorf("host file source id %s is duplicated", source.ID)
		}
		seen[source.ID] = true
		if !source.Enabled {
			result = append(result, source)
			continue
		}
		if source.Name == "" {
			return nil, errors.New("host file source name is required")
		}
		if !filepath.IsAbs(source.Path) {
			return nil, fmt.Errorf("host file path must be absolute: %s", source.Path)
		}
		if source.FilePattern == "" || strings.ContainsAny(source.FilePattern, `/\\`) {
			return nil, fmt.Errorf("file pattern for %s must be a file name pattern", source.Name)
		}
		if _, err := filepath.Match(source.FilePattern, "candidate.log"); err != nil {
			return nil, fmt.Errorf("invalid file pattern for %s: %v", source.Name, err)
		}
		if source.MaxDirSearchDepth < 0 || source.MaxDirSearchDepth > 1000 {
			return nil, fmt.Errorf("directory depth for %s must be between 0 and 1000", source.Name)
		}
		fileID := safeSourceID(source.ID)
		if seenFileIDs[fileID] {
			return nil, fmt.Errorf("host file source ids produce duplicate file name %s", fileID)
		}
		seenFileIDs[fileID] = true
		result = append(result, source)
	}
	return result, nil
}

func normalizeContainerStdout(input []ContainerStdoutSource) ([]ContainerStdoutSource, error) {
	if len(input) > 20 {
		return nil, errors.New("at most 20 container stdout sources are allowed")
	}
	result := make([]ContainerStdoutSource, 0, len(input))
	seen := map[string]bool{}
	seenFileIDs := map[string]bool{}
	for _, source := range input {
		source.ID = normalizeSourceID(source.ID)
		source.Name = strings.TrimSpace(source.Name)
		source.NamespaceRegex = strings.TrimSpace(source.NamespaceRegex)
		source.PodRegex = strings.TrimSpace(source.PodRegex)
		source.ContainerRegex = strings.TrimSpace(source.ContainerRegex)
		source.IncludeK8sLabel = strings.TrimSpace(source.IncludeK8sLabel)
		source.ExcludeK8sLabel = strings.TrimSpace(source.ExcludeK8sLabel)
		source.IncludeEnv = strings.TrimSpace(source.IncludeEnv)
		source.ExcludeEnv = strings.TrimSpace(source.ExcludeEnv)
		source.IncludeContainerLabel = strings.TrimSpace(source.IncludeContainerLabel)
		source.ExcludeContainerLabel = strings.TrimSpace(source.ExcludeContainerLabel)
		if seen[source.ID] {
			return nil, fmt.Errorf("container stdout source id %s is duplicated", source.ID)
		}
		seen[source.ID] = true
		if !source.Enabled {
			result = append(result, source)
			continue
		}
		if source.Name == "" {
			return nil, errors.New("container stdout source name is required")
		}
		if !source.CollectStdout && !source.CollectStderr {
			return nil, fmt.Errorf("%s must collect stdout or stderr", source.Name)
		}
		for label, expression := range map[string]string{
			"namespace": source.NamespaceRegex,
			"pod":       source.PodRegex,
			"container": source.ContainerRegex,
		} {
			if expression != "" {
				if _, err := regexp.Compile(expression); err != nil {
					return nil, fmt.Errorf("invalid %s regular expression for %s: %v", label, source.Name, err)
				}
			}
		}
		for label, value := range map[string]string{
			"include Pod label":       source.IncludeK8sLabel,
			"exclude Pod label":       source.ExcludeK8sLabel,
			"include environment":     source.IncludeEnv,
			"exclude environment":     source.ExcludeEnv,
			"include container label": source.IncludeContainerLabel,
			"exclude container label": source.ExcludeContainerLabel,
		} {
			if _, err := parseFilterMap(label, value); err != nil {
				return nil, fmt.Errorf("invalid filter for %s: %v", source.Name, err)
			}
		}
		fileID := safeSourceID(source.ID)
		if seenFileIDs[fileID] {
			return nil, fmt.Errorf("container stdout source ids produce duplicate file name %s", fileID)
		}
		seenFileIDs[fileID] = true
		result = append(result, source)
	}
	return result, nil
}

func renderHostFiles(sources []HostFileSource) ([]RenderedLogCollectionFile, error) {
	files := make([]RenderedLogCollectionFile, 0, len(sources))
	for _, source := range sources {
		if !source.Enabled {
			continue
		}
		pattern := filepath.Join(source.Path, source.FilePattern)
		if source.MaxDirSearchDepth > 0 {
			pattern = filepath.Join(source.Path, "**", source.FilePattern)
		}
		content, err := yaml.Marshal(hostFilePipeline{Enable: true, Inputs: []hostFileInput{{
			Type: "input_file", FilePaths: []string{pattern}, MaxDirSearchDepth: source.MaxDirSearchDepth,
			AllowingIncludedByMultiConfigs: true,
		}}})
		if err != nil {
			return nil, err
		}
		files = append(files, RenderedLogCollectionFile{Path: "config/local/apo-host-file-" + safeSourceID(source.ID) + ".yaml", Content: string(content)})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func containerFiltersForSource(source ContainerStdoutSource) (*containerFilters, error) {
	includeK8sLabel, err := parseFilterMap("include Pod label", source.IncludeK8sLabel)
	if err != nil {
		return nil, err
	}
	excludeK8sLabel, err := parseFilterMap("exclude Pod label", source.ExcludeK8sLabel)
	if err != nil {
		return nil, err
	}
	includeEnv, err := parseFilterMap("include environment", source.IncludeEnv)
	if err != nil {
		return nil, err
	}
	excludeEnv, err := parseFilterMap("exclude environment", source.ExcludeEnv)
	if err != nil {
		return nil, err
	}
	includeContainerLabel, err := parseFilterMap("include container label", source.IncludeContainerLabel)
	if err != nil {
		return nil, err
	}
	excludeContainerLabel, err := parseFilterMap("exclude container label", source.ExcludeContainerLabel)
	if err != nil {
		return nil, err
	}
	filters := &containerFilters{
		K8sNamespaceRegex:     source.NamespaceRegex,
		K8sPodRegex:           source.PodRegex,
		K8sContainerRegex:     source.ContainerRegex,
		IncludeK8sLabel:       includeK8sLabel,
		ExcludeK8sLabel:       excludeK8sLabel,
		IncludeEnv:            includeEnv,
		ExcludeEnv:            excludeEnv,
		IncludeContainerLabel: includeContainerLabel,
		ExcludeContainerLabel: excludeContainerLabel,
	}
	if filters.empty() {
		return nil, nil
	}
	return filters, nil
}

func renderContainerStdout(sources []ContainerStdoutSource) ([]RenderedLogCollectionFile, error) {
	files := make([]RenderedLogCollectionFile, 0, len(sources))
	for _, source := range sources {
		if !source.Enabled {
			continue
		}
		filters, err := containerFiltersForSource(source)
		if err != nil {
			return nil, err
		}
		content, err := yaml.Marshal(containerStdoutPipeline{Enable: true, Inputs: []containerStdoutInput{{
			Type: "input_container_stdio", IgnoringStdout: !source.CollectStdout, IgnoringStderr: !source.CollectStderr,
			ContainerFilter: filters,
		}}})
		if err != nil {
			return nil, err
		}
		files = append(files, RenderedLogCollectionFile{Path: "config/local/apo-container-stdout-" + safeSourceID(source.ID) + ".yaml", Content: string(content)})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func (s *Service) validateLogCollectionScope(ctx core.Context, scopeType, scopeID string) error {
	scopeID = strings.TrimSpace(scopeID)
	if scopeID == "" {
		return errors.New("scope id is required")
	}
	if scopeType == logCollectionScopeNode {
		_, err := s.repo.GetAgentNode(ctx, scopeID)
		return err
	}
	if scopeType != logCollectionScopeCluster {
		return fmt.Errorf("unsupported log collection scope: %s", scopeType)
	}
	nodes, err := s.repo.ListAgentNodes(ctx)
	if err != nil {
		return err
	}
	found := false
	for _, node := range nodes {
		if node.ClusterID == scopeID {
			found = true
		}
	}
	if found {
		return nil
	}
	return fmt.Errorf("cluster %s has no registered nodes", scopeID)
}

func (s *Service) GetHostFileCollection(ctx core.Context, nodeID string) (*HostFileCollection, error) {
	if err := s.validateLogCollectionScope(ctx, logCollectionScopeNode, nodeID); err != nil {
		return nil, err
	}
	repo, err := s.logCollectionRepo()
	if err != nil {
		return nil, err
	}
	stored, err := repo.GetLogCollectionConfig(ctx, logCollectionScopeNode, nodeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HostFileCollection{ScopeType: logCollectionScopeNode, ScopeID: nodeID, HostFiles: []HostFileSource{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var input HostFileCollectionInput
	if err := json.Unmarshal(stored.Config, &input); err != nil {
		return nil, err
	}
	return &HostFileCollection{ScopeType: stored.ScopeType, ScopeID: stored.ScopeID, Version: stored.Version, HostFiles: input.HostFiles}, nil
}

func (s *Service) SaveHostFileCollection(ctx core.Context, nodeID string, input HostFileCollectionInput, user int64) (*HostFileCollection, error) {
	preview, normalized, err := s.previewHostFileCollection(ctx, nodeID, input)
	_ = preview
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(HostFileCollectionInput{HostFiles: normalized})
	if err != nil {
		return nil, err
	}
	repo, err := s.logCollectionRepo()
	if err != nil {
		return nil, err
	}
	stored := &ac.LogCollectionConfig{ScopeType: logCollectionScopeNode, ScopeID: nodeID, Config: ac.JSON(data), Version: 1, UpdatedBy: user}
	if err := repo.SaveLogCollectionConfig(ctx, stored); err != nil {
		return nil, err
	}
	return &HostFileCollection{ScopeType: stored.ScopeType, ScopeID: stored.ScopeID, Version: stored.Version, HostFiles: normalized}, nil
}

func (s *Service) previewHostFileCollection(ctx core.Context, nodeID string, input HostFileCollectionInput) (*LogCollectionPreview, []HostFileSource, error) {
	if err := s.validateLogCollectionScope(ctx, logCollectionScopeNode, nodeID); err != nil {
		return nil, nil, err
	}
	normalized, err := normalizeHostFiles(input.HostFiles)
	if err != nil {
		return nil, nil, err
	}
	files, err := renderHostFiles(normalized)
	return &LogCollectionPreview{ScopeType: logCollectionScopeNode, ScopeID: nodeID, Files: files}, normalized, err
}

func (s *Service) PreviewHostFileCollection(ctx core.Context, nodeID string, input HostFileCollectionInput) (*LogCollectionPreview, error) {
	preview, _, err := s.previewHostFileCollection(ctx, nodeID, input)
	return preview, err
}

func (s *Service) GetContainerStdoutCollection(ctx core.Context, clusterID string) (*ContainerStdoutCollection, error) {
	if err := s.validateLogCollectionScope(ctx, logCollectionScopeCluster, clusterID); err != nil {
		return nil, err
	}
	repo, err := s.logCollectionRepo()
	if err != nil {
		return nil, err
	}
	stored, err := repo.GetLogCollectionConfig(ctx, logCollectionScopeCluster, clusterID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &ContainerStdoutCollection{ScopeType: logCollectionScopeCluster, ScopeID: clusterID, ContainerStdout: []ContainerStdoutSource{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var input ContainerStdoutCollectionInput
	if err := json.Unmarshal(stored.Config, &input); err != nil {
		return nil, err
	}
	return &ContainerStdoutCollection{ScopeType: stored.ScopeType, ScopeID: stored.ScopeID, Version: stored.Version, ContainerStdout: input.ContainerStdout}, nil
}

func (s *Service) SaveContainerStdoutCollection(ctx core.Context, clusterID string, input ContainerStdoutCollectionInput, user int64) (*ContainerStdoutCollection, error) {
	_, normalized, err := s.previewContainerStdoutCollection(ctx, clusterID, input)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(ContainerStdoutCollectionInput{ContainerStdout: normalized})
	if err != nil {
		return nil, err
	}
	repo, err := s.logCollectionRepo()
	if err != nil {
		return nil, err
	}
	stored := &ac.LogCollectionConfig{ScopeType: logCollectionScopeCluster, ScopeID: clusterID, Config: ac.JSON(data), Version: 1, UpdatedBy: user}
	if err := repo.SaveLogCollectionConfig(ctx, stored); err != nil {
		return nil, err
	}
	return &ContainerStdoutCollection{ScopeType: stored.ScopeType, ScopeID: stored.ScopeID, Version: stored.Version, ContainerStdout: normalized}, nil
}

func (s *Service) previewContainerStdoutCollection(ctx core.Context, clusterID string, input ContainerStdoutCollectionInput) (*LogCollectionPreview, []ContainerStdoutSource, error) {
	if err := s.validateLogCollectionScope(ctx, logCollectionScopeCluster, clusterID); err != nil {
		return nil, nil, err
	}
	normalized, err := normalizeContainerStdout(input.ContainerStdout)
	if err != nil {
		return nil, nil, err
	}
	files, err := renderContainerStdout(normalized)
	return &LogCollectionPreview{ScopeType: logCollectionScopeCluster, ScopeID: clusterID, Files: files}, normalized, err
}

func (s *Service) PreviewContainerStdoutCollection(ctx core.Context, clusterID string, input ContainerStdoutCollectionInput) (*LogCollectionPreview, error) {
	preview, _, err := s.previewContainerStdoutCollection(ctx, clusterID, input)
	return preview, err
}
