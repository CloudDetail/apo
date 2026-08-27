package agentconfig

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	logCollectionScopeNode      = "node"
	logCollectionScopeCluster   = "cluster"
	logCollectionScopeGlobal    = "global"
	logCollectionDefaultScopeID = "default"
	defaultK8sContainerSourceID = "k8s-default-container-stdout"
	maxLogEventsPerSecond       = int64(1_000_000_000)
)

type HostFileSource struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Enabled           bool    `json:"enabled"`
	Path              string  `json:"path"`
	MaxDirSearchDepth int     `json:"maxDirSearchDepth"`
	FilePattern       string  `json:"filePattern"`
	ProcessingRule    *string `json:"processingRule,omitempty"`
	OutputRule        *string `json:"outputRule,omitempty"`
}

type FilterMap map[string]string

func (filterMap *FilterMap) UnmarshalJSON(data []byte) error {
	var values map[string]string
	if err := json.Unmarshal(data, &values); err == nil {
		*filterMap = values
		return nil
	}

	var legacy string
	if err := json.Unmarshal(data, &legacy); err != nil {
		return errors.New("filter map must be an object")
	}
	values, err := parseFilterMap("legacy", legacy)
	if err != nil {
		return err
	}
	*filterMap = values
	return nil
}

type ContainerStdoutSource struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	Enabled               bool      `json:"enabled"`
	NamespaceRegex        string    `json:"namespaceRegex"`
	PodRegex              string    `json:"podRegex"`
	ContainerRegex        string    `json:"containerRegex"`
	IncludeK8sLabel       string    `json:"includeK8sLabel"`
	ExcludeK8sLabel       FilterMap `json:"excludeK8sLabel"`
	IncludeEnv            string    `json:"includeEnv"`
	ExcludeEnv            string    `json:"excludeEnv"`
	IncludeContainerLabel string    `json:"includeContainerLabel"`
	ExcludeContainerLabel string    `json:"excludeContainerLabel"`
	CollectStdout         bool      `json:"collectStdout"`
	CollectStderr         bool      `json:"collectStderr"`
	MaxEventsPerSecond    int64     `json:"maxEventsPerSecond"`
	ProcessingRule        *string   `json:"processingRule,omitempty"`
	OutputRule            *string   `json:"outputRule,omitempty"`
}

type HostFileCollectionInput struct {
	ContainerStdout     []ContainerStdoutSource `json:"containerStdout"`
	HostFiles           []HostFileSource        `json:"hostFiles"`
	DisableClusterRules bool                    `json:"disableClusterRules"`
	ProcessingRule      *string                 `json:"processingRule,omitempty"`
	OutputRule          *string                 `json:"outputRule,omitempty"`
}

type ContainerStdoutCollectionInput struct {
	ContainerStdout []ContainerStdoutSource `json:"containerStdout"`
	HostFiles       []HostFileSource        `json:"hostFiles"`
	ProcessingRule  *string                 `json:"processingRule,omitempty"`
	OutputRule      *string                 `json:"outputRule,omitempty"`
}

type RenderedLogCollectionFile struct {
	Path           string `json:"path"`
	Content        string `json:"content"`
	InputCount     int    `json:"inputCount"`
	ProcessorCount int    `json:"processorCount"`
	FlusherCount   int    `json:"flusherCount"`
}

type LogCollectionPreview struct {
	ScopeType string                      `json:"scopeType"`
	ScopeID   string                      `json:"scopeId"`
	Files     []RenderedLogCollectionFile `json:"files"`
}

type HostFileCollection struct {
	ScopeType           string                  `json:"scopeType"`
	ScopeID             string                  `json:"scopeId"`
	Version             int64                   `json:"version"`
	ContainerStdout     []ContainerStdoutSource `json:"containerStdout"`
	HostFiles           []HostFileSource        `json:"hostFiles"`
	DisableClusterRules bool                    `json:"disableClusterRules"`
	ProcessingRule      string                  `json:"processingRule"`
	OutputRule          string                  `json:"outputRule"`
	ProcessorCount      int                     `json:"processorCount"`
	OutputCount         int                     `json:"outputCount"`
}

type ContainerStdoutCollection struct {
	ScopeType       string                  `json:"scopeType"`
	ScopeID         string                  `json:"scopeId"`
	Version         int64                   `json:"version"`
	ContainerStdout []ContainerStdoutSource `json:"containerStdout"`
	HostFiles       []HostFileSource        `json:"hostFiles"`
	ProcessingRule  string                  `json:"processingRule"`
	OutputRule      string                  `json:"outputRule"`
	ProcessorCount  int                     `json:"processorCount"`
	OutputCount     int                     `json:"outputCount"`
}

type PipelineRuleDefaults struct {
	ProcessingRule string `json:"processingRule"`
	OutputRule     string `json:"outputRule"`
	ProcessorCount int    `json:"processorCount"`
	OutputCount    int    `json:"outputCount"`
}

type LogCollectionDefaultsInput struct {
	File      PipelineRuleDefaults `json:"file"`
	Container PipelineRuleDefaults `json:"container"`
}

type LogCollectionDefaults struct {
	ScopeType string               `json:"scopeType"`
	ScopeID   string               `json:"scopeId"`
	Version   int64                `json:"version"`
	File      PipelineRuleDefaults `json:"file"`
	Container PipelineRuleDefaults `json:"container"`
}

