// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/itchyny/gojq"
)

type Incident struct {
	ID string `json:"id" gorm:"id;primaryKey"`

	Desc string `json:"desc" gorm:"desc"`

	IncidentKey  string `json:"incidentKey" gorm:"incidentKey"`
	Status       string `json:"status" gorm:"status"`
	CreateTime   int64  `json:"createTime" gorm:"createTime"`
	UpdateTime   int64  `json:"updateTime" gorm:"updateTime"`
	ResolvedTime int64  `json:"resolvedTime" gorm:"resolvedTime"`
	// AlertCount   int               `json:"alertCount"`
	// Tags map[string]string `json:"tags"`

	AlertEvents []AlertEvent `json:"alertEvents" gorm:"-"`
}

func (in Incident) TableName() string {
	return "incident"
}

func (in *Incident) GetFiringAlertCount() int {
	var seen = make(map[string]struct{})
	firingCount := 0
	for i := len(in.AlertEvents) - 1; i >= 0; i++ {
		if _, find := seen[in.AlertEvents[i].AlertID]; !find {
			seen[in.AlertEvents[i].AlertID] = struct{}{}
			if in.AlertEvents[i].Status == StatusFiring {
				firingCount++
			}
		}
	}
	return firingCount
}

type Incident2Alert struct {
	IncidentId   string    `json:"incidentId" ch:"incident_id"`
	AlertEventID string    `json:"alertEventId" ch:"alert_event_id"`
	Timestamp    time.Time `json:"timestamp" ch:"timestamp"`
}

func (in Incident2Alert) TableName() string {
	return "incident2alert"
}

type IncidentKeyTemp struct {
	ID string `json:"id" gorm:"id;primaryKey"`

	Order int `json:"order" gorm:"order"`
	// payload；
	// generated_labels;
	// labels;
	IncidentKeyTemplate string              `json:"incidentKeyTemplate" gorm:"incidentKeyTemplate"`
	CompiledTemp        *template.Template  `json:"-" gorm:"-"`
	CompiledErr         error               `json:"-" gorm:"-"`
	Conditions          []IncidentCondition `json:"conditions" gorm:"-"`
	ConditionJQParser   *gojq.Query         `json:"-" gorm:"-"`

	AlertSourceID string `json:"alertSourceId" gorm:"alert_source_id"`
}

func (t *IncidentKeyTemp) TableName() string {
	return "incident_key_temp"
}

func (t *IncidentKeyTemp) Compile() error {
	var errs []error

	if len(t.ID) == 0 {
		t.ID = uuid.New().String()
	}

	for i := 0; i < len(t.Conditions); i++ {
		t.Conditions[i].IncidentTempID = t.ID
	}

	if len(t.Conditions) > 0 {
		var conditionErr error
		t.ConditionJQParser, conditionErr = buildJQConditionExpr(t.Conditions)
		if conditionErr != nil {
			errs = append(errs, conditionErr)
		}
	}

	if len(t.IncidentKeyTemplate) == 0 {
		errs = append(errs, errors.New("incident key template is empty"))
		return errors.Join(errs...)
	}

	var compileErr error
	t.CompiledTemp, compileErr = template.New("incidentKey").Parse(t.IncidentKeyTemplate)
	if compileErr != nil {
		errs = append(errs, compileErr)
	}

	if len(errs) > 0 {
		t.CompiledErr = fmt.Errorf("compile incident key template failed: %w", errors.Join(errs...))
	}
	return t.CompiledErr
}

func (t *IncidentKeyTemp) IsValid() (bool, error) {
	if err := t.Compile(); err != nil {
		return false, err
	}
	return true, nil
}

func (t *IncidentKeyTemp) Equal(other *IncidentKeyTemp) bool {
	if t.ID != other.ID {
		return false
	}

	if t.Order != other.Order {
		return false
	}

	if t.IncidentKeyTemplate != other.IncidentKeyTemplate {
		return false
	}

	if t.AlertSourceID != other.AlertSourceID {
		return false
	}

	if len(t.Conditions) != len(other.Conditions) {
		return false
	}

	for i := range t.Conditions {
		if t.Conditions[i] != other.Conditions[i] {
			return false
		}
	}

	return true
}

func (t *IncidentKeyTemp) CheckConditions(data map[string]any) (bool, error) {
	if t.ConditionJQParser == nil {
		return true, nil
	}

	iter := t.ConditionJQParser.Run(data)
	v, ok := iter.Next()
	if !ok {
		return false, nil
	}
	if err, isErr := v.(error); isErr {
		return false, err
	}

	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		return val == "true", nil
	case float64:
		return val != 0, nil
	default:
		return false, nil
	}
}

func (t *IncidentKeyTemp) GenerateIncidentKey(data map[string]any) (string, error) {
	if t.CompiledErr != nil {
		return "", t.CompiledErr
	}

	var buf bytes.Buffer
	err := t.CompiledTemp.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

type IncidentCondition struct {
	IncidentTempID string `json:"-" gorm:"type:varchar(255);column:incident_temp_id;index"`

	FromField string `json:"fromField" gorm:"type:varchar(255);column:from_field"`

	Operation string `json:"operation" gorm:"type:varchar(255);column:operation"` // support match,not match,gt,lt,ge,le,eq
	Expr      string `json:"expr" gorm:"type:varchar(255);column:expr"`           // Regex
}

func (c IncidentCondition) TableName() string {
	return "incident_condition"
}

func (c IncidentCondition) Equal(other IncidentCondition) bool {
	if c.IncidentTempID != other.IncidentTempID {
		return false
	}

	if c.FromField != other.FromField {
		return false
	}

	if c.Operation != other.Operation {
		return false
	}

	if c.Expr != other.Expr {
		return false
	}

	return true
}

func buildJQConditionExpr(conditions []IncidentCondition) (*gojq.Query, error) {
	var parts []string
	for _, condition := range conditions {
		var part string
		switch condition.Operation {
		case "match":
			part = fmt.Sprintf(`(%s | test("%s"))`, condition.FromField, condition.Expr)
		case "notMatch":
			part = fmt.Sprintf(`(%s | test("%s") | not)`, condition.FromField, condition.Expr)
		case "le":
			part = fmt.Sprintf(`(%s | tonumber? <= (%s | tonumber?))`, condition.FromField, condition.Expr)
		case "ge":
			part = fmt.Sprintf(`(%s | tonumber? >= (%s | tonumber?))`, condition.FromField, condition.Expr)
		case "lt":
			part = fmt.Sprintf(`(%s | tonumber? < (%s | tonumber?))`, condition.FromField, condition.Expr)
		case "gt":
			part = fmt.Sprintf(`(%s | tonumber? > (%s | tonumber?))`, condition.FromField, condition.Expr)
		case "eq":
			part = fmt.Sprintf(`(%s == %s)`, condition.FromField, condition.Expr)
		default:
			continue
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return nil, nil
	}
	expr := strings.Join(parts, " and ")
	return gojq.Parse(expr)
}
