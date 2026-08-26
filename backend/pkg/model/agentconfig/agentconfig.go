package agentconfig

import (
	"bytes"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type JSON []byte

func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	if !json.Valid(j) {
		return nil, errors.New("invalid JSON value")
	}
	return j, nil
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*j = JSON("{}")
		return nil
	}
	if !json.Valid(data) {
		return errors.New("invalid JSON value")
	}
	*j = append((*j)[:0], data...)
	return nil
}

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	return string(j), nil
}
func (j *JSON) Scan(v any) error {
	if v == nil {
		*j = JSON("{}")
		return nil
	}
	switch x := v.(type) {
	case []byte:
		*j = append((*j)[:0], x...)
	case string:
		*j = append((*j)[:0], x...)
	default:
		return json.Unmarshal([]byte("{}"), j)
	}
	return nil
}

type Template struct {
	ID            uuid.UUID          `gorm:"type:uuid;primaryKey" json:"id"`
	Configs       JSON               `gorm:"type:jsonb;not null;default:'[]'" json:"configs"`
	Name          string             `gorm:"size:255;not null" json:"name"`
	Description   string             `gorm:"type:text" json:"description"`
	ClusterType   string             `gorm:"size:32;not null;default:kubernetes" json:"clusterType"`
	Version       int64              `gorm:"not null;default:1" json:"version"`
	ContentSHA256 string             `gorm:"size:64;not null" json:"contentSha256"`
	Variables     []TemplateVariable `gorm:"foreignKey:TemplateID;constraint:OnDelete:CASCADE" json:"variables,omitempty"`
	CreatedBy     int64              `json:"createdBy"`
	UpdatedBy     int64              `json:"updatedBy"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
}

type TemplateProbeConfig struct {
	ProbeType string              `json:"probeType"`
	Files     []InitialConfigFile `json:"files"`
}

type TemplateVariable struct {
	ID              uint64    `gorm:"primaryKey" json:"id"`
	TemplateID      uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_template_var" json:"templateId"`
	Name            string    `gorm:"size:128;not null;uniqueIndex:idx_template_var" json:"name"`
	Required        bool      `json:"required"`
	DefaultValue    string    `gorm:"type:text" json:"defaultValue"`
	Secret          bool      `json:"secret"`
	Description     string    `gorm:"type:text" json:"description"`
	ValidationRegex string    `gorm:"type:text" json:"validationRegex"`
}