type logCollectionRepository interface {
	GetLogCollectionConfig(core.Context, string, string) (*ac.LogCollectionConfig, error)
	SaveLogCollectionConfig(core.Context, *ac.LogCollectionConfig) error
}

type hostFileInput struct {
	Type                           string   `yaml:"Type"`
	FilePaths                      []string `yaml:"FilePaths"`
	MaxDirSearchDepth              int      `yaml:"MaxDirSearchDepth,omitempty"`
	AllowingIncludedByMultiConfigs bool     `yaml:"AllowingIncludedByMultiConfigs"`
}

type containerStdoutProcessor struct {
	Type  string `yaml:"Type"`
	Limit string `yaml:"Limit"`
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

type logPipelineBase struct {
	Path       string
	Global     *yaml.Node
	Processors *yaml.Node
	Flushers   *yaml.Node
	InputTypes []string
}

var (
	logCollectorProbeTypes = []string{"ilogtail", "loongcollector"}
	hostFileInputTypes     = map[string]struct{}{
		"input_file": {},
		"file_log":   {},
	}
	containerStdoutInputTypes = map[string]struct{}{
		"input_container_stdio": {},
		"service_docker_stdout": {},
	}
)

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

func normalizeFilterMap(label string, values FilterMap) (FilterMap, error) {
	if len(values) == 0 {
		return nil, nil
	}
	result := make(FilterMap, len(values))
	for key, value := range values {
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("%s filter key is required", label)
		}
		if _, exists := result[key]; exists {
			return nil, fmt.Errorf("%s filter key %s is duplicated", label, key)
		}
		result[key] = strings.TrimSpace(value)
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
		if err := validateEditablePipelineRules(source.ProcessingRule, source.OutputRule); err != nil {
			return nil, fmt.Errorf("invalid pipeline for %s: %v", source.Name, err)
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
		var err error
		source.ExcludeK8sLabel, err = normalizeFilterMap("exclude Pod label", source.ExcludeK8sLabel)
		if err != nil {
			return nil, fmt.Errorf("invalid filter for %s: %v", source.Name, err)
		}
		if !source.CollectStdout && !source.CollectStderr {
			return nil, fmt.Errorf("%s must collect stdout or stderr", source.Name)
		}
		if source.MaxEventsPerSecond < 0 || source.MaxEventsPerSecond > maxLogEventsPerSecond {
			return nil, fmt.Errorf("maximum log events per second for %s must be between 0 and %d", source.Name, maxLogEventsPerSecond)
		}
		if err := validateEditablePipelineRules(source.ProcessingRule, source.OutputRule); err != nil {
			return nil, fmt.Errorf("invalid pipeline for %s: %v", source.Name, err)
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

func hasEnabledHostFileSource(sources []HostFileSource) bool {
	for _, source := range sources {
		if source.Enabled {
			return true
		}
	}
	return false
}

func hasEnabledContainerStdoutSource(sources []ContainerStdoutSource) bool {
	for _, source := range sources {
		if source.Enabled {
			return true
		}
	}
	return false
}

func yamlMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			return nil
		}
		return yamlMappingValue(node.Content[0], key)
	}
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == key {
			return node.Content[index+1]
		}
	}
	return nil
}

func cloneYAMLNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}
	copy := *node
	copy.Content = make([]*yaml.Node, len(node.Content))
	for index, child := range node.Content {
		copy.Content[index] = cloneYAMLNode(child)
	}
	return &copy
}

func clearYAMLComments(node *yaml.Node) {
	if node == nil {
		return
	}
	node.HeadComment = ""
	node.LineComment = ""
	node.FootComment = ""
	for _, child := range node.Content {
		clearYAMLComments(child)
	}
}

func yamlValueNode(value any) (*yaml.Node, error) {
	content, err := yaml.Marshal(value)
	if err != nil {
		return nil, err
	}
	var document yaml.Node
	if err := yaml.Unmarshal(content, &document); err != nil {
		return nil, err
	}
	if len(document.Content) != 1 {
		return nil, errors.New("generated YAML value is empty")
	}
	return document.Content[0], nil
}

func yamlScalar(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
}

func yamlBool(value bool) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: fmt.Sprintf("%t", value)}
}

func parseLogPipelineBase(path, content string) (*logPipelineBase, error) {
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(content), &document); err != nil {
		return nil, err
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, nil
	}
	root := document.Content[0]
	flushers := yamlMappingValue(root, "flushers")
	if flushers == nil || flushers.Kind != yaml.SequenceNode || len(flushers.Content) == 0 {
		return nil, nil
	}
	inputTypes := []string{}
	if inputs := yamlMappingValue(root, "inputs"); inputs != nil && inputs.Kind == yaml.SequenceNode {
		for _, input := range inputs.Content {
			if inputType := yamlMappingValue(input, "Type"); inputType != nil {
				inputTypes = append(inputTypes, inputType.Value)
			}
		}
	}
	processors := yamlMappingValue(root, "processors")
	if processors != nil && processors.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("processors in %s must be a list", path)
	}
	return &logPipelineBase{
		Path:       path,
		Global:     cloneYAMLNode(yamlMappingValue(root, "global")),
		Processors: cloneYAMLNode(processors),
		Flushers:   cloneYAMLNode(flushers),
		InputTypes: inputTypes,
	}, nil
}

