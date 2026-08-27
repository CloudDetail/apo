package agentconfig

import (
	"fmt"
	"strings"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
	"gopkg.in/yaml.v3"
)

type TemplateImportInput struct {
	YAML string `json:"yaml" binding:"required"`
}

type templateImportDocument struct {
	Name        string                      `yaml:"name"`
	Description string                      `yaml:"description"`
	ClusterType string                      `yaml:"clusterType"`
	Variables   []templateImportVariable    `yaml:"variables"`
	Configs     []templateImportProbeConfig `yaml:"configs"`
}

type templateImportVariable struct {
	Name            string `yaml:"name"`
	Required        bool   `yaml:"required"`
	DefaultValue    string `yaml:"defaultValue"`
	ValidationRegex string `yaml:"validationRegex"`
}

type templateImportProbeConfig struct {
	ProbeType string               `yaml:"probeType"`
	Files     []templateImportFile `yaml:"files"`
}

type templateImportFile struct {
	Path     string `yaml:"path"`
	Template string `yaml:"template"`
	Content  string `yaml:"content"`
}

func (s *Service) ImportTemplate(ctx core.Context, input TemplateImportInput, user int64) (*ac.Template, error) {
	var document templateImportDocument
	decoder := yaml.NewDecoder(strings.NewReader(input.YAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode template YAML: %w", err)
	}
	configs := make([]ac.TemplateProbeConfig, 0, len(document.Configs))
	for _, config := range document.Configs {
		files := make([]ac.InitialConfigFile, 0, len(config.Files))
		for _, file := range config.Files {
			content := file.Template
			if content == "" {
				content = file.Content
			}
			files = append(files, ac.InitialConfigFile{Path: file.Path, Content: content})
		}
		configs = append(configs, ac.TemplateProbeConfig{ProbeType: config.ProbeType, Files: files})
	}
	variables := make([]ac.TemplateVariable, 0, len(document.Variables))
	for _, variable := range document.Variables {
		variables = append(variables, ac.TemplateVariable{
			Name:            variable.Name,
			Required:        variable.Required,
			DefaultValue:    variable.DefaultValue,
			ValidationRegex: variable.ValidationRegex,
		})
	}
	return s.CreateTemplate(ctx, TemplateInput{
		Name:        document.Name,
		Description: document.Description,
		ClusterType: document.ClusterType,
		Configs:     configs,
		Variables:   variables,
	}, user)
}