type Node struct {
	ID            string     `gorm:"size:255;primaryKey" json:"id"`
	Name          string     `gorm:"size:255" json:"name"`
	ClusterID     string     `gorm:"size:255;index" json:"clusterId"`
	Environment   string     `gorm:"size:64" json:"environment"`
	OdigletReady  *bool      `json:"odigletReady,omitempty"`
	Labels        JSON       `gorm:"type:jsonb" json:"labels"`
	DaemonVersion string     `gorm:"size:128" json:"daemonVersion"`
	LastSeenAt    *time.Time `json:"lastSeenAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type InitialConfigFile struct {
	Path          string `json:"path"`
	Content       string `json:"content"`
	ContentSHA256 string `json:"contentSha256"`
}

type InitialConfig struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	NodeID        string    `gorm:"size:255;not null;uniqueIndex:idx_initial_config_node_probe" json:"nodeId"`
	ProbeType     string    `gorm:"size:128;not null;uniqueIndex:idx_initial_config_node_probe" json:"probeType"`
	Files         JSON      `gorm:"type:jsonb;not null" json:"files"`
	ContentSHA256 string    `gorm:"size:64;not null" json:"contentSha256"`
	CapturedAt    time.Time `json:"capturedAt"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Binding struct {
	ID                uint64    `gorm:"primaryKey" json:"id"`
	NodeID            string    `gorm:"size:255;not null;uniqueIndex:idx_node_probe" json:"nodeId"`
	ProbeType         string    `gorm:"size:128;not null;uniqueIndex:idx_node_probe" json:"probeType"`
	TemplateID        uuid.UUID `gorm:"type:uuid;not null;index" json:"templateId"`
	DesiredGeneration int64     `gorm:"not null;default:0" json:"desiredGeneration"`
	Enabled           bool      `gorm:"not null;default:true" json:"enabled"`
	Variables         JSON      `gorm:"type:jsonb;not null;default:'{}'" json:"variables"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type Release struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	NodeID          string    `gorm:"size:255;not null;uniqueIndex:idx_release_generation" json:"nodeId"`
	ProbeType       string    `gorm:"size:128;not null;uniqueIndex:idx_release_generation" json:"probeType"`
	Generation      int64     `gorm:"not null;uniqueIndex:idx_release_generation" json:"generation"`
	TemplateID      uuid.UUID `gorm:"type:uuid;not null" json:"templateId"`
	TemplateVersion int64     `gorm:"not null" json:"templateVersion"`
	Variables       JSON      `gorm:"type:jsonb;not null" json:"variables"`
	RenderedContent string    `gorm:"type:text;not null" json:"renderedContent"`
	ContentSHA256   string    `gorm:"size:64;not null" json:"contentSha256"`
	Files           JSON      `gorm:"type:jsonb;not null;default:'null'" json:"files,omitempty"`
	LogSources      JSON      `gorm:"type:jsonb;not null;default:'null'" json:"logSources"`
	CreatedAt       time.Time `json:"createdAt"`
}

type LogSource struct {
	Type     string `json:"type,omitempty"`
	Path     string `json:"path"`
	Pattern  string `json:"pattern"`
	MaxBytes int64  `json:"maxBytes,omitempty"`
}

type CircuitBreakerSettings struct {
	Enabled                     bool                              `json:"enabled"`
	MaxCPUCores                 float64                           `json:"maxCpuCores"`
	MaxMemoryBytes              uint64                            `json:"maxMemoryBytes"`
	DurationSeconds             uint64                            `json:"durationSeconds"`
	AutoRecover                 bool                              `json:"autoRecover"`
	RecoveryInitialDelaySeconds uint64                            `json:"recoveryInitialDelaySeconds"`
	RecoveryMaxDelaySeconds     uint64                            `json:"recoveryMaxDelaySeconds"`
	SystemPressure              *SystemPressureSettings           `json:"systemPressure,omitempty"`
	ComponentPolicies           map[string]CircuitBreakerSettings `json:"componentPolicies,omitempty"`
}

type SystemPressureSettings struct {
	Enabled          bool    `json:"enabled"`
	MaxCPUPercent    float64 `json:"maxCpuPercent"`
	MaxMemoryPercent float64 `json:"maxMemoryPercent"`
}

type AgentSettings struct {
	LogSources             []LogSource             `json:"logSources"`
	UseLocalDefaults       bool                    `json:"useLocalDefaults"`
	UseLocalCircuitBreaker bool                    `json:"useLocalCircuitBreaker"`
	CircuitBreaker         *CircuitBreakerSettings `json:"circuitBreaker,omitempty"`
	NamespaceInjection     *NamespaceInjection     `json:"namespaceInjection,omitempty"`
	OdigletReady           *bool                   `json:"odigletReady,omitempty"`
}

type NamespaceInjection struct {
	InstrumentAll        bool     `json:"instrumentAll"`
	InstrumentNS         []string `json:"instrumentNS"`
	InstrumentDisabledNS []string `json:"instrumentDisabledNS"`
}

type GlobalCircuitConfig struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Settings  JSON      `gorm:"type:jsonb;not null" json:"settings"`
	UpdatedBy int64     `json:"updatedBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type NodeCircuitConfig struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	NodeID    string    `gorm:"size:255;not null;uniqueIndex:idx_node_circuit" json:"nodeId"`
	Settings  JSON      `gorm:"type:jsonb;not null" json:"settings"`
	Version   int64     `gorm:"not null;default:1" json:"version"`
	UpdatedBy int64     `json:"updatedBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type NodeLogConfig struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	NodeID    string    `gorm:"size:255;not null;uniqueIndex:idx_node_probe_logs" json:"nodeId"`
	ProbeType string    `gorm:"size:128;not null;uniqueIndex:idx_node_probe_logs" json:"probeType"`
	Sources   JSON      `gorm:"type:jsonb;not null" json:"sources"`
	Version   int64     `gorm:"not null;default:1" json:"version"`
	UpdatedBy int64     `json:"updatedBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type LogCollectionConfig struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	ScopeType string    `gorm:"size:32;not null;uniqueIndex:idx_log_collection_scope" json:"scopeType"`
	ScopeID   string    `gorm:"size:255;not null;uniqueIndex:idx_log_collection_scope" json:"scopeId"`
	Config    JSON      `gorm:"type:jsonb;not null" json:"config"`
	Version   int64     `gorm:"not null;default:1" json:"version"`
	UpdatedBy int64     `json:"updatedBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ApplyState struct {
	ID                uint64     `gorm:"primaryKey" json:"id"`
	NodeID            string     `gorm:"size:255;not null;uniqueIndex:idx_state_node_probe" json:"nodeId"`
	ProbeType         string     `gorm:"size:128;not null;uniqueIndex:idx_state_node_probe" json:"probeType"`
	DesiredGeneration int64      `json:"desiredGeneration"`
	AppliedGeneration int64      `json:"appliedGeneration"`
	Status            string     `gorm:"size:32" json:"status"`
	Reason            string     `gorm:"type:text" json:"reason"`
	Error             JSON       `gorm:"type:jsonb" json:"error,omitempty"`
	Runtime           JSON       `gorm:"type:jsonb" json:"runtime"`
	Circuit           JSON       `gorm:"type:jsonb" json:"circuit"`
	ReportedAt        *time.Time `json:"reportedAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type Command struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	NodeID               string     `gorm:"size:255;not null;index" json:"nodeId"`
	ProbeType            string     `gorm:"size:128;not null" json:"probeType"`
	Type                 string     `gorm:"size:32;not null" json:"type"`
	ExpectedRuntimeEpoch *uint64    `json:"expectedRuntimeEpoch,omitempty"`
	ExpectedCircuitEpoch *uint64    `json:"expectedCircuitEpoch,omitempty"`
	Reason               string     `gorm:"type:text" json:"reason,omitempty"`
	FilePath             string     `gorm:"type:text" json:"filePath,omitempty"`
	TailLines            int        `json:"tailLines,omitempty"`
	Status               string     `gorm:"size:32;not null;default:'PENDING';index" json:"status"`
	ClaimantID           string     `gorm:"size:255" json:"claimantId,omitempty"`
	LeaseToken           string     `gorm:"size:64" json:"leaseToken,omitempty"`
	LeaseExpiresAt       *time.Time `json:"leaseExpiresAt,omitempty"`
	DeliveryCount        int        `gorm:"not null;default:0" json:"deliveryCount"`
	StartedAt            *time.Time `json:"startedAt,omitempty"`
	FinishedAt           *time.Time `json:"finishedAt,omitempty"`
	RequestedBy          int64      `json:"requestedBy"`
	Result               string     `gorm:"type:text" json:"result"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

func NewTemplate() *Template {
	return &Template{ID: uuid.New(), ClusterType: "kubernetes", Version: 1}
}
func HashContent(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func HashInitialConfigFiles(files []InitialConfigFile) string {
	ordered := append([]InitialConfigFile(nil), files...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Path < ordered[j].Path })
	var payload bytes.Buffer
	for _, file := range ordered {
		_ = binary.Write(&payload, binary.BigEndian, uint64(len(file.Path)))
		payload.WriteString(file.Path)
		_ = binary.Write(&payload, binary.BigEndian, uint64(len(file.Content)))
		payload.WriteString(file.Content)
	}
	return HashContent(payload.String())
}
func NormalizeProbe(s string) string       { return strings.TrimSpace(strings.ToLower(s)) }
func IsCommandTerminal(status string) bool { return status == "SUCCEEDED" || status == "FAILED" }
func ValidateCommandTransition(current, target string) error {
	switch target {
	case "STARTED":
		if current == "CLAIMED" || current == "STARTED" {
			return nil
		}
	case "SUCCEEDED", "FAILED":
		if current == "STARTED" {
			return nil
		}
	}
	return fmt.Errorf("invalid command transition %s -> %s", current, target)
}