func pipelineMatchScore(base *logPipelineBase, preferredInputTypes map[string]struct{}) int {
	score := 1
	for _, inputType := range base.InputTypes {
		if _, matched := preferredInputTypes[inputType]; matched {
			return 2
		}
	}
	return score
}

func (s *Service) nodeLogPipelineBase(ctx core.Context, nodeID string, preferredInputTypes map[string]struct{}) (*logPipelineBase, error) {
	var selected *logPipelineBase
	selectedScore := 0
	selectedProbeType := ""
	foundBaseline := false
	for _, probeType := range logCollectorProbeTypes {
		initial, err := s.repo.GetAgentInitialConfig(ctx, nodeID, probeType)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("get %s initial config for node %s: %w", probeType, nodeID, err)
		}
		foundBaseline = true
		var files []ac.InitialConfigFile
		if err := json.Unmarshal(initial.Files, &files); err != nil {
			return nil, fmt.Errorf("decode %s initial config for node %s: %w", probeType, nodeID, err)
		}
		sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
		for _, file := range files {
			candidate, parseErr := parseLogPipelineBase(file.Path, file.Content)
			if parseErr != nil {
				log.Printf("agentconfig: skip invalid log collector pipeline candidate node=%s probe=%s path=%s error=%v", nodeID, probeType, file.Path, parseErr)
				continue
			}
			if candidate == nil {
				continue
			}
			score := pipelineMatchScore(candidate, preferredInputTypes)
			if score > selectedScore {
				selected = candidate
				selectedScore = score
				selectedProbeType = probeType
			}
		}
	}
	if !foundBaseline {
		return nil, fmt.Errorf("node %s has no registered log collector initial config (tried %s)", nodeID, strings.Join(logCollectorProbeTypes, ", "))
	}
	if selected == nil {
		return nil, fmt.Errorf("node %s log collector initial config has no pipeline with flushers", nodeID)
	}
	if selectedScore == 1 {
		log.Printf("agentconfig: use non-matching log collector pipeline as common preview base node=%s probe=%s path=%s", nodeID, selectedProbeType, selected.Path)
	}
	return selected, nil
}

func pipelineBaseSignature(base *logPipelineBase) (string, error) {
	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	if base.Global != nil {
		root.Content = append(root.Content, yamlScalar("global"), cloneYAMLNode(base.Global))
	}
	if base.Processors != nil {
		root.Content = append(root.Content, yamlScalar("processors"), cloneYAMLNode(base.Processors))
	}
	root.Content = append(root.Content, yamlScalar("flushers"), cloneYAMLNode(base.Flushers))
	clearYAMLComments(root)
	content, err := yaml.Marshal(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}})
	return string(content), err
}

func renderPipelineRule(key string, value *yaml.Node) (string, int, error) {
	if value == nil || value.Kind != yaml.SequenceNode || len(value.Content) == 0 {
		return "", 0, nil
	}
	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Content: []*yaml.Node{
		yamlScalar(key), cloneYAMLNode(value),
	}}
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}); err != nil {
		return "", 0, err
	}
	if err := encoder.Close(); err != nil {
		return "", 0, err
	}
	return buffer.String(), len(value.Content), nil
}

func parseEditablePipelineRule(key string, content *string, required bool) (*yaml.Node, error) {
	if content == nil {
		return nil, nil
	}
	if strings.TrimSpace(*content) == "" {
		if required {
			return nil, fmt.Errorf("%s must contain at least one item", key)
		}
		return nil, nil
	}
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(*content), &document); err != nil {
		return nil, fmt.Errorf("invalid %s YAML: %w", key, err)
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s must be a YAML mapping", key)
	}
	root := document.Content[0]
	if len(root.Content) != 2 || root.Content[0].Value != key {
		return nil, fmt.Errorf("%s YAML must only contain the top-level %s field", key, key)
	}
	value := root.Content[1]
	if value.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("%s must be a list", key)
	}
	if required && len(value.Content) == 0 {
		return nil, fmt.Errorf("%s must contain at least one item", key)
	}
	for index, item := range value.Content {
		if item.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("%s item %d must be a mapping", key, index+1)
		}
		itemType := yamlMappingValue(item, "Type")
		if itemType == nil || itemType.Kind != yaml.ScalarNode || strings.TrimSpace(itemType.Value) == "" {
			return nil, fmt.Errorf("%s item %d must contain Type", key, index+1)
		}
	}
	return cloneYAMLNode(value), nil
}

func pipelineBaseWithOverrides(base *logPipelineBase, processingRule, outputRule *string) (*logPipelineBase, error) {
	overridden := &logPipelineBase{
		Path:       base.Path,
		Global:     cloneYAMLNode(base.Global),
		Processors: cloneYAMLNode(base.Processors),
		Flushers:   cloneYAMLNode(base.Flushers),
		InputTypes: append([]string(nil), base.InputTypes...),
	}
	if processingRule != nil {
		processors, err := parseEditablePipelineRule("processors", processingRule, false)
		if err != nil {
			return nil, err
		}
		overridden.Processors = processors
	}
	if outputRule != nil {
		flushers, err := parseEditablePipelineRule("flushers", outputRule, true)
		if err != nil {
			return nil, err
		}
		overridden.Flushers = flushers
	}
	return overridden, nil
}

func validateEditablePipelineRules(processingRule, outputRule *string) error {
	if _, err := parseEditablePipelineRule("processors", processingRule, false); err != nil {
		return err
	}
	_, err := parseEditablePipelineRule("flushers", outputRule, true)
	return err
}

func pipelineRules(base *logPipelineBase) (string, string, int, int, error) {
	processingRule, processorCount, err := renderPipelineRule("processors", base.Processors)
	if err != nil {
		return "", "", 0, 0, err
	}
	outputRule, outputCount, err := renderPipelineRule("flushers", base.Flushers)
	if err != nil {
		return "", "", 0, 0, err
	}
	return processingRule, outputRule, processorCount, outputCount, nil
}

func pipelineRulePointers(sourceProcessing, sourceOutput, legacyProcessing, legacyOutput *string, defaults PipelineRuleDefaults) (*string, *string) {
	processing := sourceProcessing
	if processing == nil {
		processing = legacyProcessing
	}
	if processing == nil && strings.TrimSpace(defaults.ProcessingRule) != "" {
		processing = &defaults.ProcessingRule
	}
	output := sourceOutput
	if output == nil {
		output = legacyOutput
	}
	if output == nil && strings.TrimSpace(defaults.OutputRule) != "" {
		output = &defaults.OutputRule
	}
	return processing, output
}

func migrateLegacyHostPipelines(sources []HostFileSource, processingRule, outputRule *string) []HostFileSource {
	for index := range sources {
		if sources[index].ProcessingRule == nil {
			sources[index].ProcessingRule = processingRule
		}
		if sources[index].OutputRule == nil {
			sources[index].OutputRule = outputRule
		}
	}
	return sources
}

func migrateLegacyContainerPipelines(sources []ContainerStdoutSource, processingRule, outputRule *string) []ContainerStdoutSource {
	for index := range sources {
		if sources[index].ProcessingRule == nil {
			sources[index].ProcessingRule = processingRule
		}
		if sources[index].OutputRule == nil {
			sources[index].OutputRule = outputRule
		}
	}
	return sources
}

func pipelineRuleDefaults(base *logPipelineBase) (PipelineRuleDefaults, error) {
	processingRule, outputRule, processorCount, outputCount, err := pipelineRules(base)
	if err != nil {
		return PipelineRuleDefaults{}, err
	}
	return PipelineRuleDefaults{
		ProcessingRule: processingRule,
		OutputRule:     outputRule,
		ProcessorCount: processorCount,
		OutputCount:    outputCount,
	}, nil
}

func pipelineRuleDefaultsWithCounts(defaults PipelineRuleDefaults) (PipelineRuleDefaults, error) {
	processors, err := parseEditablePipelineRule("processors", &defaults.ProcessingRule, false)
	if err != nil {
		return PipelineRuleDefaults{}, err
	}
	flushers, err := parseEditablePipelineRule("flushers", &defaults.OutputRule, true)
	if err != nil {
		return PipelineRuleDefaults{}, err
	}
	defaults.ProcessorCount = 0
	if processors != nil {
		defaults.ProcessorCount = len(processors.Content)
	}
	defaults.OutputCount = len(flushers.Content)
	return defaults, nil
}

func (s *Service) reportedPipelineDefaults(ctx core.Context, preferredInputTypes map[string]struct{}) (PipelineRuleDefaults, error) {
	nodes, err := s.repo.ListAgentNodes(ctx)
	if err != nil {
		return PipelineRuleDefaults{}, err
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	var lastErr error
	for _, node := range nodes {
		base, err := s.nodeLogPipelineBase(ctx, node.ID, preferredInputTypes)
		if err != nil {
			lastErr = err
			continue
		}
		return pipelineRuleDefaults(base)
	}
	if lastErr != nil {
		return PipelineRuleDefaults{}, lastErr
	}
	return PipelineRuleDefaults{}, errors.New("no reported iLogtail pipeline is available")
}

func (s *Service) storedLogCollectionDefaults(ctx core.Context) (*LogCollectionDefaults, error) {
	repo, err := s.logCollectionRepo()
	if err != nil {
		return nil, err
	}
	result := &LogCollectionDefaults{
		ScopeType: logCollectionScopeGlobal,
		ScopeID:   logCollectionDefaultScopeID,
	}
	stored, err := repo.GetLogCollectionConfig(ctx, logCollectionScopeGlobal, logCollectionDefaultScopeID)
	if err == nil {
		var input LogCollectionDefaultsInput
		if err := json.Unmarshal(stored.Config, &input); err != nil {
			return nil, err
		}
		result.Version = stored.Version
		result.File = input.File
		result.Container = input.Container
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return result, nil
}

func (s *Service) GetLogCollectionDefaults(ctx core.Context) (*LogCollectionDefaults, error) {
	result, err := s.storedLogCollectionDefaults(ctx)
	if err != nil {
		return nil, err
	}

	if result.Version == 0 {
		fallback, fallbackErr := s.reportedPipelineDefaults(ctx, hostFileInputTypes)
		if fallbackErr != nil {
			log.Printf("agentconfig: file pipeline defaults unavailable error=%v", fallbackErr)
		} else {
			if strings.TrimSpace(result.File.ProcessingRule) == "" {
				result.File.ProcessingRule = fallback.ProcessingRule
			}
			if strings.TrimSpace(result.File.OutputRule) == "" {
				result.File.OutputRule = fallback.OutputRule
			}
		}
	}
	if result.Version == 0 {
		fallback, fallbackErr := s.reportedPipelineDefaults(ctx, containerStdoutInputTypes)
		if fallbackErr != nil {
			log.Printf("agentconfig: container pipeline defaults unavailable error=%v", fallbackErr)
		} else {
			if strings.TrimSpace(result.Container.ProcessingRule) == "" {
				result.Container.ProcessingRule = fallback.ProcessingRule
			}
			if strings.TrimSpace(result.Container.OutputRule) == "" {
				result.Container.OutputRule = fallback.OutputRule
			}
		}
	}
	if strings.TrimSpace(result.File.OutputRule) != "" {
		withCounts, countErr := pipelineRuleDefaultsWithCounts(result.File)
		if countErr != nil {
			return nil, fmt.Errorf("invalid stored file defaults: %v", countErr)
		}
		result.File = withCounts
	}
	if strings.TrimSpace(result.Container.OutputRule) != "" {
		withCounts, countErr := pipelineRuleDefaultsWithCounts(result.Container)
		if countErr != nil {
			return nil, fmt.Errorf("invalid stored container defaults: %v", countErr)
		}
		result.Container = withCounts
	}
	return result, nil
}

func (s *Service) SaveLogCollectionDefaults(ctx core.Context, input LogCollectionDefaultsInput, user int64) (*LogCollectionDefaults, error) {
	if err := validateEditablePipelineRules(&input.File.ProcessingRule, &input.File.OutputRule); err != nil {
		return nil, fmt.Errorf("invalid file defaults: %v", err)
	}
	if err := validateEditablePipelineRules(&input.Container.ProcessingRule, &input.Container.OutputRule); err != nil {
		return nil, fmt.Errorf("invalid container defaults: %v", err)
	}
	data, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	repo, err := s.logCollectionRepo()
	if err != nil {
		return nil, err
	}
	stored := &ac.LogCollectionConfig{
		ScopeType: logCollectionScopeGlobal,
		ScopeID:   logCollectionDefaultScopeID,
		Config:    ac.JSON(data),
		Version:   1,
		UpdatedBy: user,
	}
	if err := repo.SaveLogCollectionConfig(ctx, stored); err != nil {
		return nil, err
	}
	result, err := s.GetLogCollectionDefaults(ctx)
	if err != nil {
		return nil, err
	}
	result.Version = stored.Version
	return result, nil
}

func (s *Service) attachHostPipelineRules(ctx core.Context, collection *HostFileCollection, overrideProcessingRule, overrideOutputRule *string) {
	base, err := s.nodeLogPipelineBase(ctx, collection.ScopeID, hostFileInputTypes)
	if err != nil {
		log.Printf("agentconfig: host collection pipeline defaults unavailable node=%s error=%v", collection.ScopeID, err)
		return
	}
	base, err = pipelineBaseWithOverrides(base, overrideProcessingRule, overrideOutputRule)
	if err != nil {
		log.Printf("agentconfig: host collection pipeline overrides invalid node=%s error=%v", collection.ScopeID, err)
		return
	}
	processingRule, outputRule, processorCount, outputCount, err := pipelineRules(base)
	if err != nil {
		log.Printf("agentconfig: render host collection pipeline defaults node=%s error=%v", collection.ScopeID, err)
		return
	}
	collection.ProcessingRule = processingRule
	collection.OutputRule = outputRule
	collection.ProcessorCount = processorCount
	collection.OutputCount = outputCount
}

func (s *Service) attachContainerPipelineRules(ctx core.Context, collection *ContainerStdoutCollection, overrideProcessingRule, overrideOutputRule *string) {
	base, err := s.clusterLogPipelineBase(ctx, collection.ScopeID, containerStdoutInputTypes, overrideProcessingRule, overrideOutputRule)
	if err != nil {
		log.Printf("agentconfig: container collection pipeline defaults unavailable cluster=%s error=%v", collection.ScopeID, err)
		return
	}
	processingRule, outputRule, processorCount, outputCount, err := pipelineRules(base)
	if err != nil {
		log.Printf("agentconfig: render container collection pipeline defaults cluster=%s error=%v", collection.ScopeID, err)
		return
	}
	collection.ProcessingRule = processingRule
	collection.OutputRule = outputRule
	collection.ProcessorCount = processorCount
	collection.OutputCount = outputCount
}

func (s *Service) clusterLogPipelineBase(ctx core.Context, clusterID string, preferredInputTypes map[string]struct{}, processingRule, outputRule *string) (*logPipelineBase, error) {
	nodes, err := s.repo.ListAgentNodes(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	var selected *logPipelineBase
	selectedSignature := ""
	selectedNodeID := ""
	for _, node := range nodes {
		if node.ClusterID != clusterID {
			continue
		}
		base, err := s.nodeLogPipelineBase(ctx, node.ID, preferredInputTypes)
		if err != nil {
			return nil, err
		}
		base, err = pipelineBaseWithOverrides(base, processingRule, outputRule)
		if err != nil {
			return nil, err
		}
		signature, err := pipelineBaseSignature(base)
		if err != nil {
			return nil, err
		}
		if selected == nil {
			selected = base
			selectedSignature = signature
			selectedNodeID = node.ID
			continue
		}
		if signature != selectedSignature {
			return nil, fmt.Errorf("cluster %s has inconsistent iLogtail processors or flushers between nodes %s and %s", clusterID, selectedNodeID, node.ID)
		}
	}
	if selected == nil {
		return nil, fmt.Errorf("cluster %s has no iLogtail pipeline available for preview", clusterID)
	}
	return selected, nil
}

func renderLogPipeline(inputs any, base *logPipelineBase, maxEventsPerSecond int64) (string, int, int, error) {
	inputNode, err := yamlValueNode(inputs)
	if err != nil {
		return "", 0, 0, err
	}
	if inputNode.Kind != yaml.SequenceNode {
		return "", 0, 0, errors.New("generated inputs must be a list")
	}
	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	root.Content = append(root.Content, yamlScalar("enable"), yamlBool(true))
	if base.Global != nil {
		root.Content = append(root.Content, yamlScalar("global"), cloneYAMLNode(base.Global))
	}
	root.Content = append(root.Content, yamlScalar("inputs"), inputNode)

	processors := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	if base.Processors != nil {
		for _, processor := range base.Processors.Content {
			if processorType := yamlMappingValue(processor, "Type"); processorType != nil && processorType.Value == "processor_rate_limit" {
				continue
			}
			processors.Content = append(processors.Content, cloneYAMLNode(processor))
		}
	}
	if maxEventsPerSecond > 0 {
		rateLimit, err := yamlValueNode(containerStdoutProcessor{Type: "processor_rate_limit", Limit: fmt.Sprintf("%d/s", maxEventsPerSecond)})
		if err != nil {
			return "", 0, 0, err
		}
		processors.Content = append(processors.Content, rateLimit)
	}
	if len(processors.Content) > 0 {
		root.Content = append(root.Content, yamlScalar("processors"), processors)
	}
	root.Content = append(root.Content, yamlScalar("flushers"), cloneYAMLNode(base.Flushers))

	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(&yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}); err != nil {
		return "", 0, 0, err
	}
	if err := encoder.Close(); err != nil {
		return "", 0, 0, err
	}
	return buffer.String(), len(processors.Content), len(base.Flushers.Content), nil
}

func renderHostFiles(sources []HostFileSource, base *logPipelineBase) ([]RenderedLogCollectionFile, error) {
	files := make([]RenderedLogCollectionFile, 0, len(sources))
	for _, source := range sources {
		if !source.Enabled {
			continue
		}
		pattern := filepath.Join(source.Path, source.FilePattern)
		if source.MaxDirSearchDepth > 0 {
			pattern = filepath.Join(source.Path, "**", source.FilePattern)
		}
		content, processorCount, flusherCount, err := renderLogPipeline([]hostFileInput{{
			Type: "input_file", FilePaths: []string{pattern}, MaxDirSearchDepth: source.MaxDirSearchDepth,
			AllowingIncludedByMultiConfigs: true,
		}}, base, 0)
		if err != nil {
			return nil, err
		}
		files = append(files, RenderedLogCollectionFile{
			Path: "config/local/apo-host-file-" + safeSourceID(source.ID) + ".yaml", Content: content,
			InputCount: 1, ProcessorCount: processorCount, FlusherCount: flusherCount,
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func containerFiltersForSource(source ContainerStdoutSource) (*containerFilters, error) {
	includeK8sLabel, err := parseFilterMap("include Pod label", source.IncludeK8sLabel)
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
		ExcludeK8sLabel:       map[string]string(source.ExcludeK8sLabel),
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

func renderContainerStdout(sources []ContainerStdoutSource, base *logPipelineBase) ([]RenderedLogCollectionFile, error) {
	files := make([]RenderedLogCollectionFile, 0, len(sources))
	for _, source := range sources {
		if !source.Enabled {
			continue
		}
		filters, err := containerFiltersForSource(source)
		if err != nil {
			return nil, err
		}
		inputs := []containerStdoutInput{{
			Type: "input_container_stdio", IgnoringStdout: !source.CollectStdout, IgnoringStderr: !source.CollectStderr,
			ContainerFilter: filters,
		}}
		content, processorCount, flusherCount, err := renderLogPipeline(inputs, base, source.MaxEventsPerSecond)
		if err != nil {
			return nil, err
		}
		files = append(files, RenderedLogCollectionFile{
			Path: "config/local/apo-container-stdout-" + safeSourceID(source.ID) + ".yaml", Content: content,
			InputCount: 1, ProcessorCount: processorCount, FlusherCount: flusherCount,
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func defaultContainerStdoutSources(nodes []ac.Node, clusterID string) []ContainerStdoutSource {
	for _, node := range nodes {
		if node.ClusterID != clusterID {
			continue
		}
		environment := strings.ToLower(strings.TrimSpace(node.Environment))
		if environment != "k8s" && environment != "kubernetes" {
			continue
		}
		return []ContainerStdoutSource{{
			ID: defaultK8sContainerSourceID, Name: "Kubernetes default container stdout", Enabled: true,
			ExcludeK8sLabel: FilterMap{"app.kubernetes.io/instance": "^(syncause|syncause-agent)$"},
			CollectStdout:   true, CollectStderr: true,
		}}
	}
	return []ContainerStdoutSource{}
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
		collection := &HostFileCollection{
			ScopeType: logCollectionScopeNode, ScopeID: nodeID,
			ContainerStdout: []ContainerStdoutSource{}, HostFiles: []HostFileSource{},
		}
		s.attachHostPipelineRules(ctx, collection, nil, nil)
		return collection, nil
	}
	if err != nil {
		return nil, err
	}
	var input HostFileCollectionInput
	if err := json.Unmarshal(stored.Config, &input); err != nil {
		return nil, err
	}
	collection := &HostFileCollection{
		ScopeType: stored.ScopeType, ScopeID: stored.ScopeID, Version: stored.Version,
		ContainerStdout:     migrateLegacyContainerPipelines(input.ContainerStdout, input.ProcessingRule, input.OutputRule),
		HostFiles:           migrateLegacyHostPipelines(input.HostFiles, input.ProcessingRule, input.OutputRule),
		DisableClusterRules: input.DisableClusterRules,
	}
	s.attachHostPipelineRules(ctx, collection, input.ProcessingRule, input.OutputRule)
	return collection, nil
}

func (s *Service) SaveHostFileCollection(ctx core.Context, nodeID string, input HostFileCollectionInput, user int64) (*HostFileCollection, error) {
	if err := s.ensureNodeAllowsExplicitDelivery(ctx, nodeID); err != nil {
		return nil, err
	}
	preview, normalizedContainer, normalizedFiles, err := s.previewHostFileCollection(ctx, nodeID, input)
	_ = preview
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(HostFileCollectionInput{
		ContainerStdout: normalizedContainer, HostFiles: normalizedFiles,
		DisableClusterRules: input.DisableClusterRules,
		ProcessingRule:      input.ProcessingRule, OutputRule: input.OutputRule,
	})
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
	collection := &HostFileCollection{
		ScopeType: stored.ScopeType, ScopeID: stored.ScopeID, Version: stored.Version,
		ContainerStdout: normalizedContainer, HostFiles: normalizedFiles,
		DisableClusterRules: input.DisableClusterRules,
	}
	s.attachHostPipelineRules(ctx, collection, input.ProcessingRule, input.OutputRule)
	return collection, nil
}

func (s *Service) previewHostFileCollection(ctx core.Context, nodeID string, input HostFileCollectionInput) (*LogCollectionPreview, []ContainerStdoutSource, []HostFileSource, error) {
	if err := s.validateLogCollectionScope(ctx, logCollectionScopeNode, nodeID); err != nil {
		return nil, nil, nil, err
	}
	normalizedContainer, err := normalizeContainerStdout(input.ContainerStdout)
	if err != nil {
		return nil, nil, nil, err
	}
	normalized, err := normalizeHostFiles(input.HostFiles)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := validateEditablePipelineRules(input.ProcessingRule, input.OutputRule); err != nil {
		return nil, nil, nil, err
	}
	if !hasEnabledHostFileSource(normalized) && !hasEnabledContainerStdoutSource(normalizedContainer) {
		return &LogCollectionPreview{ScopeType: logCollectionScopeNode, ScopeID: nodeID, Files: []RenderedLogCollectionFile{}}, normalizedContainer, normalized, nil
	}
	defaults, err := s.storedLogCollectionDefaults(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	files := make([]RenderedLogCollectionFile, 0, len(normalized)+len(normalizedContainer))
	if hasEnabledHostFileSource(normalized) {
		base, err := s.nodeLogPipelineBase(ctx, nodeID, hostFileInputTypes)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, source := range normalized {
			if !source.Enabled {
				continue
			}
			processing, output := pipelineRulePointers(source.ProcessingRule, source.OutputRule, input.ProcessingRule, input.OutputRule, defaults.File)
			sourceBase, err := pipelineBaseWithOverrides(base, processing, output)
			if err != nil {
				return nil, nil, nil, err
			}
			rendered, err := renderHostFiles([]HostFileSource{source}, sourceBase)
			if err != nil {
				return nil, nil, nil, err
			}
			files = append(files, rendered...)
		}
	}
	if hasEnabledContainerStdoutSource(normalizedContainer) {
		base, err := s.nodeLogPipelineBase(ctx, nodeID, containerStdoutInputTypes)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, source := range normalizedContainer {
			if !source.Enabled {
				continue
			}
			processing, output := pipelineRulePointers(source.ProcessingRule, source.OutputRule, input.ProcessingRule, input.OutputRule, defaults.Container)
			sourceBase, err := pipelineBaseWithOverrides(base, processing, output)
			if err != nil {
				return nil, nil, nil, err
			}
			rendered, err := renderContainerStdout([]ContainerStdoutSource{source}, sourceBase)
			if err != nil {
				return nil, nil, nil, err
			}
			files = append(files, rendered...)
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return &LogCollectionPreview{ScopeType: logCollectionScopeNode, ScopeID: nodeID, Files: files}, normalizedContainer, normalized, nil
}

func (s *Service) PreviewHostFileCollection(ctx core.Context, nodeID string, input HostFileCollectionInput) (*LogCollectionPreview, error) {
	preview, _, _, err := s.previewHostFileCollection(ctx, nodeID, input)
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
		nodes, listErr := s.repo.ListAgentNodes(ctx)
		if listErr != nil {
			return nil, listErr
		}
		collection := &ContainerStdoutCollection{
			ScopeType: logCollectionScopeCluster, ScopeID: clusterID,
			ContainerStdout: defaultContainerStdoutSources(nodes, clusterID), HostFiles: []HostFileSource{},
		}
		s.attachContainerPipelineRules(ctx, collection, nil, nil)
		return collection, nil
	}
	if err != nil {
		return nil, err
	}
	var input ContainerStdoutCollectionInput
	if err := json.Unmarshal(stored.Config, &input); err != nil {
		return nil, err
	}
	collection := &ContainerStdoutCollection{
		ScopeType: stored.ScopeType, ScopeID: stored.ScopeID, Version: stored.Version,
		ContainerStdout: migrateLegacyContainerPipelines(input.ContainerStdout, input.ProcessingRule, input.OutputRule),
		HostFiles:       migrateLegacyHostPipelines(input.HostFiles, input.ProcessingRule, input.OutputRule),
	}
	s.attachContainerPipelineRules(ctx, collection, input.ProcessingRule, input.OutputRule)
	return collection, nil
}

func (s *Service) SaveContainerStdoutCollection(ctx core.Context, clusterID string, input ContainerStdoutCollectionInput, user int64) (*ContainerStdoutCollection, error) {
	_, normalizedContainer, normalizedFiles, err := s.previewContainerStdoutCollection(ctx, clusterID, input)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(ContainerStdoutCollectionInput{
		ContainerStdout: normalizedContainer, HostFiles: normalizedFiles,
		ProcessingRule: input.ProcessingRule, OutputRule: input.OutputRule,
	})
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
	collection := &ContainerStdoutCollection{
		ScopeType: stored.ScopeType, ScopeID: stored.ScopeID, Version: stored.Version,
		ContainerStdout: normalizedContainer, HostFiles: normalizedFiles,
	}
	s.attachContainerPipelineRules(ctx, collection, input.ProcessingRule, input.OutputRule)
	return collection, nil
}

func (s *Service) previewContainerStdoutCollection(ctx core.Context, clusterID string, input ContainerStdoutCollectionInput) (*LogCollectionPreview, []ContainerStdoutSource, []HostFileSource, error) {
	if err := s.validateLogCollectionScope(ctx, logCollectionScopeCluster, clusterID); err != nil {
		return nil, nil, nil, err
	}
	normalizedContainer, err := normalizeContainerStdout(input.ContainerStdout)
	if err != nil {
		return nil, nil, nil, err
	}
	normalizedFiles, err := normalizeHostFiles(input.HostFiles)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := validateEditablePipelineRules(input.ProcessingRule, input.OutputRule); err != nil {
		return nil, nil, nil, err
	}
	if !hasEnabledContainerStdoutSource(normalizedContainer) && !hasEnabledHostFileSource(normalizedFiles) {
		return &LogCollectionPreview{ScopeType: logCollectionScopeCluster, ScopeID: clusterID, Files: []RenderedLogCollectionFile{}}, normalizedContainer, normalizedFiles, nil
	}
	defaults, err := s.storedLogCollectionDefaults(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	files := make([]RenderedLogCollectionFile, 0, len(normalizedContainer)+len(normalizedFiles))
	if hasEnabledContainerStdoutSource(normalizedContainer) {
		base, err := s.clusterLogPipelineBase(ctx, clusterID, containerStdoutInputTypes, nil, nil)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, source := range normalizedContainer {
			if !source.Enabled {
				continue
			}
			processing, output := pipelineRulePointers(source.ProcessingRule, source.OutputRule, input.ProcessingRule, input.OutputRule, defaults.Container)
			sourceBase, err := pipelineBaseWithOverrides(base, processing, output)
			if err != nil {
				return nil, nil, nil, err
			}
			rendered, err := renderContainerStdout([]ContainerStdoutSource{source}, sourceBase)
			if err != nil {
				return nil, nil, nil, err
			}
			files = append(files, rendered...)
		}
	}
	if hasEnabledHostFileSource(normalizedFiles) {
		base, err := s.clusterLogPipelineBase(ctx, clusterID, hostFileInputTypes, nil, nil)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, source := range normalizedFiles {
			if !source.Enabled {
				continue
			}
			processing, output := pipelineRulePointers(source.ProcessingRule, source.OutputRule, input.ProcessingRule, input.OutputRule, defaults.File)
			sourceBase, err := pipelineBaseWithOverrides(base, processing, output)
			if err != nil {
				return nil, nil, nil, err
			}
			rendered, err := renderHostFiles([]HostFileSource{source}, sourceBase)
			if err != nil {
				return nil, nil, nil, err
			}
			files = append(files, rendered...)
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return &LogCollectionPreview{ScopeType: logCollectionScopeCluster, ScopeID: clusterID, Files: files}, normalizedContainer, normalizedFiles, nil
}

func (s *Service) PreviewContainerStdoutCollection(ctx core.Context, clusterID string, input ContainerStdoutCollectionInput) (*LogCollectionPreview, error) {
	preview, _, _, err := s.previewContainerStdoutCollection(ctx, clusterID, input)
	return preview, err
}
